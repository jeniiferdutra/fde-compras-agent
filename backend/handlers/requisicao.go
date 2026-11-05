package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"backend/database"
	"backend/llm/llm"
	"backend/mcp"
)

type PedidoRequest struct {
	Mensagem string `json:"mensagem"`
}

type RequisicaoRecord struct {
	ID       int    `json:"id"`
	Mensagem string `json:"mensagem"`
	Resposta string `json:"resposta"`
	CriadoEm string `json:"criado_em"`
}

func CriarRequisicaoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req PedidoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Payload inválido", http.StatusBadRequest)
		return
	}

	// busca os dados no Servidor MCP
	mcpClient := mcp.NovoCliente("http://localhost:8001")
	produtosResp, err := mcpClient.BuscarProdutos(req.Mensagem)
	if err != nil {
		http.Error(w, "Erro ao interagir com o MCP: "+err.Error(), http.StatusInternalServerError)
		return
	}

	promptContexto := fmt.Sprintf(
		"O usuário pediu: '%s'. Com base nos seguintes produtos encontrados no estoque: %+v, escreva uma resposta amigável e resumida recomendando as melhores opções.",
		req.Mensagem, produtosResp.Produtos,
	)

	// envia para a OpenRouter processar
	respostaLLM, err := llm.PerguntarAoLLM(promptContexto)
	if err != nil {
		respostaLLM = fmt.Sprintf("Encontramos %d produtos para sua busca.", len(produtosResp.Produtos))
	}

	// salva no postgres a mensagem do usuário e a resposta processada pela LLM
	query := "INSERT INTO requisicoes (mensagem, resposta) VALUES ($1, $2) RETURNING id"
	var novoID int
	err = database.DB.QueryRow(query, req.Mensagem, respostaLLM).Scan(&novoID)
	if err != nil {
		http.Error(w, "Erro ao salvar no banco de dados: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// resposta completa para o cliente
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"id":       novoID,
		"status":   "sucesso",
		"mensagem": req.Mensagem,
		"resposta": respostaLLM,
		"produtos": produtosResp.Produtos,
	})
}

func ListarRequisicoesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	rows, err := database.DB.Query("SELECT id, mensagem, resposta, criado_em FROM requisicoes ORDER BY id DESC")
	if err != nil {
		http.Error(w, "Erro ao buscar requisições: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var requisicoes []RequisicaoRecord
	for rows.Next() {
		var req RequisicaoRecord
		if err := rows.Scan(&req.ID, &req.Mensagem, &req.Resposta, &req.CriadoEm); err != nil {
			continue
		}
		requisicoes = append(requisicoes, req)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(requisicoes)
}