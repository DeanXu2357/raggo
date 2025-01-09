package unstructured

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
)

type UnstructuredService struct {
	baseURL string
}

type UnstructuredElement struct {
	Type      string   `json:"type"`
	Text      string   `json:"text"`
	ElementID string   `json:"element_id"`
	Metadata  Metadata `json:"metadata"`
}

type Metadata struct {
	Filename    string      `json:"filename,omitempty"`
	Filetype    string      `json:"filetype,omitempty"`
	PageNumber  int         `json:"page_number,omitempty"`
	Coordinates Coordinates `json:"coordinates,omitempty"`
	TableHTML   string      `json:"table_html,omitempty"`
}

type Coordinates struct {
	Points [][]float64 `json:"points"`
	System string      `json:"system"`
}

func NewUnstructuredService(baseURL string) *UnstructuredService {
	return &UnstructuredService{
		baseURL: baseURL,
	}
}

func (s *UnstructuredService) ConvertPDFToText(filename string, content []byte) ([]UnstructuredElement, error) {
	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)

	// Create form file
	fileWriter, err := multipartWriter.CreateFormFile("files", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %v", err)
	}

	// Write file content
	if _, err = io.Copy(fileWriter, bytes.NewReader(content)); err != nil {
		return nil, fmt.Errorf("failed to write file content: %v", err)
	}

	// Write additional fields
	if err := multipartWriter.WriteField("chunking_strategy", "by_title"); err != nil {
		log.Printf("Failed to write chunking strategy: %v", err)
		return nil, fmt.Errorf("failed to write chunking strategy: %v", err)
	}

	if err := multipartWriter.WriteField("max_characters", "5000"); err != nil {
		log.Printf("Failed to write max characters: %v", err)
		return nil, fmt.Errorf("failed to write max characters: %v", err)
	}

	if err := multipartWriter.WriteField("combine_under_n_chars", "3500"); err != nil {
		log.Printf("Failed to write output format: %v", err)
		return nil, fmt.Errorf("failed to write output format: %v", err)
	}

	if err := multipartWriter.WriteField("output_format", "application/json"); err != nil {
		log.Printf("Failed to write output format: %v", err)
		return nil, fmt.Errorf("failed to write output format: %v", err)
	}

	multipartWriter.Close()

	// Create request
	httpReq, err := http.NewRequest("POST", s.baseURL+"/general/v0/general", &requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	// Send request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to convert PDF: %s", resp.Status)
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Response: %s", string(body))
		return nil, fmt.Errorf("conversion service error: %s", resp.Status)
	}

	// Parse response
	var elements []UnstructuredElement
	if err := json.NewDecoder(resp.Body).Decode(&elements); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return elements, nil
}
