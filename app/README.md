# GoLedger Challenge - App Solution

This service is a REST API in Go that bridges a local Hyperledger Besu network and PostgreSQL.

## Architecture

The application is split into 3 layers:

1. API layer: `main.go`
2. Blockchain client layer: `blockchain/client.go`
3. Database layer: `db/postgres.go`

Main flow:

1. `POST /set` writes a value to the smart contract
2. `GET /get` reads the value from blockchain
3. `POST /sync` reads from blockchain and stores the value in PostgreSQL
4. `GET /check` compares blockchain value vs database value

## Prerequisites

1. Besu devnet running and contract deployed
2. Go installed
3. Docker installed (for PostgreSQL)

From repository root, start devnet and deploy contract:

```bash
make devnet-deploy
```

Copy the contract address printed at the end of deployment.

## Environment Configuration

Create `app/.env`:

```bash
cd app
cp .env.example .env 2>/dev/null || touch .env
```

Use this template:

```env
# Besu RPC (if node-1 is stalled, use 8547, 8549, or 8551)
RPC_URL=http://localhost:8545

# Deployed address from `make devnet-deploy`
CONTRACT_ADDRESS=0xYOUR_DEPLOYED_CONTRACT_ADDRESS

# Pre-funded key from SimpleStorage/.env.example
PRIVATE_KEY=0x8f2a55949038a9610f50fb23b5883af3b4ecb3c3bb792cbcefbd1542c692be63

# PostgreSQL connection used by db.NewDB()
DB_URL=postgres://admin:password@localhost:5432/goledger_db?sslmode=disable
```

Important notes:

1. The API currently loads only `app/.env`.
2. If `PRIVATE_KEY` is not present in `app/.env`, `POST /set` will fail.

## Start PostgreSQL

Inside `app/`:

```bash
docker compose up -d
```

Check container status:

```bash
docker ps --filter name=app-db-1
```

## Run the API

Inside `app/`:

```bash
go run main.go
```

Server listens on `http://localhost:8080`.

## Endpoints

### 1. Set value on chain

```bash
curl -X POST http://localhost:8080/set \
	-H "Content-Type: application/json" \
	-d '{"value": 150}'
```

### 2. Read value from chain

```bash
curl http://localhost:8080/get
```

### 3. Sync chain value to PostgreSQL

```bash
curl -X POST http://localhost:8080/sync
```

### 4. Check sync status

```bash
curl http://localhost:8080/check
```

## Troubleshooting

### Error: `erro ao conectar no banco: connect: connection refused`

Cause: PostgreSQL container is not running or not ready.

Fix:

```bash
cd app
docker compose up -d
```

### Error: `PRIVATE_KEY não definida`

Cause: `PRIVATE_KEY` missing in `app/.env`.

Fix: add the key from `SimpleStorage/.env.example` to `app/.env`.

### Error: `erro aguardando mineração: context deadline exceeded`

Cause: selected RPC node is connected but not producing blocks (or is isolated).

Fix:

1. Try another Besu node in `RPC_URL`: `http://localhost:8547`, `http://localhost:8549`, or `http://localhost:8551`.
2. If needed, restart the devnet and redeploy the contract:

```bash
make stop-devnet
make devnet-deploy
```