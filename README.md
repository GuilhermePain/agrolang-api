# AgroResiliente API & Plataforma

Backend em Go e Frontend Next.js da plataforma AgroResiliente — inteligência e adaptação climática voltada a pequenos e médios agricultores. Ver [PRD-AgroResiliente.md](./PRD-AgroResiliente.md) para detalhes de produto.

## Stack Tecnológica

- **Backend:** Go 1.22 (Standard Library Mux, Goroutines, Pgx/v5)
- **Banco de Dados:** PostgreSQL 16 com extensão espacial PostGIS 3.4 (SRID 4326)
- **Frontend:** Next.js 15, React 19, Tailwind CSS, Leaflet / React-Leaflet
- **Serviços Externos:** Open-Meteo API (clima horário gratuito) e Evolution API (WhatsApp)
- **Infraestrutura:** Docker & Docker Compose com auto-migrations

---

## Estrutura de Diretórios

```
cmd/api            Entrypoint HTTP e ciclo de varredura
internal/app       Orquestrador de ciclos (RunCycle, auto-resolução e notificações)
internal/config    Configuração via variáveis de ambiente
internal/handler   Handlers REST HTTP (Produtores, Propriedades, Alertas, Webhooks)
internal/model     Entidades de domínio (Producer, Property, Alert, NotificationLog)
internal/notifier  Cliente Evolution API e gerador de mensagens acionáveis
internal/repository Camada de persistência PostgreSQL / PostGIS
internal/service   Motor puro de regras de risco agronômico (Risk Engine)
internal/weather   Cliente Open-Meteo com retry exponencial e resiliência
internal/worker    Varredura concorrente de talhões (Goroutines)
migrations         Scripts SQL versionados de migração do banco
web/               Frontend web completo em Next.js
```

---

## Endpoints da API REST

### Produtores
- `GET /producers`: Lista todos os produtores rurais cadastrados.
- `GET /producers/{id}`: Retorna os dados cadastrais de um produtor específico.
- `POST /producers`: Cadastra um novo produtor (`name`, `whatsapp_phone`, `city`, `state`).
- `PUT /producers/{id}`: Atualiza os dados de um produtor existente.
- `DELETE /producers/{id}`: Remove o produtor e seus talhões vinculados em cascata.
- `GET /producers/{id}/properties`: Retorna todos os talhões pertencentes ao produtor.
- `GET /producers/{id}/notifications`: Histórico de mensagens WhatsApp enviadas ao produtor.

### Propriedades (Talhões)
- `GET /properties`: Lista os talhões monitorados (suporta `?producer_id={id}`).
- `GET /properties/{id}`: Ficha cadastral do talhão.
- `POST /properties`: Registra novo talhão (`producer_id`, `latitude`, `longitude`, `crop`, `soil_type`, `crop_stage`).
- `PUT /properties/{id}`: Atualiza informações da lavoura (ex: transição de fase do cultivo ou coordenadas).
- `DELETE /properties/{id}`: Exclui a propriedade e seu histórico de alertas.
- `GET /properties/{id}/alerts`: Histórico de alertas disparados para a propriedade.
- `GET /properties/{id}/forecast`: Previsão horária de 7 dias obtida via Open-Meteo.

### Alertas e Varredura
- `GET /alerts`: Lista os alertas mais recentes do sistema.
- `POST /alerts/{id}/resolve`: Marca manualmente um alerta como resolvido.
- `POST /scan/run`: Força a execução imediata da varredura climática para todas as lavouras.

### Auditoria e Webhooks
- `POST /webhooks/evolution`: Recebe eventos de status de entrega (`DELIVERY_ACK`, `READ`) do WhatsApp.
- `GET /notifications`: Histórico geral de notificações enviadas pelo sistema.

---

## Como Executar

### 1. Com Docker Compose (Recomendado)

```bash
# Na raiz de agrolang-api-main/agrolang-api-main
cp .env.example .env

# Sobe o banco PostGIS, roda as migrações automaticamente e inicia a API
docker compose up --build
```

A API estará disponível em `http://localhost:8080`.

### 2. Frontend Next.js

```bash
cd web
echo "API_URL=http://localhost:8080" > .env.local   # opcional; este já é o padrão
npm install
npm run dev
```

Acesse o painel em `http://localhost:3000` (é preciso entrar ou criar conta em `/cadastro`).

O frontend fala com a API **apenas pelo servidor Next** (Server Components e Server Actions), enviando o token da sessão guardado no cookie httpOnly `session`. Por isso a variável é `API_URL` (só servidor); o navegador nunca chama a API diretamente e não existe `NEXT_PUBLIC_API_URL`.
