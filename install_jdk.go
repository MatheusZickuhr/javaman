package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

const (
	// adoptiumAPIURL is the URL pattern for the Adoptium API to get the latest GA release.
	// Parameters: {version}, {os}, {arch}, {image_type}, {jvm_impl}
	adoptiumAPIURL = "https://api.adoptium.net/v3/binary/latest/%s/ga/windows/x64/jdk/hotspot/normal/eclipse"
)

func installJdk(version string) (string, error) {
	fmt.Printf("Starting JDK %s installation...\n", version)

	homeDir, err := getUserHomeDir()
	if err != nil {
		return "", err
	}

	installRootPath := filepath.Join(homeDir, ".jdks")

	if err := os.MkdirAll(installRootPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create base directory %s: %w", installRootPath, err)
	}

	archivePath, err := downloadJdkArchive(version, installRootPath)
	if err != nil {
		return "", err
	}
	defer cleanupArchive(archivePath)

	fmt.Printf("Extracting files to %s...\n", installRootPath)
	jdkPath, err := unzip(archivePath, installRootPath)
	if err != nil {
		return "", fmt.Errorf("failed to extract JDK: %w", err)
	}

	fmt.Printf("JDK %s installed successfully at: %s\n", version, jdkPath)
	return jdkPath, nil
}

func downloadJdkArchive(version, downloadDir string) (string, error) {
	fmt.Println("--> Resolving final download URL...")
	finalURL, err := getFinalJdkDownloadURL(version)
	if err != nil {
		return "", fmt.Errorf("failed to get download URL: %w", err)
	}
	fmt.Printf("--> URL found: %s\n", finalURL)

	fileName := filepath.Base(finalURL)
	destinationPath := filepath.Join(downloadDir, fileName)

	fmt.Printf("--> Downloading to: %s\n", destinationPath)
	if err := downloadFile(finalURL, destinationPath); err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}

	fmt.Printf("\nJDK %s download completed!\n", version)
	return destinationPath, nil
}

// getFinalJdkDownloadURL queries the Adoptium API to find the direct download URL.
// The API responds with a redirect (HTTP 307) to the actual file.
func getFinalJdkDownloadURL(version string) (string, error) {
	apiURL := fmt.Sprintf(adoptiumAPIURL, version)

	// Create a custom HTTP client that does NOT automatically follow redirects.
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Stop following redirects
		},
	}

	resp, err := client.Get(apiURL)
	if err != nil {
		// If the error is a url.Error, check if the cause was our intended stop.
		if urlErr, ok := err.(*url.Error); ok && urlErr.Err == http.ErrUseLastResponse {
			// This is the expected behavior, ignore this specific error.
		} else {
			return "", fmt.Errorf("http request to the API failed: %w", err)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTemporaryRedirect {
		return "", fmt.Errorf("adoptium API did not return a redirect. Status: %s", resp.Status)
	}

	finalURL := resp.Header.Get("Location")
	if finalURL == "" {
		return "", fmt.Errorf("'Location' header not found in the API response")
	}

	return finalURL, nil
}
