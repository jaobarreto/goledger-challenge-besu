package main

import (
	"fmt"
	"log"
	"math/big"

	"goledger-challenge/blockchain"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env", "../SimpleStorage/.env"); err != nil {
		log.Println("Aviso: nenhum arquivo .env encontrado em app/.env ou ../SimpleStorage/.env")
	}

	fmt.Println("Conectando à Blockchain Besu...")
	eth, err := blockchain.NewClient()
	if err != nil {
		log.Fatalf("Falha crítica ao conectar no Besu: %v", err)
	}
	fmt.Println("Conectado ao nó do Besu com sucesso!")

	// 1. Lê o valor atual
	val, err := eth.GetValue()
	if err != nil {
		log.Fatalf("Erro ao ler valor inicial: %v", err)
	}
	fmt.Printf("Valor inicial no contrato: %s\n", val.String())

	// 2. Grava um novo valor (ex: 50)
	novoValor := big.NewInt(50)
	fmt.Printf("Enviando transação para gravar o valor: %s\n", novoValor.String())
	err = eth.SetValue(novoValor)
	if err != nil {
		log.Fatalf("Erro ao gravar valor: %v", err)
	}
	fmt.Println("Valor gravado e minerado com sucesso!")

	// 3. Lê o valor novamente para confirmar
	valFinal, err := eth.GetValue()
	if err != nil {
		log.Fatalf("Erro ao ler valor final: %v", err)
	}
	fmt.Printf("Valor final confirmado no contrato: %s\n", valFinal.String())
}
