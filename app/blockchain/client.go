package blockchain

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	ethClient *ethclient.Client
	contract  common.Address
	parsedABI abi.ABI
}

// Estrutura para extrair apenas a ABI do JSON gigante do Foundry
type FoundryArtifact struct {
	ABI json.RawMessage `json:"abi"`
}

func NewClient() (*Client, error) {
	rpcURL := os.Getenv("RPC_URL")
	contractAddr := os.Getenv("CONTRACT_ADDRESS")

	// 1. Conecta no nó do Besu
	client, err := ethclient.DialContext(context.Background(), rpcURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar no Besu: %w", err)
	}

	// 2. Lê o arquivo do contrato compilado
	artifactData, err := os.ReadFile("../SimpleStorage/out/SimpleStorage.sol/SimpleStorage.json")
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo do contrato: %w", err)
	}

	var artifact FoundryArtifact
	if err := json.Unmarshal(artifactData, &artifact); err != nil {
		return nil, fmt.Errorf("erro no parse do JSON do contrato: %w", err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(string(artifact.ABI)))
	if err != nil {
		return nil, fmt.Errorf("erro no parse da ABI: %w", err)
	}

	return &Client{
		ethClient: client,
		contract:  common.HexToAddress(contractAddr),
		parsedABI: parsedABI,
	}, nil
}

// GET: Lê o valor atual na Blockchain
func (c *Client) GetValue() (*big.Int, error) {
	contract := bind.NewBoundContract(c.contract, c.parsedABI, c.ethClient, c.ethClient, c.ethClient)

	var output []interface{}
	err := contract.Call(&bind.CallOpts{Context: context.Background()}, &output, "get")
	if err != nil {
		return nil, fmt.Errorf("erro ao chamar get(): %w", err)
	}

	val, ok := output[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("erro ao converter output para *big.Int")
	}

	return val, nil
}

// SET: Grava um novo valor na Blockchain
func (c *Client) SetValue(value *big.Int) error {
	privateKeyHex := strings.TrimSpace(os.Getenv("PRIVATE_KEY"))
	if privateKeyHex == "" {
		return fmt.Errorf("PRIVATE_KEY não definida (configure em app/.env ou ../SimpleStorage/.env)")
	}
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return fmt.Errorf("erro ao carregar chave privada: %w", err)
	}

	chainID, err := c.ethClient.ChainID(context.Background())
	if err != nil {
		return fmt.Errorf("erro ao obter ChainID: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return fmt.Errorf("erro ao criar transactor: %w", err)
	}

	contract := bind.NewBoundContract(c.contract, c.parsedABI, c.ethClient, c.ethClient, c.ethClient)

	tx, err := contract.Transact(auth, "set", value)
	if err != nil {
		return fmt.Errorf("erro ao enviar transação: %w", err)
	}

	fmt.Printf("⏳ Transação enviada! Hash: %s\n⏳ Aguardando ser minerada pelo Besu...\n", tx.Hash().Hex())

	// Aguarda a transação ser incluída num bloco (mineração)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	receipt, err := bind.WaitMined(ctx, c.ethClient, tx)
	if err != nil {
		return fmt.Errorf("erro aguardando mineração: %w", err)
	}

	if receipt.Status != 1 {
		return fmt.Errorf("transação falhou na blockchain")
	}

	return nil
}
