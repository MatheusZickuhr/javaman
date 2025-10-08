package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// installMaven baixa e extrai uma versão específica do Apache Maven.
// A instalação é feita na pasta .mvn dentro do diretório do usuário atual do Windows.
func installMaven(version string) (string, error) {
	// 1. Garante que o código só execute no Windows
	if runtime.GOOS != "windows" {
		return "", fmt.Errorf("este script foi projetado para rodar apenas no Windows")
	}

	// 2. Obter o diretório home do usuário
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("não foi possível obter o diretório do usuário: %w", err)
	}

	// 3. Definir os caminhos de instalação
	mvnRootPath := filepath.Join(homeDir, ".mvn")
	versionFolderName := fmt.Sprintf("apache-maven-%s", version)
	finalInstallPath := filepath.Join(mvnRootPath, versionFolderName)
	zipFileName := fmt.Sprintf("maven-%s-bin.zip", version)
	zipFilePath := filepath.Join(mvnRootPath, zipFileName)

	// 4. Verificar se a versão já está instalada
	if _, err := os.Stat(finalInstallPath); err == nil {
		fmt.Printf("✅ A versão %s do Maven já está instalada em %s\n", version, finalInstallPath)
		return "", nil
	}

	// 5. Criar o diretório .mvn se não existir
	if err := os.MkdirAll(mvnRootPath, 0755); err != nil {
		return "", fmt.Errorf("falha ao criar o diretório base %s: %w", mvnRootPath, err)
	}
	fmt.Printf("Diretório de instalação: %s\n", mvnRootPath)

	// 6. Construir a URL de download
	// Usamos o archive oficial da Apache para garantir acesso a versões mais antigas
	downloadURL := fmt.Sprintf("https://archive.apache.org/dist/maven/maven-3/%s/binaries/apache-maven-%s-bin.zip", version, version)

	// 7. Baixar o arquivo
	fmt.Printf("Baixando Maven %s de %s\n", version, downloadURL)
	resp, err := http.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("falha ao iniciar o download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("falha ao baixar o arquivo: status code %d. Verifique se a versão '%s' é válida", resp.StatusCode, version)
	}

	// Criar o arquivo zip temporário
	out, err := os.Create(zipFilePath)
	if err != nil {
		return "", fmt.Errorf("falha ao criar o arquivo zip local: %w", err)
	}
	defer out.Close()

	// Escrever o conteúdo no arquivo
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("falha ao salvar o arquivo baixado: %w", err)
	}
	fmt.Println("Download concluído com sucesso.")

	// 8. Extrair o arquivo zip
	fmt.Printf("Extraindo arquivos para %s...\n", mvnRootPath)
	if err := unzip(zipFilePath, mvnRootPath); err != nil {
		return "", fmt.Errorf("falha ao extrair o arquivo zip: %w", err)
	}
	fmt.Println("Extração concluída.")

	// 9. Limpar o arquivo zip baixado
	fmt.Printf("Limpando arquivo %s...\n", zipFileName)
	if err := os.Remove(zipFilePath); err != nil {
		// Não retorna erro aqui, pois a instalação principal funcionou. Apenas avisa.
		fmt.Fprintf(os.Stderr, "Aviso: não foi possível remover o arquivo zip: %v\n", err)
	}

	fmt.Printf("✅ Maven %s instalado com sucesso em: %s\n", version, finalInstallPath)
	return finalInstallPath, nil
}

// unzip extrai um arquivo zip para um diretório de destino.
// Implementa uma verificação de segurança contra "Zip Slip".
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		// Verificação de segurança contra "Zip Slip"
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("caminho de arquivo inválido: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
