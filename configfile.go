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
	// Itere usando o índice (i) e o valor (inst)
	for i := range c.Jdks {
		// Acesse o elemento diretamente pelo índice para obter seu endereço
		if version == c.Jdks[i].Version {
			return &c.Jdks[i] // CORRETO: Retorna o endereço do elemento na lista
		}
	}

	return nil
}

func (c ConfigFile) GetMavenByVersion(version string) *Installation {

	// Itere usando o índice (i) e o valor (inst)
	for i := range c.Mavens {
		// Acesse o elemento diretamente pelo índice para obter seu endereço
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

// Retornos:
//
//	error: Um erro, se ocorrer algum durante o processo.
func saveConfigFile(config *ConfigFile) error {
	// 1. Converter (Marshal) a struct para um slice de bytes no formato JSON.
	// Usamos MarshalIndent para que o JSON fique formatado (com quebras de linha e indentação),
	// o que o torna mais legível para humanos.
	// O "" significa sem prefixo por linha, e "  " significa usar 2 espaços para indentação.
	dadosJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		// Retorna um erro mais descritivo envolvendo o erro original.
		return fmt.Errorf("erro ao converter a struct para JSON: %w", err)
	}

	// 2. Gravar o slice de bytes no arquivo "javaman.json".
	// A função os.WriteFile lida com a criação do arquivo se ele não existir
	// ou com a substituição do conteúdo se ele já existir.
	// 0644 é uma permissão de arquivo padrão (leitura/escrita para o proprietário,
	// e apenas leitura para os outros).
	err = os.WriteFile("javaman.json", dadosJSON, 0644)
	if err != nil {
		return fmt.Errorf("erro ao gravar o arquivo JSON: %w", err)
	}

	// 3. Se tudo correu bem, retorna nil (sem erro).
	return nil
}
