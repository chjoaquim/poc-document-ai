package main

import (
	"encoding/json"
	"fmt"
	"github.com/chjoaquim/poc-document-ai/documentai"
	"io"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/upload", uploadFileHandler)

	port := ":8080"
	fmt.Printf("Servidor rodando na porta %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Erro ao recuperar o arquivo", http.StatusBadRequest)
		return
	}
	defer file.Close()

	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil && err != io.EOF {
		http.Error(w, "Erro ao ler o arquivo", http.StatusInternalServerError)
		return
	}

	mimeType := http.DetectContentType(buf)
	fmt.Printf("MIME type detectado: %s\n", mimeType)

	// Reiniciar o leitor do arquivo (necessário porque já lemos alguns bytes)
	file.Seek(0, io.SeekStart)

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Erro ao ler o arquivo", http.StatusInternalServerError)
		return
	}

	processor := documentai.NewFileProcessor()
	fileToProcess := &documentai.FileRequest{
		Content:  fileBytes,
		MimeType: mimeType,
	}

	response, err := processor.ProcessDocument(fileToProcess)
	if err != nil {
		http.Error(w, "Erro ao processar arquivo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
