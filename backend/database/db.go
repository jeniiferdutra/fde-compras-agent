package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // Driver do Postgres
)

var DB *sql.DB

// inicializa a conexão com o PostgreSQL
func Conectar() {
	connStr := "host=localhost port=5432 user=postgres password=2738 dbname=fde_compras sslmode=disable"

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Erro ao abrir conexão com o banco: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Erro ao conectar com o Postgres: %v", err)
	}

	fmt.Println("Conexão com o PostgreSQL estabelecida com sucesso!")
	
	// Cria a tabela de requisições caso ela não exista
	criarTabela()
}

func criarTabela() {
	query := `
	CREATE TABLE IF NOT EXISTS requisicoes (
		id SERIAL PRIMARY KEY,
		mensagem TEXT NOT NULL,
		resposta TEXT NOT NULL,
		criado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := DB.Exec(query)
	if err != nil {
		log.Fatalf("Erro ao criar tabela no banco: %v", err)
	}
}