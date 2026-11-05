# FDE Compras Agent

Agente de compras desenvolvido em Go, React e PostgreSQL, utilizando LLM via OpenRouter e ferramentas MCP para consultar produtos na API DummyJSON e registrar requisições de compra.

## Como executar

**Pré-requisitos:** Docker Desktop e uma chave da OpenRouter.

1. Clone o repositório e acesse a pasta:

```bash
git clone <URL_DO_REPOSITORIO>
cd fde-compras-agent
```

2. Crie o arquivo `.env` na raiz do projeto:

```env
POSTGRES_USER=postgres
POSTGRES_PASSWORD=sua_senha
POSTGRES_DB=fde_compras

DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=sua_senha
DB_NAME=fde_compras

OPENROUTER_API_KEY=sua_chave
```

3. Inicie os serviços:

```bash
docker compose up --build -d
```

4. Verifique se os containers estão ativos:

```bash
docker compose ps
```

Para acompanhar os logs:

```bash
docker compose logs -f
```

## Arquitetura e decisões técnicas

O projeto possui quatro componentes:

- **Frontend (React):** interface de chat para receber solicitações de compra.
- **Backend:** coordena o agente, processa os pedidos e disponibiliza os endpoints de requisições.
- **Servidor MCP (Go):** disponibiliza as ferramentas `buscar_produtos(termo)` e `detalhar_produto(id)`, integradas à API DummyJSON.
- **PostgreSQL:** armazena as requisições de compra e seus respectivos itens.

O modelo de linguagem, integrado via OpenRouter, interpreta os pedidos e decide quando utilizar as ferramentas MCP, sem depender de um fluxo fixo de chamadas definido no código.

O PostgreSQL foi escolhido para garantir a persistência das requisições e permitir sua consulta posterior. A inicialização do banco e dos serviços é realizada pelo Docker Compose.

## Endpoints

| Método | Endpoint | Descrição |
|---|---|---|
| POST | `/requisicoes` | Processa um pedido e registra a requisição de compra. |
| GET | `/requisicoes` | Lista as requisições armazenadas. |

## Testes

Os testes das ferramentas MCP podem ser executados sem chave de LLM:

```bash
go test ./...
```

Execute o comando no diretório do módulo Go.

## Uso de IA

ChatGPT e Gemini foi utilizado como apoio à programação, depuração e documentação do projeto.
