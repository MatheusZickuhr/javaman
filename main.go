package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func main() {

	configFile := loadConfigFile()

	args := os.Args[1:]
	parsedArgs := parseArgs(args)

	if parsedArgs.UseJdkCommand {
		updateJdkVersion(parsedArgs.JdkVersion, configFile)
	}

	if parsedArgs.UseMvnCommand {
		updateMavenVersion(parsedArgs.MvnVersion, configFile)
	}

	if parsedArgs.ListCommand != "" {
		executeListCommand(parsedArgs.ListCommand, configFile)
	}

	if parsedArgs.InstallJdkCommand {
		installJdk(parsedArgs.JdkVersion, configFile)
	}

	if parsedArgs.InstallMvnCommand {
		installMvn(parsedArgs.MvnVersion, configFile)
	}

	if parsedArgs.UninstallJdkCommand {
		uninstallJdk(parsedArgs.JdkVersion, configFile)
	}

	if parsedArgs.UninstallMvnCommand {
		uninstallMvn(parsedArgs.MvnVersion, configFile)
	}

}

func uninstallJdk(version string, config *ConfigFile) error {

	// Variável para armazenar o índice da instalação a ser removida.
	// Iniciamos com -1 para indicar que a versão não foi encontrada.
	foundIndex := -1
	var installationToRemove Installation

	// Itera sobre a lista de JDKs para encontrar a versão correspondente.
	for i, jdk := range config.Jdks {
		if jdk.Version == version {
			foundIndex = i
			installationToRemove = jdk
			break
		}
	}

	// Se a versão não for encontrada na configuração, retorna um erro.
	if foundIndex == -1 {
		return fmt.Errorf("versão do JDK '%s' não encontrada", version)
	}

	// Verifica se o HomePath não está vazio para evitar a exclusão da raiz do sistema.
	if installationToRemove.HomePath == "" {
		return fmt.Errorf("HomePath para a versão '%s' está vazio, remoção abortada por segurança", version)
	}

	// Remove o diretório da instalação e todo o seu conteúdo.
	fmt.Printf("Removendo diretório: %s\n", installationToRemove.HomePath)
	err := os.RemoveAll(installationToRemove.HomePath)
	if err != nil {
		return fmt.Errorf("falha ao remover o diretório '%s': %w", installationToRemove.HomePath, err)
	}

	// se tiver em uso, remove do path
	if installationToRemove.InUse {
		systemPath := readEnvVariable("Path")
		newPath := removeInstallationsFromPath(systemPath, &config.Jdks)
		setEnvVariable("Path", newPath)
		removeEnvVariable("JAVA_HOME")
	}

	// Remove o elemento da slice 'Jdks' utilizando a técnica de slicing.
	// Isso cria uma nova slice que contém todos os elementos, exceto o do índice encontrado.
	config.Jdks = append(config.Jdks[:foundIndex], config.Jdks[foundIndex+1:]...)

	// atualiza o arquivo
	saveConfigFile(config)

	fmt.Printf("Versão '%s' do JDK desinstalada com sucesso.\n", version)
	return nil

}

func uninstallMvn(version string, config *ConfigFile) error {

	// Variável para armazenar o índice da instalação a ser removida.
	// Iniciamos com -1 para indicar que a versão não foi encontrada.
	foundIndex := -1
	var installationToRemove Installation

	// Itera sobre a lista de Mavens para encontrar a versão correspondente.
	for i, mvn := range config.Mavens {
		if mvn.Version == version {
			foundIndex = i
			installationToRemove = mvn
			break
		}
	}

	// Se a versão não for encontrada na configuração, retorna um erro.
	if foundIndex == -1 {
		return fmt.Errorf("versão do Maven '%s' não encontrada", version)
	}

	// Verifica se o HomePath não está vazio para evitar a exclusão da raiz do sistema.
	if installationToRemove.HomePath == "" {
		return fmt.Errorf("HomePath para a versão '%s' está vazio, remoção abortada por segurança", version)
	}

	// Remove o diretório da instalação e todo o seu conteúdo.
	fmt.Printf("Removendo diretório: %s\n", installationToRemove.HomePath)
	err := os.RemoveAll(installationToRemove.HomePath)
	if err != nil {
		return fmt.Errorf("falha ao remover o diretório '%s': %w", installationToRemove.HomePath, err)
	}

	// se tiver em uso, reomve do path
	if installationToRemove.InUse {
		systemPath := readEnvVariable("Path")
		newPath := removeInstallationsFromPath(systemPath, &config.Mavens)
		setEnvVariable("Path", newPath)
		removeEnvVariable("MAVEN_HOME")
	}

	// Remove o elemento da slice 'Mavens' utilizando a técnica de slicing.
	// Isso cria uma nova slice que contém todos os elementos, exceto o do índice encontrado.
	config.Mavens = append(config.Mavens[:foundIndex], config.Mavens[foundIndex+1:]...)

	// atualiza o arquivo
	saveConfigFile(config)

	fmt.Printf("Versão '%s' do Maven desinstalada com sucesso.\n", version)
	return nil
}

func installMvn(mvnVersion string, configFile *ConfigFile) {
	installPath, err := installMaven(mvnVersion)

	if err != nil {
		panic(err)
	}

	configFile.Mavens = append(configFile.Mavens, Installation{HomePath: installPath, BinPath: installPath + "\\bin", Version: mvnVersion})

	err = saveConfigFile(configFile)

	if err != nil {
		fmt.Println(err)
	}

}

func installJdk(jdkVersion string, configFile *ConfigFile) {

	for _, inst := range configFile.Jdks {
		if inst.Version == jdkVersion {
			fmt.Println("Jdk " + jdkVersion + " already installed")
			return
		}
	}

	fmt.Println("Installing jdk " + jdkVersion)

	zipFileDir := downloadJdkOnTempDir(jdkVersion)
	jdkPath, err := extrairJDK(zipFileDir)

	if err != nil {
		fmt.Println(err)
	}

	inst := Installation{Version: jdkVersion, HomePath: jdkPath, BinPath: jdkPath + "\\bin"}
	configFile.Jdks = append(configFile.Jdks, inst)

	err = saveConfigFile(configFile)

	if err != nil {
		fmt.Println(err)
	}

}

func updateJdkVersion(jdkVersion string, configFile *ConfigFile) {
	installations := &configFile.Jdks
	selected := configFile.GetJdkByVersion(jdkVersion)

	if selected == nil {
		fmt.Printf("No installation found for version %s in file %s\n", jdkVersion, configFile)
		return
	}

	systemPath := readEnvVariable("Path")
	newPath := removeInstallationsFromPath(systemPath, installations)
	newPath = addNewInstallationToPath(newPath, selected)

	setEnvVariable("Path", newPath)
	setEnvVariable("JAVA_HOME", selected.HomePath)

	selected.InUse = true
	saveConfigFile(configFile)

	fmt.Println("JDK updated to version " + jdkVersion)
}

func updateMavenVersion(mavenVersion string, configFile *ConfigFile) {
	installations := &configFile.Mavens
	selected := configFile.GetMavenByVersion(mavenVersion)

	if selected == nil {
		fmt.Printf("No installation found for version %s in file %s\n", mavenVersion, configFile)
		return
	}

	systemPath := readEnvVariable("Path")
	newPath := removeInstallationsFromPath(systemPath, installations)
	newPath = addNewInstallationToPath(newPath, selected)

	setEnvVariable("Path", newPath)
	setEnvVariable("MAVEN_HOME", selected.HomePath)

	selected.InUse = true
	saveConfigFile(configFile)

	fmt.Println("Maven updated to version " + mavenVersion)
}

func parseArgs(args []string) ParsedArgs {
	var parsed ParsedArgs
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "use-jdk":
			parsed.UseJdkCommand = true
			parsed.JdkVersion = args[i+1]
		case "use-mvn":
			parsed.UseMvnCommand = true
			parsed.MvnVersion = args[i+1]
		case "list":
			parsed.ListCommand = args[i+1]
		case "install-jdk":
			parsed.InstallJdkCommand = true
			parsed.JdkVersion = args[i+1]
		case "install-mvn":
			parsed.InstallMvnCommand = true
			parsed.MvnVersion = args[i+1]
		case "uninstall-jdk":
			parsed.UninstallJdkCommand = true
			parsed.JdkVersion = args[i+1]
		case "uninstall-mvn":
			parsed.UninstallMvnCommand = true
			parsed.MvnVersion = args[i+1]
		}
	}

	return parsed
}

func executeListCommand(command string, configFile *ConfigFile) {
	switch command {
	case "jdk":
		for _, inst := range configFile.Jdks {
			fmt.Println("Jdk Version " + inst.Version + ", Home path: " + inst.HomePath)
		}
	case "mvn":
		for _, inst := range configFile.Mavens {
			fmt.Println("Maven Version " + inst.Version + ", Home path: " + inst.HomePath)
		}
	default:
		fmt.Println("Invalid list command")
		return
	}
}

func removeInstallationsFromPath(path string, installations *[]Installation) string {
	parts := strings.Split(path, ";")
	var newParts []string
	for _, part := range parts {
		if !isBinFolderInInstallations(part, installations) {
			newParts = append(newParts, part)
		}
	}
	return strings.Join(newParts, ";")
}

func isBinFolderInInstallations(folder string, installations *[]Installation) bool {
	for _, inst := range *installations {
		if strings.Contains(folder, inst.HomePath) {
			return true
		}
	}
	return false
}

func addNewInstallationToPath(path string, inst *Installation) string {
	return path + ";" + inst.BinPath
}

func readEnvVariable(name string) string {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE)
	if err != nil {
		fmt.Println("Error reading registry:", err)
		return ""
	}
	defer key.Close()

	val, _, err := key.GetStringValue(name)
	if err != nil {
		return ""
	}
	return val
}

func setEnvVariable(name, value string) {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, `Environment`, registry.SET_VALUE)
	if err != nil {
		fmt.Println("Error setting registry:", err)
		return
	}
	defer key.Close()

	err = key.SetStringValue(name, value)
	if err != nil {
		fmt.Println("Error writing value:", err)
	}
}

func removeEnvVariable(name string) error {
	// 1. Abrir a chave do registro onde as variáveis de ambiente do usuário estão localizadas.
	// registry.CURRENT_USER corresponde a HKEY_CURRENT_USER.
	// O segundo argumento é o caminho para a subchave.
	// O terceiro argumento especifica o acesso desejado. Precisamos de acesso de escrita (WRITE) para deletar.
	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.WRITE)
	if err != nil {
		// Se a chave 'Environment' não existir, o que é muito raro, o erro será tratado aqui.
		return fmt.Errorf("não foi possível abrir a chave do registro 'Environment': %w", err)
	}
	// 2. É uma boa prática fechar a chave quando terminamos de usá-la.
	defer key.Close()

	// 3. Deletar o valor (que representa a variável de ambiente) dentro da chave aberta.
	err = key.DeleteValue(name)
	if err != nil {
		// Este erro pode ocorrer se a variável de ambiente não existir no registro,
		// o que pode ser um caso esperado. Você pode querer tratar este erro especificamente.
		// Exemplo: if err == registry.ErrNotExist { ... }
		return fmt.Errorf("não foi possível deletar a variável '%s' do registro: %w", name, err)
	}

	fmt.Printf("Variável de ambiente '%s' removida com sucesso do registro.\n", name)
	fmt.Println("Atenção: A maioria dos programas precisa ser reiniciada para que a mudança tenha efeito.")

	return nil
}
