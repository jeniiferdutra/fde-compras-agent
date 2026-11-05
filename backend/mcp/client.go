package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// Estruturas de dados para mapear a resposta do MCP
type Produto struct {
	ID    int     `json:"id"`
	Title string  `json:"title"`
	Price float64 `json:"price"`
}

type RespostaMCP struct {
	Status   string    `json:"status"`
	Produtos []Produto `json:"produtos"`
}

type Client struct {
	BaseURL string
}

func NovoCliente(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
	}
}

// BuscarProdutos faz a requisição HTTP GET para o Servidor MCP na porta 8001
func (c *Client) BuscarProdutos(query string) (*RespostaMCP, error) {
	urlFinal := fmt.Sprintf("%s/tools/buscar_produtos?q=%s", c.BaseURL, url.QueryEscape(query))

	resp, err := http.Get(urlFinal)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao servidor MCP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mcp retornou código de erro: %d", resp.StatusCode)
	}

	var resultado RespostaMCP
	if err := json.NewDecoder(resp.Body).Decode(&resultado); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta JSON do MCP: %w", err)
	}

	return &resultado, nil
}