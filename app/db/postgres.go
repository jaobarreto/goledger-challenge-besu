package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

type DB struct {
	Conn *sql.DB
}

func NewDB() (*DB, error) {
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		return nil, fmt.Errorf("DB_URL não encontrada no .env")
	}

	conn, err := sql.Open("postgres", dbUrl)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir banco: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("erro ao conectar no banco: %w", err)
	}

	// Cria a tabela.
	query := `CREATE TABLE IF NOT EXISTS contract_data (
		id INT PRIMARY KEY, 
		val VARCHAR(255)
	)`
	if _, err := conn.Exec(query); err != nil {
		return nil, fmt.Errorf("erro ao criar tabela: %w", err)
	}

	return &DB{Conn: conn}, nil
}

func (db *DB) SaveValue(val string) error {
	query := `INSERT INTO contract_data (id, val) VALUES (1, $1) 
	          ON CONFLICT (id) DO UPDATE SET val = EXCLUDED.val`
	_, err := db.Conn.Exec(query, val)
	return err
}

func (db *DB) GetSavedValue() (string, error) {
	var val string
	err := db.Conn.QueryRow(`SELECT val FROM contract_data WHERE id = 1`).Scan(&val)
	if err == sql.ErrNoRows {
		return "", nil // Se ainda não sincronizou nada, retorna vazio
	}
	return val, err
}
