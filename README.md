# AgroResiliente API

Backend em Go da plataforma AgroResiliente — inteligência e adaptação climática para pequenos e médios agricultores. Ver [PRD-AgroResiliente.md](./PRD-AgroResiliente.md) para detalhes de produto.

## Stack

- Go 1.22
- PostgreSQL + PostGIS
- Docker / Docker Compose

## Estrutura

```
cmd/api            entrypoint HTTP
internal/config     configuração e env
internal/handler    HTTP handlers
internal/model      entidades de domínio
internal/repository acesso a dados (PostgreSQL/PostGIS)
internal/service    regras de negócio (motor de risco, etc.)
internal/worker      jobs/cron (ingestão climática)
migrations          migrations SQL
pkg                 pacotes compartilhados
```

## Setup local

```bash
cp .env.example .env
docker compose up --build
```

API sobe em `http://localhost:8080`. Health check: `GET /health`.

## Desenvolvimento sem Docker

```bash
go run ./cmd/api
```

Requer Postgres com PostGIS acessível conforme variáveis em `.env`.
