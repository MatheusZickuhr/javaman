package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// Constantes para a API do Adoptium
const (
	// Padrão de URL da API para obter a versão estável mais recente (GA - General Availability) de uma feature release.
	// Parâmetros: {versão}, {os}, {arch}, {image_type}, {jvm_impl}
	adoptiumAPIURL = "https://api.adoptium.net/v3/binary/latest/%s/ga/windows/x64/jdk/hotspot/normal/eclipse"
)

func downloadJdkOnTempDir(versaoJDK string) string {

	// Cria um diretório temporário para o download
	diretorio, err := os.MkdirTemp("", "jdk-downloads")
	if err != nil {
		fmt.Printf("Erro ao criar diretório temporário: %v\n", err)
		return ""
	}

	fmt.Printf("O JDK será baixado em: %s\n\n", diretorio)

	caminhoArquivo, err := DownloadOpenJDK(versaoJDK, diretorio)
	if err != nil {
		fmt.Printf("\n❌ Ocorreu um erro: %v\n", err)
		return ""
	}

	fmt.Printf("Arquivo salvo em: %s\n", caminhoArquivo)

	return caminhoArquivo
}

// DownloadOpenJDK baixa uma versão específica do OpenJDK (Temurin) para um diretório.
//
// Parâmetros:
//   - versao: A versão principal do JDK a ser baixada (ex: "11", "17", "21").
//   - diretorioDownload: O caminho para o diretório onde o arquivo .zip será salvo.
//
// Retorna:
//   - O caminho completo do arquivo baixado e um erro, se ocorrer.
func DownloadOpenJDK(versao, diretorioDownload string) (string, error) {
	fmt.Printf("Iniciando o download do OpenJDK versão %s...\n", versao)

	// 1. Obter a URL final de download usando a API do Adoptium
	fmt.Println("--> Resolvendo o URL de download final...")
	urlFinal, err := obterURLFinalDownload(versao)
	if err != nil {
		return "", fmt.Errorf("falha ao obter a URL de download: %w", err)
	}
	fmt.Printf("--> URL encontrada: %s\n", urlFinal)

	// 2. Extrair o nome do arquivo do URL
	nomeArquivo := filepath.Base(urlFinal)
	caminhoDestino := filepath.Join(diretorioDownload, nomeArquivo)

	// 3. Garantir que o diretório de destino exista
	if err := os.MkdirAll(diretorioDownload, os.ModePerm); err != nil {
		return "", fmt.Errorf("falha ao criar o diretório de download '%s': %w", diretorioDownload, err)
	}

	// 4. Baixar o arquivo
	fmt.Printf("--> Baixando para: %s\n", caminhoDestino)
	if err := baixarArquivo(urlFinal, caminhoDestino); err != nil {
		return "", fmt.Errorf("falha ao baixar o arquivo: %w", err)
	}

	fmt.Printf("\n✅ Download do OpenJDK %s concluído com sucesso!\n", versao)
	return caminhoDestino, nil
}

// obterURLFinalDownload consulta a API do Adoptium para encontrar o URL de download direto.
// A API responde com um redirecionamento (HTTP 307) para o arquivo real.
// Esta função captura o cabeçalho 'Location' do redirecionamento.
func obterURLFinalDownload(versao string) (string, error) {
	apiURL := fmt.Sprintf(adoptiumAPIURL, versao)

	// Criamos um cliente HTTP customizado que NÃO segue redirecionamentos automaticamente.
	// Isso nos permite inspecionar a resposta de redirecionamento.
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Sinaliza para parar de seguir redirecionamentos e retornar a resposta atual.
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(apiURL)
	if err != nil {
		// Se o erro for do tipo url.Error, verificamos se a causa foi o nosso
		// "http.ErrUseLastResponse", o que é o comportamento esperado.
		if urlErr, ok := err.(*url.Error); ok && urlErr.Err == http.ErrUseLastResponse {
			// Ignoramos este erro específico e continuamos
		} else {
			return "", fmt.Errorf("falha na requisição HTTP para a API: %w", err)
		}
	}
	defer resp.Body.Close()

	// A API deve responder com um status de redirecionamento (307 Temporary Redirect)
	if resp.StatusCode != http.StatusTemporaryRedirect {
		return "", fmt.Errorf("a API do Adoptium não retornou um redirecionamento. Status: %s", resp.Status)
	}

	// O URL final está no cabeçalho "Location" da resposta.
	urlFinal := resp.Header.Get("Location")
	if urlFinal == "" {
		return "", fmt.Errorf("o cabeçalho 'Location' não foi encontrado na resposta da API")
	}

	return urlFinal, nil
}

// baixarArquivo realiza o download de um arquivo de um URL para um caminho local.
func baixarArquivo(url, caminhoDestino string) error {
	// Cria o arquivo de destino
	arquivo, err := os.Create(caminhoDestino)
	if err != nil {
		return fmt.Errorf("não foi possível criar o arquivo de destino: %w", err)
	}
	defer arquivo.Close()

	// Faz a requisição GET para o URL do arquivo
	resp, err := http.Get(url)
	if err != nil {
		// Se houver erro, remove o arquivo criado que está vazio.
		os.Remove(caminhoDestino)
		return fmt.Errorf("não foi possível fazer a requisição HTTP para o download: %w", err)
	}
	defer resp.Body.Close()

	// Verifica se a requisição foi bem-sucedida
	if resp.StatusCode != http.StatusOK {
		os.Remove(caminhoDestino)
		return fmt.Errorf("falha ao baixar o arquivo, status do servidor: %s", resp.Status)
	}

	// Copia o corpo da resposta (o arquivo) para o arquivo local
	_, err = io.Copy(arquivo, resp.Body)
	if err != nil {
		os.Remove(caminhoDestino)
		return fmt.Errorf("falha ao salvar o conteúdo do arquivo: %w", err)
	}

	return nil
}

// extrairJDK extrai um arquivo ZIP de uma JDK para a pasta "jdks" no diretório do usuário.
// Retorna o caminho completo para o diretório de extração em caso de sucesso e um erro caso contrário.
//
// Parâmetros:
//
//	zipFilePath: O caminho para o arquivo ZIP da JDK.
//
// Retornos:
//
//	string: O caminho onde os arquivos foram extraídos.
//	error:  Um erro, se ocorrer algum durante o processo.
func extrairJDK(zipFilePath string) (string, error) {
	// 1. Obter o diretório home do usuário atual.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("não foi possível obter o diretório do usuário: %w", err)
	}

	// 2. Criar o caminho completo para a pasta ".jdks". Este será nosso valor de retorno.
	destDir := filepath.Join(homeDir, ".jdks")

	// 3. Criar a pasta "jdks" se ela não existir.
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("não foi possível criar o diretório de destino '%s': %w", destDir, err)
	}

	// 4. Abrir o arquivo ZIP para leitura.
	r, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return "", fmt.Errorf("não foi possível abrir o arquivo ZIP '%s': %w", zipFilePath, err)
	}
	defer r.Close()

	// 5. Iterar sobre cada arquivo dentro do ZIP.
	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)

		if !filepath.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return "", fmt.Errorf("caminho de arquivo inválido: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return "", fmt.Errorf("não foi possível criar o diretório para o arquivo '%s': %w", fpath, err)
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return "", fmt.Errorf("não foi possível criar o arquivo de destino '%s': %w", fpath, err)
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return "", fmt.Errorf("não foi possível abrir o arquivo '%s' dentro do ZIP: %w", f.Name, err)
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return "", fmt.Errorf("não foi possível extrair o arquivo '%s': %w", f.Name, err)
		}
	}

	// Em caso de sucesso, retorna o caminho de destino e nenhum erro (nil).
	return filepath.Join(destDir, r.File[0].Name), nil
}
