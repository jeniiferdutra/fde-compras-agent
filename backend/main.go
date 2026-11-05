package main

import (
	"log"
	"net/http"

	"backend/database"
	"backend/handlers"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK) // "Tudo 100% por aqui!"
	w.Write([]byte("Backend Agent está a rodar com sucesso!"))
}

func main() {
	// conexão com o banco
	database.Conectar()

	// registra as rotas
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/requisicoes", handlers.CriarRequisicaoHandler)

	log.Println("Backend/Agent a rodar na porta 8080...")
	
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}