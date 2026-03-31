package main

import (
	"fmt"
	"log"

	"goledger-challenge/db"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: Nenhum arquivo .env encontrado")
	}

	database, err := db.NewDB()
	if err != nil {
		log.Fatalf("Falha crítica: %v", err)
	}
	fmt.Println("Conectado ao PostgreSQL com sucesso!")

	err = database.SaveValue("100")
	if err != nil {
		log.Fatalf("Erro ao salvar: %v", err)
	}
	fmt.Println("Valor '100' salvo no banco.")

	val, err := database.GetSavedValue()
	if err != nil {
		log.Fatalf("Erro ao ler: %v", err)
	}
	fmt.Printf("Valor lido do banco: %s\n", val)
}
