# Quiz de cadastro com login — Plano de Implementação

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Entregar um quiz mobile-first por perfil (produtor / técnico) que termina em criação de conta (e-mail + senha), isola os dados por conta e usa as respostas para personalizar o dashboard e os alertas de WhatsApp.

**Architecture:** O Go ganha `users`/`sessions`, endpoints `/auth/*` (signup transacional), middleware `RequireSession` e escopo por dono em todas as rotas. O Next guarda o token da sessão num cookie `httpOnly`, chama o Go no servidor com `Authorization: Bearer` e protege rotas com `proxy.ts`. O quiz roda no navegador (estado em `localStorage`, sem senha) e envia tudo de uma vez no último passo.

**Tech Stack:** Go 1.25 (`net/http`, pgx v5, `golang.org/x/crypto/bcrypt`), PostgreSQL + PostGIS, golang-migrate, Next 16.3 (App Router, server actions, `proxy.ts`), React 19, Tailwind 4, react-leaflet, Vitest.

**Spec:** `docs/superpowers/specs/2026-10-30-onboarding-quiz-design.md`

**Convenções do repositório:**
- Commits terminam com a linha `Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>`.
- Trabalhe na branch `feat/onboarding-quiz`.
- Módulo Go: `github.com/GuilhermePain/agrolang-api`.
- Testes de repositório usam o Postgres local (`agrolang-db-1`, porta 5432) e **truncam as tabelas**.
- O Next desta versão tem mudanças (`web/AGENTS.md`): `proxy.ts` no lugar de `middleware`, `cookies()` assíncrono. Docs em `web/node_modules/next/dist/docs/`.
- Comandos Go rodam em `/home/lukas-melo/PROJETOS/agrolang`; comandos npm em `/home/lukas-melo/PROJETOS/agrolang/web`.

---

## Mapa de arquivos

**Backend (criar)**
- `migrations/000003_create_users_and_sessions.up.sql` / `.down.sql` — tabelas `users`, `sessions`, coluna `producers.user_id`.
- `internal/auth/password.go`, `internal/auth/token.go` (+ testes) — bcrypt e tokens de sessão.
- `internal/model/user.go` — `User`, `Profile`.
- `internal/repository/user_repository.go` — usuários e sessões.
- `internal/repository/signup_repository.go` — cadastro transacional.
- `internal/handler/context.go` — usuário no `context`, middleware `RequireSession`.
- `internal/handler/signup_validation.go` — validação do corpo do signup.
- `internal/handler/auth_handler.go` — `/auth/signup|login|logout|me`.
- `internal/app/notify.go` — `ShouldNotify` (filtro por preocupações).
- Testes: `user_repository_test.go`, `signup_repository_test.go`, `scoped_test.go`, `session_middleware_test.go`, `auth_handler_test.go`, `helpers_test.go`, `notify_test.go`.

**Backend (modificar)**
- `internal/model/producer.go` — campo `Concerns`.
- `internal/repository/{producer,property,alert}_repository.go` — métodos por usuário.
- `internal/handler/{producer,property,alert}_handler.go` + testes — escopo por usuário.
- `internal/handler/json.go` — `writeFieldErrors`.
- `internal/app/cycle.go` — filtro de notificação.
- `cmd/api/main.go` — fiação.
- `internal/repository/producer_repository_test.go` — `TRUNCATE` inclui `users`, `sessions`.

**Frontend (criar)** — em `web/`
- `lib/types.ts`, `lib/config.ts`, `lib/session.ts`
- `lib/quiz/{types,options,steps,account,state,payload}.ts` (+ testes)
- `lib/geocode.ts`, `lib/personalize.ts` (+ testes)
- `app/actions/{auth,data}.ts`, `proxy.ts`, `app/sair/route.ts`
- `app/entrar/page.tsx`, `components/LoginForm.tsx`
- `app/cadastro/page.tsx`, `app/bem-vindo/page.tsx`, `components/ClearQuiz.tsx`
- `components/quiz/{styles.ts,QuizShell,ChoiceList,IdentityStep,CropStep,CitySearch,LocationStep,LocationMap,LocationMapClient,AccountStep,QuizFlow}.tsx`

**Frontend (modificar)**
- `lib/api.ts` (reescrito, só servidor), `components/ProducerForm.tsx`, `components/PropertyForm.tsx`, `app/properties/new/page.tsx`, `app/page.tsx`, `.env.local`.

---

## Parte A — Backend (Go)

### Task 1: Migration de usuários e sessões

**Files:**
- Create: `migrations/000003_create_users_and_sessions.up.sql`
- Create: `migrations/000003_create_users_and_sessions.down.sql`
- Modify: `internal/repository/producer_repository_test.go` (função `testPool`)

- [ ] **Step 1: Escrever a migration `up`**

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    profile TEXT NOT NULL CHECK (profile IN ('producer', 'technician')),
    name TEXT NOT NULL,
    whatsapp_phone TEXT NOT NULL,
    organization TEXT,
    concerns TEXT[] NOT NULL DEFAULT '{}',
    onboarding JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX users_email_idx ON users (lower(email));

CREATE TABLE sessions (
    token_hash TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX sessions_user_id_idx ON sessions (user_id);

ALTER TABLE producers ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE CASCADE;

CREATE INDEX producers_user_id_idx ON producers (user_id);
```

- [ ] **Step 2: Escrever a migration `down`**

```sql
ALTER TABLE producers DROP COLUMN user_id;

DROP TABLE sessions;

DROP TABLE users;
```

- [ ] **Step 3: Aplicar no banco local e conferir**

Run:
```bash
cd /home/lukas-melo/PROJETOS/agrolang
DB="postgres://agrolang:agrolang@localhost:5432/agrolang?sslmode=disable"
migrate -path migrations -database "$DB" version
migrate -path migrations -database "$DB" up
docker exec agrolang-db-1 psql -U agrolang -d agrolang -c '\d users' -c '\d sessions' -c '\d producers'
```
Expected: `version` mostra `2`; `up` imprime `3/u create_users_and_sessions`; `\d users` lista as colunas acima; `\d producers` mostra `user_id`.

- [ ] **Step 4: Testar o `down` e reaplicar**

Run:
```bash
migrate -path migrations -database "$DB" down 1
migrate -path migrations -database "$DB" up
```
Expected: `3/d create_users_and_sessions` e depois `3/u create_users_and_sessions`, sem erro.

- [ ] **Step 5: Incluir as novas tabelas no `TRUNCATE` dos testes**

Em `internal/repository/producer_repository_test.go`, trocar a linha do `TRUNCATE` dentro de `testPool`:

```go
	if _, err := pool.Exec(context.Background(), "TRUNCATE alerts, properties, producers, sessions, users CASCADE"); err != nil {
```

- [ ] **Step 6: Rodar os testes existentes de repositório**

Run: `go test ./internal/repository/ -v`
Expected: PASS (todos os testes atuais continuam passando).

- [ ] **Step 7: Commit**

```bash
git add migrations/000003_create_users_and_sessions.up.sql migrations/000003_create_users_and_sessions.down.sql internal/repository/producer_repository_test.go
git commit -m "feat: add users and sessions tables" -m "Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>"
```

---

### Task 2: Pacote `auth` (senha e token de sessão)

**Files:**
- Create: `internal/auth/password.go`, `internal/auth/token.go`
- Test: `internal/auth/password_test.go`, `internal/auth/token_test.go`
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1: Escrever os testes que falham**

`internal/auth/password_test.go`:
```go
package auth_test

import (
	"strings"
	"testing"

	"github.com/GuilhermePain/agrolang-api/internal/auth"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := auth.HashPassword("segredo123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "segredo123" {
		t.Fatal("hash must not equal the plain password")
	}
	if !auth.CheckPassword(hash, "segredo123") {
		t.Error("expected correct password to match")
	}
	if auth.CheckPassword(hash, "errada123") {
		t.Error("expected wrong password not to match")
	}
}

func TestHashPassword_TooLongReturnsError(t *testing.T) {
	_, err := auth.HashPassword(strings.Repeat("a", auth.MaxPasswordLength+1))
	if err == nil {
		t.Fatal("expected error for password longer than the bcrypt limit")
	}
}
```

`internal/auth/token_test.go`:
```go
package auth_test

import (
	"testing"

	"github.com/GuilhermePain/agrolang-api/internal/auth"
)

func TestNewToken_ReturnsDistinctTokensWithMatchingHash(t *testing.T) {
	token1, hash1, err := auth.NewToken()
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}
	token2, _, err := auth.NewToken()
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}

	if token1 == "" || token1 == token2 {
		t.Errorf("expected distinct non-empty tokens, got %q and %q", token1, token2)
	}
	if auth.HashToken(token1) != hash1 {
		t.Error("expected HashToken(token) to equal the returned hash")
	}
	if len(hash1) != 64 {
		t.Errorf("expected 64-char hex SHA-256, got %d chars", len(hash1))
	}
}

func TestHashToken_IsDeterministic(t *testing.T) {
	if auth.HashToken("abc") != auth.HashToken("abc") {
		t.Error("expected same input to give same hash")
	}
	if auth.HashToken("abc") == auth.HashToken("abd") {
		t.Error("expected different inputs to give different hashes")
	}
}
```

- [ ] **Step 2: Rodar e ver falhar**

Run: `go test ./internal/auth/`
Expected: FAIL (pacote sem arquivos de código / símbolos indefinidos).

- [ ] **Step 3: Implementar**

`internal/auth/password.go`:
```go
package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	MinPasswordLength = 8
	// bcrypt ignores or rejects anything past 72 bytes.
	MaxPasswordLength = 72
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("auth: hash password: %w", err)
	}
	return string(hash), nil
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
```

`internal/auth/token.go`:
```go
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// NewToken returns a random session token and its SHA-256 hex hash. Only
// the hash is stored; the token itself is handed to the client once.
func NewToken() (token, hash string, err error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", fmt.Errorf("auth: generate token: %w", err)
	}
	token = base64.RawURLEncoding.EncodeToString(raw)
	return token, HashToken(token), nil
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
```

- [ ] **Step 4: Adicionar a dependência e rodar os testes**

Run:
```bash
go get golang.org/x/crypto@latest
go mod tidy
go test ./internal/auth/ -v
```
Expected: PASS (4 testes). `go.mod` passa a listar `golang.org/x/crypto` em `require`.

- [ ] **Step 5: Commit**

```bash
git add internal/auth go.mod go.sum
git commit -m "feat: add password hashing and session token helpers" -m "Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>"
```

---

### Task 3: Modelos `User` e `Producer.Concerns`

**Files:**
- Create: `internal/model/user.go`
- Modify: `internal/model/producer.go`

- [ ] **Step 1: Criar `internal/model/user.go`**

```go
package model

import "time"

type Profile string

const (
	ProfileProducer   Profile = "producer"
	ProfileTechnician Profile = "technician"
)

type User struct {
	ID            string            `json:"id"`
	Email         string            `json:"email"`
	PasswordHash  string            `json:"-"`
	Profile       Profile           `json:"profile"`
	Name          string            `json:"name"`
	WhatsAppPhone string            `json:"whatsapp_phone"`
	Organization  *string           `json:"organization"`
	Concerns      []string          `json:"concerns"`
	Onboarding    map[string]string `json:"onboarding"`
	CreatedAt     time.Time         `json:"created_at"`
}
```

- [ ] **Step 2: Adicionar `Concerns` ao `Producer`** (substituir o arquivo `internal/model/producer.go`)

```go
package model

import "time"

type Producer struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	WhatsAppPhone string    `json:"whatsapp_phone"`
	City          string    `json:"city"`
	State         string    `json:"state"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Concerns are the risk types the producer's account owner asked to be
	// alerted about; empty means "notify everything". Filled by GetByID only
	// and never serialized.
	Concerns []string `json:"-"`
}
```

- [ ] **Step 3: Compilar**

Run: `go build ./...`
Expected: sem erros.

- [ ] **Step 4: Commit**

```bash
git add internal/model
git commit -m "feat: add user model and producer concerns field" -m "Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>"
```

---

### Task 4: `UserRepository` (usuários e sessões)

**Files:**
- Create: `internal/repository/user_repository.go`
- Test: `internal/repository/user_repository_test.go`

- [ ] **Step 1: Escrever os testes que falham**

`internal/repository/user_repository_test.go`:
```go
package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

// createTestUser inserts a bare producer-profile user and returns its ID.
func createTestUser(t *testing.T, pool *pgxpool.Pool, email string) string {
	t.Helper()

	var id string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO users (email, password_hash, profile, name, whatsapp_phone)
		 VALUES ($1, 'hash', 'producer', 'Teste', '5519999999999')
		 RETURNING id`, email).Scan(&id)
	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}
	return id
}

func TestUserRepository_GetByEmail_IsCaseInsensitive(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewUserRepository(pool)
	id := createTestUser(t, pool, "Ana@Exemplo.com")

	got, err := repo.GetByEmail(context.Background(), "ana@exemplo.com")
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if got.ID != id {
		t.Errorf("expected ID %q, got %q", id, got.ID)
	}
	if got.Concerns == nil || got.Onboarding == nil {
		t.Error("expected Concerns and Onboarding to be non-nil")
	}
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewUserRepository(pool)

	_, err := repo.GetByEmail(context.Background(), "ninguem@exemplo.com")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUserRepository_SessionLifecycle(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewUserRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, pool, "ana@exemplo.com")

	if err := repo.CreateSession(ctx, userID, "hash-1", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	got, err := repo.UserBySession(ctx, "hash-1")
	if err != nil {
		t.Fatalf("UserBySession: %v", err)
	}
	if got.ID != userID {
		t.Errorf("expected user %q, got %q", userID, got.ID)
	}

	if err := repo.DeleteSession(ctx, "hash-1"); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}
	if _, err := repo.UserBySession(ctx, "hash-1"); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestUserRepository_ExpiredSessionIsIgnoredAndPurged(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewUserRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, pool, "ana@exemplo.com")

	if err := repo.CreateSession(ctx, userID, "old", time.Now().Add(-time.Hour)); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	if _, err := repo.UserBySession(ctx, "old"); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for expired session, got %v", err)
	}

	if err := repo.DeleteExpiredSessions(ctx); err != nil {
		t.Fatalf("DeleteExpiredSessions: %v", err)
	}
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM sessions").Scan(&count); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if count != 0 {
		t.Errorf("expected expired session purged, %d remain", count)
	}
}
```

- [ ] **Step 2: Rodar e ver falhar**

Run: `go test ./internal/repository/ -run UserRepository`
Expected: FAIL (`undefined: repository.NewUserRepository`).

- [ ] **Step 3: Implementar**

`internal/repository/user_repository.go`:
```go
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/GuilhermePain/agrolang-api/internal/model"
)

// userColumns is qualified with the "u" alias for SELECTs; userReturning is
// the same list, unqualified, for INSERT ... RETURNING. Keep the order in
// sync with scanUser.
const (
	userColumns   = `u.id, u.email, u.password_hash, u.profile, u.name, u.whatsapp_phone, u.organization, u.concerns, u.onboarding, u.created_at`
	userReturning = `id, email, password_hash, profile, name, whatsapp_phone, organization, concerns, onboarding, created_at`
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func scanUser(row pgx.Row) (model.User, error) {
	var u model.User
	if err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Profile, &u.Name, &u.WhatsAppPhone,
		&u.Organization, &u.Concerns, &u.Onboarding, &u.CreatedAt,
	); err != nil {
		return model.User{}, err
	}
	if u.Concerns == nil {
		u.Concerns = []string{}
	}
	if u.Onboarding == nil {
		u.Onboarding = map[string]string{}
	}
	return u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (model.User, error) {
	const query = `SELECT ` + userColumns + ` FROM users u WHERE lower(u.email) = lower($1)`

	u, err := scanUser(r.pool.QueryRow(ctx, query, email))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("repository: get user by email: %w", err)
	}
	return u, nil
}

// UserBySession returns the owner of a live (non-expired) session.
func (r *UserRepository) UserBySession(ctx context.Context, tokenHash string) (model.User, error) {
	const query = `SELECT ` + userColumns + `
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.expires_at > now()`

	u, err := scanUser(r.pool.QueryRow(ctx, query, tokenHash))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("repository: get user by session: %w", err)
	}
	return u, nil
}

func (r *UserRepository) CreateSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	const query = `INSERT INTO sessions (token_hash, user_id, expires_at) VALUES ($1, $2, $3)`

	if _, err := r.pool.Exec(ctx, query, tokenHash, userID, expiresAt); err != nil {
		return fmt.Errorf("repository: create session: %w", err)
	}
	return nil
}

func (r *UserRepository) DeleteSession(ctx context.Context, tokenHash string) error {
	if _, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, tokenHash); err != nil {
		return fmt.Errorf("repository: delete session: %w", err)
	}
	return nil
}

func (r *UserRepository) DeleteExpiredSessions(ctx context.Context) error {
	if _, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at <= now()`); err != nil {
		return fmt.Errorf("repository: delete expired sessions: %w", err)
	}
	return nil
}
```

- [ ] **Step 4: Rodar e ver passar**

Run: `go test ./internal/repository/ -run UserRepository -v`
Expected: PASS (4 testes).

- [ ] **Step 5: Commit**

```bash
git add internal/repository/user_repository.go internal/repository/user_repository_test.go
git commit -m "feat: add user and session repository" -m "Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>"
```

---

### Task 5: Repositórios com escopo por usuário

**Files:**
- Modify: `internal/repository/producer_repository.go`, `property_repository.go`, `alert_repository.go`
- Test: `internal/repository/scoped_test.go`

- [ ] **Step 1: Escrever os testes que falham**

`internal/repository/scoped_test.go`:
```go
package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

func TestProducerRepository_ScopedByUser(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewProducerRepository(pool)
	ctx := context.Background()
	userA := createTestUser(t, pool, "a@exemplo.com")
	userB := createTestUser(t, pool, "b@exemplo.com")

	mine, err := repo.CreateForUser(ctx, userA, model.Producer{Name: "Meu", WhatsAppPhone: "5511", City: "Campinas", State: "SP"})
	if err != nil {
		t.Fatalf("CreateForUser: %v", err)
	}
	if _, err := repo.CreateForUser(ctx, userB, model.Producer{Name: "Dele", WhatsAppPhone: "5522", City: "Recife", State: "PE"}); err != nil {
		t.Fatalf("CreateForUser: %v", err)
	}

	list, err := repo.ListByUser(ctx, userA)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(list) != 1 || list[0].ID != mine.ID {
		t.Fatalf("expected only user A's producer, got %+v", list)
	}

	if _, err := repo.GetByIDForUser(ctx, mine.ID, userB); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for another user's producer, got %v", err)
	}
	if got, err := repo.GetByIDForUser(ctx, mine.ID, userA); err != nil || got.ID != mine.ID {
		t.Fatalf("expected own producer, got %+v, %v", got, err)
	}
}

func TestProducerRepository_ListByUser_EmptyIsNotNil(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewProducerRepository(pool)
	userID := createTestUser(t, pool, "a@exemplo.com")

	list, err := repo.ListByUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if list == nil || len(list) != 0 {
		t.Fatalf("expected empty non-nil slice, got %#v", list)
	}
}

func TestProducerRepository_GetByID_IncludesOwnerConcerns(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewProducerRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, pool, "a@exemplo.com")
	if _, err := pool.Exec(ctx, `UPDATE users SET concerns = '{frost,heavy_rain}' WHERE id = $1`, userID); err != nil {
		t.Fatalf("set concerns: %v", err)
	}

	owned, err := repo.CreateForUser(ctx, userID, model.Producer{Name: "Meu", WhatsAppPhone: "5511", City: "Campinas", State: "SP"})
	if err != nil {
		t.Fatalf("CreateForUser: %v", err)
	}
	legacy, err := repo.Create(ctx, model.Producer{Name: "Antigo", WhatsAppPhone: "5533", City: "Recife", State: "PE"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, owned.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if len(got.Concerns) != 2 || got.Concerns[0] != "frost" {
		t.Errorf("expected owner concerns [frost heavy_rain], got %v", got.Concerns)
	}

	legacyGot, err := repo.GetByID(ctx, legacy.ID)
	if err != nil {
		t.Fatalf("GetByID legacy: %v", err)
	}
	if len(legacyGot.Concerns) != 0 {
		t.Errorf("expected no concerns for a producer without owner, got %v", legacyGot.Concerns)
	}
}

func TestPropertyRepository_ScopedByUser(t *testing.T) {
	pool := testPool(t)
	producers := repository.NewProducerRepository(pool)
	properties := repository.NewPropertyRepository(pool)
	ctx := context.Background()
	userA := createTestUser(t, pool, "a@exemplo.com")
	userB := createTestUser(t, pool, "b@exemplo.com")

	producerA, err := producers.CreateForUser(ctx, userA, model.Producer{Name: "A", WhatsAppPhone: "5511", City: "Campinas", State: "SP"})
	if err != nil {
		t.Fatalf("CreateForUser: %v", err)
	}

	newProperty := model.Property{
		ProducerID: producerA.ID, Latitude: -22.9, Longitude: -47.06,
		Crop: "Milho", SoilType: "argiloso", CropStage: model.CropStageFlowering,
	}

	if _, err := properties.CreateForUser(ctx, userB, newProperty); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound creating under another user's producer, got %v", err)
	}

	created, err := properties.CreateForUser(ctx, userA, newProperty)
	if err != nil {
		t.Fatalf("CreateForUser: %v", err)
	}
	if created.Latitude != -22.9 || created.Longitude != -47.06 {
		t.Errorf("expected coordinates round-trip, got %v,%v", created.Latitude, created.Longitude)
	}

	list, err := properties.ListByUser(ctx, userA)
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 property for user A, got %v, %v", list, err)
	}
	other, err := properties.ListByUser(ctx, userB)
	if err != nil || other == nil || len(other) != 0 {
		t.Fatalf("expected empty non-nil list for user B, got %#v, %v", other, err)
	}

	if _, err := properties.GetByIDForUser(ctx, created.ID, userB); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for another user's property, got %v", err)
	}
	if got, err := properties.GetByIDForUser(ctx, created.ID, userA); err != nil || got.ID != created.ID {
		t.Fatalf("expected own property, got %+v, %v", got, err)
	}
}

func TestAlertRepository_ListRecentForUser_OnlyOwnAlerts(t *testing.T) {
	pool := testPool(t)
	producers := repository.NewProducerRepository(pool)
	properties := repository.NewPropertyRepository(pool)
	alerts := repository.NewAlertRepository(pool)
	ctx := context.Background()
	userA := createTestUser(t, pool, "a@exemplo.com")
	userB := createTestUser(t, pool, "b@exemplo.com")

	makeAlert := func(userID, email string) {
		t.Helper()
		producer, err := producers.CreateForUser(ctx, userID, model.Producer{Name: email, WhatsAppPhone: "5511", City: "X", State: "SP"})
		if err != nil {
			t.Fatalf("CreateForUser producer: %v", err)
		}
		property, err := properties.CreateForUser(ctx, userID, model.Property{
			ProducerID: producer.ID, Latitude: -22.9, Longitude: -47.06,
			Crop: "Milho", SoilType: "argiloso", CropStage: model.CropStageFlowering,
		})
		if err != nil {
			t.Fatalf("CreateForUser property: %v", err)
		}
		now := time.Now()
		if _, err := alerts.Create(ctx, model.Alert{
			PropertyID: property.ID, Level: "critical", AlertType: "frost",
			PeriodStart: now, PeriodEnd: now.Add(time.Hour),
		}); err != nil {
			t.Fatalf("Create alert: %v", err)
		}
	}
	makeAlert(userA, "a")
	makeAlert(userB, "b")

	got, err := alerts.ListRecentForUser(ctx, userA, 50)
	if err != nil {
		t.Fatalf("ListRecentForUser: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 alert for user A, got %d", len(got))
	}
}
```

- [ ] **Step 2: Rodar e ver falhar**

Run: `go test ./internal/repository/ -run "ScopedByUser|ListByUser|IncludesOwnerConcerns|ListRecentForUser"`
Expected: FAIL (métodos indefinidos).

- [ ] **Step 3: `ProducerRepository` — adicionar métodos e substituir `GetByID`**

Em `internal/repository/producer_repository.go`, **substituir** a função `GetByID` inteira por esta (agora traz as preocupações do dono) e **acrescentar** as três funções novas ao final do arquivo:

```go
func (r *ProducerRepository) GetByID(ctx context.Context, id string) (model.Producer, error) {
	const query = `
		SELECT p.id, p.name, p.whatsapp_phone, p.city, p.state, p.created_at, p.updated_at,
			COALESCE(u.concerns, '{}')
		FROM producers p
		LEFT JOIN users u ON u.id = p.user_id
		WHERE p.id = $1`

	var p model.Producer
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.WhatsAppPhone, &p.City, &p.State, &p.CreatedAt, &p.UpdatedAt,
		&p.Concerns,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Producer{}, ErrNotFound
	}
	if err != nil {
		return model.Producer{}, fmt.Errorf("repository: get producer: %w", err)
	}
	return p, nil
}

func (r *ProducerRepository) CreateForUser(ctx context.Context, userID string, p model.Producer) (model.Producer, error) {
	const query = `
		INSERT INTO producers (user_id, name, whatsapp_phone, city, state)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, whatsapp_phone, city, state, created_at, updated_at`

	var created model.Producer
	err := r.pool.QueryRow(ctx, query, userID, p.Name, p.WhatsAppPhone, p.City, p.State).Scan(
		&created.ID, &created.Name, &created.WhatsAppPhone, &created.City, &created.State,
		&created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return model.Producer{}, fmt.Errorf("repository: create producer for user: %w", err)
	}
	return created, nil
}

func (r *ProducerRepository) ListByUser(ctx context.Context, userID string) ([]model.Producer, error) {
	const query = `
		SELECT id, name, whatsapp_phone, city, state, created_at, updated_at
		FROM producers
		WHERE user_id = $1
		ORDER BY created_at`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: list producers by user: %w", err)
	}
	defer rows.Close()

	all := make([]model.Producer, 0)
	for rows.Next() {
		var p model.Producer
		if err := rows.Scan(
			&p.ID, &p.Name, &p.WhatsAppPhone, &p.City, &p.State, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan producer: %w", err)
		}
		all = append(all, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate producers: %w", err)
	}
	return all, nil
}

func (r *ProducerRepository) GetByIDForUser(ctx context.Context, id, userID string) (model.Producer, error) {
	const query = `
		SELECT id, name, whatsapp_phone, city, state, created_at, updated_at
		FROM producers
		WHERE id = $1 AND user_id = $2`

	var p model.Producer
	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&p.ID, &p.Name, &p.WhatsAppPhone, &p.City, &p.State, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Producer{}, ErrNotFound
	}
	if err != nil {
		return model.Producer{}, fmt.Errorf("repository: get producer for user: %w", err)
	}
	return p, nil
}
```

- [ ] **Step 4: `PropertyRepository` — acrescentar ao final de `property_repository.go`**

```go
// CreateForUser inserts the property only if the producer belongs to the
// user; otherwise it returns ErrNotFound.
func (r *PropertyRepository) CreateForUser(ctx context.Context, userID string, p model.Property) (model.Property, error) {
	const query = `
		INSERT INTO properties (producer_id, location, crop, soil_type, crop_stage)
		SELECT pr.id, ST_SetSRID(ST_MakePoint($3, $4), 4326), $5, $6, $7
		FROM producers pr
		WHERE pr.id = $1 AND pr.user_id = $2
		RETURNING id, producer_id, ST_X(location::geometry), ST_Y(location::geometry),
			crop, soil_type, crop_stage, created_at, updated_at`

	var created model.Property
	err := r.pool.QueryRow(ctx, query, p.ProducerID, userID, p.Longitude, p.Latitude, p.Crop, p.SoilType, p.CropStage).Scan(
		&created.ID, &created.ProducerID, &created.Longitude, &created.Latitude,
		&created.Crop, &created.SoilType, &created.CropStage, &created.CreatedAt, &created.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Property{}, ErrNotFound
	}
	if err != nil {
		return model.Property{}, fmt.Errorf("repository: create property for user: %w", err)
	}
	return created, nil
}

func (r *PropertyRepository) ListByUser(ctx context.Context, userID string) ([]model.Property, error) {
	const query = `
		SELECT p.id, p.producer_id, ST_X(p.location::geometry), ST_Y(p.location::geometry),
			p.crop, p.soil_type, p.crop_stage, p.created_at, p.updated_at
		FROM properties p
		JOIN producers pr ON pr.id = p.producer_id
		WHERE pr.user_id = $1
		ORDER BY p.created_at`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: list properties by user: %w", err)
	}
	defer rows.Close()

	all := make([]model.Property, 0)
	for rows.Next() {
		var p model.Property
		if err := rows.Scan(
			&p.ID, &p.ProducerID, &p.Longitude, &p.Latitude,
			&p.Crop, &p.SoilType, &p.CropStage, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan property: %w", err)
		}
		all = append(all, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate properties: %w", err)
	}
	return all, nil
}

func (r *PropertyRepository) GetByIDForUser(ctx context.Context, id, userID string) (model.Property, error) {
	const query = `
		SELECT p.id, p.producer_id, ST_X(p.location::geometry), ST_Y(p.location::geometry),
			p.crop, p.soil_type, p.crop_stage, p.created_at, p.updated_at
		FROM properties p
		JOIN producers pr ON pr.id = p.producer_id
		WHERE p.id = $1 AND pr.user_id = $2`

	var p model.Property
	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&p.ID, &p.ProducerID, &p.Longitude, &p.Latitude,
		&p.Crop, &p.SoilType, &p.CropStage, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Property{}, ErrNotFound
	}
	if err != nil {
		return model.Property{}, fmt.Errorf("repository: get property for user: %w", err)
	}
	return p, nil
}
```

- [ ] **Step 5: `AlertRepository` — acrescentar ao final de `alert_repository.go`**

```go
func (r *AlertRepository) ListRecentForUser(ctx context.Context, userID string, limit int) ([]model.Alert, error) {
	const query = `
		SELECT a.id, a.property_id, a.level, a.alert_type, a.period_start, a.period_end, a.triggered_at, a.resolved_at
		FROM alerts a
		JOIN properties p ON p.id = a.property_id
		JOIN producers pr ON pr.id = p.producer_id
		WHERE pr.user_id = $1
		ORDER BY a.triggered_at DESC
		LIMIT $2`

	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("repository: query alerts for user: %w", err)
	}
	defer rows.Close()

	all := make([]model.Alert, 0)
	for rows.Next() {
		var a model.Alert
		if err := rows.Scan(
			&a.ID, &a.PropertyID, &a.Level, &a.AlertType,
			&a.PeriodStart, &a.PeriodEnd, &a.TriggeredAt, &a.ResolvedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan alert: %w", err)
		}
		all = append(all, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate alerts: %w", err)
	}
	return all, nil
}
```

- [ ] **Step 6: Rodar todos os testes do pacote**

Run: `go test ./internal/repository/ -v`
Expected: PASS (testes novos e antigos).

- [ ] **Step 7: Commit**

```bash
git add internal/repository
git commit -m "feat: scope producer, property and alert queries by owner" -m "Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>"
```

---

### Task 6: `SignupRepository` (cadastro transacional)

**Files:**
- Create: `internal/repository/signup_repository.go`
- Test: `internal/repository/signup_repository_test.go`

- [ ] **Step 1: Escrever os testes que falham**

`internal/repository/signup_repository_test.go`:
```go
package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

func producerSignup(email string) repository.SignupInput {
	return repository.SignupInput{
		User: model.User{
			Email: email, PasswordHash: "hash", Profile: model.ProfileProducer,
			Name: "José", WhatsAppPhone: "5519999999999",
			Concerns:   []string{"frost", "severe_water_stress"},
			Onboarding: map[string]string{"work_style": "family", "areas_bucket": "1"},
		},
		Producer: &model.Producer{City: "Campinas", State: "SP"},
		Property: &model.Property{
			Latitude: -22.9099, Longitude: -47.0626,
			Crop: "Milho", SoilType: "argiloso", CropStage: model.CropStageFlowering,
		},
		SessionTokenHash: "token-hash",
		SessionExpiresAt: time.Now().Add(time.Hour),
	}
}

func TestSignupRepository_Producer_CreatesUserProducerPropertyAndSession(t *testing.T) {
	pool := testPool(t)
	signups := repository.NewSignupRepository(pool)
	users := repository.NewUserRepository(pool)
	ctx := context.Background()

	result, err := signups.Signup(ctx, producerSignup("Jose@Exemplo.com"))
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}

	if result.User.ID == "" || result.User.Email != "jose@exemplo.com" {
		t.Errorf("expected stored lowercase email, got %+v", result.User)
	}
	if len(result.User.Concerns) != 2 || result.User.Onboarding["areas_bucket"] != "1" {
		t.Errorf("expected concerns and onboarding saved, got %+v", result.User)
	}
	if result.Producer == nil || result.Producer.Name != "José" || result.Producer.WhatsAppPhone != "5519999999999" {
		t.Errorf("expected producer copied from the user, got %+v", result.Producer)
	}
	if result.Property == nil || result.Property.ProducerID != result.Producer.ID {
		t.Errorf("expected property linked to producer, got %+v", result.Property)
	}

	owned, err := repository.NewProducerRepository(pool).ListByUser(ctx, result.User.ID)
	if err != nil || len(owned) != 1 {
		t.Fatalf("expected the producer owned by the new user, got %v, %v", owned, err)
	}
	if _, err := users.UserBySession(ctx, "token-hash"); err != nil {
		t.Errorf("expected live session, got %v", err)
	}
}

func TestSignupRepository_Technician_CreatesOnlyUserAndSession(t *testing.T) {
	pool := testPool(t)
	signups := repository.NewSignupRepository(pool)
	ctx := context.Background()
	org := "Coopera Campo"

	result, err := signups.Signup(ctx, repository.SignupInput{
		User: model.User{
			Email: "ana@coop.com", PasswordHash: "hash", Profile: model.ProfileTechnician,
			Name: "Ana", WhatsAppPhone: "5519988887777", Organization: &org,
		},
		SessionTokenHash: "tech-token",
		SessionExpiresAt: time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}
	if result.Producer != nil || result.Property != nil {
		t.Errorf("expected no producer or property for a technician, got %+v / %+v", result.Producer, result.Property)
	}
	if result.User.Organization == nil || *result.User.Organization != org {
		t.Errorf("expected organization saved, got %v", result.User.Organization)
	}

	var producers int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM producers").Scan(&producers); err != nil {
		t.Fatalf("count producers: %v", err)
	}
	if producers != 0 {
		t.Errorf("expected 0 producers, got %d", producers)
	}
}

func TestSignupRepository_DuplicateEmail_ReturnsErrEmailTaken(t *testing.T) {
	pool := testPool(t)
	signups := repository.NewSignupRepository(pool)
	ctx := context.Background()

	if _, err := signups.Signup(ctx, producerSignup("jose@exemplo.com")); err != nil {
		t.Fatalf("first Signup: %v", err)
	}

	second := producerSignup("JOSE@exemplo.com")
	second.SessionTokenHash = "other-token"
	if _, err := signups.Signup(ctx, second); !errors.Is(err, repository.ErrEmailTaken) {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestSignupRepository_FailureRollsBackEverything(t *testing.T) {
	pool := testPool(t)
	signups := repository.NewSignupRepository(pool)
	ctx := context.Background()

	input := producerSignup("rollback@exemplo.com")
	input.Property.CropStage = "not-a-stage" // violates the crop_stage enum

	if _, err := signups.Signup(ctx, input); err == nil {
		t.Fatal("expected an error for an invalid crop stage")
	}

	var users, producers int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM users").Scan(&users); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM producers").Scan(&producers); err != nil {
		t.Fatalf("count producers: %v", err)
	}
	if users != 0 || producers != 0 {
		t.Errorf("expected rollback, found %d users and %d producers", users, producers)
	}
}
```

- [ ] **Step 2: Rodar e ver falhar**

Run: `go test ./internal/repository/ -run SignupRepository`
Expected: FAIL (`undefined: repository.NewSignupRepository`).

- [ ] **Step 3: Implementar**

`internal/repository/signup_repository.go`:
```go
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/GuilhermePain/agrolang-api/internal/model"
)

var ErrEmailTaken = errors.New("repository: email already registered")

// SignupInput carries everything created at account signup. Producer and
// Property are set for the producer profile and nil for technicians. The
// producer's name and phone are copied from User.
type SignupInput struct {
	User             model.User
	Producer         *model.Producer
	Property         *model.Property
	SessionTokenHash string
	SessionExpiresAt time.Time
}

type SignupResult struct {
	User     model.User
	Producer *model.Producer
	Property *model.Property
}

type SignupRepository struct {
	pool *pgxpool.Pool
}

func NewSignupRepository(pool *pgxpool.Pool) *SignupRepository {
	return &SignupRepository{pool: pool}
}

// Signup creates the user, optional producer and property, and the first
// session in one transaction: either all of it exists or none of it.
func (r *SignupRepository) Signup(ctx context.Context, in SignupInput) (SignupResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return SignupResult{}, fmt.Errorf("repository: begin signup: %w", err)
	}
	defer tx.Rollback(ctx)

	concerns := in.User.Concerns
	if concerns == nil {
		concerns = []string{}
	}
	onboarding := in.User.Onboarding
	if onboarding == nil {
		onboarding = map[string]string{}
	}

	const insertUser = `
		INSERT INTO users (email, password_hash, profile, name, whatsapp_phone, organization, concerns, onboarding)
		VALUES (lower($1), $2, $3, $4, $5, $6, $7, $8)
		RETURNING ` + userReturning

	user, err := scanUser(tx.QueryRow(ctx, insertUser,
		in.User.Email, in.User.PasswordHash, in.User.Profile, in.User.Name, in.User.WhatsAppPhone,
		in.User.Organization, concerns, onboarding,
	))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return SignupResult{}, ErrEmailTaken
		}
		return SignupResult{}, fmt.Errorf("repository: signup insert user: %w", err)
	}

	result := SignupResult{User: user}

	if in.Producer != nil {
		const insertProducer = `
			INSERT INTO producers (user_id, name, whatsapp_phone, city, state)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, name, whatsapp_phone, city, state, created_at, updated_at`

		var producer model.Producer
		if err := tx.QueryRow(ctx, insertProducer, user.ID, user.Name, user.WhatsAppPhone, in.Producer.City, in.Producer.State).Scan(
			&producer.ID, &producer.Name, &producer.WhatsAppPhone, &producer.City, &producer.State,
			&producer.CreatedAt, &producer.UpdatedAt,
		); err != nil {
			return SignupResult{}, fmt.Errorf("repository: signup insert producer: %w", err)
		}
		result.Producer = &producer

		if in.Property != nil {
			const insertProperty = `
				INSERT INTO properties (producer_id, location, crop, soil_type, crop_stage)
				VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326), $4, $5, $6)
				RETURNING id, producer_id, ST_X(location::geometry), ST_Y(location::geometry),
					crop, soil_type, crop_stage, created_at, updated_at`

			var property model.Property
			if err := tx.QueryRow(ctx, insertProperty,
				producer.ID, in.Property.Longitude, in.Property.Latitude,
				in.Property.Crop, in.Property.SoilType, in.Property.CropStage,
			).Scan(
				&property.ID, &property.ProducerID, &property.Longitude, &property.Latitude,
				&property.Crop, &property.SoilType, &property.CropStage, &property.CreatedAt, &property.UpdatedAt,
			); err != nil {
				return SignupResult{}, fmt.Errorf("repository: signup insert property: %w", err)
			}
			result.Property = &property
		}
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO sessions (token_hash, user_id, expires_at) VALUES ($1, $2, $3)`,
		in.SessionTokenHash, user.ID, in.SessionExpiresAt,
	); err != nil {
		return SignupResult{}, fmt.Errorf("repository: signup insert session: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return SignupResult{}, fmt.Errorf("repository: commit signup: %w", err)
	}
	return result, nil
}
```

- [ ] **Step 4: Rodar e ver passar**

Run: `go test ./internal/repository/ -v`
Expected: PASS (todos).

- [ ] **Step 5: Commit**

```bash
git add internal/repository/signup_repository.go internal/repository/signup_repository_test.go
git commit -m "feat: add transactional signup repository" -m "Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>"
```

---

### Task 7: Sessão no handler, middleware e `AuthHandler`

**Files:**
- Create: `internal/handler/context.go`, `internal/handler/signup_validation.go`, `internal/handler/auth_handler.go`
- Modify: `internal/handler/json.go`
- Test: `internal/handler/session_middleware_test.go`, `internal/handler/auth_handler_test.go`

- [ ] **Step 1: Acrescentar `writeFieldErrors` em `internal/handler/json.go`** (ao final do arquivo)

```go
func writeFieldErrors(w http.ResponseWriter, status int, message string, fields map[string]string) {
	writeJSON(w, status, map[string]any{"error": message, "fields": fields})
}
```

- [ ] **Step 2: Escrever o teste do middleware (falha)**

`internal/handler/session_middleware_test.go`:
```go
package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GuilhermePain/agrolang-api/internal/auth"
	"github.com/GuilhermePain/agrolang-api/internal/handler"
	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

type fakeSessionLookup struct {
	byHash map[string]model.User
}

func (f *fakeSessionLookup) UserBySession(ctx context.Context, tokenHash string) (model.User, error) {
	u, ok := f.byHash[tokenHash]
	if !ok {
		return model.User{}, repository.ErrNotFound
	}
	return u, nil
}

func serveThroughSession(lookup handler.SessionLookup, req *http.Request) (*httptest.ResponseRecorder, *model.User) {
	var seen *model.User
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, ok := handler.UserFromContext(r.Context()); ok {
			seen = &u
		}
		w.WriteHeader(http.StatusNoContent)
	})
	rec := httptest.NewRecorder()
	handler.RequireSession(lookup)(next).ServeHTTP(rec, req)
	return rec, seen
}

func TestRequireSession_ValidTokenPutsUserInContext(t *testing.T) {
	lookup := &fakeSessionLookup{byHash: map[string]model.User{
		auth.HashToken("good-token"): {ID: "user-1"},
	}}
	req := httptest.NewRequest(http.MethodGet, "/anything", nil)
	req.Header.Set("Authorization", "Bearer good-token")

	rec, seen := serveThroughSession(lookup, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected next handler to run, got %d", rec.Code)
	}
	if seen == nil || seen.ID != "user-1" {
		t.Fatalf("expected user-1 in context, got %+v", seen)
	}
}

func TestRequireSession_RejectsMissingOrUnknownToken(t *testing.T) {
	lookup := &fakeSessionLookup{byHash: map[string]model.User{}}

	cases := map[string]string{
		"no header":      "",
		"wrong scheme":   "Basic abc",
		"empty bearer":   "Bearer ",
		"unknown bearer": "Bearer nope",
	}
	for name, header := range cases {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/anything", nil)
			if header != "" {
				req.Header.Set("Authorization", header)
			}
			rec, _ := serveThroughSession(lookup, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", rec.Code)
			}
		})
	}
}
```

- [ ] **Step 3: Rodar e ver falhar**

Run: `go test ./internal/handler/ -run RequireSession`
Expected: FAIL (`undefined: handler.RequireSession`).

- [ ] **Step 4: Implementar `internal/handler/context.go`**

```go
package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/GuilhermePain/agrolang-api/internal/auth"
	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

type ctxKey int

const (
	userCtxKey ctxKey = iota
	tokenHashCtxKey
)

// ContextWithUser is exported so tests can authenticate a request without a
// session store.
func ContextWithUser(ctx context.Context, u model.User) context.Context {
	return context.WithValue(ctx, userCtxKey, u)
}

func UserFromContext(ctx context.Context) (model.User, bool) {
	u, ok := ctx.Value(userCtxKey).(model.User)
	return u, ok
}

func contextWithTokenHash(ctx context.Context, hash string) context.Context {
	return context.WithValue(ctx, tokenHashCtxKey, hash)
}

func tokenHashFromContext(ctx context.Context) (string, bool) {
	h, ok := ctx.Value(tokenHashCtxKey).(string)
	return h, ok
}

// requireUser reads the authenticated user set by RequireSession and writes
// a 401 when it is missing, so handlers never run unauthenticated.
func requireUser(w http.ResponseWriter, r *http.Request) (model.User, bool) {
	u, ok := UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return model.User{}, false
	}
	return u, true
}

type SessionLookup interface {
	UserBySession(ctx context.Context, tokenHash string) (model.User, error)
}

// RequireSession authenticates "Authorization: Bearer <token>" against live
// sessions and stores the user (and the token hash, for logout) in the
// request context.
func RequireSession(lookup SessionLookup) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r)
			if !ok {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			hash := auth.HashToken(token)
			user, err := lookup.UserBySession(r.Context(), hash)
			if errors.Is(err, repository.ErrNotFound) {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to check session")
				return
			}

			ctx := contextWithTokenHash(ContextWithUser(r.Context(), user), hash)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(r *http.Request) (string, bool) {
	const prefix = "Bearer "
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	return token, token != ""
}
```

- [ ] **Step 5: Rodar o teste do middleware**

Run: `go test ./internal/handler/ -run RequireSession -v`
Expected: PASS.

- [ ] **Step 6: Escrever os testes do `AuthHandler` (falham)**

`internal/handler/auth_handler_test.go`:
```go
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/auth"
	"github.com/GuilhermePain/agrolang-api/internal/handler"
	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

type fakeSignuper struct {
	got    repository.SignupInput
	result repository.SignupResult
	err    error
}

func (f *fakeSignuper) Signup(ctx context.Context, in repository.SignupInput) (repository.SignupResult, error) {
	f.got = in
	if f.err != nil {
		return repository.SignupResult{}, f.err
	}
	user := in.User
	user.ID = "user-1"
	return repository.SignupResult{User: user, Producer: in.Producer, Property: in.Property}, nil
}

type fakeAuthStore struct {
	users          map[string]model.User
	createdSession string
	deletedSession string
	purged         bool
}

func (f *fakeAuthStore) GetByEmail(ctx context.Context, email string) (model.User, error) {
	u, ok := f.users[email]
	if !ok {
		return model.User{}, repository.ErrNotFound
	}
	return u, nil
}

func (f *fakeAuthStore) CreateSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	f.createdSession = tokenHash
	return nil
}

func (f *fakeAuthStore) DeleteSession(ctx context.Context, tokenHash string) error {
	f.deletedSession = tokenHash
	return nil
}

func (f *fakeAuthStore) DeleteExpiredSessions(ctx context.Context) error {
	f.purged = true
	return nil
}

const validProducerSignup = `{
	"profile":"producer","name":"José da Silva","whatsapp_phone":"5519999999999",
	"email":"Jose@Exemplo.com","password":"segredo123","organization":null,
	"concerns":["frost","severe_water_stress"],
	"onboarding":{"work_style":"family","areas_bucket":"1"},
	"producer":{"city":"Campinas","state":"sp"},
	"property":{"latitude":-22.9099,"longitude":-47.0626,"crop":"Milho","soil_type":"argiloso","crop_stage":"flowering"}
}`

const validTechnicianSignup = `{
	"profile":"technician","name":"Ana","whatsapp_phone":"5519988887777",
	"email":"ana@coop.com","password":"segredo123","organization":"Coopera Campo",
	"concerns":[],"onboarding":{"work_style":"technician","producers_bucket":"11-40"}
}`

func newAuthMux(signuper *fakeSignuper, store *fakeAuthStore) *http.ServeMux {
	h := handler.NewAuthHandler(signuper, store)
	mux := http.NewServeMux()
	h.RegisterPublic(mux)
	h.RegisterProtected(mux)
	return mux
}

func postJSON(mux http.Handler, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestAuthHandler_Signup_Producer_Returns201WithSession(t *testing.T) {
	signuper := &fakeSignuper{}
	mux := newAuthMux(signuper, &fakeAuthStore{})

	rec := postJSON(mux, "/auth/signup", validProducerSignup)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Token     string     `json:"token"`
		ExpiresAt time.Time  `json:"expires_at"`
		User      model.User `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Token == "" || resp.User.ID != "user-1" {
		t.Errorf("expected token and user, got %+v", resp)
	}
	if !resp.ExpiresAt.After(time.Now().Add(29 * 24 * time.Hour)) {
		t.Errorf("expected ~30 day expiry, got %v", resp.ExpiresAt)
	}
	if strings.Contains(rec.Body.String(), "password") {
		t.Errorf("response must not expose password data: %s", rec.Body.String())
	}

	in := signuper.got
	if in.User.Email != "jose@exemplo.com" {
		t.Errorf("expected normalized email, got %q", in.User.Email)
	}
	if in.User.PasswordHash == "" || in.User.PasswordHash == "segredo123" || !auth.CheckPassword(in.User.PasswordHash, "segredo123") {
		t.Error("expected a bcrypt hash of the password")
	}
	if in.SessionTokenHash != auth.HashToken(resp.Token) {
		t.Error("expected stored session hash to match the returned token")
	}
	if in.Producer == nil || in.Producer.State != "SP" {
		t.Errorf("expected producer with upper-cased state, got %+v", in.Producer)
	}
	if in.Property == nil || in.Property.CropStage != model.CropStageFlowering {
		t.Errorf("expected property with flowering stage, got %+v", in.Property)
	}
}

func TestAuthHandler_Signup_Technician_HasNoProducerOrProperty(t *testing.T) {
	signuper := &fakeSignuper{}
	mux := newAuthMux(signuper, &fakeAuthStore{})

	rec := postJSON(mux, "/auth/signup", validTechnicianSignup)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if signuper.got.Producer != nil || signuper.got.Property != nil {
		t.Errorf("expected no producer/property, got %+v / %+v", signuper.got.Producer, signuper.got.Property)
	}
	if signuper.got.User.Organization == nil || *signuper.got.User.Organization != "Coopera Campo" {
		t.Errorf("expected organization saved, got %v", signuper.got.User.Organization)
	}
}

func TestAuthHandler_Signup_InvalidBody_Returns400WithFieldMessages(t *testing.T) {
	mux := newAuthMux(&fakeSignuper{}, &fakeAuthStore{})

	body := `{"profile":"producer","name":"J","whatsapp_phone":"123","email":"nope","password":"curta",
		"concerns":["tornado"],"onboarding":{"work_style":"alien"}}`
	rec := postJSON(mux, "/auth/signup", body)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Fields map[string]string `json:"fields"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, key := range []string{"name", "whatsapp_phone", "email", "password", "concerns", "onboarding", "producer", "property"} {
		if resp.Fields[key] == "" {
			t.Errorf("expected a message for field %q, got fields %v", key, resp.Fields)
		}
	}
}

func TestAuthHandler_Signup_TechnicianWithProducer_Returns400(t *testing.T) {
	mux := newAuthMux(&fakeSignuper{}, &fakeAuthStore{})

	body := strings.Replace(validTechnicianSignup, `"concerns":[]`, `"concerns":[],"producer":{"city":"X","state":"SP"}`, 1)
	rec := postJSON(mux, "/auth/signup", body)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"producer"`) {
		t.Errorf("expected a producer field error, got %s", rec.Body.String())
	}
}

func TestAuthHandler_Signup_EmailTaken_Returns409(t *testing.T) {
	mux := newAuthMux(&fakeSignuper{err: repository.ErrEmailTaken}, &fakeAuthStore{})

	rec := postJSON(mux, "/auth/signup", validProducerSignup)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"email"`) {
		t.Errorf("expected an email field error, got %s", rec.Body.String())
	}
}

func TestAuthHandler_Login(t *testing.T) {
	hash, err := auth.HashPassword("segredo123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	store := &fakeAuthStore{users: map[string]model.User{
		"jose@exemplo.com": {ID: "user-1", Email: "jose@exemplo.com", PasswordHash: hash},
	}}
	mux := newAuthMux(&fakeSignuper{}, store)

	t.Run("correct credentials return a session", func(t *testing.T) {
		rec := postJSON(mux, "/auth/login", `{"email":" Jose@Exemplo.com ","password":"segredo123"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var resp struct {
			Token string     `json:"token"`
			User  model.User `json:"user"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.Token == "" || resp.User.ID != "user-1" {
			t.Errorf("expected token and user, got %+v", resp)
		}
		if store.createdSession != auth.HashToken(resp.Token) {
			t.Error("expected the session hash to be stored")
		}
		if !store.purged {
			t.Error("expected expired sessions to be purged on login")
		}
	})

	t.Run("wrong password and unknown email give the same 401", func(t *testing.T) {
		wrong := postJSON(mux, "/auth/login", `{"email":"jose@exemplo.com","password":"errada123"}`)
		unknown := postJSON(mux, "/auth/login", `{"email":"ninguem@exemplo.com","password":"segredo123"}`)
		if wrong.Code != http.StatusUnauthorized || unknown.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for both, got %d and %d", wrong.Code, unknown.Code)
		}
		if wrong.Body.String() != unknown.Body.String() {
			t.Errorf("expected identical bodies, got %q and %q", wrong.Body.String(), unknown.Body.String())
		}
	})
}

func TestAuthHandler_Me_ReturnsUserFromContext(t *testing.T) {
	mux := newAuthMux(&fakeSignuper{}, &fakeAuthStore{})

	req := authed(httptest.NewRequest(http.MethodGet, "/auth/me", nil))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got model.User
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ID != testUserID {
		t.Errorf("expected %q, got %q", testUserID, got.ID)
	}
}

func TestAuthHandler_Me_WithoutUser_Returns401(t *testing.T) {
	mux := newAuthMux(&fakeSignuper{}, &fakeAuthStore{})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/auth/me", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthHandler_Logout_DeletesCurrentSession(t *testing.T) {
	store := &fakeAuthStore{}
	h := handler.NewAuthHandler(&fakeSignuper{}, store)
	mux := http.NewServeMux()
	h.RegisterProtected(mux)

	lookup := &fakeSessionLookup{byHash: map[string]model.User{
		auth.HashToken("tok"): {ID: testUserID},
	}}
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer tok")
	rec := httptest.NewRecorder()
	handler.RequireSession(lookup)(mux).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
	if store.deletedSession != auth.HashToken("tok") {
		t.Errorf("expected session %q deleted, got %q", auth.HashToken("tok"), store.deletedSession)
	}
}
```

Criar também o helper compartilhado `internal/handler/helpers_test.go`:
```go
package handler_test

import (
	"net/http"

	"github.com/GuilhermePain/agrolang-api/internal/handler"
	"github.com/GuilhermePain/agrolang-api/internal/model"
)

const testUserID = "user-1"

// authed marks the request as made by the test user, as RequireSession
// would in production.
func authed(req *http.Request) *http.Request {
	return req.WithContext(handler.ContextWithUser(req.Context(), model.User{ID: testUserID}))
}
```

- [ ] **Step 7: Rodar e ver falhar**

Run: `go test ./internal/handler/ -run AuthHandler`
Expected: FAIL (`undefined: handler.NewAuthHandler`).

- [ ] **Step 8: Implementar `internal/handler/signup_validation.go`**

```go
package handler

import (
	"net/mail"
	"regexp"
	"strings"

	"github.com/GuilhermePain/agrolang-api/internal/auth"
	"github.com/GuilhermePain/agrolang-api/internal/model"
)

var (
	phonePattern = regexp.MustCompile(`^55\d{10,11}$`)
	statePattern = regexp.MustCompile(`^[A-Z]{2}$`)

	validConcerns = map[string]bool{
		"frost": true, "heat_wave": true, "heavy_rain": true, "severe_water_stress": true,
	}
	validSoils  = map[string]bool{"argiloso": true, "arenoso": true, "misto": true, "nao_informado": true}
	validStages = map[model.CropStage]bool{
		model.CropStageGermination: true, model.CropStageFlowering: true, model.CropStageHarvest: true,
	}
	onboardingValues = map[string]map[string]bool{
		"work_style":       {"family": true, "multi_area": true, "technician": true},
		"areas_bucket":     {"1": true, "2-5": true, "6+": true},
		"producers_bucket": {"ate-10": true, "11-40": true, "mais-40": true},
	}
)

type signupRequest struct {
	Profile       model.Profile     `json:"profile"`
	Name          string            `json:"name"`
	WhatsAppPhone string            `json:"whatsapp_phone"`
	Email         string            `json:"email"`
	Password      string            `json:"password"`
	Organization  *string           `json:"organization"`
	Concerns      []string          `json:"concerns"`
	Onboarding    map[string]string `json:"onboarding"`
	Producer      *struct {
		City  string `json:"city"`
		State string `json:"state"`
	} `json:"producer"`
	Property *struct {
		Latitude  *float64        `json:"latitude"`
		Longitude *float64        `json:"longitude"`
		Crop      string          `json:"crop"`
		SoilType  string          `json:"soil_type"`
		CropStage model.CropStage `json:"crop_stage"`
	} `json:"property"`
}

// validate returns user-facing (pt-BR) messages keyed by field; "producer.*"
// and "property.*" keys point at fields nested in those objects.
func (req *signupRequest) validate() map[string]string {
	fields := map[string]string{}

	if req.Profile != model.ProfileProducer && req.Profile != model.ProfileTechnician {
		fields["profile"] = "Escolha como você trabalha."
	}
	if len([]rune(strings.TrimSpace(req.Name))) < 2 {
		fields["name"] = "Digite seu nome."
	}
	if !phonePattern.MatchString(req.WhatsAppPhone) {
		fields["whatsapp_phone"] = "Digite o DDD e o número do WhatsApp."
	}
	if !validEmail(req.Email) {
		fields["email"] = "Digite um e-mail válido."
	}
	if n := len(req.Password); n < auth.MinPasswordLength || n > auth.MaxPasswordLength {
		fields["password"] = "A senha precisa ter de 8 a 72 caracteres."
	}
	for _, c := range req.Concerns {
		if !validConcerns[c] {
			fields["concerns"] = "Escolha riscos da lista."
			break
		}
	}
	for key, value := range req.Onboarding {
		allowed, known := onboardingValues[key]
		if !known || !allowed[value] {
			fields["onboarding"] = "As respostas do quiz estão inválidas."
			break
		}
	}

	organization := ""
	if req.Organization != nil {
		organization = strings.TrimSpace(*req.Organization)
	}

	switch req.Profile {
	case model.ProfileTechnician:
		if organization == "" {
			fields["organization"] = "Digite o nome da sua organização."
		}
		if req.Producer != nil {
			fields["producer"] = "Técnicos cadastram produtores depois, no painel."
		}
		if req.Property != nil {
			fields["property"] = "Técnicos cadastram propriedades depois, no painel."
		}
	case model.ProfileProducer:
		if organization != "" {
			fields["organization"] = "Organização é só para técnicos."
		}
		req.validateProducerAndProperty(fields)
	}

	return fields
}

func (req *signupRequest) validateProducerAndProperty(fields map[string]string) {
	if req.Producer == nil {
		fields["producer"] = "Informe a cidade."
	} else {
		if strings.TrimSpace(req.Producer.City) == "" {
			fields["producer.city"] = "Escolha a cidade."
		}
		if !statePattern.MatchString(strings.ToUpper(strings.TrimSpace(req.Producer.State))) {
			fields["producer.state"] = "Informe o estado com 2 letras."
		}
	}

	if req.Property == nil {
		fields["property"] = "Marque a sua terra no mapa."
		return
	}
	p := req.Property
	if p.Latitude == nil || *p.Latitude < -90 || *p.Latitude > 90 {
		fields["property.latitude"] = "Marque a sua terra no mapa."
	}
	if p.Longitude == nil || *p.Longitude < -180 || *p.Longitude > 180 {
		fields["property.longitude"] = "Marque a sua terra no mapa."
	}
	if strings.TrimSpace(p.Crop) == "" {
		fields["property.crop"] = "Escolha ou digite a cultura."
	}
	if !validSoils[p.SoilType] {
		fields["property.soil_type"] = "Escolha o tipo de solo."
	}
	if !validStages[p.CropStage] {
		fields["property.crop_stage"] = "Escolha em que fase está."
	}
}

func validEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}
```

- [ ] **Step 9: Implementar `internal/handler/auth_handler.go`**

```go
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/auth"
	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

const (
	sessionTTL       = 30 * 24 * time.Hour
	maxAuthBodyBytes = 1 << 20
)

type Signuper interface {
	Signup(ctx context.Context, in repository.SignupInput) (repository.SignupResult, error)
}

type AuthStore interface {
	GetByEmail(ctx context.Context, email string) (model.User, error)
	CreateSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
	DeleteSession(ctx context.Context, tokenHash string) error
	DeleteExpiredSessions(ctx context.Context) error
}

type AuthHandler struct {
	signuper Signuper
	store    AuthStore
	// dummyHash lets login spend the same bcrypt time for unknown emails,
	// so response timing doesn't reveal which emails are registered.
	dummyHash string
}

func NewAuthHandler(signuper Signuper, store AuthStore) *AuthHandler {
	dummy, _ := auth.HashPassword("not-a-real-password")
	return &AuthHandler{signuper: signuper, store: store, dummyHash: dummy}
}

func (h *AuthHandler) RegisterPublic(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/signup", h.signup)
	mux.HandleFunc("POST /auth/login", h.login)
}

func (h *AuthHandler) RegisterProtected(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/logout", h.logout)
	mux.HandleFunc("GET /auth/me", h.me)
}

type authResponse struct {
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	User      model.User `json:"user"`
}

func (h *AuthHandler) signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxAuthBodyBytes)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if fields := req.validate(); len(fields) > 0 {
		writeFieldErrors(w, http.StatusBadRequest, "invalid signup data", fields)
		return
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create account")
		return
	}
	token, tokenHash, err := auth.NewToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create account")
		return
	}
	expiresAt := time.Now().Add(sessionTTL)

	user := model.User{
		Email:         strings.ToLower(strings.TrimSpace(req.Email)),
		PasswordHash:  passwordHash,
		Profile:       req.Profile,
		Name:          strings.TrimSpace(req.Name),
		WhatsAppPhone: req.WhatsAppPhone,
		Concerns:      req.Concerns,
		Onboarding:    req.Onboarding,
	}
	if req.Organization != nil && strings.TrimSpace(*req.Organization) != "" {
		org := strings.TrimSpace(*req.Organization)
		user.Organization = &org
	}

	input := repository.SignupInput{
		User:             user,
		SessionTokenHash: tokenHash,
		SessionExpiresAt: expiresAt,
	}
	if req.Profile == model.ProfileProducer {
		input.Producer = &model.Producer{
			City:  strings.TrimSpace(req.Producer.City),
			State: strings.ToUpper(strings.TrimSpace(req.Producer.State)),
		}
		input.Property = &model.Property{
			Latitude:  *req.Property.Latitude,
			Longitude: *req.Property.Longitude,
			Crop:      strings.TrimSpace(req.Property.Crop),
			SoilType:  req.Property.SoilType,
			CropStage: req.Property.CropStage,
		}
	}

	result, err := h.signuper.Signup(r.Context(), input)
	if errors.Is(err, repository.ErrEmailTaken) {
		writeFieldErrors(w, http.StatusConflict, "email already registered",
			map[string]string{"email": "Este e-mail já tem cadastro."})
		return
	}
	if err != nil {
		log.Printf("handler: signup: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create account")
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{Token: token, ExpiresAt: expiresAt, User: result.User})
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxAuthBodyBytes)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.store.GetByEmail(r.Context(), strings.ToLower(strings.TrimSpace(req.Email)))
	found := err == nil
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		log.Printf("handler: login lookup: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to log in")
		return
	}

	hash := h.dummyHash
	if found {
		hash = user.PasswordHash
	}
	passwordOK := auth.CheckPassword(hash, req.Password)
	if !found || !passwordOK {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if err := h.store.DeleteExpiredSessions(r.Context()); err != nil {
		log.Printf("handler: purge expired sessions: %v", err)
	}

	token, tokenHash, err := auth.NewToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to log in")
		return
	}
	expiresAt := time.Now().Add(sessionTTL)
	if err := h.store.CreateSession(r.Context(), user.ID, tokenHash, expiresAt); err != nil {
		log.Printf("handler: create session: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to log in")
		return
	}

	writeJSON(w, http.StatusOK, authResponse{Token: token, ExpiresAt: expiresAt, User: user})
}

func (h *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	hash, ok := tokenHashFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.store.DeleteSession(r.Context(), hash); err != nil {
		log.Printf("handler: delete session: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to log out")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) me(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, user)
}
```

- [ ] **Step 10: Rodar e ver passar**

Run: `go test ./internal/handler/ -run "AuthHandler|RequireSession" -v`
Expected: PASS. (Os demais testes do pacote ainda não foram alterados e continuam compilando, pois só adicionamos arquivos.)

Run: `go build ./... && go vet ./internal/handler/`
Expected: sem erros.

- [ ] **Step 11: Commit**

```bash
git add internal/handler
git commit -m "feat: add session middleware and auth endpoints" -m "Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>"
```

---

### Task 8: Isolar handlers existentes por usuário e ligar tudo no `main.go`

Os handlers de produtor, propriedade e alerta passam a usar os métodos por usuário. Os testes e o `main.go` mudam no mesmo commit para o projeto compilar.

**Files:**
- Modify (substituir): `internal/handler/producer_handler.go`, `property_handler.go`, `alert_handler.go`
- Modify (substituir): `internal/handler/producer_handler_test.go`, `property_handler_test.go`, `alert_handler_test.go`
- Modify (substituir): `cmd/api/main.go`

- [ ] **Step 1: Substituir `internal/handler/producer_handler.go`**

```go
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

type ProducerStore interface {
	CreateForUser(ctx context.Context, userID string, p model.Producer) (model.Producer, error)
	GetByIDForUser(ctx context.Context, id, userID string) (model.Producer, error)
	ListByUser(ctx context.Context, userID string) ([]model.Producer, error)
}

type ProducerHandler struct {
	store ProducerStore
}

func NewProducerHandler(store ProducerStore) *ProducerHandler {
	return &ProducerHandler{store: store}
}

func (h *ProducerHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /producers", h.create)
	mux.HandleFunc("GET /producers", h.list)
	mux.HandleFunc("GET /producers/{id}", h.getByID)
}

func (h *ProducerHandler) create(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	var p model.Producer
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.store.CreateForUser(r.Context(), user.ID, p)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create producer")
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (h *ProducerHandler) list(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	producers, err := h.store.ListByUser(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list producers")
		return
	}

	writeJSON(w, http.StatusOK, producers)
}

func (h *ProducerHandler) getByID(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	p, err := h.store.GetByIDForUser(r.Context(), r.PathValue("id"), user.ID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "producer not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get producer")
		return
	}

	writeJSON(w, http.StatusOK, p)
}
```

- [ ] **Step 2: Substituir `internal/handler/property_handler.go`**

```go
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
	"github.com/GuilhermePain/agrolang-api/internal/service/risk"
	"github.com/GuilhermePain/agrolang-api/internal/weather"
)

type PropertyStore interface {
	CreateForUser(ctx context.Context, userID string, p model.Property) (model.Property, error)
	GetByIDForUser(ctx context.Context, id, userID string) (model.Property, error)
	ListByUser(ctx context.Context, userID string) ([]model.Property, error)
}

type PropertyAlertLister interface {
	ListByProperty(ctx context.Context, propertyID string) ([]model.Alert, error)
}

type PropertyForecastFetcher interface {
	FetchForecast(ctx context.Context, coords weather.Coordinates) ([]risk.WeatherReading, error)
}

type PropertyHandler struct {
	store    PropertyStore
	alerts   PropertyAlertLister
	forecast PropertyForecastFetcher
}

func NewPropertyHandler(store PropertyStore, alerts PropertyAlertLister, forecast PropertyForecastFetcher) *PropertyHandler {
	return &PropertyHandler{store: store, alerts: alerts, forecast: forecast}
}

func (h *PropertyHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /properties", h.create)
	mux.HandleFunc("GET /properties", h.list)
	mux.HandleFunc("GET /properties/{id}", h.getByID)
	mux.HandleFunc("GET /properties/{id}/alerts", h.listAlerts)
	mux.HandleFunc("GET /properties/{id}/forecast", h.forecastFor)
}

func (h *PropertyHandler) create(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	var p model.Property
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.store.CreateForUser(r.Context(), user.ID, p)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "producer not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create property")
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

// ownedProperty loads the property only if it belongs to the user, writing
// the 404/500 response itself when it can't.
func (h *PropertyHandler) ownedProperty(w http.ResponseWriter, r *http.Request) (model.Property, bool) {
	user, ok := requireUser(w, r)
	if !ok {
		return model.Property{}, false
	}

	p, err := h.store.GetByIDForUser(r.Context(), r.PathValue("id"), user.ID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "property not found")
		return model.Property{}, false
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get property")
		return model.Property{}, false
	}
	return p, true
}

func (h *PropertyHandler) getByID(w http.ResponseWriter, r *http.Request) {
	p, ok := h.ownedProperty(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *PropertyHandler) list(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	properties, err := h.store.ListByUser(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list properties")
		return
	}

	writeJSON(w, http.StatusOK, properties)
}

func (h *PropertyHandler) listAlerts(w http.ResponseWriter, r *http.Request) {
	p, ok := h.ownedProperty(w, r)
	if !ok {
		return
	}

	alerts, err := h.alerts.ListByProperty(r.Context(), p.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list alerts")
		return
	}

	writeJSON(w, http.StatusOK, alerts)
}

func (h *PropertyHandler) forecastFor(w http.ResponseWriter, r *http.Request) {
	p, ok := h.ownedProperty(w, r)
	if !ok {
		return
	}

	readings, err := h.forecast.FetchForecast(r.Context(), weather.Coordinates{Lat: p.Latitude, Lon: p.Longitude})
	if err != nil {
		writeError(w, http.StatusBadGateway, "failed to fetch forecast")
		return
	}

	writeJSON(w, http.StatusOK, readings)
}
```

- [ ] **Step 3: Substituir `internal/handler/alert_handler.go`**

```go
package handler

import (
	"context"
	"net/http"

	"github.com/GuilhermePain/agrolang-api/internal/model"
)

const defaultRecentAlertsLimit = 50

type RecentAlertLister interface {
	ListRecentForUser(ctx context.Context, userID string, limit int) ([]model.Alert, error)
}

type AlertHandler struct {
	store RecentAlertLister
}

func NewAlertHandler(store RecentAlertLister) *AlertHandler {
	return &AlertHandler{store: store}
}

func (h *AlertHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /alerts", h.list)
}

func (h *AlertHandler) list(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	alerts, err := h.store.ListRecentForUser(r.Context(), user.ID, defaultRecentAlertsLimit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list alerts")
		return
	}

	writeJSON(w, http.StatusOK, alerts)
}
```

- [ ] **Step 4: Substituir `internal/handler/producer_handler_test.go`**

```go
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GuilhermePain/agrolang-api/internal/handler"
	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

type fakeProducerStore struct {
	created      model.Producer
	createdFor   string
	lookedUpFor  string
	listedFor    string
	byID         map[string]model.Producer
	all          []model.Producer
	createErr    error
	listErr      error
}

func (f *fakeProducerStore) CreateForUser(ctx context.Context, userID string, p model.Producer) (model.Producer, error) {
	if f.createErr != nil {
		return model.Producer{}, f.createErr
	}
	p.ID = "generated-id"
	f.created = p
	f.createdFor = userID
	return p, nil
}

func (f *fakeProducerStore) GetByIDForUser(ctx context.Context, id, userID string) (model.Producer, error) {
	f.lookedUpFor = userID
	p, ok := f.byID[id]
	if !ok {
		return model.Producer{}, repository.ErrNotFound
	}
	return p, nil
}

func (f *fakeProducerStore) ListByUser(ctx context.Context, userID string) ([]model.Producer, error) {
	f.listedFor = userID
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.all, nil
}

func newProducerMux(store *fakeProducerStore) *http.ServeMux {
	mux := http.NewServeMux()
	handler.NewProducerHandler(store).Register(mux)
	return mux
}

func TestProducerHandler_Create_ReturnsCreatedProducerOwnedByUser(t *testing.T) {
	store := &fakeProducerStore{}
	mux := newProducerMux(store)

	body := `{"name":"Joao","whatsapp_phone":"5511999999999","city":"Campinas","state":"SP"}`
	req := authed(httptest.NewRequest(http.MethodPost, "/producers", bytes.NewBufferString(body)))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var got model.Producer
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.ID != "generated-id" || got.Name != "Joao" || got.WhatsAppPhone != "5511999999999" {
		t.Errorf("unexpected producer: %+v", got)
	}
	if store.createdFor != testUserID {
		t.Errorf("expected producer created for %q, got %q", testUserID, store.createdFor)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"whatsapp_phone":"5511999999999"`)) {
		t.Errorf("expected snake_case whatsapp_phone key, got %s", rec.Body.String())
	}
}

func TestProducerHandler_GetByID_ReturnsProducerScopedToUser(t *testing.T) {
	store := &fakeProducerStore{byID: map[string]model.Producer{"prod-1": {ID: "prod-1", Name: "Maria"}}}
	mux := newProducerMux(store)

	req := authed(httptest.NewRequest(http.MethodGet, "/producers/prod-1", nil))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got model.Producer
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Name != "Maria" {
		t.Errorf("expected Name Maria, got %q", got.Name)
	}
	if store.lookedUpFor != testUserID {
		t.Errorf("expected lookup scoped to %q, got %q", testUserID, store.lookedUpFor)
	}
}

func TestProducerHandler_GetByID_NotFound_Returns404(t *testing.T) {
	mux := newProducerMux(&fakeProducerStore{byID: map[string]model.Producer{}})

	req := authed(httptest.NewRequest(http.MethodGet, "/producers/missing", nil))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestProducerHandler_List_ReturnsUsersProducers(t *testing.T) {
	store := &fakeProducerStore{all: []model.Producer{{ID: "prod-1", Name: "Maria"}, {ID: "prod-2", Name: "Joao"}}}
	mux := newProducerMux(store)

	req := authed(httptest.NewRequest(http.MethodGet, "/producers", nil))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got []model.Producer
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 producers, got %d", len(got))
	}
	if store.listedFor != testUserID {
		t.Errorf("expected list scoped to %q, got %q", testUserID, store.listedFor)
	}
}

func TestProducerHandler_Create_InvalidJSON_Returns400(t *testing.T) {
	mux := newProducerMux(&fakeProducerStore{})

	req := authed(httptest.NewRequest(http.MethodPost, "/producers", bytes.NewBufferString("not json")))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestProducerHandler_WithoutUser_Returns401(t *testing.T) {
	mux := newProducerMux(&fakeProducerStore{})

	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/producers"},
		{http.MethodGet, "/producers/prod-1"},
		{http.MethodPost, "/producers"},
	} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString("{}")))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: expected 401, got %d", tc.method, tc.path, rec.Code)
		}
	}
}
```

- [ ] **Step 5: Substituir `internal/handler/property_handler_test.go`**

```go
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/handler"
	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
	"github.com/GuilhermePain/agrolang-api/internal/service/risk"
	"github.com/GuilhermePain/agrolang-api/internal/weather"
)

type fakePropertyStore struct {
	byID       map[string]model.Property
	all        []model.Property
	createErr  error
	createdFor string
	listedFor  string
}

func (f *fakePropertyStore) CreateForUser(ctx context.Context, userID string, p model.Property) (model.Property, error) {
	if f.createErr != nil {
		return model.Property{}, f.createErr
	}
	p.ID = "generated-id"
	f.createdFor = userID
	return p, nil
}

func (f *fakePropertyStore) GetByIDForUser(ctx context.Context, id, userID string) (model.Property, error) {
	p, ok := f.byID[id]
	if !ok {
		return model.Property{}, repository.ErrNotFound
	}
	return p, nil
}

func (f *fakePropertyStore) ListByUser(ctx context.Context, userID string) ([]model.Property, error) {
	f.listedFor = userID
	return f.all, nil
}

type fakePropertyAlertLister struct {
	byProperty map[string][]model.Alert
}

func (f *fakePropertyAlertLister) ListByProperty(ctx context.Context, propertyID string) ([]model.Alert, error) {
	return f.byProperty[propertyID], nil
}

type fakePropertyForecastFetcher struct {
	readings []risk.WeatherReading
	err      error
}

func (f *fakePropertyForecastFetcher) FetchForecast(ctx context.Context, coords weather.Coordinates) ([]risk.WeatherReading, error) {
	return f.readings, f.err
}

func newPropertyMux(store *fakePropertyStore, alerts *fakePropertyAlertLister, fetcher *fakePropertyForecastFetcher) *http.ServeMux {
	mux := http.NewServeMux()
	handler.NewPropertyHandler(store, alerts, fetcher).Register(mux)
	return mux
}

func serve(mux http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := authed(httptest.NewRequest(method, path, bytes.NewBufferString(body)))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestPropertyHandler_Create_ReturnsCreatedProperty(t *testing.T) {
	store := &fakePropertyStore{}
	mux := newPropertyMux(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	body := `{"producer_id":"prod-1","latitude":-22.9,"longitude":-47.06,"crop":"corn","soil_type":"clay","crop_stage":"flowering"}`
	rec := serve(mux, http.MethodPost, "/properties", body)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var got model.Property
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.ID != "generated-id" {
		t.Errorf("expected ID generated-id, got %q", got.ID)
	}
	if store.createdFor != testUserID {
		t.Errorf("expected property created for %q, got %q", testUserID, store.createdFor)
	}
}

func TestPropertyHandler_Create_ForeignProducer_Returns404(t *testing.T) {
	store := &fakePropertyStore{createErr: repository.ErrNotFound}
	mux := newPropertyMux(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	rec := serve(mux, http.MethodPost, "/properties", `{"producer_id":"not-mine","latitude":0,"longitude":0,"crop":"x","soil_type":"y","crop_stage":"harvest"}`)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPropertyHandler_GetByID_ReturnsProperty(t *testing.T) {
	store := &fakePropertyStore{byID: map[string]model.Property{"prop-1": {ID: "prop-1", Crop: "corn"}}}
	mux := newPropertyMux(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	rec := serve(mux, http.MethodGet, "/properties/prop-1", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPropertyHandler_GetByID_NotFound_Returns404(t *testing.T) {
	mux := newPropertyMux(&fakePropertyStore{byID: map[string]model.Property{}}, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	rec := serve(mux, http.MethodGet, "/properties/missing", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPropertyHandler_List_ReturnsUsersProperties(t *testing.T) {
	store := &fakePropertyStore{all: []model.Property{{ID: "p1"}, {ID: "p2"}}}
	mux := newPropertyMux(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	rec := serve(mux, http.MethodGet, "/properties", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got []model.Property
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 properties, got %d", len(got))
	}
	if store.listedFor != testUserID {
		t.Errorf("expected list scoped to %q, got %q", testUserID, store.listedFor)
	}
}

func TestPropertyHandler_ListAlerts_ReturnsAlertsForOwnedProperty(t *testing.T) {
	store := &fakePropertyStore{byID: map[string]model.Property{"prop-1": {ID: "prop-1"}}}
	alertLister := &fakePropertyAlertLister{byProperty: map[string][]model.Alert{
		"prop-1": {{ID: "a1", PropertyID: "prop-1", AlertType: "frost"}},
	}}
	mux := newPropertyMux(store, alertLister, &fakePropertyForecastFetcher{})

	rec := serve(mux, http.MethodGet, "/properties/prop-1/alerts", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got []model.Alert
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 1 || got[0].AlertType != "frost" {
		t.Fatalf("expected 1 frost alert, got %v", got)
	}
}

func TestPropertyHandler_ListAlerts_ForeignProperty_Returns404(t *testing.T) {
	alertLister := &fakePropertyAlertLister{byProperty: map[string][]model.Alert{
		"prop-9": {{ID: "a1", PropertyID: "prop-9"}},
	}}
	mux := newPropertyMux(&fakePropertyStore{byID: map[string]model.Property{}}, alertLister, &fakePropertyForecastFetcher{})

	rec := serve(mux, http.MethodGet, "/properties/prop-9/alerts", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for a property the user does not own, got %d", rec.Code)
	}
}

func TestPropertyHandler_Forecast_ReturnsReadingsForProperty(t *testing.T) {
	store := &fakePropertyStore{byID: map[string]model.Property{"prop-1": {ID: "prop-1", Latitude: -22.9, Longitude: -47.06}}}
	fetcher := &fakePropertyForecastFetcher{readings: []risk.WeatherReading{{Time: time.Now(), TempAvgC: 25, HumidityPct: 50}}}
	mux := newPropertyMux(store, &fakePropertyAlertLister{}, fetcher)

	rec := serve(mux, http.MethodGet, "/properties/prop-1/forecast", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got []risk.WeatherReading
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 1 || got[0].TempAvgC != 25 {
		t.Fatalf("expected 1 reading with TempAvgC 25, got %v", got)
	}
}

func TestPropertyHandler_Forecast_PropertyNotFound_Returns404(t *testing.T) {
	mux := newPropertyMux(&fakePropertyStore{byID: map[string]model.Property{}}, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	rec := serve(mux, http.MethodGet, "/properties/missing/forecast", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPropertyHandler_WithoutUser_Returns401(t *testing.T) {
	mux := newPropertyMux(&fakePropertyStore{}, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/properties", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
```

- [ ] **Step 6: Substituir `internal/handler/alert_handler_test.go`**

```go
package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GuilhermePain/agrolang-api/internal/handler"
	"github.com/GuilhermePain/agrolang-api/internal/model"
)

type fakeRecentAlertLister struct {
	recent   []model.Alert
	listedFor string
}

func (f *fakeRecentAlertLister) ListRecentForUser(ctx context.Context, userID string, limit int) ([]model.Alert, error) {
	f.listedFor = userID
	if limit < len(f.recent) {
		return f.recent[:limit], nil
	}
	return f.recent, nil
}

func TestAlertHandler_List_ReturnsRecentAlertsForUser(t *testing.T) {
	store := &fakeRecentAlertLister{recent: []model.Alert{
		{ID: "a1", AlertType: "frost"},
		{ID: "a2", AlertType: "heavy_rain"},
	}}
	mux := http.NewServeMux()
	handler.NewAlertHandler(store).Register(mux)

	req := authed(httptest.NewRequest(http.MethodGet, "/alerts", nil))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got []model.Alert
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 alerts, got %d", len(got))
	}
	if store.listedFor != testUserID {
		t.Errorf("expected list scoped to %q, got %q", testUserID, store.listedFor)
	}
}

func TestAlertHandler_WithoutUser_Returns401(t *testing.T) {
	mux := http.NewServeMux()
	handler.NewAlertHandler(&fakeRecentAlertLister{}).Register(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/alerts", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
```

- [ ] **Step 7: Substituir `cmd/api/main.go`**

```go
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/GuilhermePain/agrolang-api/internal/app"
	"github.com/GuilhermePain/agrolang-api/internal/config"
	"github.com/GuilhermePain/agrolang-api/internal/handler"
	"github.com/GuilhermePain/agrolang-api/internal/notifier"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
	"github.com/GuilhermePain/agrolang-api/internal/weather"
)

const scanConcurrency = 5

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DBDSN())
	if err != nil {
		log.Fatalf("db: connect: %v", err)
	}
	defer pool.Close()

	propertyRepo := repository.NewPropertyRepository(pool)
	alertRepo := repository.NewAlertRepository(pool)
	producerRepo := repository.NewProducerRepository(pool)
	userRepo := repository.NewUserRepository(pool)
	signupRepo := repository.NewSignupRepository(pool)
	weatherClient := weather.NewOpenMeteoClient(weather.WithBaseURL(cfg.OpenMeteoBaseURL))
	notifierClient := notifier.NewClient(cfg.EvolutionAPIURL, cfg.EvolutionAPIKey)

	go runScanLoop(ctx, propertyRepo, weatherClient, alertRepo, producerRepo, notifierClient, cfg.ScanInterval)

	root := http.NewServeMux()
	root.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	authHandler := handler.NewAuthHandler(signupRepo, userRepo)
	authHandler.RegisterPublic(root)

	// Everything else requires a live session.
	protected := http.NewServeMux()
	authHandler.RegisterProtected(protected)
	handler.NewProducerHandler(producerRepo).Register(protected)
	handler.NewPropertyHandler(propertyRepo, alertRepo, weatherClient).Register(protected)
	handler.NewAlertHandler(alertRepo).Register(protected)
	root.Handle("/", handler.RequireSession(userRepo)(protected))

	addr := ":" + cfg.HTTPPort
	log.Printf("agrolang-api listening on %s", addr)
	if err := http.ListenAndServe(addr, handler.WithCORS(root)); err != nil {
		log.Fatal(err)
	}
}

func runScanLoop(ctx context.Context, propertyRepo *repository.PropertyRepository, weatherClient *weather.OpenMeteoClient, alertRepo *repository.AlertRepository, producerRepo *repository.ProducerRepository, notifierClient *notifier.Client, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	runOnce := func() {
		if err := app.RunCycle(ctx, propertyRepo, weatherClient, alertRepo, producerRepo, notifierClient, scanConcurrency); err != nil {
			log.Printf("app: scan cycle failed: %v", err)
		}
	}

	runOnce()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runOnce()
		}
	}
}
```

- [ ] **Step 8: Formatar, compilar e testar tudo**

Run:
```bash
gofmt -l . ; gofmt -w internal cmd
go build ./... && go vet ./...
go test ./... 2>&1 | tail -20
```
Expected: `gofmt -l` pode listar arquivos antes do `-w`; depois do build e vet, sem erros; `go test ./...` imprime `ok` para todos os pacotes.

- [ ] **Step 9: Smoke test manual com curl**

Parar a API que está rodando na porta 8080 (binário antigo) e subir a nova:

```bash
kill $(ss -ltnp 2>/dev/null | grep ':8080' | grep -o 'pid=[0-9]*' | cut -d= -f2) 2>/dev/null
cd /home/lukas-melo/PROJETOS/agrolang
set -a; source .env; set +a
DB_HOST=localhost go run ./cmd/api > /tmp/agrolang-api.log 2>&1 &
sleep 3 && curl -s localhost:8080/health
```
Expected: `{"status":"ok"}`.

```bash
# 401 sem sessão
curl -s -o /dev/null -w "%{http_code}\n" localhost:8080/properties

# signup do produtor
curl -s -X POST localhost:8080/auth/signup -H 'Content-Type: application/json' -d '{
 "profile":"producer","name":"José da Silva","whatsapp_phone":"5519999999999",
 "email":"smoke@exemplo.com","password":"segredo123","organization":null,
 "concerns":["frost"],"onboarding":{"work_style":"family","areas_bucket":"1"},
 "producer":{"city":"Campinas","state":"SP"},
 "property":{"latitude":-22.9099,"longitude":-47.0626,"crop":"Milho","soil_type":"argiloso","crop_stage":"flowering"}}' | tee /tmp/signup.json
```
Expected: primeiro `401`; o signup devolve JSON com `token`, `expires_at` e `user` (sem `password_hash`).

```bash
TOKEN=$(python3 -c 'import json;print(json.load(open("/tmp/signup.json"))["token"])')
curl -s localhost:8080/auth/me -H "Authorization: Bearer $TOKEN"
curl -s localhost:8080/properties -H "Authorization: Bearer $TOKEN"
curl -s -X POST localhost:8080/auth/logout -H "Authorization: Bearer $TOKEN" -o /dev/null -w "%{http_code}\n"
curl -s localhost:8080/auth/me -H "Authorization: Bearer $TOKEN" -o /dev/null -w "%{http_code}\n"
```
Expected: `/auth/me` devolve o usuário; `/properties` devolve 1 propriedade; logout `204`; `/auth/me` depois do logout `401`.

Limpar o usuário de teste: `docker exec agrolang-db-1 psql -U agrolang -d agrolang -c "DELETE FROM users WHERE email='smoke@exemplo.com'"` (apaga em cascata o produtor e a propriedade).

- [ ] **Step 10: Commit**

```bash
git add internal/handler cmd/api/main.go
git commit -m "feat: require a session and scope producers, properties and alerts by owner" -m "Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>"
```

---

### Task 9: Alertas respeitam as preocupações (crítico sempre notifica)

**Files:**
- Create: `internal/app/notify.go`
- Modify: `internal/app/cycle.go`
- Test: `internal/app/notify_test.go`

- [ ] **Step 1: Escrever os testes que falham**

`internal/app/notify_test.go` (reaproveita os fakes de `cycle_test.go`, mesmo pacote `app_test`):
```go
package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/app"
	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/service/risk"
)

func TestShouldNotify(t *testing.T) {
	cases := []struct {
		name     string
		level    risk.Level
		alert    risk.Alert
		concerns []string
		want     bool
	}{
		{"critical always notifies even when not chosen", risk.LevelCritical, risk.AlertFrost, []string{"heavy_rain"}, true},
		{"medium notifies when the type was chosen", risk.LevelMedium, risk.AlertFrost, []string{"frost", "heavy_rain"}, true},
		{"medium is skipped when the type was not chosen", risk.LevelMedium, risk.AlertHeatWave, []string{"frost"}, false},
		{"medium notifies when no concerns were chosen", risk.LevelMedium, risk.AlertHeatWave, nil, true},
		{"medium notifies for an empty concerns list", risk.LevelMedium, risk.AlertHeavyRain, []string{}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := app.ShouldNotify(tc.level, tc.alert, tc.concerns); got != tc.want {
				t.Errorf("ShouldNotify(%s, %s, %v) = %v, want %v", tc.level, tc.alert, tc.concerns, got, tc.want)
			}
		})
	}
}

func TestRunCycle_CriticalAlertNotifiesEvenWhenOwnerDidNotChooseIt(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	lister := &fakeLister{properties: []model.Property{
		{ID: "prop-1", ProducerID: "producer-1", Crop: "tomato", CropStage: model.CropStageHarvest},
	}}
	fetcher := &fakeFetcher{readings: []risk.WeatherReading{
		{Time: start, TempAvgC: 4, HumidityPct: 60},
		{Time: start.Add(2 * time.Hour), TempAvgC: 4, HumidityPct: 60},
	}}
	creator := &fakeAlertCreator{}
	producers := &fakeProducerFinder{producers: map[string]model.Producer{
		"producer-1": {ID: "producer-1", WhatsAppPhone: "5511999999999", Concerns: []string{"heavy_rain"}},
	}}
	notif := &fakeNotifier{}

	if err := app.RunCycle(context.Background(), lister, fetcher, creator, producers, notif, 2); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	if len(creator.created) != 1 {
		t.Fatalf("expected the alert to be persisted, got %d", len(creator.created))
	}
	if len(notif.sent) != 1 {
		t.Fatalf("expected a critical frost alert to notify despite concerns, got %d notifications", len(notif.sent))
	}
}
```

- [ ] **Step 2: Rodar e ver falhar**

Run: `go test ./internal/app/ -run "ShouldNotify|CriticalAlertNotifies"`
Expected: FAIL (`undefined: app.ShouldNotify`).

- [ ] **Step 3: Implementar `internal/app/notify.go`**

```go
package app

import (
	"slices"

	"github.com/GuilhermePain/agrolang-api/internal/service/risk"
)

// ShouldNotify decides whether a WhatsApp alert goes out. Critical alerts
// always do, because the crop is at stake. Medium alerts are filtered by the
// risk types the account owner chose; an empty list means "notify everything".
func ShouldNotify(level risk.Level, alert risk.Alert, concerns []string) bool {
	if level == risk.LevelCritical {
		return true
	}
	if len(concerns) == 0 {
		return true
	}
	return slices.Contains(concerns, string(alert))
}
```

- [ ] **Step 4: Usar o filtro em `notifyProducer` (`internal/app/cycle.go`)**

Substituir a função `notifyProducer` por:

```go
func notifyProducer(ctx context.Context, producers ProducerFinder, notif Notifier, property model.Property, res risk.Result, periodStart, periodEnd time.Time) {
	producer, err := producers.GetByID(ctx, property.ProducerID)
	if err != nil {
		log.Printf("app: lookup producer for property %s: %v", property.ID, err)
		return
	}

	if !ShouldNotify(res.Level, res.Alert, producer.Concerns) {
		return
	}

	message := notifier.BuildMessage(res, periodStart, periodEnd)
	if err := notif.Notify(ctx, producer.WhatsAppPhone, message); err != nil {
		log.Printf("app: notify producer %s: %v", producer.ID, err)
	}
}
```

- [ ] **Step 5: Rodar todos os testes do pacote**

Run: `go test ./internal/app/ -v`
Expected: PASS (testes novos e antigos de `cycle_test.go`).

- [ ] **Step 6: Commit**

```bash
git add internal/app
git commit -m "feat: filter medium WhatsApp alerts by the owner's chosen risks" -m "Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>"
```

---

## Parte B — Frontend (Next 16)

> Rodar comandos em `/home/lukas-melo/PROJETOS/agrolang/web`. Testes: `npm test`. Imports dentro de `lib/` usam caminho **relativo** (como os testes atuais); componentes e páginas usam o alias `@/`.

### Task 10: Bibliotecas puras do quiz (tipos, passos, estado, payload)

**Files:**
- Create: `web/lib/types.ts`
- Create: `web/lib/quiz/{types,options,steps,account,state,payload}.ts`
- Test: `web/lib/quiz/{steps,account,state,payload}.test.ts`

- [ ] **Step 1: Criar `web/lib/types.ts`** (tipos compartilhados; inclui `AuthUser`)

```ts
export type RiskLevel = "low" | "medium" | "critical";

export type CropStage = "germination" | "flowering" | "harvest";

export type Producer = {
  id: string;
  name: string;
  whatsapp_phone: string;
  city: string;
  state: string;
  created_at: string;
  updated_at: string;
};

export type Property = {
  id: string;
  producer_id: string;
  latitude: number;
  longitude: number;
  crop: string;
  soil_type: string;
  crop_stage: string;
  created_at: string;
  updated_at: string;
};

export type Alert = {
  id: string;
  property_id: string;
  level: RiskLevel;
  alert_type: string;
  period_start: string;
  period_end: string;
  triggered_at: string;
  resolved_at: string | null;
};

export type WeatherReading = {
  time: string;
  temp_avg_c: number;
  humidity_pct: number;
  precipitation_mm: number;
};

export type NewProducer = {
  name: string;
  whatsapp_phone: string;
  city: string;
  state: string;
};

export type NewProperty = {
  producer_id: string;
  latitude: number;
  longitude: number;
  crop: string;
  soil_type: string;
  crop_stage: CropStage;
};

export type AuthUser = {
  id: string;
  email: string;
  profile: "producer" | "technician";
  name: string;
  whatsapp_phone: string;
  organization: string | null;
  concerns: string[];
  onboarding: Record<string, string>;
  created_at: string;
};
```

- [ ] **Step 2: Criar `web/lib/quiz/types.ts`**

```ts
import type { CropStage } from "../types";

export type { CropStage };

export type Profile = "producer" | "technician";
export type WorkStyle = "family" | "multi_area" | "technician";
export type AreasBucket = "1" | "2-5" | "6+";
export type ProducersBucket = "ate-10" | "11-40" | "mais-40";
export type Concern = "frost" | "severe_water_stress" | "heat_wave" | "heavy_rain";
export type SoilType = "argiloso" | "arenoso" | "misto" | "nao_informado";

/** Everything the quiz collects. The password is deliberately not part of it, so it can never be persisted. */
export type QuizAnswers = {
  work_style: WorkStyle | null;
  areas_bucket: AreasBucket | null;
  producers_bucket: ProducersBucket | null;
  name: string;
  /** What the person typed (DDD + number); normalized only when building the payload. */
  phone: string;
  organization: string;
  /** One of CROP_OPTIONS or "outra"; the typed name lives in crop_custom. */
  crop_choice: string;
  crop_custom: string;
  crop_stage: CropStage | null;
  city: string;
  state: string;
  latitude: number | null;
  longitude: number | null;
  soil_type: SoilType | null;
  concerns: Concern[];
  email: string;
};
```

- [ ] **Step 3: Criar `web/lib/quiz/options.ts`**

```ts
import type {
  AreasBucket,
  Concern,
  CropStage,
  ProducersBucket,
  SoilType,
  WorkStyle,
} from "./types";

export type Option<T extends string> = { value: T; label: string; hint?: string };

export const STYLE_OPTIONS: Option<WorkStyle>[] = [
  { value: "family", label: "Planto na minha terra, com a família" },
  { value: "multi_area", label: "Tenho várias áreas e contrato gente" },
  { value: "technician", label: "Sou agrônomo ou técnico de cooperativa" },
];

export const AREAS_OPTIONS: Option<AreasBucket>[] = [
  { value: "1", label: "Uma só" },
  { value: "2-5", label: "De 2 a 5" },
  { value: "6+", label: "Mais de 5" },
];

export const PRODUCERS_OPTIONS: Option<ProducersBucket>[] = [
  { value: "ate-10", label: "Até 10 produtores" },
  { value: "11-40", label: "De 11 a 40" },
  { value: "mais-40", label: "Mais de 40" },
];

export const CROP_OPTIONS = ["Milho", "Tomate", "Café", "Soja", "Feijão"] as const;
export const CROP_OTHER = "outra";

export const STAGE_OPTIONS: Option<CropStage>[] = [
  { value: "germination", label: "Acabei de plantar", hint: "Germinação" },
  { value: "flowering", label: "Está florindo", hint: "Floração" },
  { value: "harvest", label: "Perto de colher", hint: "Colheita" },
];

export const SOIL_OPTIONS: Option<SoilType>[] = [
  { value: "argiloso", label: "Argiloso", hint: "Terra vermelha ou barrenta, gruda quando molhada" },
  { value: "arenoso", label: "Arenoso", hint: "Solta e clara, a água escorre rápido" },
  { value: "misto", label: "Misto", hint: "Meio a meio, ou não é bem de um tipo só" },
  { value: "nao_informado", label: "Não sei", hint: "Tudo bem, dá para ajustar depois" },
];

export const CONCERN_OPTIONS: Option<Concern>[] = [
  { value: "frost", label: "Geada", hint: "Frio forte que queima a planta" },
  { value: "severe_water_stress", label: "Seca", hint: "Falta de chuva e ar muito seco" },
  { value: "heat_wave", label: "Calor forte", hint: "Dias seguidos muito quentes" },
  { value: "heavy_rain", label: "Chuva forte", hint: "Temporal, alagamento nas áreas baixas" },
];
```

- [ ] **Step 4: Escrever os testes de `steps` (falham)**

`web/lib/quiz/steps.test.ts`:
```ts
import { describe, expect, it } from "vitest";
import {
  cropValue,
  firstIncompleteIndex,
  normalizePhone,
  profileOf,
  stepForField,
  stepsFor,
  validateStep,
} from "./steps";
import { initialAnswers } from "./state";
import type { QuizAnswers } from "./types";

const answers = (patch: Partial<QuizAnswers> = {}): QuizAnswers => ({ ...initialAnswers, ...patch });

describe("stepsFor / profileOf", () => {
  it("only offers the first step until a work style is chosen", () => {
    expect(stepsFor(null)).toEqual(["style"]);
    expect(profileOf(null)).toBeNull();
  });

  it("gives producers 8 steps ending in the account", () => {
    expect(stepsFor("family")).toEqual([
      "style", "areas", "identity", "crop", "location", "soil", "concerns", "account",
    ]);
    expect(stepsFor("multi_area")).toHaveLength(8);
    expect(profileOf("family")).toBe("producer");
    expect(profileOf("multi_area")).toBe("producer");
  });

  it("gives technicians 5 steps", () => {
    expect(stepsFor("technician")).toEqual(["style", "identity", "producers", "concerns", "account"]);
    expect(profileOf("technician")).toBe("technician");
  });
});

describe("normalizePhone", () => {
  it("adds the Brazil country code to DDD + number", () => {
    expect(normalizePhone("19 99999-9999")).toBe("5519999999999");
    expect(normalizePhone("(19) 3333-4444")).toBe("551933334444");
  });

  it("keeps a number that already has the country code", () => {
    expect(normalizePhone("+55 19 99999-9999")).toBe("5519999999999");
  });

  it("rejects numbers of the wrong length", () => {
    expect(normalizePhone("123")).toBeNull();
    expect(normalizePhone("")).toBeNull();
  });
});

describe("cropValue", () => {
  it("returns the chosen crop", () => {
    expect(cropValue(answers({ crop_choice: "Milho" }))).toBe("Milho");
  });

  it("returns the typed crop when 'outra' is chosen", () => {
    expect(cropValue(answers({ crop_choice: "outra", crop_custom: "  Mandioca " }))).toBe("Mandioca");
  });
});

describe("validateStep", () => {
  it("requires a choice on single-choice steps", () => {
    expect(validateStep("style", answers())).toBe("Escolha uma das opções.");
    expect(validateStep("style", answers({ work_style: "family" }))).toBeNull();
    expect(validateStep("areas", answers())).toBe("Escolha uma das opções.");
    expect(validateStep("producers", answers())).toBe("Escolha uma das opções.");
    expect(validateStep("soil", answers())).toBe("Escolha o tipo de solo.");
  });

  it("validates the identity step and asks for an organization only from technicians", () => {
    const base = { name: "José", phone: "19 99999-9999" };
    expect(validateStep("identity", answers({ ...base, work_style: "family" }))).toBeNull();
    expect(validateStep("identity", answers({ ...base, work_style: "technician" }))).toBe(
      "Digite o nome da sua organização."
    );
    expect(
      validateStep("identity", answers({ ...base, work_style: "technician", organization: "Coop" }))
    ).toBeNull();
    expect(validateStep("identity", answers({ name: "J", phone: base.phone }))).toBe("Digite seu nome.");
    expect(validateStep("identity", answers({ name: "José", phone: "12" }))).toBe(
      "Digite o DDD e o número, só números. Ex.: 19 99999-9999."
    );
  });

  it("validates the crop step", () => {
    expect(validateStep("crop", answers())).toBe("Escolha ou digite a cultura.");
    expect(validateStep("crop", answers({ crop_choice: "Milho" }))).toBe("Escolha em que fase está.");
    expect(validateStep("crop", answers({ crop_choice: "Milho", crop_stage: "flowering" }))).toBeNull();
    expect(validateStep("crop", answers({ crop_choice: "outra", crop_stage: "harvest" }))).toBe(
      "Escolha ou digite a cultura."
    );
  });

  it("validates the location step", () => {
    expect(validateStep("location", answers())).toBe("Marque sua terra no mapa.");
    expect(validateStep("location", answers({ latitude: -22.9, longitude: -47.06 }))).toBe(
      "Escolha a cidade na busca."
    );
    expect(
      validateStep("location", answers({ latitude: -22.9, longitude: -47.06, city: "Campinas", state: "SP" }))
    ).toBeNull();
  });

  it("never blocks the optional concerns step", () => {
    expect(validateStep("concerns", answers())).toBeNull();
  });
});

describe("firstIncompleteIndex", () => {
  it("points at the first step that still has an error", () => {
    const steps = stepsFor("family");
    expect(firstIncompleteIndex(steps, answers({ work_style: "family" }))).toBe(1);
    expect(firstIncompleteIndex(steps, answers({ work_style: "family", areas_bucket: "1" }))).toBe(2);
  });

  it("returns the last step when everything is filled", () => {
    const steps = stepsFor("technician");
    const full = answers({
      work_style: "technician", name: "Ana", phone: "19 98888-7777",
      organization: "Coop", producers_bucket: "11-40",
    });
    expect(firstIncompleteIndex(steps, full)).toBe(steps.length - 1);
  });
});

describe("stepForField", () => {
  it("maps API field names back to the quiz step", () => {
    expect(stepForField("name")).toBe("identity");
    expect(stepForField("whatsapp_phone")).toBe("identity");
    expect(stepForField("email")).toBe("account");
    expect(stepForField("producer.city")).toBe("location");
    expect(stepForField("property.latitude")).toBe("location");
    expect(stepForField("property.crop_stage")).toBe("crop");
    expect(stepForField("property.soil_type")).toBe("soil");
    expect(stepForField("concerns")).toBe("concerns");
    expect(stepForField("unknown")).toBeNull();
  });
});
```

- [ ] **Step 5: Escrever os testes de `account` (falham)**

`web/lib/quiz/account.test.ts`:
```ts
import { describe, expect, it } from "vitest";
import { validateAccount } from "./account";

describe("validateAccount", () => {
  it("accepts a valid email and password", () => {
    expect(validateAccount("jose@exemplo.com", "segredo123")).toEqual({});
  });

  it("rejects a malformed email", () => {
    expect(validateAccount("jose", "segredo123").email).toBe("Digite um e-mail válido.");
    expect(validateAccount("jose@exemplo", "segredo123").email).toBe("Digite um e-mail válido.");
  });

  it("rejects passwords that are too short or too long", () => {
    expect(validateAccount("jose@exemplo.com", "curta").password).toBe(
      "A senha precisa ter pelo menos 8 caracteres."
    );
    expect(validateAccount("jose@exemplo.com", "a".repeat(73)).password).toBe(
      "A senha pode ter no máximo 72 caracteres."
    );
  });
});
```

- [ ] **Step 6: Escrever os testes de `state` (falham)**

`web/lib/quiz/state.test.ts`:
```ts
import { describe, expect, it } from "vitest";
import {
  STORAGE_KEY,
  clearAnswers,
  initialAnswers,
  loadAnswers,
  quizReducer,
  saveAnswers,
} from "./state";

function fakeStorage(initial: Record<string, string> = {}) {
  const data = { ...initial };
  return {
    data,
    getItem: (key: string) => data[key] ?? null,
    setItem: (key: string, value: string) => {
      data[key] = value;
    },
    removeItem: (key: string) => {
      delete data[key];
    },
  };
}

describe("quizReducer", () => {
  it("merges a patch into the answers", () => {
    const next = quizReducer(initialAnswers, { type: "patch", patch: { name: "José" } });
    expect(next.name).toBe("José");
    expect(next.city).toBe("");
  });

  it("replaces and resets the answers", () => {
    const filled = { ...initialAnswers, name: "José" };
    expect(quizReducer(initialAnswers, { type: "replace", answers: filled })).toBe(filled);
    expect(quizReducer(filled, { type: "reset" })).toEqual(initialAnswers);
  });
});

describe("persistence", () => {
  it("round-trips answers through storage", () => {
    const storage = fakeStorage();
    const answers = { ...initialAnswers, name: "José", concerns: ["frost" as const] };
    saveAnswers(answers, storage);
    expect(loadAnswers(storage)).toEqual(answers);
  });

  it("returns the initial answers when storage is empty, null or corrupted", () => {
    expect(loadAnswers(fakeStorage())).toEqual(initialAnswers);
    expect(loadAnswers(null)).toEqual(initialAnswers);
    expect(loadAnswers(fakeStorage({ [STORAGE_KEY]: "{not json" }))).toEqual(initialAnswers);
  });

  it("drops unknown keys and values of the wrong kind", () => {
    const storage = fakeStorage({
      [STORAGE_KEY]: JSON.stringify({ name: "José", password: "oops", concerns: "frost", latitude: "x", city: 5 }),
    });
    const loaded = loadAnswers(storage);
    expect(loaded.name).toBe("José");
    expect(loaded.concerns).toEqual([]);
    expect(loaded.city).toBe("");
    expect(loaded.latitude).toBeNull();
    expect("password" in loaded).toBe(false);
  });

  it("clears the stored answers", () => {
    const storage = fakeStorage({ [STORAGE_KEY]: "{}" });
    clearAnswers(storage);
    expect(storage.data[STORAGE_KEY]).toBeUndefined();
  });

  it("does not throw when storage fails", () => {
    const broken = {
      getItem: () => {
        throw new Error("blocked");
      },
      setItem: () => {
        throw new Error("blocked");
      },
      removeItem: () => {
        throw new Error("blocked");
      },
    };
    expect(loadAnswers(broken)).toEqual(initialAnswers);
    expect(() => saveAnswers(initialAnswers, broken)).not.toThrow();
    expect(() => clearAnswers(broken)).not.toThrow();
  });
});
```

- [ ] **Step 7: Escrever os testes de `payload` (falham)**

`web/lib/quiz/payload.test.ts`:
```ts
import { describe, expect, it } from "vitest";
import { buildSignupPayload } from "./payload";
import { initialAnswers } from "./state";
import type { QuizAnswers } from "./types";

const producer: QuizAnswers = {
  ...initialAnswers,
  work_style: "family",
  areas_bucket: "2-5",
  name: " José da Silva ",
  phone: "19 99999-9999",
  crop_choice: "outra",
  crop_custom: "Mandioca",
  crop_stage: "flowering",
  city: "Campinas",
  state: "SP",
  latitude: -22.9099,
  longitude: -47.0626,
  soil_type: "argiloso",
  concerns: ["frost", "severe_water_stress"],
  email: " Jose@Exemplo.com ",
  organization: "ignorado",
  producers_bucket: "11-40",
};

const technician: QuizAnswers = {
  ...initialAnswers,
  work_style: "technician",
  name: "Ana",
  phone: "(19) 98888-7777",
  organization: " Coopera Campo ",
  producers_bucket: "11-40",
  concerns: ["heavy_rain"],
  email: "ana@coop.com",
  city: "ignorada",
  latitude: 1,
  longitude: 2,
};

describe("buildSignupPayload", () => {
  it("builds the producer payload with producer and property", () => {
    expect(buildSignupPayload(producer, "segredo123")).toEqual({
      profile: "producer",
      name: "José da Silva",
      whatsapp_phone: "5519999999999",
      email: "jose@exemplo.com",
      password: "segredo123",
      organization: null,
      concerns: ["frost", "severe_water_stress"],
      onboarding: { work_style: "family", areas_bucket: "2-5" },
      producer: { city: "Campinas", state: "SP" },
      property: {
        latitude: -22.9099,
        longitude: -47.0626,
        crop: "Mandioca",
        soil_type: "argiloso",
        crop_stage: "flowering",
      },
    });
  });

  it("builds the technician payload without producer or property", () => {
    const payload = buildSignupPayload(technician, "segredo123");
    expect(payload).toEqual({
      profile: "technician",
      name: "Ana",
      whatsapp_phone: "5519988887777",
      email: "ana@coop.com",
      password: "segredo123",
      organization: "Coopera Campo",
      concerns: ["heavy_rain"],
      onboarding: { work_style: "technician", producers_bucket: "11-40" },
    });
    expect("producer" in payload).toBe(false);
    expect("property" in payload).toBe(false);
  });

  it("throws when the answers are incomplete", () => {
    expect(() => buildSignupPayload(initialAnswers, "segredo123")).toThrow("incomplete quiz answers");
    expect(() => buildSignupPayload({ ...producer, latitude: null }, "segredo123")).toThrow(
      "incomplete quiz answers"
    );
    expect(() => buildSignupPayload({ ...technician, producers_bucket: null }, "segredo123")).toThrow(
      "incomplete quiz answers"
    );
  });
});
```

- [ ] **Step 8: Rodar e ver falhar**

Run: `npx vitest run lib/quiz`
Expected: FAIL (módulos `./steps`, `./state`, `./account`, `./payload` não existem).

- [ ] **Step 9: Implementar `web/lib/quiz/state.ts`**

```ts
import type { QuizAnswers } from "./types";

export const STORAGE_KEY = "agrolang.quiz.v1";

export const initialAnswers: QuizAnswers = {
  work_style: null,
  areas_bucket: null,
  producers_bucket: null,
  name: "",
  phone: "",
  organization: "",
  crop_choice: "",
  crop_custom: "",
  crop_stage: null,
  city: "",
  state: "",
  latitude: null,
  longitude: null,
  soil_type: null,
  concerns: [],
  email: "",
};

export type QuizAction =
  | { type: "patch"; patch: Partial<QuizAnswers> }
  | { type: "replace"; answers: QuizAnswers }
  | { type: "reset" };

export function quizReducer(state: QuizAnswers, action: QuizAction): QuizAnswers {
  switch (action.type) {
    case "patch":
      return { ...state, ...action.patch };
    case "replace":
      return action.answers;
    case "reset":
      return initialAnswers;
  }
}

type StorageLike = Pick<Storage, "getItem" | "setItem" | "removeItem">;

function defaultStorage(): StorageLike | null {
  try {
    return typeof window === "undefined" ? null : window.localStorage;
  } catch {
    return null;
  }
}

const NUMERIC_KEYS = new Set(["latitude", "longitude"]);

function sameKind(key: string, initial: unknown, value: unknown): boolean {
  if (Array.isArray(initial)) return Array.isArray(value);
  if (NUMERIC_KEYS.has(key)) return value === null || typeof value === "number";
  if (initial === null) return value === null || typeof value === "string";
  return typeof value === typeof initial;
}

/** Reads saved answers, ignoring anything unknown or of the wrong kind. Never throws. */
export function loadAnswers(storage: StorageLike | null = defaultStorage()): QuizAnswers {
  if (!storage) return initialAnswers;
  try {
    const raw = storage.getItem(STORAGE_KEY);
    if (!raw) return initialAnswers;
    const parsed: unknown = JSON.parse(raw);
    if (typeof parsed !== "object" || parsed === null) return initialAnswers;

    const merged: Record<string, unknown> = { ...initialAnswers };
    for (const key of Object.keys(initialAnswers)) {
      const value = (parsed as Record<string, unknown>)[key];
      if (value !== undefined && sameKind(key, (initialAnswers as Record<string, unknown>)[key], value)) {
        merged[key] = value;
      }
    }
    return merged as QuizAnswers;
  } catch {
    return initialAnswers;
  }
}

export function saveAnswers(answers: QuizAnswers, storage: StorageLike | null = defaultStorage()): void {
  if (!storage) return;
  try {
    storage.setItem(STORAGE_KEY, JSON.stringify(answers));
  } catch {
    // storage is optional: the quiz keeps working from memory
  }
}

export function clearAnswers(storage: StorageLike | null = defaultStorage()): void {
  if (!storage) return;
  try {
    storage.removeItem(STORAGE_KEY);
  } catch {
    // nothing to clean up if storage is unavailable
  }
}
```

- [ ] **Step 10: Implementar `web/lib/quiz/steps.ts`**

```ts
import { CROP_OTHER } from "./options";
import type { Profile, QuizAnswers, WorkStyle } from "./types";

export type StepId =
  | "style"
  | "areas"
  | "identity"
  | "crop"
  | "location"
  | "soil"
  | "concerns"
  | "producers"
  | "account";

const PRODUCER_STEPS: StepId[] = ["style", "areas", "identity", "crop", "location", "soil", "concerns", "account"];
const TECHNICIAN_STEPS: StepId[] = ["style", "identity", "producers", "concerns", "account"];

export function profileOf(workStyle: WorkStyle | null): Profile | null {
  if (workStyle === null) return null;
  return workStyle === "technician" ? "technician" : "producer";
}

export function stepsFor(workStyle: WorkStyle | null): StepId[] {
  const profile = profileOf(workStyle);
  if (profile === null) return ["style"];
  return profile === "technician" ? TECHNICIAN_STEPS : PRODUCER_STEPS;
}

/** DDD + number (10 or 11 digits), with or without the 55 country code. Returns digits starting with 55, or null. */
export function normalizePhone(raw: string): string | null {
  const digits = raw.replace(/\D/g, "");
  if (/^55\d{10,11}$/.test(digits)) return digits;
  if (/^\d{10,11}$/.test(digits)) return `55${digits}`;
  return null;
}

export function cropValue(a: QuizAnswers): string {
  return a.crop_choice === CROP_OTHER ? a.crop_custom.trim() : a.crop_choice.trim();
}

/** Returns the message to show when the step can't advance yet, or null when it's complete. */
export function validateStep(id: StepId, a: QuizAnswers): string | null {
  switch (id) {
    case "style":
      return a.work_style === null ? "Escolha uma das opções." : null;
    case "areas":
      return a.areas_bucket === null ? "Escolha uma das opções." : null;
    case "producers":
      return a.producers_bucket === null ? "Escolha uma das opções." : null;
    case "identity":
      if (a.name.trim().length < 2) return "Digite seu nome.";
      if (normalizePhone(a.phone) === null) return "Digite o DDD e o número, só números. Ex.: 19 99999-9999.";
      if (a.work_style === "technician" && a.organization.trim() === "") {
        return "Digite o nome da sua organização.";
      }
      return null;
    case "crop":
      if (cropValue(a) === "") return "Escolha ou digite a cultura.";
      if (a.crop_stage === null) return "Escolha em que fase está.";
      return null;
    case "location":
      if (a.latitude === null || a.longitude === null) return "Marque sua terra no mapa.";
      if (a.city.trim() === "" || !/^[A-Z]{2}$/.test(a.state)) return "Escolha a cidade na busca.";
      return null;
    case "soil":
      return a.soil_type === null ? "Escolha o tipo de solo." : null;
    case "concerns":
    case "account":
      return null;
  }
}

/** Index of the first step that still has an error; the last step when all are complete. */
export function firstIncompleteIndex(steps: StepId[], a: QuizAnswers): number {
  const index = steps.findIndex((id) => validateStep(id, a) !== null);
  return index === -1 ? steps.length - 1 : index;
}

const FIELD_TO_STEP: Record<string, StepId> = {
  profile: "style",
  onboarding: "style",
  name: "identity",
  whatsapp_phone: "identity",
  organization: "identity",
  email: "account",
  password: "account",
  concerns: "concerns",
  producer: "location",
  "producer.city": "location",
  "producer.state": "location",
  property: "location",
  "property.latitude": "location",
  "property.longitude": "location",
  "property.crop": "crop",
  "property.crop_stage": "crop",
  "property.soil_type": "soil",
};

export function stepForField(field: string): StepId | null {
  return FIELD_TO_STEP[field] ?? null;
}
```

- [ ] **Step 11: Implementar `web/lib/quiz/account.ts`**

```ts
export type AccountErrors = { email?: string; password?: string };

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const MIN_PASSWORD = 8;
const MAX_PASSWORD_BYTES = 72;

export function validateAccount(email: string, password: string): AccountErrors {
  const errors: AccountErrors = {};
  if (!EMAIL_PATTERN.test(email.trim())) errors.email = "Digite um e-mail válido.";
  if (password.length < MIN_PASSWORD) {
    errors.password = "A senha precisa ter pelo menos 8 caracteres.";
  } else if (new TextEncoder().encode(password).length > MAX_PASSWORD_BYTES) {
    errors.password = "A senha pode ter no máximo 72 caracteres.";
  }
  return errors;
}
```

- [ ] **Step 12: Implementar `web/lib/quiz/payload.ts`**

```ts
import { cropValue, normalizePhone, profileOf } from "./steps";
import type { Concern, CropStage, Profile, QuizAnswers, SoilType } from "./types";

export type SignupPayload = {
  profile: Profile;
  name: string;
  whatsapp_phone: string;
  email: string;
  password: string;
  organization: string | null;
  concerns: Concern[];
  onboarding: Record<string, string>;
  producer?: { city: string; state: string };
  property?: {
    latitude: number;
    longitude: number;
    crop: string;
    soil_type: SoilType;
    crop_stage: CropStage;
  };
};

const INCOMPLETE = "incomplete quiz answers";

/** Turns the answers into the POST /auth/signup body; answers of the other profile are ignored. */
export function buildSignupPayload(a: QuizAnswers, password: string): SignupPayload {
  const profile = profileOf(a.work_style);
  const phone = normalizePhone(a.phone);
  if (profile === null || a.work_style === null || phone === null) throw new Error(INCOMPLETE);

  const base = {
    profile,
    name: a.name.trim(),
    whatsapp_phone: phone,
    email: a.email.trim().toLowerCase(),
    password,
    concerns: a.concerns,
  };

  if (profile === "technician") {
    if (a.producers_bucket === null) throw new Error(INCOMPLETE);
    return {
      ...base,
      organization: a.organization.trim(),
      onboarding: { work_style: a.work_style, producers_bucket: a.producers_bucket },
    };
  }

  const crop = cropValue(a);
  if (
    a.areas_bucket === null ||
    a.crop_stage === null ||
    a.soil_type === null ||
    a.latitude === null ||
    a.longitude === null ||
    crop === ""
  ) {
    throw new Error(INCOMPLETE);
  }

  return {
    ...base,
    organization: null,
    onboarding: { work_style: a.work_style, areas_bucket: a.areas_bucket },
    producer: { city: a.city.trim(), state: a.state },
    property: {
      latitude: a.latitude,
      longitude: a.longitude,
      crop,
      soil_type: a.soil_type,
      crop_stage: a.crop_stage,
    },
  };
}
```

- [ ] **Step 13: Rodar e ver passar**

Run: `npx vitest run lib/quiz`
Expected: PASS (4 arquivos de teste).

- [ ] **Step 14: Commit**

```bash
git add web/lib/types.ts web/lib/quiz
git commit -m "feat(web): add quiz steps, state and signup payload libs" -m "Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>"
```

---

### Task 11: Geocodificação e personalização (bibliotecas puras)

**Files:**
- Create: `web/lib/geocode.ts`, `web/lib/personalize.ts`
- Test: `web/lib/geocode.test.ts`, `web/lib/personalize.test.ts`

- [ ] **Step 1: Escrever os testes de `geocode` (falham)**

`web/lib/geocode.test.ts`:
```ts
import { describe, expect, it, vi } from "vitest";
import { parseGeocodingResponse, searchCities } from "./geocode";

const response = {
  results: [
    { name: "Campinas", admin1: "São Paulo", latitude: -22.90556, longitude: -47.06083, country_code: "BR" },
    { name: "Campinas", admin1: "Distrito do Porto", latitude: 41.16667, longitude: -8.65, country_code: "PT" },
    { name: "Lugar Sem Estado", admin1: "Estado Inexistente", latitude: 1, longitude: 2, country_code: "BR" },
  ],
};

describe("parseGeocodingResponse", () => {
  it("keeps only Brazilian results whose state maps to a UF", () => {
    expect(parseGeocodingResponse(response)).toEqual([
      {
        name: "Campinas",
        state: "SP",
        latitude: -22.90556,
        longitude: -47.06083,
        label: "Campinas, SP",
      },
    ]);
  });

  it("returns an empty list for unexpected shapes", () => {
    expect(parseGeocodingResponse(null)).toEqual([]);
    expect(parseGeocodingResponse({})).toEqual([]);
    expect(parseGeocodingResponse({ results: "x" })).toEqual([]);
  });
});

describe("searchCities", () => {
  it("does not call the API for very short queries", async () => {
    const fetchImpl = vi.fn();
    expect(await searchCities("c", fetchImpl as unknown as typeof fetch)).toEqual([]);
    expect(fetchImpl).not.toHaveBeenCalled();
  });

  it("queries Open-Meteo restricted to Brazil and parses the result", async () => {
    const fetchImpl = vi.fn().mockResolvedValue({ ok: true, json: async () => response });

    const cities = await searchCities("Campinas", fetchImpl as unknown as typeof fetch);

    const url = String(fetchImpl.mock.calls[0][0]);
    expect(url).toContain("geocoding-api.open-meteo.com/v1/search");
    expect(url).toContain("name=Campinas");
    expect(url).toContain("countryCode=BR");
    expect(cities).toHaveLength(1);
  });

  it("throws when the API fails", async () => {
    const fetchImpl = vi.fn().mockResolvedValue({ ok: false, status: 500 });
    await expect(searchCities("Campinas", fetchImpl as unknown as typeof fetch)).rejects.toThrow();
  });
});
```

- [ ] **Step 2: Escrever os testes de `personalize` (falham)**

`web/lib/personalize.test.ts`:
```ts
import { describe, expect, it } from "vitest";
import {
  currentHourBR,
  greeting,
  joinList,
  matchesConcern,
  missingAreasNotice,
  sortByPriority,
  welcomeMessage,
} from "./personalize";
import type { Alert } from "./types";

function alert(level: Alert["level"], type: string, resolved = false): Alert {
  return {
    id: `${level}-${type}`,
    property_id: "p",
    level,
    alert_type: type,
    period_start: "2026-10-30T00:00:00Z",
    period_end: "2026-10-31T00:00:00Z",
    triggered_at: "2026-10-30T00:00:00Z",
    resolved_at: resolved ? "2026-10-30T12:00:00Z" : null,
  };
}

describe("matchesConcern", () => {
  it("is true when an active alert has a chosen type", () => {
    expect(matchesConcern([alert("medium", "frost")], ["frost"])).toBe(true);
  });

  it("ignores resolved alerts, other types and empty concerns", () => {
    expect(matchesConcern([alert("medium", "frost", true)], ["frost"])).toBe(false);
    expect(matchesConcern([alert("medium", "heat_wave")], ["frost"])).toBe(false);
    expect(matchesConcern([alert("medium", "frost")], [])).toBe(false);
  });
});

describe("sortByPriority", () => {
  const quiet = { id: "quiet", alerts: [] as Alert[] };
  const mediumFrost = { id: "medium-frost", alerts: [alert("medium", "frost")] };
  const mediumHeat = { id: "medium-heat", alerts: [alert("medium", "heat_wave")] };
  const critical = { id: "critical", alerts: [alert("critical", "heavy_rain")] };

  it("orders by risk level first", () => {
    const sorted = sortByPriority([quiet, mediumHeat, critical], []);
    expect(sorted.map((i) => i.id)).toEqual(["critical", "medium-heat", "quiet"]);
  });

  it("puts chosen risk types ahead within the same level", () => {
    const sorted = sortByPriority([mediumHeat, mediumFrost], ["frost"]);
    expect(sorted.map((i) => i.id)).toEqual(["medium-frost", "medium-heat"]);
  });

  it("does not mutate the input", () => {
    const input = [quiet, critical];
    sortByPriority(input, []);
    expect(input.map((i) => i.id)).toEqual(["quiet", "critical"]);
  });
});

describe("greeting / currentHourBR", () => {
  it("greets by time of day using the first name", () => {
    expect(greeting("José da Silva", 8)).toBe("Bom dia, José");
    expect(greeting("José da Silva", 15)).toBe("Boa tarde, José");
    expect(greeting("José da Silva", 21)).toBe("Boa noite, José");
  });

  it("computes the hour in Brasília time", () => {
    expect(currentHourBR(new Date("2026-10-30T15:00:00Z"))).toBe(12);
    expect(currentHourBR(new Date("2026-10-30T02:30:00Z"))).toBe(23);
  });
});

describe("missingAreasNotice", () => {
  it("warns when fewer areas than announced are registered", () => {
    expect(missingAreasNotice("2-5", 1)).toBe(
      "Você disse ter de 2 a 5 áreas e há 1 cadastrada. Cadastre as outras para acompanhar todas."
    );
    expect(missingAreasNotice("6+", 3)).toBe(
      "Você disse ter mais de 5 áreas e há 3 cadastradas. Cadastre as outras para acompanhar todas."
    );
  });

  it("stays quiet when enough areas exist or the bucket is a single area", () => {
    expect(missingAreasNotice("2-5", 2)).toBeNull();
    expect(missingAreasNotice("1", 0)).toBeNull();
    expect(missingAreasNotice(undefined, 0)).toBeNull();
  });
});

describe("welcomeMessage / joinList", () => {
  it("joins lists in Portuguese", () => {
    expect(joinList([])).toBe("");
    expect(joinList(["geada"])).toBe("geada");
    expect(joinList(["geada", "seca"])).toBe("geada e seca");
    expect(joinList(["geada", "seca", "calor forte"])).toBe("geada, seca e calor forte");
  });

  it("tells producers what they will be alerted about and where", () => {
    expect(welcomeMessage({ profile: "producer", concerns: ["frost", "severe_water_stress"] }, "Campinas")).toBe(
      "Vamos te avisar de geada e seca em Campinas."
    );
    expect(welcomeMessage({ profile: "producer", concerns: [] }, "Campinas")).toBe(
      "Vamos te avisar de qualquer risco do clima em Campinas."
    );
    expect(welcomeMessage({ profile: "producer", concerns: ["frost"] })).toBe(
      "Vamos te avisar de geada na sua terra."
    );
  });

  it("tells technicians what the panel will highlight", () => {
    expect(welcomeMessage({ profile: "technician", concerns: ["heavy_rain"] })).toBe(
      "Vamos destacar chuva forte no seu painel."
    );
    expect(welcomeMessage({ profile: "technician", concerns: [] })).toBe(
      "Vamos acompanhar todos os riscos do clima no seu painel."
    );
  });
});
```

- [ ] **Step 3: Rodar e ver falhar**

Run: `npx vitest run lib/geocode.test.ts lib/personalize.test.ts`
Expected: FAIL (módulos inexistentes).

- [ ] **Step 4: Implementar `web/lib/geocode.ts`**

```ts
export type CityResult = {
  name: string;
  state: string;
  latitude: number;
  longitude: number;
  label: string;
};

// Open-Meteo returns the state name (admin1), not the UF code.
const STATE_UF: Record<string, string> = {
  Acre: "AC",
  Alagoas: "AL",
  Amapá: "AP",
  Amazonas: "AM",
  Bahia: "BA",
  Ceará: "CE",
  "Distrito Federal": "DF",
  "Espírito Santo": "ES",
  Goiás: "GO",
  Maranhão: "MA",
  "Mato Grosso": "MT",
  "Mato Grosso do Sul": "MS",
  "Minas Gerais": "MG",
  Pará: "PA",
  Paraíba: "PB",
  Paraná: "PR",
  Pernambuco: "PE",
  Piauí: "PI",
  "Rio de Janeiro": "RJ",
  "Rio Grande do Norte": "RN",
  "Rio Grande do Sul": "RS",
  Rondônia: "RO",
  Roraima: "RR",
  "Santa Catarina": "SC",
  "São Paulo": "SP",
  Sergipe: "SE",
  Tocantins: "TO",
};

const GEOCODING_URL = "https://geocoding-api.open-meteo.com/v1/search";

export function parseGeocodingResponse(json: unknown): CityResult[] {
  if (typeof json !== "object" || json === null) return [];
  const results = (json as { results?: unknown }).results;
  if (!Array.isArray(results)) return [];

  const cities: CityResult[] = [];
  for (const item of results) {
    if (typeof item !== "object" || item === null) continue;
    const r = item as Record<string, unknown>;
    if (r.country_code !== "BR") continue;
    if (typeof r.name !== "string" || typeof r.admin1 !== "string") continue;
    if (typeof r.latitude !== "number" || typeof r.longitude !== "number") continue;
    const state = STATE_UF[r.admin1];
    if (!state) continue;
    cities.push({
      name: r.name,
      state,
      latitude: r.latitude,
      longitude: r.longitude,
      label: `${r.name}, ${state}`,
    });
  }
  return cities;
}

export async function searchCities(query: string, fetchImpl: typeof fetch = fetch): Promise<CityResult[]> {
  const name = query.trim();
  if (name.length < 2) return [];

  const url = `${GEOCODING_URL}?name=${encodeURIComponent(name)}&count=8&language=pt&countryCode=BR`;
  const res = await fetchImpl(url);
  if (!res.ok) throw new Error(`geocoding failed: ${res.status}`);
  return parseGeocodingResponse(await res.json());
}
```

- [ ] **Step 5: Implementar `web/lib/personalize.ts`**

```ts
import { currentRiskLevel } from "./risk";
import type { Alert, AuthUser } from "./types";

const LEVEL_RANK = { low: 0, medium: 1, critical: 2 } as const;

export const CONCERN_LABELS: Record<string, string> = {
  frost: "Geada",
  severe_water_stress: "Seca",
  heat_wave: "Calor forte",
  heavy_rain: "Chuva forte",
};

/** True when an unresolved alert is of a type the account owner asked to prioritize. */
export function matchesConcern(alerts: Alert[], concerns: string[]): boolean {
  if (concerns.length === 0) return false;
  return alerts.some((a) => a.resolved_at === null && concerns.includes(a.alert_type));
}

/** Highest risk first; within a level, areas with a chosen risk type come first. */
export function sortByPriority<T extends { alerts: Alert[] }>(items: T[], concerns: string[]): T[] {
  const score = (item: T) =>
    LEVEL_RANK[currentRiskLevel(item.alerts)] * 2 + (matchesConcern(item.alerts, concerns) ? 1 : 0);
  return [...items].sort((a, b) => score(b) - score(a));
}

export function currentHourBR(now: Date = new Date()): number {
  const hour = new Intl.DateTimeFormat("en-US", {
    hour: "numeric",
    hourCycle: "h23",
    timeZone: "America/Sao_Paulo",
  }).format(now);
  return Number(hour);
}

export function greeting(name: string, hour: number): string {
  const salutation = hour < 12 ? "Bom dia" : hour < 18 ? "Boa tarde" : "Boa noite";
  return `${salutation}, ${name.trim().split(/\s+/)[0]}`;
}

const AREA_BUCKETS: Record<string, { min: number; label: string }> = {
  "2-5": { min: 2, label: "de 2 a 5" },
  "6+": { min: 6, label: "mais de 5" },
};

/** Reminder for producers who announced more areas than they have registered. */
export function missingAreasNotice(bucket: string | undefined, registered: number): string | null {
  const info = bucket ? AREA_BUCKETS[bucket] : undefined;
  if (!info || registered >= info.min) return null;
  const registeredText = registered === 1 ? "1 cadastrada" : `${registered} cadastradas`;
  return `Você disse ter ${info.label} áreas e há ${registeredText}. Cadastre as outras para acompanhar todas.`;
}

export function joinList(items: string[]): string {
  if (items.length <= 1) return items.join("");
  return `${items.slice(0, -1).join(", ")} e ${items[items.length - 1]}`;
}

export function welcomeMessage(user: Pick<AuthUser, "profile" | "concerns">, city?: string): string {
  const labels = user.concerns.map((c) => CONCERN_LABELS[c]).filter(Boolean).map((l) => l.toLowerCase());

  if (user.profile === "technician") {
    return labels.length > 0
      ? `Vamos destacar ${joinList(labels)} no seu painel.`
      : "Vamos acompanhar todos os riscos do clima no seu painel.";
  }

  const what = labels.length > 0 ? `de ${joinList(labels)}` : "de qualquer risco do clima";
  return city ? `Vamos te avisar ${what} em ${city}.` : `Vamos te avisar ${what} na sua terra.`;
}
```

- [ ] **Step 6: Rodar tudo**

Run: `npm test`
Expected: PASS (testes novos e os existentes `risk`, `forecast`, `validation`).

- [ ] **Step 7: Commit**

```bash
git add web/lib/geocode.ts web/lib/geocode.test.ts web/lib/personalize.ts web/lib/personalize.test.ts
git commit -m "feat(web): add city geocoding and dashboard personalization libs" -m "Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>"
```

---

### Task 12: Servidor — sessão, API autenticada, server actions e `proxy.ts`

Esta tarefa troca a forma de chamar a API (agora só no servidor, com token). Os formulários antigos e a página de nova propriedade mudam junto, para o projeto continuar compilando.

**Files:**
- Create: `web/lib/config.ts`, `web/lib/session.ts`, `web/app/actions/auth.ts`, `web/app/actions/data.ts`, `web/proxy.ts`, `web/app/sair/route.ts`
- Modify (substituir): `web/lib/api.ts`, `web/components/ProducerForm.tsx` (trecho), `web/components/PropertyForm.tsx`, `web/app/properties/new/page.tsx`
- Modify: `web/.env.local`

- [ ] **Step 1: Trocar a variável de ambiente**

Run:
```bash
cd /home/lukas-melo/PROJETOS/agrolang/web
sed -i 's/^NEXT_PUBLIC_API_URL=/API_URL=/' .env.local
grep -n "API_URL" .env.local | sed 's/=.*/=<ok>/'
```
Expected: `1:API_URL=<ok>`. (O arquivo é ignorado pelo git; o valor, por exemplo `http://localhost:8080`, não muda.)

- [ ] **Step 2: Criar `web/lib/config.ts`**

```ts
export const API_URL = process.env.API_URL ?? "http://localhost:8080";
```

- [ ] **Step 3: Criar `web/lib/session.ts`**

```ts
import { cache } from "react";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { API_URL } from "./config";
import type { AuthUser } from "./types";

export const SESSION_COOKIE = "session";

export async function getSessionToken(): Promise<string | null> {
  return (await cookies()).get(SESSION_COOKIE)?.value ?? null;
}

export async function setSessionCookie(token: string, expiresAt: Date): Promise<void> {
  (await cookies()).set(SESSION_COOKIE, token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    expires: expiresAt,
  });
}

export async function clearSessionCookie(): Promise<void> {
  (await cookies()).delete(SESSION_COOKIE);
}

/** The logged-in account, or null when there's no cookie or the API says the session is gone. */
export const getCurrentUser = cache(async (): Promise<AuthUser | null> => {
  const token = await getSessionToken();
  if (!token) return null;

  const res = await fetch(`${API_URL}/auth/me`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (res.status === 401) return null;
  if (!res.ok) throw new Error(`API /auth/me failed: ${res.status}`);
  return (await res.json()) as AuthUser;
});

/** For protected pages: sends stale or missing sessions through /sair, which clears the cookie. */
export async function requireUser(): Promise<AuthUser> {
  const user = await getCurrentUser();
  if (!user) redirect("/sair");
  return user;
}
```

- [ ] **Step 4: Substituir `web/lib/api.ts`** (agora só para o servidor)

```ts
import { redirect } from "next/navigation";
import { API_URL } from "./config";
import { getSessionToken } from "./session";
import type {
  Alert,
  AuthUser,
  NewProducer,
  NewProperty,
  Producer,
  Property,
  WeatherReading,
} from "./types";

export type {
  Alert,
  AuthUser,
  CropStage,
  NewProducer,
  NewProperty,
  Producer,
  Property,
  RiskLevel,
  WeatherReading,
} from "./types";

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public fields?: Record<string, string>
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export type SessionResponse = { token: string; expires_at: string; user: AuthUser };

type RequestOptions = { method?: string; body?: unknown; auth?: boolean };

async function apiRequest<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const auth = opts.auth !== false;
  const headers: Record<string, string> = {};
  if (opts.body !== undefined) headers["Content-Type"] = "application/json";
  if (auth) {
    const token = await getSessionToken();
    if (token) headers.Authorization = `Bearer ${token}`;
  }

  const res = await fetch(`${API_URL}${path}`, {
    method: opts.method ?? "GET",
    headers,
    body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined,
    cache: "no-store",
  });

  if (res.status === 401 && auth) redirect("/sair");

  if (!res.ok) {
    let message = `API ${path} failed: ${res.status}`;
    let fields: Record<string, string> | undefined;
    try {
      const payload = (await res.json()) as { error?: string; fields?: Record<string, string> };
      if (payload.error) message = payload.error;
      fields = payload.fields;
    } catch {
      // response body wasn't JSON; keep the generic message
    }
    throw new ApiError(message, res.status, fields);
  }

  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export function listProperties(): Promise<Property[]> {
  return apiRequest<Property[]>("/properties");
}

export function listProducers(): Promise<Producer[]> {
  return apiRequest<Producer[]>("/producers");
}

export function createProducer(producer: NewProducer): Promise<Producer> {
  return apiRequest<Producer>("/producers", { method: "POST", body: producer });
}

export function createProperty(property: NewProperty): Promise<Property> {
  return apiRequest<Property>("/properties", { method: "POST", body: property });
}

export function getPropertyAlerts(propertyId: string): Promise<Alert[]> {
  return apiRequest<Alert[]>(`/properties/${propertyId}/alerts`);
}

export function getPropertyForecast(propertyId: string): Promise<WeatherReading[]> {
  return apiRequest<WeatherReading[]>(`/properties/${propertyId}/forecast`);
}

export function listRecentAlerts(): Promise<Alert[]> {
  return apiRequest<Alert[]>("/alerts");
}

export function signup(payload: unknown): Promise<SessionResponse> {
  return apiRequest<SessionResponse>("/auth/signup", { method: "POST", body: payload, auth: false });
}

export function login(email: string, password: string): Promise<SessionResponse> {
  return apiRequest<SessionResponse>("/auth/login", {
    method: "POST",
    body: { email, password },
    auth: false,
  });
}

export function logout(): Promise<void> {
  return apiRequest<void>("/auth/logout", { method: "POST" });
}
```

- [ ] **Step 5: Criar `web/app/actions/auth.ts`**

```ts
"use server";

import { redirect } from "next/navigation";
import { ApiError, login, logout, signup } from "@/lib/api";
import { clearSessionCookie, setSessionCookie } from "@/lib/session";

export type AuthFormState =
  | { error?: string; fields?: Record<string, string>; emailTaken?: boolean }
  | undefined;

const NETWORK_ERROR = "Sem conexão com o servidor. Tente de novo.";

export async function signupAction(_prev: AuthFormState, formData: FormData): Promise<AuthFormState> {
  const raw = formData.get("payload");
  if (typeof raw !== "string" || raw === "") {
    return { error: "Não foi possível enviar o cadastro. Tente de novo." };
  }

  let payload: unknown;
  try {
    payload = JSON.parse(raw);
  } catch {
    return { error: "Não foi possível enviar o cadastro. Tente de novo." };
  }

  let session;
  try {
    session = await signup(payload);
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 409) {
        return { emailTaken: true, error: "Este e-mail já tem cadastro.", fields: err.fields };
      }
      if (err.status === 400) {
        return { error: "Revise as respostas abaixo.", fields: err.fields };
      }
      return { error: "Não foi possível criar a conta agora. Tente de novo." };
    }
    return { error: NETWORK_ERROR };
  }

  await setSessionCookie(session.token, new Date(session.expires_at));
  redirect("/bem-vindo");
}

export async function loginAction(_prev: AuthFormState, formData: FormData): Promise<AuthFormState> {
  const email = String(formData.get("email") ?? "");
  const password = String(formData.get("password") ?? "");
  if (email.trim() === "" || password === "") {
    return { error: "Digite o e-mail e a senha." };
  }

  let session;
  try {
    session = await login(email, password);
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 401) return { error: "E-mail ou senha incorretos." };
      return { error: "Não foi possível entrar agora. Tente de novo." };
    }
    return { error: NETWORK_ERROR };
  }

  await setSessionCookie(session.token, new Date(session.expires_at));
  redirect("/");
}

export async function logoutAction(): Promise<void> {
  try {
    await logout();
  } catch (err) {
    // an already-expired session is fine; anything else is not worth blocking the exit
    if (!(err instanceof ApiError)) throw err;
  }
  await clearSessionCookie();
  redirect("/entrar");
}
```

> Nota: `logout()` chama `apiRequest` com `auth` ligado; se a sessão já expirou o Go responde 401 e `apiRequest` faz `redirect("/sair")`, que limpa o cookie e leva a `/entrar`, o mesmo resultado final.

- [ ] **Step 6: Criar `web/app/actions/data.ts`**

```ts
"use server";

import { ApiError, createProducer, createProperty } from "@/lib/api";
import type { NewProducer, NewProperty, Producer, Property } from "@/lib/types";

export type ActionResult<T> = { ok: true; data: T } | { ok: false; error: string };

export async function createProducerAction(values: NewProducer): Promise<ActionResult<Producer>> {
  try {
    return { ok: true, data: await createProducer(values) };
  } catch (err) {
    if (err instanceof ApiError) return { ok: false, error: err.message };
    throw err;
  }
}

export async function createPropertyAction(values: NewProperty): Promise<ActionResult<Property>> {
  try {
    return { ok: true, data: await createProperty(values) };
  } catch (err) {
    if (err instanceof ApiError) return { ok: false, error: err.message };
    throw err;
  }
}
```

- [ ] **Step 7: Criar `web/proxy.ts`**

```ts
import { NextResponse, type NextRequest } from "next/server";

const SESSION_COOKIE = "session";
const PUBLIC_PATHS = ["/entrar", "/cadastro", "/sair"];
const GUEST_ONLY_PATHS = ["/entrar", "/cadastro"];

const matches = (pathname: string, paths: string[]) =>
  paths.some((p) => pathname === p || pathname.startsWith(`${p}/`));

// Optimistic check only: the real validation is requireUser() on the pages.
export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const hasSession = request.cookies.has(SESSION_COOKIE);

  if (!hasSession && !matches(pathname, PUBLIC_PATHS)) {
    return NextResponse.redirect(new URL("/entrar", request.url));
  }
  if (hasSession && matches(pathname, GUEST_ONLY_PATHS)) {
    return NextResponse.redirect(new URL("/", request.url));
  }
  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico|.*\\.(?:png|jpg|jpeg|svg|ico|webp)$).*)"],
};
```

- [ ] **Step 8: Criar `web/app/sair/route.ts`**

```ts
import { NextResponse } from "next/server";
import { clearSessionCookie } from "@/lib/session";

// Landing point for stale sessions: drops the cookie so /entrar doesn't bounce back.
export async function GET(request: Request) {
  await clearSessionCookie();
  return NextResponse.redirect(new URL("/entrar", request.url));
}
```

- [ ] **Step 9: Migrar `ProducerForm` para a server action**

Em `web/components/ProducerForm.tsx`:

1. Trocar a linha de import
```tsx
import { createProducer, ApiError } from "@/lib/api";
```
por
```tsx
import { createProducerAction } from "@/app/actions/data";
```

2. Trocar o bloco `try/catch/finally` de `handleSubmit`
```tsx
    try {
      const created = await createProducer(values);
      setCreatedId(created.id);
    } catch (err) {
      setSubmitError(
        err instanceof ApiError ? err.message : "Não foi possível cadastrar o produtor."
      );
    } finally {
      setSubmitting(false);
    }
```
por
```tsx
    try {
      const result = await createProducerAction(values);
      if (result.ok) setCreatedId(result.data.id);
      else setSubmitError(result.error);
    } catch {
      setSubmitError("Não foi possível cadastrar o produtor.");
    } finally {
      setSubmitting(false);
    }
```

- [ ] **Step 10: Substituir `web/components/PropertyForm.tsx`** (produtores agora vêm por props, vindos do servidor)

```tsx
"use client";

import { useState, type FormEvent } from "react";
import Link from "next/link";
import { createPropertyAction } from "@/app/actions/data";
import type { CropStage, Producer } from "@/lib/types";
import { validateProperty, type PropertyFormValues, type PropertyFormErrors } from "@/lib/validation";
import { FieldShell, fieldClass } from "./FieldShell";

const CROP_STAGE_OPTIONS: { value: CropStage; label: string }[] = [
  { value: "germination", label: "Germinação" },
  { value: "flowering", label: "Floração" },
  { value: "harvest", label: "Colheita" },
];

export default function PropertyForm({
  initialProducerId,
  producers,
}: {
  initialProducerId?: string;
  producers: Producer[];
}) {
  const [values, setValues] = useState<PropertyFormValues>({
    producer_id: initialProducerId ?? "",
    latitude: "",
    longitude: "",
    crop: "",
    soil_type: "",
    crop_stage: "",
  });
  const [errors, setErrors] = useState<PropertyFormErrors>({});
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [created, setCreated] = useState<string | null>(null);

  function update<K extends keyof PropertyFormValues>(key: K, value: PropertyFormValues[K]) {
    setValues((v) => ({ ...v, [key]: value }));
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const validation = validateProperty(values);
    setErrors(validation);
    if (Object.keys(validation).length > 0) return;

    setSubmitting(true);
    setSubmitError(null);
    try {
      const result = await createPropertyAction({
        producer_id: values.producer_id,
        latitude: Number(values.latitude),
        longitude: Number(values.longitude),
        crop: values.crop,
        soil_type: values.soil_type,
        crop_stage: values.crop_stage as CropStage,
      });
      if (result.ok) setCreated(result.data.id);
      else setSubmitError(result.error);
    } catch {
      setSubmitError("Não foi possível cadastrar a propriedade.");
    } finally {
      setSubmitting(false);
    }
  }

  if (created) {
    return (
      <div className="rise-in border-2 border-risk-low bg-risk-low-bg px-5 py-6 font-mono text-sm text-risk-low">
        <p className="font-display text-lg italic text-ink">Propriedade cadastrada.</p>
        <p className="mt-2 text-ink-soft">
          O talhão já entra na próxima varredura de risco climático.
        </p>
        <div className="mt-4 flex flex-wrap gap-3">
          <Link
            href={`/properties/${created}`}
            className="border-2 border-ink bg-accent px-4 py-2 font-semibold uppercase tracking-wide text-paper hover:bg-ink"
          >
            Ver ficha do talhão
          </Link>
          <Link href="/" className="border-2 border-ink px-4 py-2 font-semibold uppercase tracking-wide text-ink">
            Voltar ao mapa
          </Link>
        </div>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-5 rise-in">
      <FieldShell label="Produtor" htmlFor="producer_id" error={errors.producer_id}>
        <select
          id="producer_id"
          className={fieldClass}
          value={values.producer_id}
          onChange={(e) => update("producer_id", e.target.value)}
        >
          <option value="">Selecione um produtor</option>
          {producers.map((p) => (
            <option key={p.id} value={p.id}>
              {p.name} &middot; {p.city}/{p.state}
            </option>
          ))}
        </select>
        {producers.length === 0 && (
          <p className="font-mono text-xs text-ink-soft">
            Nenhum produtor cadastrado ainda.{" "}
            <Link href="/producers/new" className="underline text-accent">
              Cadastrar o primeiro
            </Link>
          </p>
        )}
      </FieldShell>

      <div className="grid grid-cols-2 gap-4">
        <FieldShell label="Latitude" htmlFor="latitude" error={errors.latitude}>
          <input
            id="latitude"
            className={`${fieldClass} tabular`}
            value={values.latitude}
            onChange={(e) => update("latitude", e.target.value)}
            placeholder="-22.9099"
            inputMode="decimal"
          />
        </FieldShell>

        <FieldShell label="Longitude" htmlFor="longitude" error={errors.longitude}>
          <input
            id="longitude"
            className={`${fieldClass} tabular`}
            value={values.longitude}
            onChange={(e) => update("longitude", e.target.value)}
            placeholder="-47.0626"
            inputMode="decimal"
          />
        </FieldShell>
      </div>

      <FieldShell label="Cultura" htmlFor="crop" error={errors.crop}>
        <input
          id="crop"
          className={fieldClass}
          value={values.crop}
          onChange={(e) => update("crop", e.target.value)}
          placeholder="Ex: Milho, Tomate, Café"
        />
      </FieldShell>

      <div className="grid grid-cols-2 gap-4">
        <FieldShell label="Tipo de solo" htmlFor="soil_type" error={errors.soil_type}>
          <input
            id="soil_type"
            className={fieldClass}
            value={values.soil_type}
            onChange={(e) => update("soil_type", e.target.value)}
            placeholder="Ex: Argiloso"
          />
        </FieldShell>

        <FieldShell label="Fase do cultivo" htmlFor="crop_stage" error={errors.crop_stage}>
          <select
            id="crop_stage"
            className={fieldClass}
            value={values.crop_stage}
            onChange={(e) => update("crop_stage", e.target.value)}
          >
            <option value="">Selecione</option>
            {CROP_STAGE_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </FieldShell>
      </div>

      {submitError && (
        <p className="border-2 border-risk-critical bg-risk-critical-bg px-3 py-2 font-mono text-sm text-risk-critical">
          {submitError}
        </p>
      )}

      <button
        type="submit"
        disabled={submitting}
        className="w-fit border-2 border-ink bg-accent px-5 py-2.5 font-mono text-sm font-semibold uppercase tracking-wide text-paper transition hover:bg-ink disabled:opacity-50"
      >
        {submitting ? "Cadastrando..." : "Cadastrar propriedade"}
      </button>
    </form>
  );
}
```

- [ ] **Step 11: Atualizar `web/app/properties/new/page.tsx`** (busca os produtores no servidor)

```tsx
import Link from "next/link";
import PropertyForm from "@/components/PropertyForm";
import { listProducers } from "@/lib/api";

export default async function NewPropertyPage({
  searchParams,
}: {
  searchParams: Promise<{ producer_id?: string }>;
}) {
  const { producer_id } = await searchParams;
  const producers = await listProducers();

  return (
    <div className="mx-auto flex w-full max-w-xl flex-1 flex-col gap-6 p-5 sm:p-8">
      <div className="rise-in">
        <Link
          href="/"
          className="inline-flex items-center gap-1.5 font-mono text-xs font-semibold uppercase tracking-wide text-accent"
        >
          &larr; Voltar ao mapa
        </Link>
        <p className="mt-2 font-mono text-[11px] uppercase tracking-[0.25em] text-ink-soft">
          Novo cadastro &middot; RF-01
        </p>
        <h1 className="font-display text-2xl italic text-ink">Cadastrar propriedade</h1>
      </div>

      <PropertyForm initialProducerId={producer_id} producers={producers} />
    </div>
  );
}
```

- [ ] **Step 12: Conferir tipos, testes e lint**

`lib/api.ts` agora importa `next/headers` (via `session.ts`), então nenhum arquivo de cliente nem teste pode importar **valores** dele (só `import type`). Conferir:

```bash
cd /home/lukas-melo/PROJETOS/agrolang/web
grep -rn 'from "./api"\|from "@/lib/api"' lib components app | grep -v "import type"
```
Expected: só aparecem arquivos de servidor (`app/page.tsx`, `app/bem-vindo/page.tsx`, `app/properties/**/page.tsx`, `app/actions/*.ts`). Se aparecer algo em `components/*` ou `lib/*` (exceto `lib/api.ts`), trocar por `import type { ... } from "@/lib/types"` (ou `"./types"`).

Run:
```bash
cd /home/lukas-melo/PROJETOS/agrolang/web
npx tsc --noEmit
npm test
npm run lint
```
Expected: `tsc` sem erros; `npm test` PASS; lint sem erros. (`tsc` pode apontar `app/page.tsx` ainda usando só o que `api.ts` exporta; ele continua válido porque as funções `listProperties` e `getPropertyAlerts` existem.)

- [ ] **Step 13: Commit**

```bash
git add web/lib web/app web/components web/proxy.ts
git commit -m "feat(web): authenticate API calls on the server with a session cookie" -m "Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>"
```

---

### Task 13: Tela de login (`/entrar`)

**Files:**
- Create: `web/components/LoginForm.tsx`, `web/app/entrar/page.tsx`

- [ ] **Step 1: Criar `web/components/LoginForm.tsx`**

```tsx
"use client";

import { useActionState } from "react";
import Link from "next/link";
import { loginAction } from "@/app/actions/auth";
import { FieldShell, fieldClass } from "./FieldShell";

const bigField = `${fieldClass} min-h-12 text-base`;

export default function LoginForm() {
  const [state, formAction, pending] = useActionState(loginAction, undefined);

  return (
    <form action={formAction} className="flex flex-col gap-5 rise-in">
      <FieldShell label="E-mail" htmlFor="email">
        <input
          id="email"
          name="email"
          type="email"
          inputMode="email"
          autoComplete="email"
          className={bigField}
          placeholder="voce@exemplo.com"
          required
        />
      </FieldShell>

      <FieldShell label="Senha" htmlFor="password">
        <input
          id="password"
          name="password"
          type="password"
          autoComplete="current-password"
          className={bigField}
          required
        />
      </FieldShell>

      {state?.error && (
        <p
          role="alert"
          className="border-2 border-risk-critical bg-risk-critical-bg px-3 py-2 font-mono text-sm text-risk-critical"
        >
          {state.error}
        </p>
      )}

      <button
        type="submit"
        disabled={pending}
        className="min-h-12 border-2 border-ink bg-accent px-5 font-mono text-sm font-semibold uppercase tracking-wide text-paper transition hover:bg-ink disabled:opacity-50"
      >
        {pending ? "Entrando..." : "Entrar"}
      </button>

      <p className="font-mono text-xs text-ink-soft">
        Ainda não tem conta?{" "}
        <Link href="/cadastro" className="text-accent underline">
          Fazer o cadastro
        </Link>
      </p>
    </form>
  );
}
```

- [ ] **Step 2: Criar `web/app/entrar/page.tsx`**

```tsx
import type { Metadata } from "next";
import LoginForm from "@/components/LoginForm";

export const metadata: Metadata = { title: "Entrar · AgroResiliente" };

export default function LoginPage() {
  return (
    <div className="mx-auto flex w-full max-w-md flex-1 flex-col justify-center gap-6 p-5 sm:p-8">
      <div className="rise-in">
        <p className="mb-1 font-mono text-[11px] uppercase tracking-[0.25em] text-ink-soft">
          Estação de campo
        </p>
        <h1 className="font-display text-3xl italic text-ink">Entrar</h1>
      </div>

      <LoginForm />
    </div>
  );
}
```

- [ ] **Step 3: Conferir tipos**

Run: `npx tsc --noEmit`
Expected: sem erros.

- [ ] **Step 4: Commit**

```bash
git add web/components/LoginForm.tsx web/app/entrar
git commit -m "feat(web): add login page" -m "Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>"
```

---

### Task 14: Interface do quiz (`/cadastro`) e boas-vindas (`/bem-vindo`)

**Files:**
- Create: `web/components/quiz/{styles.ts,QuizShell.tsx,ChoiceList.tsx,IdentityStep.tsx,CropStep.tsx,CitySearch.tsx,LocationStep.tsx,LocationMap.tsx,LocationMapClient.tsx,AccountStep.tsx,QuizFlow.tsx}`
- Create: `web/app/cadastro/page.tsx`, `web/app/bem-vindo/page.tsx`, `web/components/ClearQuiz.tsx`

- [ ] **Step 1: Criar `web/components/quiz/styles.ts`**

```ts
import { fieldClass } from "../FieldShell";

export const bigFieldClass = `${fieldClass} min-h-12 text-base`;

export function optionClass(selected: boolean): string {
  return [
    "flex min-h-14 w-full flex-col justify-center border-2 px-4 py-3 text-left font-mono text-sm transition active:translate-y-px",
    selected
      ? "border-ink bg-ink text-paper shadow-[var(--shadow-hard)]"
      : "border-ink bg-card text-ink hover:shadow-[var(--shadow-hard)]",
  ].join(" ");
}

export const primaryButtonClass =
  "min-h-12 w-full border-2 border-ink bg-accent px-5 font-mono text-sm font-semibold uppercase tracking-wide text-paper transition hover:bg-ink disabled:opacity-50";
```

- [ ] **Step 2: Criar `web/components/quiz/QuizShell.tsx`**

```tsx
import type { ReactNode } from "react";

export default function QuizShell({
  index,
  total,
  title,
  hint,
  onBack,
  footer,
  children,
}: {
  index: number;
  /** null while the path (and so the step count) isn't known yet. */
  total: number | null;
  title: string;
  hint?: string;
  onBack?: () => void;
  footer?: ReactNode;
  children: ReactNode;
}) {
  const progress = total ? ((index + 1) / total) * 100 : 8;

  return (
    <div className="mx-auto flex min-h-dvh w-full max-w-md flex-1 flex-col">
      <header className="flex flex-col gap-3 px-5 pb-2 pt-5">
        <div className="flex items-center justify-between">
          {onBack ? (
            <button
              type="button"
              onClick={onBack}
              className="-ml-2 min-h-11 px-2 font-mono text-xs font-semibold uppercase tracking-wide text-accent"
            >
              &larr; Voltar
            </button>
          ) : (
            <span />
          )}
          <span className="font-mono text-[11px] uppercase tracking-[0.2em] text-ink-soft">
            {total ? `Passo ${index + 1} de ${total}` : "Passo 1"}
          </span>
        </div>
        <div
          role="progressbar"
          aria-valuemin={0}
          aria-valuemax={100}
          aria-valuenow={Math.round(progress)}
          className="h-2 border-2 border-ink bg-card"
        >
          <div className="h-full bg-accent transition-all" style={{ width: `${progress}%` }} />
        </div>
      </header>

      <main className="flex flex-1 flex-col gap-6 px-5 py-4">
        <div className="rise-in">
          <h1 className="font-display text-2xl italic leading-tight text-ink">{title}</h1>
          {hint && <p className="mt-1 text-sm text-ink-soft">{hint}</p>}
        </div>
        <div className="flex flex-col gap-4">{children}</div>
      </main>

      {footer && (
        <footer className="sticky bottom-0 border-t-2 border-ink bg-paper px-5 py-4">{footer}</footer>
      )}
    </div>
  );
}
```

- [ ] **Step 3: Criar `web/components/quiz/ChoiceList.tsx`**

```tsx
"use client";

import type { Option } from "@/lib/quiz/options";
import { optionClass } from "./styles";

function OptionBody({ label, hint, selected }: { label: string; hint?: string; selected: boolean }) {
  return (
    <>
      <span className="font-semibold">{label}</span>
      {hint && <span className={`mt-0.5 text-xs ${selected ? "text-paper/80" : "text-ink-soft"}`}>{hint}</span>}
    </>
  );
}

export function ChoiceList<T extends string>({
  options,
  value,
  onChange,
}: {
  options: Option<T>[];
  value: T | null;
  onChange: (value: T) => void;
}) {
  return (
    <div role="radiogroup" className="flex flex-col gap-3">
      {options.map((o) => {
        const selected = o.value === value;
        return (
          <button
            key={o.value}
            type="button"
            role="radio"
            aria-checked={selected}
            onClick={() => onChange(o.value)}
            className={optionClass(selected)}
          >
            <OptionBody label={o.label} hint={o.hint} selected={selected} />
          </button>
        );
      })}
    </div>
  );
}

export function MultiChoiceList<T extends string>({
  options,
  values,
  onChange,
}: {
  options: Option<T>[];
  values: T[];
  onChange: (values: T[]) => void;
}) {
  function toggle(value: T) {
    onChange(values.includes(value) ? values.filter((v) => v !== value) : [...values, value]);
  }

  return (
    <div role="group" className="flex flex-col gap-3">
      {options.map((o) => {
        const selected = values.includes(o.value);
        return (
          <button
            key={o.value}
            type="button"
            aria-pressed={selected}
            onClick={() => toggle(o.value)}
            className={optionClass(selected)}
          >
            <OptionBody label={o.label} hint={o.hint} selected={selected} />
          </button>
        );
      })}
    </div>
  );
}
```

- [ ] **Step 4: Criar `web/components/quiz/IdentityStep.tsx`**

```tsx
"use client";

import type { QuizAnswers } from "@/lib/quiz/types";
import { FieldShell } from "../FieldShell";
import { bigFieldClass } from "./styles";

export default function IdentityStep({
  answers,
  technician,
  onChange,
}: {
  answers: QuizAnswers;
  technician: boolean;
  onChange: (patch: Partial<QuizAnswers>) => void;
}) {
  return (
    <>
      <FieldShell label="Seu nome" htmlFor="name">
        <input
          id="name"
          className={bigFieldClass}
          value={answers.name}
          onChange={(e) => onChange({ name: e.target.value })}
          autoComplete="name"
          placeholder="Ex: José da Silva"
        />
      </FieldShell>

      <FieldShell label="WhatsApp (com DDD)" htmlFor="phone">
        <input
          id="phone"
          className={bigFieldClass}
          value={answers.phone}
          onChange={(e) => onChange({ phone: e.target.value })}
          inputMode="numeric"
          autoComplete="tel-national"
          placeholder="19 99999-9999"
        />
      </FieldShell>

      {technician && (
        <FieldShell label="Organização" htmlFor="organization">
          <input
            id="organization"
            className={bigFieldClass}
            value={answers.organization}
            onChange={(e) => onChange({ organization: e.target.value })}
            autoComplete="organization"
            placeholder="Ex: Cooperativa Campo Verde"
          />
        </FieldShell>
      )}
    </>
  );
}
```

- [ ] **Step 5: Criar `web/components/quiz/CropStep.tsx`**

```tsx
"use client";

import { CROP_OPTIONS, CROP_OTHER, STAGE_OPTIONS } from "@/lib/quiz/options";
import type { QuizAnswers } from "@/lib/quiz/types";
import { FieldShell } from "../FieldShell";
import { ChoiceList } from "./ChoiceList";
import { bigFieldClass, optionClass } from "./styles";

export default function CropStep({
  answers,
  onChange,
}: {
  answers: QuizAnswers;
  onChange: (patch: Partial<QuizAnswers>) => void;
}) {
  const choices = [...CROP_OPTIONS, CROP_OTHER] as string[];

  return (
    <>
      <div role="radiogroup" aria-label="Cultura" className="grid grid-cols-2 gap-3">
        {choices.map((choice) => {
          const selected = answers.crop_choice === choice;
          return (
            <button
              key={choice}
              type="button"
              role="radio"
              aria-checked={selected}
              onClick={() => onChange({ crop_choice: choice })}
              className={optionClass(selected)}
            >
              <span className="font-semibold">{choice === CROP_OTHER ? "Outra" : choice}</span>
            </button>
          );
        })}
      </div>

      {answers.crop_choice === CROP_OTHER && (
        <FieldShell label="Qual cultura?" htmlFor="crop_custom">
          <input
            id="crop_custom"
            className={bigFieldClass}
            value={answers.crop_custom}
            onChange={(e) => onChange({ crop_custom: e.target.value })}
            placeholder="Ex: Mandioca"
            autoFocus
          />
        </FieldShell>
      )}

      <p className="mt-2 font-mono text-[11px] font-semibold uppercase tracking-wide text-ink-soft">
        Em que fase está?
      </p>
      <ChoiceList
        options={STAGE_OPTIONS}
        value={answers.crop_stage}
        onChange={(crop_stage) => onChange({ crop_stage })}
      />
    </>
  );
}
```

- [ ] **Step 6: Criar `web/components/quiz/CitySearch.tsx`**

```tsx
"use client";

import { useEffect, useRef, useState } from "react";
import { searchCities, type CityResult } from "@/lib/geocode";
import { bigFieldClass } from "./styles";

export default function CitySearch({
  initialLabel,
  onPick,
}: {
  initialLabel: string;
  onPick: (city: CityResult) => void;
}) {
  const [query, setQuery] = useState(initialLabel);
  const [picked, setPicked] = useState(initialLabel !== "");
  const [results, setResults] = useState<CityResult[]>([]);
  const [status, setStatus] = useState<"idle" | "loading" | "error">("idle");
  const requestId = useRef(0);

  useEffect(() => {
    const q = query.trim();
    if (picked || q.length < 2) {
      requestId.current += 1; // drop any in-flight answer
      return;
    }

    const id = ++requestId.current;
    const timer = setTimeout(async () => {
      setStatus("loading");
      try {
        const found = await searchCities(q);
        if (id !== requestId.current) return;
        setResults(found);
        setStatus("idle");
      } catch {
        if (id !== requestId.current) return;
        setResults([]);
        setStatus("error");
      }
    }, 300);
    return () => clearTimeout(timer);
  }, [query, picked]);

  function choose(city: CityResult) {
    setQuery(city.label);
    setPicked(true);
    setResults([]);
    setStatus("idle");
    onPick(city);
  }

  return (
    <div className="flex flex-col gap-2">
      <label
        htmlFor="city-search"
        className="font-mono text-[11px] font-semibold uppercase tracking-wide text-ink-soft"
      >
        Cidade
      </label>
      <input
        id="city-search"
        className={bigFieldClass}
        value={query}
        onChange={(e) => {
          setQuery(e.target.value);
          setPicked(false);
        }}
        autoComplete="off"
        placeholder="Digite o nome da cidade"
      />

      {status === "loading" && <p className="font-mono text-xs text-ink-soft">Buscando...</p>}
      {status === "error" && (
        <p className="font-mono text-xs text-risk-critical">
          Não deu para buscar agora. Use &ldquo;Digitar à mão&rdquo; abaixo.
        </p>
      )}
      {!picked && status === "idle" && query.trim().length >= 2 && results.length === 0 && (
        <p className="font-mono text-xs text-ink-soft">Nenhuma cidade encontrada.</p>
      )}

      {results.length > 0 && (
        <ul className="flex flex-col border-2 border-ink bg-card">
          {results.map((city) => (
            <li key={`${city.label}-${city.latitude}`}>
              <button
                type="button"
                onClick={() => choose(city)}
                className="min-h-12 w-full border-b border-line px-4 py-2 text-left font-mono text-sm text-ink last:border-b-0 hover:bg-paper-deep"
              >
                {city.label}
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
```

- [ ] **Step 7: Criar `web/components/quiz/LocationMap.tsx`** (cliente apenas, Leaflet)

```tsx
"use client";

import { useEffect } from "react";
import L from "leaflet";
import { MapContainer, Marker, TileLayer, useMap, useMapEvents } from "react-leaflet";
import "leaflet/dist/leaflet.css";

const BRAZIL_CENTER: [number, number] = [-14.235, -51.9253];

const pinIcon = L.divIcon({
  className: "",
  html: '<div style="width:22px;height:22px;border:3px solid #2a2420;background:#8a4a2f;border-radius:50% 50% 50% 0;transform:rotate(-45deg);box-shadow:3px 3px 0 #2a2420"></div>',
  iconSize: [22, 22],
  iconAnchor: [11, 22],
});

function Recenter({ pin }: { pin: [number, number] | null }) {
  const map = useMap();
  const lat = pin?.[0];
  const lon = pin?.[1];
  useEffect(() => {
    if (lat !== undefined && lon !== undefined) map.setView([lat, lon], Math.max(map.getZoom(), 11));
  }, [lat, lon, map]);
  return null;
}

function ClickToPlace({ onChange }: { onChange: (lat: number, lon: number) => void }) {
  useMapEvents({
    click(e) {
      onChange(e.latlng.lat, e.latlng.lng);
    },
  });
  return null;
}

export default function LocationMap({
  latitude,
  longitude,
  onChange,
}: {
  latitude: number | null;
  longitude: number | null;
  onChange: (lat: number, lon: number) => void;
}) {
  const pin: [number, number] | null = latitude !== null && longitude !== null ? [latitude, longitude] : null;

  return (
    <MapContainer center={pin ?? BRAZIL_CENTER} zoom={pin ? 12 : 4} className="h-full w-full">
      <TileLayer
        attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
      />
      <Recenter pin={pin} />
      <ClickToPlace onChange={onChange} />
      {pin && (
        <Marker
          position={pin}
          icon={pinIcon}
          draggable
          eventHandlers={{
            dragend(e) {
              const p = (e.target as L.Marker).getLatLng();
              onChange(p.lat, p.lng);
            },
          }}
        />
      )}
    </MapContainer>
  );
}
```

- [ ] **Step 8: Criar `web/components/quiz/LocationMapClient.tsx`**

```tsx
"use client";

import dynamic from "next/dynamic";

// Leaflet touches `window`, so the map is only ever loaded in the browser.
const LocationMapClient = dynamic(() => import("./LocationMap"), {
  ssr: false,
  loading: () => (
    <div className="flex h-full w-full items-center justify-center bg-paper-deep font-mono text-xs text-ink-soft">
      Carregando mapa...
    </div>
  ),
});

export default LocationMapClient;
```

- [ ] **Step 9: Criar `web/components/quiz/LocationStep.tsx`**

```tsx
"use client";

import { useState } from "react";
import type { CityResult } from "@/lib/geocode";
import type { QuizAnswers } from "@/lib/quiz/types";
import { FieldShell } from "../FieldShell";
import CitySearch from "./CitySearch";
import LocationMapClient from "./LocationMapClient";
import { bigFieldClass, optionClass } from "./styles";

const round5 = (n: number) => Math.round(n * 1e5) / 1e5;

export default function LocationStep({
  answers,
  onChange,
}: {
  answers: QuizAnswers;
  onChange: (patch: Partial<QuizAnswers>) => void;
}) {
  const [manual, setManual] = useState(false);
  const [gpsMessage, setGpsMessage] = useState<string | null>(null);

  function pickCity(city: CityResult) {
    onChange({
      city: city.name,
      state: city.state,
      latitude: round5(city.latitude),
      longitude: round5(city.longitude),
    });
  }

  function setPin(lat: number, lon: number) {
    setGpsMessage(null);
    onChange({ latitude: round5(lat), longitude: round5(lon) });
  }

  function useMyLocation() {
    if (!("geolocation" in navigator)) {
      setGpsMessage("Seu aparelho não deixa usar a localização. Arraste o pino no mapa.");
      return;
    }
    navigator.geolocation.getCurrentPosition(
      (pos) => setPin(pos.coords.latitude, pos.coords.longitude),
      () => setGpsMessage("Não deu para pegar a localização. Arraste o pino no mapa."),
      { enableHighAccuracy: true, timeout: 10000 }
    );
  }

  return (
    <>
      <CitySearch
        initialLabel={answers.city && answers.state ? `${answers.city}, ${answers.state}` : ""}
        onPick={pickCity}
      />

      <button
        type="button"
        onClick={() => setManual((m) => !m)}
        className="w-fit font-mono text-xs text-accent underline"
      >
        {manual ? "Voltar para a busca" : "Não achei, digitar à mão"}
      </button>

      {manual && (
        <div className="grid grid-cols-[1fr_5rem] gap-3">
          <FieldShell label="Cidade" htmlFor="manual-city">
            <input
              id="manual-city"
              className={bigFieldClass}
              value={answers.city}
              onChange={(e) => onChange({ city: e.target.value })}
            />
          </FieldShell>
          <FieldShell label="UF" htmlFor="manual-state">
            <input
              id="manual-state"
              className={bigFieldClass}
              value={answers.state}
              maxLength={2}
              onChange={(e) => onChange({ state: e.target.value.toUpperCase() })}
            />
          </FieldShell>
        </div>
      )}

      <div className="h-72 border-2 border-ink shadow-[var(--shadow-hard)]">
        <LocationMapClient latitude={answers.latitude} longitude={answers.longitude} onChange={setPin} />
      </div>
      <p className="font-mono text-xs text-ink-soft">
        Toque no mapa ou arraste o pino até a sua área.
      </p>

      <button type="button" onClick={useMyLocation} className={optionClass(false)}>
        <span className="font-semibold">Usar minha localização</span>
        <span className="mt-0.5 text-xs text-ink-soft">Funciona se você estiver na sua terra agora</span>
      </button>
      {gpsMessage && <p className="font-mono text-xs text-risk-critical">{gpsMessage}</p>}
    </>
  );
}
```

- [ ] **Step 10: Criar `web/components/quiz/AccountStep.tsx`**

```tsx
"use client";

import { useActionState, useState, type FormEvent } from "react";
import Link from "next/link";
import { signupAction } from "@/app/actions/auth";
import { validateAccount } from "@/lib/quiz/account";
import { buildSignupPayload } from "@/lib/quiz/payload";
import { stepForField, type StepId } from "@/lib/quiz/steps";
import type { QuizAnswers } from "@/lib/quiz/types";
import { FieldShell } from "../FieldShell";
import { bigFieldClass, primaryButtonClass } from "./styles";

export default function AccountStep({
  answers,
  onEmailChange,
  onFixStep,
}: {
  answers: QuizAnswers;
  onEmailChange: (email: string) => void;
  onFixStep: (step: StepId) => void;
}) {
  const [state, formAction, pending] = useActionState(signupAction, undefined);
  const [password, setPassword] = useState("");
  const [showErrors, setShowErrors] = useState(false);

  const accountErrors = validateAccount(answers.email, password);
  let payload = "";
  try {
    payload = JSON.stringify(buildSignupPayload(answers, password));
  } catch {
    payload = ""; // earlier steps are incomplete; the guard in QuizFlow normally prevents this
  }
  const invalid = Object.keys(accountErrors).length > 0 || payload === "";

  const fieldMessages = state?.fields ? Object.entries(state.fields) : [];
  const firstBadStep =
    fieldMessages.map(([field]) => stepForField(field)).find((s): s is StepId => s !== null && s !== "account") ??
    null;

  function handleSubmit(e: FormEvent) {
    if (invalid) {
      e.preventDefault();
      setShowErrors(true);
    }
  }

  return (
    <form action={formAction} onSubmit={handleSubmit} className="flex flex-col gap-5">
      <input type="hidden" name="payload" value={payload} />

      <FieldShell label="E-mail" htmlFor="email" error={showErrors ? accountErrors.email : undefined}>
        <input
          id="email"
          type="email"
          inputMode="email"
          autoComplete="email"
          className={bigFieldClass}
          value={answers.email}
          onChange={(e) => onEmailChange(e.target.value)}
          placeholder="voce@exemplo.com"
        />
      </FieldShell>

      <FieldShell
        label="Senha (mínimo 8 caracteres)"
        htmlFor="password"
        error={showErrors ? accountErrors.password : undefined}
      >
        <input
          id="password"
          type="password"
          autoComplete="new-password"
          className={bigFieldClass}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
      </FieldShell>

      {state?.error && (
        <div
          role="alert"
          className="flex flex-col gap-2 border-2 border-risk-critical bg-risk-critical-bg px-3 py-3 font-mono text-sm text-risk-critical"
        >
          <p>{state.error}</p>
          {fieldMessages.length > 0 && (
            <ul className="list-disc pl-5 text-xs">
              {fieldMessages.map(([field, message]) => (
                <li key={field}>{message}</li>
              ))}
            </ul>
          )}
          {state.emailTaken && (
            <Link href="/entrar" className="font-semibold underline">
              Entrar com este e-mail
            </Link>
          )}
          {firstBadStep && (
            <button type="button" onClick={() => onFixStep(firstBadStep)} className="w-fit font-semibold underline">
              Corrigir respostas
            </button>
          )}
        </div>
      )}

      <button type="submit" disabled={pending} className={primaryButtonClass}>
        {pending ? "Criando sua conta..." : "Criar minha conta"}
      </button>
    </form>
  );
}
```

- [ ] **Step 11: Criar `web/components/quiz/QuizFlow.tsx`**

```tsx
"use client";

import { useEffect, useReducer, useState, type ReactNode } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import {
  AREAS_OPTIONS,
  CONCERN_OPTIONS,
  PRODUCERS_OPTIONS,
  SOIL_OPTIONS,
  STYLE_OPTIONS,
} from "@/lib/quiz/options";
import { initialAnswers, loadAnswers, quizReducer, saveAnswers } from "@/lib/quiz/state";
import { firstIncompleteIndex, stepsFor, validateStep, type StepId } from "@/lib/quiz/steps";
import type { QuizAnswers } from "@/lib/quiz/types";
import AccountStep from "./AccountStep";
import { ChoiceList, MultiChoiceList } from "./ChoiceList";
import CropStep from "./CropStep";
import IdentityStep from "./IdentityStep";
import LocationStep from "./LocationStep";
import QuizShell from "./QuizShell";
import { primaryButtonClass } from "./styles";

const TITLES: Record<StepId, { title: string; hint?: string }> = {
  style: { title: "Qual dessas combina com você?" },
  areas: { title: "Quantas áreas de plantio você tem?" },
  identity: { title: "Como podemos te chamar?", hint: "Os avisos chegam pelo WhatsApp." },
  crop: { title: "O que você planta?" },
  location: { title: "Onde fica a sua terra?", hint: "Busque a cidade e ajuste o pino sobre a sua área." },
  soil: { title: "Como é a terra?" },
  concerns: {
    title: "O que mais te preocupa no clima?",
    hint: "Pode marcar mais de um. Se pular, avisamos de tudo.",
  },
  producers: { title: "Quantos produtores você atende?" },
  account: { title: "Crie o seu acesso", hint: "Com e-mail e senha você entra de novo quando quiser." },
};

function parseStep(raw: string | null): number {
  const n = Number(raw ?? "1");
  return Number.isInteger(n) && n >= 1 ? n - 1 : 0;
}

export default function QuizFlow() {
  const router = useRouter();
  const params = useSearchParams();
  const [answers, dispatch] = useReducer(quizReducer, initialAnswers);
  const [hydrated, setHydrated] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Restore saved answers once in the browser; storage is never read during render.
  useEffect(() => {
    dispatch({ type: "replace", answers: loadAnswers() });
    setHydrated(true);
  }, []);

  useEffect(() => {
    if (hydrated) saveAnswers(answers);
  }, [answers, hydrated]);

  const steps = stepsFor(answers.work_style);
  const index = Math.min(parseStep(params.get("passo")), steps.length - 1);
  const furthest = firstIncompleteIndex(steps, answers);

  // A URL can't skip ahead of unanswered steps (e.g. a reload or a pasted link).
  useEffect(() => {
    if (hydrated && index > furthest) router.replace(`/cadastro?passo=${furthest + 1}`);
  }, [hydrated, index, furthest, router]);

  if (!hydrated || index > furthest) {
    return (
      <QuizShell index={0} total={null} title="Carregando...">
        {null}
      </QuizShell>
    );
  }

  const stepId = steps[index];
  const { title, hint } = TITLES[stepId];
  const patch = (p: Partial<QuizAnswers>) => dispatch({ type: "patch", patch: p });

  function goTo(i: number) {
    setError(null);
    router.push(`/cadastro?passo=${i + 1}`);
  }

  function next() {
    const message = validateStep(stepId, answers);
    if (message) {
      setError(message);
      return;
    }
    goTo(index + 1);
  }

  // Single-choice steps advance on tap; the Continue button stays for returning visitors.
  function pick(p: Partial<QuizAnswers>) {
    patch(p);
    setError(null);
    window.setTimeout(() => goTo(index + 1), 180);
  }

  function renderStep(): ReactNode {
    switch (stepId) {
      case "style":
        return (
          <ChoiceList
            options={STYLE_OPTIONS}
            value={answers.work_style}
            onChange={(work_style) => pick({ work_style })}
          />
        );
      case "areas":
        return (
          <ChoiceList
            options={AREAS_OPTIONS}
            value={answers.areas_bucket}
            onChange={(areas_bucket) => pick({ areas_bucket })}
          />
        );
      case "producers":
        return (
          <ChoiceList
            options={PRODUCERS_OPTIONS}
            value={answers.producers_bucket}
            onChange={(producers_bucket) => pick({ producers_bucket })}
          />
        );
      case "soil":
        return (
          <ChoiceList
            options={SOIL_OPTIONS}
            value={answers.soil_type}
            onChange={(soil_type) => pick({ soil_type })}
          />
        );
      case "identity":
        return <IdentityStep answers={answers} technician={answers.work_style === "technician"} onChange={patch} />;
      case "crop":
        return <CropStep answers={answers} onChange={patch} />;
      case "location":
        return <LocationStep answers={answers} onChange={patch} />;
      case "concerns":
        return (
          <MultiChoiceList
            options={CONCERN_OPTIONS}
            values={answers.concerns}
            onChange={(concerns) => patch({ concerns })}
          />
        );
      case "account":
        return (
          <AccountStep
            answers={answers}
            onEmailChange={(email) => patch({ email })}
            onFixStep={(step) => {
              const target = steps.indexOf(step);
              if (target !== -1) goTo(target);
            }}
          />
        );
    }
  }

  const footer =
    stepId === "account" ? null : (
      <div className="flex flex-col gap-2">
        {error && (
          <p role="alert" className="font-mono text-sm text-risk-critical">
            {error}
          </p>
        )}
        <button type="button" onClick={next} className={primaryButtonClass}>
          Continuar
        </button>
        {index === 0 && (
          <p className="text-center font-mono text-xs text-ink-soft">
            Já tem conta?{" "}
            <Link href="/entrar" className="text-accent underline">
              Entrar
            </Link>
          </p>
        )}
      </div>
    );

  return (
    <QuizShell
      index={index}
      total={answers.work_style ? steps.length : null}
      title={title}
      hint={hint}
      onBack={index > 0 ? () => goTo(index - 1) : undefined}
      footer={footer}
    >
      {renderStep()}
    </QuizShell>
  );
}
```

> O `clearAnswers` não é usado aqui; quem limpa o storage é o `ClearQuiz` (Step 13), na tela de boas-vindas.

- [ ] **Step 12: Criar `web/app/cadastro/page.tsx`**

```tsx
import type { Metadata } from "next";
import { Suspense } from "react";
import QuizFlow from "@/components/quiz/QuizFlow";

export const metadata: Metadata = { title: "Cadastro · AgroResiliente" };

export default function SignupPage() {
  return (
    <Suspense fallback={null}>
      <QuizFlow />
    </Suspense>
  );
}
```

- [ ] **Step 13: Criar `web/components/ClearQuiz.tsx`**

```tsx
"use client";

import { useEffect } from "react";
import { clearAnswers } from "@/lib/quiz/state";

/** The quiz answers are only needed until the account exists; drop them once the welcome page loads. */
export default function ClearQuiz() {
  useEffect(() => {
    clearAnswers();
  }, []);
  return null;
}
```

- [ ] **Step 14: Criar `web/app/bem-vindo/page.tsx`**

```tsx
import type { Metadata } from "next";
import Link from "next/link";
import ClearQuiz from "@/components/ClearQuiz";
import { listProducers } from "@/lib/api";
import { CONCERN_LABELS, welcomeMessage } from "@/lib/personalize";
import { requireUser } from "@/lib/session";

export const metadata: Metadata = { title: "Tudo pronto · AgroResiliente" };

export default async function WelcomePage() {
  const user = await requireUser();
  const producers = user.profile === "producer" ? await listProducers() : [];
  const city = producers[0]?.city;
  const firstName = user.name.trim().split(/\s+/)[0];
  const concerns = user.concerns.map((c) => CONCERN_LABELS[c]).filter(Boolean);

  return (
    <div className="mx-auto flex min-h-dvh w-full max-w-md flex-1 flex-col justify-center gap-6 px-5 py-8">
      <ClearQuiz />

      <div className="rise-in">
        <p className="mb-1 font-mono text-[11px] uppercase tracking-[0.25em] text-ink-soft">Cadastro concluído</p>
        <h1 className="font-display text-3xl italic leading-tight text-ink">Tudo pronto, {firstName}.</h1>
        <p className="mt-3 text-base text-ink">{welcomeMessage(user, city)}</p>
      </div>

      {concerns.length > 0 && (
        <ul className="flex flex-wrap gap-2 rise-in" style={{ animationDelay: "80ms" }}>
          {concerns.map((label) => (
            <li
              key={label}
              className="border-2 border-ink bg-card px-3 py-1.5 font-mono text-xs font-semibold uppercase tracking-wide text-ink"
            >
              {label}
            </li>
          ))}
        </ul>
      )}

      <div className="flex flex-col gap-3 rise-in" style={{ animationDelay: "140ms" }}>
        {user.profile === "technician" && (
          <Link
            href="/producers/new"
            className="flex min-h-12 items-center justify-center border-2 border-ink bg-accent px-5 font-mono text-sm font-semibold uppercase tracking-wide text-paper transition hover:bg-ink"
          >
            Cadastrar o primeiro produtor
          </Link>
        )}
        <Link
          href="/"
          className={`flex min-h-12 items-center justify-center border-2 border-ink px-5 font-mono text-sm font-semibold uppercase tracking-wide transition ${
            user.profile === "technician" ? "text-ink hover:bg-ink hover:text-paper" : "bg-accent text-paper hover:bg-ink"
          }`}
        >
          Ir para o painel
        </Link>
      </div>
    </div>
  );
}
```

- [ ] **Step 15: Conferir tipos, lint e testes**

Run:
```bash
cd /home/lukas-melo/PROJETOS/agrolang/web
npx tsc --noEmit
npm run lint
npm test
```
Expected: sem erros de tipo e testes PASS. Se o lint apontar `react-hooks/set-state-in-effect` em `QuizFlow` (`setHydrated`/`dispatch` dentro do `useEffect` de restauração), adicionar acima da linha `dispatch({ type: "replace", answers: loadAnswers() });` o comentário `// eslint-disable-next-line react-hooks/set-state-in-effect -- restoring browser-only storage after mount`.

- [ ] **Step 16: Commit**

```bash
git add web/components web/app/cadastro web/app/bem-vindo
git commit -m "feat(web): add onboarding quiz and welcome page" -m "Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>"
```

---

### Task 15: Dashboard personalizado

**Files:**
- Modify (substituir): `web/app/page.tsx`

- [ ] **Step 1: Substituir `web/app/page.tsx`**

```tsx
import Link from "next/link";
import { listProperties, getPropertyAlerts } from "@/lib/api";
import { logoutAction } from "@/app/actions/auth";
import PropertyMapClient from "@/components/PropertyMapClient";
import {
  CONCERN_LABELS,
  currentHourBR,
  greeting,
  matchesConcern,
  missingAreasNotice,
  sortByPriority,
} from "@/lib/personalize";
import { currentRiskLevel, riskLabels } from "@/lib/risk";
import { requireUser } from "@/lib/session";

const riskBadgeClass: Record<string, string> = {
  low: "bg-risk-low-bg text-risk-low border-risk-low",
  medium: "bg-risk-medium-bg text-risk-medium border-risk-medium",
  critical: "bg-risk-critical-bg text-risk-critical border-risk-critical pulse-critical",
};

const cropStageLabels: Record<string, string> = {
  germination: "Germinação",
  flowering: "Floração",
  harvest: "Colheita",
};

export default async function Home() {
  const user = await requireUser();
  const isTechnician = user.profile === "technician";

  let items: Awaited<ReturnType<typeof loadItems>> = [];
  let error: string | null = null;

  try {
    items = await loadItems();
  } catch {
    error = "Não foi possível carregar as propriedades. Verifique se a API está no ar.";
  }

  const counts = items.reduce(
    (acc, { alerts }) => {
      acc[currentRiskLevel(alerts)] += 1;
      return acc;
    },
    { low: 0, medium: 0, critical: 0 }
  );

  const sorted = sortByPriority(items, user.concerns);
  const concernChips = user.concerns.map((c) => CONCERN_LABELS[c]).filter(Boolean);
  const areasNotice = isTechnician ? null : missingAreasNotice(user.onboarding.areas_bucket, items.length);

  return (
    <div className="flex flex-1 flex-col">
      <header className="relative overflow-hidden border-b-2 border-ink px-5 py-5 sm:px-8">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div className="rise-in">
            <p className="mb-1 font-mono text-[11px] uppercase tracking-[0.25em] text-ink-soft">
              {isTechnician && user.organization ? user.organization : "Estação de campo"} &middot; RF-05
            </p>
            <h1 className="font-display text-3xl italic tracking-tight text-ink sm:text-4xl">
              {greeting(user.name, currentHourBR())}
            </h1>
            <p className="mt-1 max-w-md text-sm text-ink-soft">
              Leitura do risco climático de cada talhão monitorado, em tempo quase real.
            </p>
            {concernChips.length > 0 && (
              <ul className="mt-3 flex flex-wrap items-center gap-2" aria-label="Riscos que você acompanha">
                <li className="font-mono text-[10px] uppercase tracking-wide text-ink-soft">Acompanhando</li>
                {concernChips.map((label) => (
                  <li
                    key={label}
                    className="border-2 border-ink bg-card px-2 py-0.5 font-mono text-[11px] font-semibold uppercase tracking-wide text-ink"
                  >
                    {label}
                  </li>
                ))}
              </ul>
            )}
          </div>

          <div className="flex items-end gap-2 rise-in" style={{ animationDelay: "80ms" }}>
            <CountPill label="Baixo" count={counts.low} tone="low" big={isTechnician} />
            <CountPill label="Médio" count={counts.medium} tone="medium" big={isTechnician} />
            <CountPill label="Crítico" count={counts.critical} tone="critical" big={isTechnician} />
            <form action={logoutAction}>
              <button
                type="submit"
                className="min-h-11 border-2 border-ink px-3 font-mono text-[11px] font-semibold uppercase tracking-wide text-ink transition hover:bg-ink hover:text-paper"
              >
                Sair
              </button>
            </form>
          </div>
        </div>
      </header>

      {error && (
        <p className="border-b-2 border-risk-critical bg-risk-critical-bg px-6 py-3 text-sm font-medium text-risk-critical sm:px-8">
          {error}
        </p>
      )}

      {areasNotice && (
        <div className="flex flex-wrap items-center justify-between gap-3 border-b-2 border-risk-medium bg-risk-medium-bg px-5 py-3 text-sm text-ink sm:px-8">
          <p>{areasNotice}</p>
          <Link
            href="/properties/new"
            className="min-h-11 border-2 border-ink bg-accent px-3 py-2 font-mono text-xs font-semibold uppercase tracking-wide text-paper transition hover:bg-ink"
          >
            + Propriedade
          </Link>
        </div>
      )}

      <div className="flex flex-1 flex-col gap-5 p-5 lg:flex-row lg:p-8">
        <div
          className="h-[420px] shrink-0 overflow-hidden border-2 border-ink shadow-[var(--shadow-hard)] lg:h-auto lg:flex-1 lg:shrink rise-in"
          style={{ animationDelay: "140ms" }}
        >
          <PropertyMapClient items={sorted} />
        </div>

        <aside className="flex w-full flex-col lg:w-96">
          <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
            <h2 className="font-display text-lg italic text-ink">Talhões</h2>
            <div className="flex gap-2">
              <Link
                href="/producers/new"
                className={`min-h-11 border-2 border-ink px-3 py-2 font-mono text-xs font-semibold uppercase tracking-wide transition ${
                  isTechnician
                    ? "bg-accent text-paper hover:bg-ink"
                    : "text-ink hover:bg-ink hover:text-paper"
                }`}
              >
                + Produtor
              </Link>
              <Link
                href="/properties/new"
                className={`min-h-11 border-2 border-ink px-3 py-2 font-mono text-xs font-semibold uppercase tracking-wide transition ${
                  isTechnician
                    ? "text-ink hover:bg-ink hover:text-paper"
                    : "bg-accent text-paper hover:bg-ink"
                }`}
              >
                + Propriedade
              </Link>
            </div>
          </div>

          <ul className="flex flex-1 flex-col gap-2.5 overflow-y-auto">
            {sorted.map(({ property, alerts }, i) => {
              const level = currentRiskLevel(alerts);
              const prioritized = matchesConcern(alerts, user.concerns);
              return (
                <li
                  key={property.id}
                  className="rise-in"
                  style={{ animationDelay: `${180 + i * 40}ms` }}
                >
                  <Link
                    href={`/properties/${property.id}`}
                    className="group flex items-center justify-between gap-3 border-2 border-line bg-card px-4 py-3 transition hover:border-ink hover:-translate-y-0.5 hover:shadow-[var(--shadow-hard)]"
                  >
                    <div className="min-w-0">
                      <p className="truncate font-display text-base text-ink">{property.crop}</p>
                      <p className="truncate font-mono text-[11px] uppercase tracking-wide text-ink-soft">
                        {cropStageLabels[property.crop_stage] ?? property.crop_stage} &middot;{" "}
                        {property.soil_type}
                      </p>
                    </div>
                    <div className="flex shrink-0 flex-col items-end gap-1">
                      {prioritized && (
                        <span className="font-mono text-[10px] font-semibold uppercase tracking-wide text-accent">
                          Prioridade
                        </span>
                      )}
                      <span
                        className={`border-2 px-2 py-1 font-mono text-[11px] font-semibold uppercase tracking-wide ${riskBadgeClass[level]}`}
                      >
                        {riskLabels[level]}
                      </span>
                    </div>
                  </Link>
                </li>
              );
            })}
            {items.length === 0 && !error && (
              <li className="border-2 border-dashed border-line px-4 py-6 text-center font-mono text-sm text-ink-soft">
                {isTechnician ? "Você ainda não cadastrou produtores." : "Nenhuma propriedade cadastrada."}
                <br />
                <Link href="/producers/new" className="underline decoration-dashed underline-offset-4">
                  {isTechnician ? "Cadastrar o primeiro produtor" : "Cadastrar a primeira"}
                </Link>
              </li>
            )}
          </ul>
        </aside>
      </div>
    </div>
  );
}

function CountPill({
  label,
  count,
  tone,
  big,
}: {
  label: string;
  count: number;
  tone: string;
  big?: boolean;
}) {
  return (
    <div className={`border-2 px-3 py-1.5 text-center font-mono ${riskBadgeClass[tone]}`}>
      <div className={`${big ? "text-2xl" : "text-lg"} font-semibold leading-none tabular`}>{count}</div>
      <div className="mt-1 text-[10px] uppercase leading-none tracking-wide">{label}</div>
    </div>
  );
}

async function loadItems() {
  const properties = await listProperties();
  return Promise.all(
    properties.map(async (property) => ({
      property,
      alerts: await getPropertyAlerts(property.id),
    }))
  );
}
```

- [ ] **Step 2: Conferir tipos, lint, testes e build**

Run:
```bash
cd /home/lukas-melo/PROJETOS/agrolang/web
npx tsc --noEmit && npm run lint && npm test && npm run build
```
Expected: tudo passa; o `build` lista as rotas `/`, `/cadastro`, `/entrar`, `/bem-vindo`, `/sair` e o `Proxy (Middleware)`.

- [ ] **Step 3: Commit**

```bash
git add web/app/page.tsx
git commit -m "feat(web): personalize the dashboard by profile and chosen risks" -m "Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>"
```

---

### Task 16: Verificação ponta a ponta no navegador (390 px)

**Files:** nenhum arquivo novo (só correções, se algo falhar).

- [ ] **Step 1: Subir a pilha local**

Run (API Go já sobe da Task 8; reinicie o Next para ler o `API_URL`):
```bash
cd /home/lukas-melo/PROJETOS/agrolang/web
kill $(ss -ltnp 2>/dev/null | grep ':3000' | grep -o 'pid=[0-9]*' | cut -d= -f2) 2>/dev/null
npm run dev > /tmp/agrolang-web.log 2>&1 &
sleep 6 && curl -s -o /dev/null -w "%{http_code}\n" localhost:3000/cadastro
curl -s localhost:8080/health
```
Expected: `200` e `{"status":"ok"}`. (Se a API não estiver de pé, repita o Step 9 da Task 8.)

- [ ] **Step 2: Rota protegida redireciona**

Run: `curl -s -o /dev/null -w "%{http_code} %{redirect_url}\n" localhost:3000/`
Expected: `307 http://localhost:3000/entrar`.

- [ ] **Step 3: Percorrer o caminho do produtor no navegador**

Com o MCP `chrome-devtools` (ou Chrome normal em modo celular), abrir `http://localhost:3000/cadastro` com viewport 390×844 e conferir, passo a passo:
1. Tela 1: tocar em "Planto na minha terra, com a família" avança sozinho; a barra mostra "Passo 2 de 8".
2. "Uma só" avança; "Voltar" volta ao passo anterior e o botão de voltar do navegador também.
3. Nome e WhatsApp: tentar "Continuar" vazio mostra "Digite seu nome."; preencher `José da Silva` / `19 99999-9999`.
4. Cultura "Milho" + fase "Está florindo" → Continuar.
5. Localização: buscar "Campinas", escolher "Campinas, SP"; o mapa centra e o pino aparece; arrastar o pino muda as coordenadas; Continuar.
6. Solo "Argiloso" avança sozinho; preocupações: marcar Geada e Seca → Continuar.
7. Recarregar a página no meio (ex.: em `?passo=6`): as respostas continuam e o passo é o mesmo.
8. Abrir `/cadastro?passo=8` numa aba anônima (sem respostas): redireciona para o primeiro passo incompleto.
9. Conta: e-mail novo + senha de 8+ caracteres → "Criar minha conta" leva a `/bem-vindo` com "Vamos te avisar de geada e seca em Campinas." e os chips; "Ir para o painel" abre o dashboard com a saudação, os chips "Acompanhando" e a propriedade.

Expected: tudo acima acontece; nenhuma rolagem horizontal; botões com altura de toque confortável.

- [ ] **Step 4: Conferir a persistência e a limpeza**

No console do navegador, em `/bem-vindo`: `localStorage.getItem("agrolang.quiz.v1")`.
Expected: `null` (a tela de boas-vindas limpa as respostas). E em `Application → Cookies`: cookie `session` com `HttpOnly` marcado.

- [ ] **Step 5: Caminho do técnico**

Sair (botão "Sair" no dashboard), abrir `/cadastro`, escolher "Sou agrônomo ou técnico de cooperativa" (a barra passa a "Passo 2 de 5"), preencher nome, WhatsApp e organização, "De 11 a 40", marcar "Chuva forte" e criar a conta com outro e-mail.
Expected: `/bem-vindo` mostra "Vamos destacar chuva forte no seu painel." e o botão "Cadastrar o primeiro produtor"; o dashboard mostra o nome da organização, contadores grandes, estado vazio "Você ainda não cadastrou produtores." e "+ Produtor" em destaque. Cadastrar um produtor e uma propriedade pelos formulários e ver que aparecem no painel.

- [ ] **Step 6: Erros**

- Criar conta de novo com o e-mail já usado: aparece "Este e-mail já tem cadastro." com o link "Entrar com este e-mail", e as respostas continuam salvas.
- `/entrar` com senha errada: "E-mail ou senha incorretos."; com a certa: abre o dashboard.
- Isolamento: logado como técnico, `GET http://localhost:3000/properties/<id-do-produtor>` (id de uma propriedade do outro usuário) deve mostrar a página sem dados (erro de carregamento), e `curl -s localhost:8080/properties/<id> -H "Authorization: Bearer <token do técnico>"` deve responder `404`.
- GPS negado (negar a permissão no navegador e tocar "Usar minha localização"): aparece o aviso e o pino segue funcionando.
- Desligar a API Go e tentar criar a conta: aparece "Sem conexão com o servidor. Tente de novo." sem perder as respostas.

Expected: cada item se comporta como descrito.

- [ ] **Step 7: Alertas por preocupação (conferência no banco)**

Run:
```bash
docker exec agrolang-db-1 psql -U agrolang -d agrolang -c "select email, profile, concerns, onboarding from users order by created_at"
```
Expected: os dois usuários de teste com `concerns` e `onboarding` como escolhidos. (O filtro de WhatsApp em si é coberto por `go test ./internal/app/`.)

- [ ] **Step 8: Limpar os dados de teste**

Run:
```bash
docker exec agrolang-db-1 psql -U agrolang -d agrolang -c "DELETE FROM users WHERE email IN ('<email do produtor de teste>', '<email do técnico de teste>')"
```
Expected: `DELETE 2` (produtores e propriedades vão junto, por cascata).

- [ ] **Step 9: Registrar a mudança de variável no README**

Acrescentar ao final de `web/README.md`:

```markdown

## Configuração

O frontend chama o backend Go **só no servidor**, com o token da sessão em cookie `httpOnly`.
Defina a URL da API em `web/.env.local`:

```
API_URL=http://localhost:8080
```

(Antes era `NEXT_PUBLIC_API_URL`; ela não é mais usada.)
```

- [ ] **Step 10: Rodar tudo uma última vez e commitar**

Run:
```bash
cd /home/lukas-melo/PROJETOS/agrolang
go build ./... && go vet ./... && go test ./... 2>&1 | tail -15
cd web && npx tsc --noEmit && npm run lint && npm test
```
Expected: tudo passa.

```bash
cd /home/lukas-melo/PROJETOS/agrolang
git add web/README.md
git commit -m "docs(web): document API_URL for the authenticated frontend" -m "Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>"
git status --short
```
Expected: `git status` vazio.

---

## Autoavaliação do plano

**Cobertura do spec**
- §3 fluxo de telas: Tasks 10 (passos/validações), 14 (UI), 16 (verificação). Passo na URL, progresso, voltar, `localStorage` sem senha, botões ≥48 px: Tasks 10 e 14.
- §4.1 migration: Task 1. §4.2 endpoints: Task 7. §4.3 sessão/senha: Tasks 2 e 7. §4.4 isolamento: Tasks 5 e 8. §4.5 alertas: Task 9.
- §5.1 rotas e `proxy.ts`: Tasks 12–14. §5.2 unidades: Tasks 10–14 (todas as listadas, inclusive `lib/session.ts`, `lib/personalize.ts`). §5.3 localização/geocodificação e fallbacks: Tasks 11 e 14. §5.4 dashboard: Task 15.
- §6 erros: Task 12 (`actions/auth.ts`), Task 14 (`AccountStep`), Task 16 (verificação).
- §7 testes: Go nas Tasks 2, 4–9; Vitest nas Tasks 10–11; manual na Task 16.

**Itens que o spec deixou implícitos e o plano fixa**
- Listas vazias da API voltam como `[]`, não `null` (`ListByUser` usa `make`), para a conta nova não quebrar o dashboard.
- `GetByID` do produtor agora traz `Concerns` (via `LEFT JOIN users`), em vez de mudar a assinatura do `RunCycle`.
- Sessão expirada/inválida passa por `/sair` (route handler), que é quem pode apagar o cookie.
- Logado, `/entrar` e `/cadastro` redirecionam para `/`.

**Consistência de nomes** (conferida): `ShouldNotify`, `ListByUser`/`GetByIDForUser`/`CreateForUser`/`ListRecentForUser`, `ErrEmailTaken`, `SignupInput`/`SignupResult`, `NewAuthHandler(signuper, store)`, `RegisterPublic`/`RegisterProtected`, `ContextWithUser`/`UserFromContext`/`RequireSession`, `stepsFor`/`validateStep`/`firstIncompleteIndex`/`stepForField`, `buildSignupPayload`, `loadAnswers`/`saveAnswers`/`clearAnswers`, `signupAction`/`loginAction`/`logoutAction`, `requireUser`/`getCurrentUser`.

**Riscos conhecidos**
- Os testes de repositório truncam o banco de desenvolvimento (`agrolang`), como já era antes; os dados locais de teste somem a cada execução.
- Dados antigos sem `user_id` ficam invisíveis na interface, mas continuam sendo varridos e notificados.
- O lint do React 19 pode exigir o `eslint-disable` descrito na Task 14, Step 16.
