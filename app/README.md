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
2. Go installed (optional, only needed for local non-container run)
3. Docker installed

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

# Besu RPC used by API container (docker compose)
RPC_URL_DOCKER=http://host.docker.internal:8547

# Deployed address from `make devnet-deploy`
CONTRACT_ADDRESS=0xYOUR_DEPLOYED_CONTRACT_ADDRESS

# Pre-funded key from SimpleStorage/.env.example
PRIVATE_KEY=0x8f2a55949038a9610f50fb23b5883af3b4ecb3c3bb792cbcefbd1542c692be63

# PostgreSQL connection used by db.NewDB()
DB_URL=postgres://admin:password@localhost:5432/goledger_db?sslmode=disable

# Optional override. Usually not needed.
CONTRACT_ARTIFACT_PATH=
```

Important notes:

1. The API first tries `app/.env` and then falls back to `../SimpleStorage/.env`.
2. If `PRIVATE_KEY` is not present in `app/.env`, `POST /set` will fail.
3. In Docker mode, `RPC_URL_DOCKER` is used by compose to avoid `localhost` networking issues inside containers.
4. In Docker mode, contract ABI is loaded from mounted path `/app/SimpleStorage/out/SimpleStorage.sol/SimpleStorage.json`.

## Run With Docker Compose (Recommended)

Inside `app/`:

```bash
docker compose up --build -d
```

This command starts:

1. `db` (PostgreSQL)
2. `api` (Go service)

API URL: `http://localhost:8080`

Swagger UI: `http://localhost:8080/swagger`

To stop all services:

```bash
docker compose down
```

## Run Locally (Without API Container)

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

## Unit Tests

Run tests from `app/`:

```bash
go test ./...
```

The test suite in `main_test.go` validates success and failure HTTP status codes for:

1. `POST /set`
2. `GET /get`
3. `POST /sync`
4. `GET /check`

## End-to-End Validation Checklist (Executed)

Validated on 2026-04-01 with API running in Docker Compose:

1. `docker compose up --build -d`: passed
2. API container healthy and listening on `:8080`: passed
3. `POST /set` returned `200`: passed
4. `GET /get` returned value written in step 3: passed
5. `POST /sync` returned `200`: passed
6. `GET /check` returned `synced: true`: passed
7. `GET /swagger/doc.json` returned `200`: passed
8. `GET /swagger` returned `200`: passed

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