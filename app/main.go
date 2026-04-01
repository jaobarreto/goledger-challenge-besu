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

type SetRequest struct {
	Value int64 `json:"value" binding:"required"`
}

type chainClient interface {
	SetValue(value *big.Int) error
	GetValue() (*big.Int, error)
}

type valueStore interface {
	SaveValue(val string) error
	GetSavedValue() (string, error)
}

func setupRouter(eth chainClient, database valueStore) *gin.Engine {
	router := gin.Default()

	router.StaticFile("/swagger/doc.json", "docs/openapi.json")
	router.GET("/swagger", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, `<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>GoLedger API Docs</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
      window.onload = () => {
        window.ui = SwaggerUIBundle({
          url: '/swagger/doc.json',
          dom_id: '#swagger-ui',
          presets: [SwaggerUIBundle.presets.apis],
          layout: 'BaseLayout'
        });
      };
    </script>
  </body>
</html>`)
	})

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

	router.GET("/get", func(c *gin.Context) {
		val, err := eth.GetValue()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler da blockchain: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"chain_value": val.String()})
	})

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

	return router
}

func loadEnv() {
	if err := godotenv.Load(); err == nil {
		return
	}

	if err := godotenv.Load("../SimpleStorage/.env"); err != nil {
		log.Println("Aviso: Nenhum arquivo .env encontrado")
	}
}

func main() {
	loadEnv()

	database, err := db.NewDB()
	if err != nil {
		log.Fatalf("Falha crítica no BD: %v", err)
	}
	log.Println("Banco de Dados conectado")

	eth, err := blockchain.NewClient()
	if err != nil {
		log.Fatalf("Falha crítica na Blockchain: %v", err)
	}
	log.Println("Blockchain Besu conectada")

	router := setupRouter(eth, database)

	log.Println("Servidor rodando na porta 8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
