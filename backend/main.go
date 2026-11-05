package main

import (
	"fmt"
	"net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK) // "Tudo 100% por aqui!"
	w.Write([]byte("Backend Agent está a rodar com sucesso!"))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	fmt.Println("Backend/Agent a rodar na porta 8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Printf("Erro ao iniciar o backend: %v\n", err)
	}
}