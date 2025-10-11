package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func getUserHomeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get user home directory: %w", err)
	}
	return homeDir, nil
}

func downloadFile(url, destinationPath string) error {

	file, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("could not create destination file: %w", err)
	}
	defer file.Close()

	resp, err := http.Get(url)
	if err != nil {
		os.Remove(destinationPath)
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.Remove(destinationPath)
		return fmt.Errorf("download failed, server status: %s", resp.Status)
	}

	// Copy the response body to the local file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		os.Remove(destinationPath)
		return fmt.Errorf("failed to save file content: %w", err)
	}

	return nil
}

func unzip(zipFilePath, destination string) (string, error) {
	zipFile, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return "", fmt.Errorf("could not open zip file '%s': %w", zipFilePath, err)
	}
	defer zipFile.Close()

	if err := os.MkdirAll(destination, os.ModePerm); err != nil {
		return "", fmt.Errorf("could not create destination directory '%s': %w", destination, err)
	}

	for _, file := range zipFile.File {
		filePath := filepath.Join(destination, file.Name)

		// Security check against "Zip Slip".
		if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(os.PathSeparator)) {
			return "", fmt.Errorf("invalid file path (Zip Slip attempt): %s", filePath)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return "", fmt.Errorf("could not create directory for file '%s': %w", filePath, err)
		}

		destinationFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return "", fmt.Errorf("could not create destination file '%s': %w", filePath, err)
		}

		sourceFile, err := file.Open()
		if err != nil {
			destinationFile.Close()
			return "", fmt.Errorf("could not open file '%s' inside zip: %w", file.Name, err)
		}

		_, err = io.Copy(destinationFile, sourceFile)

		destinationFile.Close()
		sourceFile.Close()

		if err != nil {
			return "", fmt.Errorf("could not extract file '%s': %w", file.Name, err)
		}
	}

	// Determine the path of the extracted root directory.
	var extractedPath string
	if len(zipFile.File) > 0 {
		topLevelDir := strings.Split(zipFile.File[0].Name, string(os.PathSeparator))[0]
		extractedPath = filepath.Join(destination, topLevelDir)
	} else {
		extractedPath = destination
	}

	return extractedPath, nil
}

func cleanupArchive(archivePath string) {
	err := os.Remove(archivePath)
	if err != nil {
		fmt.Printf("Warning: Error deleting temporary file: %v\n", err)
	}
}
