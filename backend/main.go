package main

import (
	"log"
	"net/http"

	"backend/database"
	"backend/handlers"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Backend Agent está a rodar com sucesso!"))
}

//  liberar requisições vindas do React (CORS)
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	// conexão com o banco
	database.InitDB()

	// registra as rotas aplicando o CORS na rota de requisições
	// registra as rotas
	http.HandleFunc("/", healthHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/requisicoes", enableCORS(handlers.CriarRequisicaoHandler))

	log.Println("Backend/Agent a rodar na porta 8080...")
	
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}