package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenRouterRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Choice struct {
	Message Message `json:"message"`
}

type OpenRouterResponse struct {
	Choices []Choice `json:"choices"`
}

func PerguntarAoLLM(prompt string) (string, error) {
	url := "https://openrouter.ai/api/v1/chat/completions"

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("variável de ambiente OPENROUTER_API_KEY não configurada")
	}

	reqBody := OpenRouterRequest{
		Model: "openai/gpt-4o-mini",
		Messages: []Message{
			{
				Role:    "system",
				Content: "Você é um assistente de e-commerce útil e direto que ajuda a estruturar respostas com base em dados de produtos.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar request para LLM: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição HTTP para OpenRouter: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "http://localhost:8080")
	req.Header.Set("X-Title", "FDE Compras Agent")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro de rede ao chamar OpenRouter: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openrouter retornou status code inesperado: %d", resp.StatusCode)
	}

	var openRouterResp OpenRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&openRouterResp); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta da OpenRouter: %w", err)
	}

	if len(openRouterResp.Choices) == 0 {
		return "", fmt.Errorf("nenhuma resposta retornada pela OpenRouter")
	}

	return openRouterResp.Choices[0].Message.Content, nil
}