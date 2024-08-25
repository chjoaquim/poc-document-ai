package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Erro ao recuperar o arquivo", http.StatusBadRequest)
		return
	}
	defer file.Close()
	outFile, err := os.Create("/uploads/" + header.Filename)
	if err != nil {
		http.Error(w, "Erro ao criar o arquivo", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		http.Error(w, "Erro ao salvar o arquivo", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Upload do arquivo %s realizado com sucesso!\n", header.Filename)
}
