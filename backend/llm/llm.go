package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type FunctionDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

type Tool struct {
	Type     string             `json:"type"`
	Function FunctionDefinition `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type OpenRouterRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools,omitempty"`
}

type Choice struct {
	Message Message `json:"message"`
}

type OpenRouterResponse struct {
	Choices []Choice `json:"choices"`
}

// envia a mensagem e as ferramentas disponíveis para a OpenRouter
func PerguntarAoLLMComFerramentas(prompt string) (string, []ToolCall, error) {
	url := "https://openrouter.ai/api/v1/chat/completions"

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return "", nil, fmt.Errorf("variável de ambiente OPENROUTER_API_KEY não configurada")
	}

	tools := []Tool{
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "buscar_produtos",
				Description: "Busca produtos no catálogo da DummyJSON com base num termo de pesquisa.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"termo": map[string]any{
							"type":        "string",
							"description": "O termo de busca do produto, ex: laptop, telemóvel",
						},
					},
					"required": []string{"termo"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "detalhar_produto",
				Description: "Obtém detalhes completos de um produto específico informando o ID.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id": map[string]any{
							"type":        "string",
							"description": "O ID numérico ou string do produto",
						},
					},
					"required": []string{"id"},
				},
			},
		},
	}

	reqBody := OpenRouterRequest{
		Model: "openai/gpt-4o",
		Messages: []Message{
			{
				Role: "system",
				Content: `Você é um agente de compras inteligente.

				Sua função é pesquisar produtos no catálogo usando a ferramenta buscar_produtos.

				REGRAS:
				1. Sempre que o usuário pedir para encontrar, pesquisar, comparar ou recomendar produtos, chame a ferramenta buscar_produtos.
				2. O catálogo utiliza termos em inglês. Traduza o nome do produto para inglês antes de pesquisar.
				3. Exemplos:
				- notebook → laptop
				- celular → smartphone
				- televisão → TV
				- geladeira → refrigerator
				- fone de ouvido → headphones
				4. Não invente produtos nem preços.
				5. Se o usuário pedir detalhes de um produto informando seu ID, utilize detalhar_produto.
				6. Para perguntas que não envolvam pesquisa de produtos, responda normalmente.`,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Tools: tools,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("erro ao serializar request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "http://localhost:8080")
	req.Header.Set("X-Title", "FDE Compras Agent")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("erro de rede: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("openrouter retornou status code: %d", resp.StatusCode)
	}

	var openRouterResp OpenRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&openRouterResp); err != nil {
		return "", nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	if len(openRouterResp.Choices) == 0 {
		return "", nil, fmt.Errorf("nenhuma resposta retornada")
	}

	msg := openRouterResp.Choices[0].Message

	log.Printf("OpenRouter retornou %d tool calls", len(msg.ToolCalls))
	log.Printf("Conteúdo retornado: %s", msg.Content)

	return msg.Content, msg.ToolCalls, nil
}
