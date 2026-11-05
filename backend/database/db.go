package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		log.Fatal("A variável DB_PASSWORD não foi configurada")
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "fde_compras"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Erro ao abrir conexão com o banco de dados: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Erro ao conectar ao PostgreSQL: %v", err)
	}

	log.Println("Conexão com o PostgreSQL estabelecida com sucesso!")

	createTables()
}

func createTables() {
	queryRequisicoes := `
	CREATE TABLE IF NOT EXISTS requisicoes (
		id SERIAL PRIMARY KEY,
		mensagem TEXT NOT NULL,
		resposta TEXT NOT NULL,
		criado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := DB.Exec(queryRequisicoes)
	if err != nil {
		log.Fatalf("Erro ao criar tabela 'requisicoes': %v", err)
	}

	queryItens := `
	CREATE TABLE IF NOT EXISTS itens_requisicao (
		id SERIAL PRIMARY KEY,
		requisicao_id INT REFERENCES requisicoes(id) ON DELETE CASCADE,
		produto_id INT,
		nome_produto TEXT NOT NULL,
		quantidade INT NOT NULL,
		preco_unitario NUMERIC(10, 2) NOT NULL,
		valor_total NUMERIC(10, 2) NOT NULL
	);
	`

	_, err = DB.Exec(queryItens)
	if err != nil {
		log.Fatalf("Erro ao criar tabela 'itens_requisicao': %v", err)
	}

	log.Println("Tabelas 'requisicoes' e 'itens_requisicao' verificadas/criadas com sucesso.")

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
    dbHost = "localhost"
	}
}