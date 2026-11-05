package main

import (
	 "encoding/json"
    "fmt"
    "net/http"
    "net/url"
)

type Product struct {
	ID    int     `json:"id"`
	Title string  `json:"title"`
	Price float64 `json:"price"`
}

type DummyResponse struct {
	Products []Product `json:"products"`
}

// busca produtos por termo na DummyJSON
func buscarProdutosHandler(w http.ResponseWriter, r *http.Request) {
	termo := r.URL.Query().Get("termo")
	if termo == "" {
		http.Error(w, "Parâmetro 'termo' é obrigatório", http.StatusBadRequest)
		return
	}

	apiURL := "https://dummyjson.com/products/search?q=" + url.QueryEscape(termo)
	
	resp, err := http.Get(apiURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, "Erro ao comunicar com a DummyJSON", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	var dummyRes DummyResponse
	if err := json.NewDecoder(resp.Body).Decode(&dummyRes); err != nil {
		http.Error(w, "Erro ao processar resposta da API", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dummyRes.Products)
}

// detalha um produto específico pelo ID
func detalharProdutoHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID do produto é obrigatório", http.StatusBadRequest)
		return
	}

	apiURL := fmt.Sprintf("https://dummyjson.com/products/%s", id)
	
	resp, err := http.Get(apiURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, "Produto não encontrado", http.StatusNotFound)
		return
	}
	defer resp.Body.Close()

	var produto Product
	if err := json.NewDecoder(resp.Body).Decode(&produto); err != nil {
		http.Error(w, "Erro ao processar produto", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(produto)
}

// liberar requisições (CORS) caso necessário
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	mux := http.NewServeMux()
	
	// registro das rotas / ferramentas do MCP
	mux.HandleFunc("GET /tools/buscar_produtos", enableCORS(buscarProdutosHandler))
	mux.HandleFunc("GET /tools/detalhar_produto/{id}", enableCORS(detalharProdutoHandler))

	fmt.Println("Servidor MCP (Go) a rodar na porta 8001...")
	if err := http.ListenAndServe(":8001", mux); err != nil {
		fmt.Printf("Erro ao iniciar servidor: %v\n", err)
	}
}