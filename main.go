package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func main() {

	configFile := loadConfigFile()

	if configFile == nil {
		fmt.Println("No config file found")
		return
	}

	args := os.Args[1:]
	parsedArgs, err := parseArgs(args)

	if err != nil {
		fmt.Println(err)
	}

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
		installJava(parsedArgs.JdkVersion, configFile)
	}

	if parsedArgs.InstallMvnCommand {
		installMvn(parsedArgs.MvnVersion, configFile)
	}

	if parsedArgs.UninstallJdkCommand {
		uninstallJava(parsedArgs.JdkVersion, configFile)
	}

	if parsedArgs.UninstallMvnCommand {
		uninstallMvn(parsedArgs.MvnVersion, configFile)
	}

}

func uninstallJava(jdkVersion string, configFile *ConfigFile) {
	err := uninstallInstallation(jdkVersion, configFile, &configFile.Jdks, "JDK", "JAVA_HOME")
	if err != nil {
		fmt.Println(err)
	}
}

func uninstallMvn(mavenVersion string, configFile *ConfigFile) {
	err := uninstallInstallation(mavenVersion, configFile, &configFile.Mavens, "Maven", "MAVEN_HOME")
	if err != nil {
		fmt.Println(err)
	}
}

func uninstallInstallation(version string, config *ConfigFile, installations *[]Installation, installationType string, homeEnvVar string) error {
	foundIndex := -1
	var installationToRemove Installation

	for i, inst := range *installations {
		if inst.Version == version {
			foundIndex = i
			installationToRemove = inst
			break
		}
	}

	if foundIndex == -1 {
		return fmt.Errorf("%s version '%s' not found", installationType, version)
	}

	if installationToRemove.HomePath == "" {
		return fmt.Errorf("HomePath for version '%s' is empty, removal aborted for safety", version)
	}

	fmt.Printf("Removing directory: %s\n", installationToRemove.HomePath)
	err := os.RemoveAll(installationToRemove.HomePath)
	if err != nil {
		return fmt.Errorf("failed to remove directory '%s': %w", installationToRemove.HomePath, err)
	}

	// remove from installations
	*installations = append((*installations)[:foundIndex], (*installations)[foundIndex+1:]...)

	saveConfigFile(config)

	fmt.Printf("%s version '%s' uninstalled successfully.\n", installationType, version)
	return nil
}

func installMvn(mvnVersion string, configFile *ConfigFile) {
	for _, inst := range configFile.Mavens {
		if inst.Version == mvnVersion {
			fmt.Println("Maven " + mvnVersion + " already installed")
			return
		}
	}

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

func installJava(jdkVersion string, configFile *ConfigFile) {
	for _, inst := range configFile.Jdks {
		if inst.Version == jdkVersion {
			fmt.Println("Jdk " + jdkVersion + " already installed")
			return
		}
	}

	jdkPath, err := installJdk(jdkVersion)

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

	setEnvVariable("JAVA_BIN_PATH", selected.BinPath)
	setEnvVariable("JAVA_HOME", selected.HomePath)

	systemPath := readEnvVariable("Path")
	newPath := removeBinFolderFromPath(systemPath, "%JAVA_BIN_PATH%")
	newPath = addNewBinFolderToPath(newPath, "%JAVA_BIN_PATH%")
	setEnvVariable("Path", newPath)

	setInUseToFalse(installations)
	selected.InUse = true
	saveConfigFile(configFile)

	fmt.Println("JDK updated to version " + jdkVersion)
	fmt.Println("Restart your terminal to apply the changes")
}

func updateMavenVersion(mavenVersion string, configFile *ConfigFile) {
	installations := &configFile.Mavens
	selected := configFile.GetMavenByVersion(mavenVersion)

	if selected == nil {
		fmt.Printf("No installation found for version %s in file %s\n", mavenVersion, configFile)
		return
	}

	setEnvVariable("MAVEN_BIN_PATH", selected.BinPath)
	setEnvVariable("MAVEN_HOME", selected.HomePath)

	systemPath := readEnvVariable("Path")
	newPath := removeBinFolderFromPath(systemPath, "%MAVEN_BIN_PATH%")
	newPath = addNewBinFolderToPath(newPath, "%MAVEN_BIN_PATH%")
	setEnvVariable("Path", newPath)

	setInUseToFalse(installations)
	selected.InUse = true
	saveConfigFile(configFile)

	fmt.Println("Maven updated to version " + mavenVersion)
	fmt.Println("Restart your terminal to apply the changes")
}

func parseArgs(args []string) (ParsedArgs, error) {
	var parsed ParsedArgs

	switch args[0] {
	case "use-jdk":
		parsed.UseJdkCommand = true
		parsed.JdkVersion = args[1]
	case "use-mvn":
		parsed.UseMvnCommand = true
		parsed.MvnVersion = args[1]
	case "list":
		parsed.ListCommand = args[1]
	case "install-jdk":
		parsed.InstallJdkCommand = true
		parsed.JdkVersion = args[1]
	case "install-mvn":
		parsed.InstallMvnCommand = true
		parsed.MvnVersion = args[1]
	case "uninstall-jdk":
		parsed.UninstallJdkCommand = true
		parsed.JdkVersion = args[1]
	case "uninstall-mvn":
		parsed.UninstallMvnCommand = true
		parsed.MvnVersion = args[1]
	default:
		return parsed, fmt.Errorf("unknown argument '%s'", args[0])
	}

	return parsed, nil
}

func executeListCommand(command string, configFile *ConfigFile) {
	var installations *[]Installation

	switch command {
	case "jdk":
		installations = &configFile.Jdks
	case "mvn":
		installations = &configFile.Mavens
	default:
		fmt.Println("Invalid list command")
		return
	}

	for _, inst := range *installations {
		if inst.InUse {
			fmt.Println(inst.Version + " (In Use)")
		} else {
			fmt.Println(inst.Version)
		}
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

func addNewBinFolderToPath(path string, binFolder string) string {
	return path + ";" + binFolder
}

func removeBinFolderFromPath(path string, toBeRemovedBinFolder string) string {
	parts := strings.Split(path, ";")

	var newParts []string
	for _, part := range parts {
		if toBeRemovedBinFolder == part {
			continue
		}
		newParts = append(newParts, part)
	}

	return strings.Join(newParts, ";")
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
	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.WRITE)
	if err != nil {
		return fmt.Errorf("error opening registry key 'Environment': %w", err)
	}

	defer key.Close()

	err = key.DeleteValue(name)
	if err != nil {
		return fmt.Errorf("error deleting variable '%s' from registry: %w", name, err)
	}

	fmt.Printf("Env variable '%s' succesfully removed.\n", name)

	return nil
}

func setInUseToFalse(installations *[]Installation) {

	for i := range *installations {
		installation := &(*installations)[i]
		installation.InUse = false
	}
}
