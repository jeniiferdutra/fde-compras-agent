package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Product struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NovoCliente(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) BuscarProdutos(termo string) ([]Product, error) {
	apiURL := fmt.Sprintf(
		"%s/tools/buscar_produtos?termo=%s",
		c.BaseURL,
		url.QueryEscape(termo),
	)

	resp, err := c.HTTPClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao MCP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"MCP retornou status %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	var produtos []Product
	if err := json.NewDecoder(resp.Body).Decode(&produtos); err != nil {
		return nil, fmt.Errorf("erro ao interpretar produtos: %w", err)
	}

	return produtos, nil
}

func (c *Client) DetalharProduto(id string) (Product, error) {
	apiURL := fmt.Sprintf(
		"%s/tools/detalhar_produto/%s",
		c.BaseURL,
		url.PathEscape(id),
	)

	resp, err := c.HTTPClient.Get(apiURL)
	if err != nil {
		return Product{}, fmt.Errorf("erro ao conectar ao MCP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return Product{}, fmt.Errorf(
			"MCP retornou status %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	var produto Product
	if err := json.NewDecoder(resp.Body).Decode(&produto); err != nil {
		return Product{}, fmt.Errorf("erro ao interpretar produto: %w", err)
	}

	return produto, nil
}