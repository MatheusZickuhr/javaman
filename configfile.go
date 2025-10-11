package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type ConfigFile struct {
	Jdks   []Installation
	Mavens []Installation
}

func (c ConfigFile) GetJdkByVersion(version string) *Installation {
	for i := range c.Jdks {
		if version == c.Jdks[i].Version {
			return &c.Jdks[i] // CORRETO: Retorna o endereço do elemento na lista
		}
	}

	return nil
}

func (c ConfigFile) GetMavenByVersion(version string) *Installation {

	for i := range c.Mavens {
		if version == c.Mavens[i].Version {
			return &c.Mavens[i] // CORRETO: Retorna o endereço do elemento na lista
		}
	}

	return nil
}

func loadConfigFile() *ConfigFile {
	jsonFile, err := os.Open("javaman.json")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var configFile ConfigFile

	err = json.Unmarshal(byteValue, &configFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return &configFile
}

func saveConfigFile(config *ConfigFile) error {

	jsonAsByteArray, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error converting struct to JSON: %w", err)
	}

	err = os.WriteFile("javaman.json", jsonAsByteArray, 0644)
	if err != nil {
		return fmt.Errorf("error writing JSON file: %w", err)
	}

	return nil
}
