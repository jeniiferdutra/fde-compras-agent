package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"backend/database"
	"backend/llm"
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

type ItemRetorno struct {
	ProdutoID     int     `json:"produto_id"`
	NomeProduto   string  `json:"nome_produto"`
	Quantidade    int     `json:"quantidade"`
	PrecoUnitario float64 `json:"preco_unitario"`
	ValorTotal    float64 `json:"valor_total"`
}

// Função auxiliar para extrair a quantidade da mensagem (ex: "3 smartphones" -> 3)
func extrairQuantidade(mensagem string) int {
	re := regexp.MustCompile(`(?i)(?:preciso de|quero|comprar)?\s*(\d+)\s*(?:unidades?|itens?)?`)
	matches := re.FindStringSubmatch(mensagem)
	if len(matches) > 1 {
		if q, err := strconv.Atoi(matches[1]); err == nil && q > 0 {
			return q
		}
	}

	reNum := regexp.MustCompile(`\d+`)
	match := reNum.FindString(mensagem)
	if match != "" {
		var q int
		fmt.Sscanf(match, "%d", &q)
		if q > 0 {
			return q
		}
	}
	return 1
}

// Função auxiliar para extrair preço máximo da mensagem (ex: "até US$ 2.000" -> 2000.0)
func extrairPrecoMaximo(mensagem string) float64 {
	// Procura padrões como "ate 2000", "ate r$ 2000", "ate us$ 2.000"
	re := regexp.MustCompile(`(?i)at[eé]\s+(?:r\$|us\$|\$)?\s*([\d\.]+(?:,\d{2})?)`)
	match := re.FindStringSubmatch(mensagem)
	if len(match) > 1 {
		valStr := strings.ReplaceAll(match[1], ".", "")
		valStr = strings.ReplaceAll(valStr, ",", ".")
		if val, err := strconv.ParseFloat(valStr, 64); err == nil {
			return val
		}
	}
	return 0
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

	respostaLLM, toolCalls, err := llm.PerguntarAoLLMComFerramentas(req.Mensagem)
	log.Printf("Resposta LLM: %s", respostaLLM)
log.Printf("Quantidade de ferramentas chamadas: %d", len(toolCalls))

for _, call := range toolCalls {
    log.Printf("Ferramenta chamada: %s", call.Function.Name)
    log.Printf("Argumentos: %s", call.Function.Arguments)
}

	if err != nil {
		http.Error(w, "Erro ao processar com o LLM: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var produtos []mcp.Product
	mcpClient := mcp.NovoCliente("http://host.docker.internal:8001")

	for _, call := range toolCalls {
		if call.Function.Name == "buscar_produtos" {
			var args struct {
				Termo string `json:"termo"`
			}
			
			if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err == nil {
				prods, err := mcpClient.BuscarProdutos(args.Termo)
					if err != nil {
    			log.Printf("ERRO ao buscar produtos no MCP: %v", err)
					} else {
    			log.Printf("Produtos encontrados pelo MCP: %d", len(prods))
    			produtos = append(produtos, prods...)
			}
			}
		} else if call.Function.Name == "detalhar_produto" {
			var args struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err == nil {
				prod, err := mcpClient.DetalharProduto(args.ID)
				if err == nil {
					produtos = append(produtos, prod)
				}
			}
		}

	}
	
	log.Printf("Total de produtos encontrados: %d", len(produtos))

	// Extrai quantidade e teto de preço da mensagem
	quantidade := extrairQuantidade(req.Mensagem)
	precoMax := extrairPrecoMaximo(req.Mensagem)

	// Filtra os produtos pelo preço máximo, caso o usuário tenha estipulado um limite
	var produtosFiltrados []mcp.Product
	for _, p := range produtos {
		if precoMax > 0 && p.Price > precoMax {
			continue // Ignora produtos acima do teto de preço informado
		}
		produtosFiltrados = append(produtosFiltrados, p)
	}

	// Se houver produtos válidos, selecionamos o melhor/primeiro candidato para atender ao pedido exato
	// (evitando duplicar vários modelos diferentes na mesma ordem de compra)
	if len(produtosFiltrados) > 0 {
		produtos = []mcp.Product{produtosFiltrados[0]}
	} else if len(produtos) > 0 && precoMax > 0 {
		// Se encontrou produtos mas todos ultrapassavam o preço, mantemos vazio para indicar restrição
		produtos = []mcp.Product{}
	}

	if respostaLLM == "" {
		if len(produtos) > 0 {
			respostaLLM = fmt.Sprintf("Produto selecionado: %s (Qtd: %d).", produtos[0].Title, quantidade)
		} else {
			respostaLLM = "Nenhum produto atendeu aos critérios de busca ou faixa de preço solicitada."
		}
	}

	tx, err := database.DB.Begin()
	if err != nil {
		http.Error(w, "Erro ao iniciar transação no banco de dados", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var novoID int
	queryReq := "INSERT INTO requisicoes (mensagem, resposta) VALUES ($1, $2) RETURNING id"
	err = tx.QueryRow(queryReq, req.Mensagem, respostaLLM).Scan(&novoID)
	if err != nil {
		http.Error(w, "Erro ao salvar no banco de dados: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf(">>> SUCESSO! Gravado na base com ID: %d\n", novoID)

	var itensSalvos []ItemRetorno
	for _, p := range produtos {
		valorTotal := float64(quantidade) * p.Price

		queryItem := `
			INSERT INTO itens_requisicao (requisicao_id, produto_id, nome_produto, quantidade, preco_unitario, valor_total)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = tx.Exec(queryItem, novoID, p.ID, p.Title, quantidade, p.Price, valorTotal)
		if err != nil {
			http.Error(w, "Erro ao salvar item da requisição: "+err.Error(), http.StatusInternalServerError)
			return
		}

		itensSalvos = append(itensSalvos, ItemRetorno{
			ProdutoID:     p.ID,
			NomeProduto:   p.Title,
			Quantidade:    quantidade,
			PrecoUnitario: p.Price,
			ValorTotal:    valorTotal,
		})
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "Erro ao confirmar transação: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"id":       novoID,
		"status":   "sucesso",
		"mensagem": req.Mensagem,
		"resposta": respostaLLM,
		"itens":    itensSalvos,
	})
}