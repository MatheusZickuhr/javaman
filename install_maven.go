package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func installMaven(version string) (string, error) {
	fmt.Printf("Starting Maven %s installation...\n", version)

	if runtime.GOOS != "windows" {
		return "", fmt.Errorf("this script is designed to run only on Windows")
	}

	homeDir, err := getUserHomeDir()
	if err != nil {
		return "", err
	}

	installRootPath := filepath.Join(homeDir, ".mvns")
	versionFolderName := fmt.Sprintf("apache-maven-%s", version)
	finalInstallPath := filepath.Join(installRootPath, versionFolderName)
	zipFileName := fmt.Sprintf("apache-maven-%s-bin.zip", version)
	archivePath := filepath.Join(installRootPath, zipFileName)

	if _, err := os.Stat(finalInstallPath); err == nil {
		fmt.Printf("Maven version %s is already installed at %s\n", version, finalInstallPath)
		return finalInstallPath, nil
	}

	if err := os.MkdirAll(installRootPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create base directory %s: %w", installRootPath, err)
	}
	fmt.Printf("Installation directory: %s\n", installRootPath)

	downloadURL := fmt.Sprintf("https://archive.apache.org/dist/maven/maven-3/%s/binaries/apache-maven-%s-bin.zip", version, version)

	fmt.Printf("Downloading Maven %s from %s\n", version, downloadURL)
	if err := downloadFile(downloadURL, archivePath); err != nil {
		return "", fmt.Errorf("failed to download: %w. Please check if version '%s' is valid", err, version)
	}
	fmt.Println("Download completed successfully.")
	defer cleanupArchive(archivePath)

	fmt.Printf("Extracting files to %s...\n", installRootPath)
	if _, err := unzip(archivePath, installRootPath); err != nil {
		return "", fmt.Errorf("failed to extract zip file: %w", err)
	}
	fmt.Println("Extraction completed.")

	fmt.Printf("Maven %s installed successfully at: %s\n", version, finalInstallPath)
	return finalInstallPath, nil
}
