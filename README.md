# FDE — Agente de compras

Aplicação para criar requisições de compra por meio de uma conversa. O agente consulta produtos da DummyJSON, confirma os dados dos itens e registra a requisição no PostgreSQL.

## Como executar

Requisitos: Docker Desktop com Docker Compose e uma chave da OpenRouter.

```bash
cp .env.example .env
```

Defina `OPENROUTER_API_KEY` no arquivo `.env` e inicie os serviços:

```bash
docker compose up --build
```

Acesse o chat em [http://localhost:5174](http://localhost:5174). Para encerrar, use `Ctrl+C`.

Os testes das ferramentas MCP rodam sem chave de LLM nem acesso à DummyJSON:

```bash
cd mcp_server
mvn test
```

## Decisões principais

- API e servidor MCP em Java 21 com Spring Boot; chat em React e Vite.
- O modelo escolhe quando usar as ferramentas MCP para buscar e detalhar produtos da DummyJSON.
- A API valida itens e preços antes de salvar as requisições no PostgreSQL; as tabelas são criadas automaticamente.
- As consultas ao catálogo têm timeout e mensagens de erro claras.

## Uso de IA

Usei o Codex como apoio durante o desenvolvimento, aplicando meu conhecimento no uso de IA para acelerar a implementação, avaliar alternativas e revisar o código. Revisei as decisões e o comportamento final da aplicação.
