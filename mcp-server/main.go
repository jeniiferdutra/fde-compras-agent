
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const dummyJSONBaseURL = "https://dummyjson.com"

type Product struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type DummyResponse struct {
	Products []Product `json:"products"`
}

type BuscarProdutosInput struct {
	Termo string `json:"termo" jsonschema:"termo utilizado para pesquisar produtos no catálogo"`
}

type DetalharProdutoInput struct {
	ID string `json:"id" jsonschema:"identificador do produto"`
}

type ServerOutput struct {
	Message string `json:"message,omitempty"`
}

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// buscarProdutos consulta o catálogo da DummyJSON.
func buscarProdutos(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input BuscarProdutosInput,
) (*mcp.CallToolResult, []Product, error) {

	if input.Termo == "" {
		return nil, nil, fmt.Errorf("o termo de busca é obrigatório")
	}

	apiURL := dummyJSONBaseURL +
		"/products/search?q=" +
		url.QueryEscape(input.Termo)

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		apiURL,
		nil,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao consultar DummyJSON: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf(
			"DummyJSON retornou status HTTP %d",
			resp.StatusCode,
		)
	}

	var resultado DummyResponse

	if err := json.NewDecoder(resp.Body).Decode(&resultado); err != nil {
		return nil, nil, fmt.Errorf(
			"erro ao interpretar resposta da DummyJSON: %w",
			err,
		)
	}

	return nil, resultado.Products, nil
}

// detalharProduto consulta um produto específico pelo ID.
func detalharProduto(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input DetalharProdutoInput,
) (*mcp.CallToolResult, Product, error) {

	if input.ID == "" {
		return nil, Product{}, fmt.Errorf("o ID do produto é obrigatório")
	}

	apiURL := dummyJSONBaseURL +
		"/products/" +
		url.PathEscape(input.ID)

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		apiURL,
		nil,
	)
	if err != nil {
		return nil, Product{}, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, Product{}, fmt.Errorf("erro ao consultar DummyJSON: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, Product{}, fmt.Errorf(
			"produto com ID %s não encontrado",
			input.ID,
		)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, Product{}, fmt.Errorf(
			"DummyJSON retornou status HTTP %d",
			resp.StatusCode,
		)
	}

	var produto Product

	if err := json.NewDecoder(resp.Body).Decode(&produto); err != nil {
		return nil, Product{}, fmt.Errorf(
			"erro ao interpretar produto: %w",
			err,
		)
	}

	return nil, produto, nil
}

func main() {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "fde-compras-server",
			Version: "1.0.0",
		},
		nil,
	)

	// Registra a ferramenta de busca no protocolo MCP.
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "buscar_produtos",
			Description: "Busca produtos no catálogo da DummyJSON usando um termo de pesquisa.",
		},
		buscarProdutos,
	)

	// Registra a ferramenta de detalhamento no protocolo MCP.
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "detalhar_produto",
			Description: "Retorna os detalhes de um produto específico a partir do seu ID.",
		},
		detalharProduto,
	)

	// Cria o transporte MCP Streamable HTTP.
	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			return server
		},
		&mcp.StreamableHTTPOptions{
			JSONResponse: true,
		},
	)

	mux := http.NewServeMux()

	// Endpoint do protocolo MCP.
	mux.Handle("/mcp", mcpHandler)

	// Endpoint simples para verificar a saúde do serviço.
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "MCP Server funcionando")
	})

	httpServer := &http.Server{
		Addr:              ":8001",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("Servidor MCP iniciado na porta 8001")
	log.Println("Endpoint MCP: http://localhost:8001/mcp")

	if err := httpServer.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatalf("Erro ao iniciar servidor MCP: %v", err)
	}
}
