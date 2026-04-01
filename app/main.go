package main

import (
	"log"
	"math/big"
	"net/http"

	"goledger-challenge/blockchain"
	"goledger-challenge/db"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Estrutura para receber o JSON na rota SET
type SetRequest struct {
	Value int64 `json:"value" binding:"required"`
}

func main() {
	// 1. Carrega variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: Nenhum arquivo .env encontrado")
	}

	// 2. Inicializa o Banco de Dados
	database, err := db.NewDB()
	if err != nil {
		log.Fatalf("Falha crítica no BD: %v", err)
	}
	log.Println("✅ Banco de Dados conectado!")

	// 3. Inicializa o Cliente Blockchain
	eth, err := blockchain.NewClient()
	if err != nil {
		log.Fatalf("Falha crítica na Blockchain: %v", err)
	}
	log.Println("✅ Blockchain Besu conectada!")

	// 4. Configura o servidor web Gin
	// gin.SetMode(gin.ReleaseMode) // Descomente para produção
	router := gin.Default()

	// 1. SET: Grava um novo valor na Blockchain
	router.POST("/set", func(c *gin.Context) {
		var req SetRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido. Envie {'value': numero}"})
			return
		}

		bigVal := big.NewInt(req.Value)
		if err := eth.SetValue(bigVal); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gravar na blockchain: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Valor salvo e minerado na blockchain!", "value": req.Value})
	})

	// 2. GET: Lê o valor atual da Blockchain
	router.GET("/get", func(c *gin.Context) {
		val, err := eth.GetValue()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler da blockchain: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"chain_value": val.String()})
	})

	// 3. SYNC: Lê da Blockchain e salva no Banco de Dados
	router.POST("/sync", func(c *gin.Context) {
		val, err := eth.GetValue()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler da chain: " + err.Error()})
			return
		}

		if err := database.SaveValue(val.String()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar no BD: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "Sincronização concluída com sucesso!",
			"synced_value": val.String(),
		})
	})

	// 4. CHECK: Compara a Blockchain com o Banco de Dados
	router.GET("/check", func(c *gin.Context) {
		chainVal, err := eth.GetValue()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler da chain"})
			return
		}

		dbVal, err := database.GetSavedValue()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler do BD"})
			return
		}

		isSynced := chainVal.String() == dbVal

		c.JSON(http.StatusOK, gin.H{
			"synced":      isSynced,
			"chain_value": chainVal.String(),
			"db_value":    dbVal,
		})
	})

	// Inicia o servidor na porta 8080
	log.Println("🚀 Servidor rodando na porta 8080...")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
