# Quiz de cadastro (funil de onboarding) com login — Design

Data: 2026-10-30 · Status: aprovado em conversa, aguardando revisão do spec escrito

## 1. Objetivo

Hoje o AgroResiliente não tem conta nem login: produtor e propriedade são cadastrados por dois formulários soltos, e todas as rotas da API são abertas. Este trabalho adiciona:

1. Um **quiz mobile-first**, uma pergunta por tela, que personaliza o caminho conforme o perfil (produtor ou técnico/cooperativa) e coleta o que o app precisa para mostrar dados já no primeiro acesso.
2. **Conta com e-mail e senha**, criada no último passo do quiz. O acesso ao app só é liberado depois dela.
3. **Personalização real**: as respostas mudam o dashboard e configuram quais alertas de WhatsApp a pessoa recebe.
4. **Isolamento de dados**: cada conta só enxerga os próprios produtores, propriedades e alertas.

## 2. Decisões tomadas

| Tema | Decisão |
|---|---|
| Perfis | Produtor e técnico/cooperativa, com caminhos diferentes. |
| Pergunta inicial | Por "jeito de trabalhar", em linguagem do dia a dia. Produtor responde depois uma pergunta de tamanho (número de áreas). |
| Login | E-mail + senha. O WhatsApp fica só como contato para alertas e **não** é verificado. |
| Uso das respostas | Personalizam a tela **e** configuram os alertas. |
| Técnico | Quiz leve (só sobre ele). Depois cadastra produtores com os formulários existentes. Sem importação em lote. |
| Arquitetura do funil | Respostas no navegador e **envio único** no fim (`POST /auth/signup` transacional). Sem rascunho no servidor. |
| Localização | Pino no mapa, abrindo na cidade informada, com atalho de GPS. |
| Sessão | Token opaco emitido pelo Go, guardado pelo Next em cookie `httpOnly`. |
| Alerta crítico | **Sempre notifica**, independente das preocupações escolhidas. |

## 3. Fluxo de telas

Tela 1 é comum. Cada opção define o perfil e o `work_style`:

| Opção na tela 1 | `profile` | `work_style` |
|---|---|---|
| Planto na minha terra, com a família | `producer` | `family` |
| Tenho várias áreas e contrato gente | `producer` | `multi_area` |
| Sou agrônomo ou técnico de cooperativa | `technician` | `technician` |

### Produtor (8 telas)

1. Jeito de trabalhar.
2. Quantas áreas: `1` · `2-5` · `6+` (grava `onboarding.areas_bucket`).
3. Nome e WhatsApp. O telefone é digitado com DDD (10 ou 11 dígitos, teclado numérico); o payload prefixa `55`.
4. Cultura (botões: Milho, Tomate, Café, Soja, Feijão, Outra com texto livre) e fase em linguagem simples: "Acabei de plantar" (`germination`), "Está florindo" (`flowering`), "Perto de colher" (`harvest`).
5. Onde fica: campo de cidade com sugestões (geocodificação Open-Meteo) → mapa centrado com pino arrastável e botão "Usar minha localização". Preenche `city`, `state`, `latitude`, `longitude`.
6. Tipo de solo: `argiloso` · `arenoso` · `misto` · "Não sei" (`nao_informado`), com descrição curta de cada um.
7. O que mais te preocupa (múltipla escolha): Geada (`frost`), Seca (`severe_water_stress`), Calor forte (`heat_wave`), Chuva forte (`heavy_rain`). Pular sem marcar significa "avisar de tudo".
8. Criar conta: e-mail e senha (mínimo 8 caracteres).

Depois do cadastro: `/bem-vindo` (resumo personalizado montado a partir da conta, ex.: "Vamos te avisar de geada e seca em Campinas") e botão para o dashboard.

### Técnico / cooperativa (5 telas)

1. Jeito de trabalhar.
2. Nome, WhatsApp e organização.
3. Quantos produtores atende: `ate-10` · `11-40` · `mais-40` (grava `onboarding.producers_bucket`).
4. Riscos a priorizar (mesma lista do passo 7 do produtor).
5. Criar conta.

Depois: `/bem-vindo` e dashboard vazio com "+ Produtor" como chamada principal. Nenhum `producer` nem `property` é criado no quiz.

### Comportamento comum

- Barra de progresso, botão "Voltar" sempre visível, passo atual na URL (`/cadastro?passo=N`) para o botão de voltar do celular funcionar.
- Botões com no mínimo 48 px de altura, uma ação principal por tela.
- Respostas persistidas em `localStorage` (chave versionada, leitura e escrita em `try/catch`, o quiz funciona sem storage). **A senha nunca é persistida.** O storage é limpo após cadastro bem-sucedido.
- Visual "Field Instrument" já usado no dashboard (tokens de `web/app/globals.css`, Fraunces + IBM Plex Mono).

## 4. Backend (Go)

### 4.1 Migration `000003`

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

O `down` desfaz na ordem inversa. `concerns` guarda os valores de `risk.Alert` já existentes. `onboarding` guarda `work_style`, `areas_bucket` e `producers_bucket`, usados para personalizar e segmentar, sem lógica de negócio.

### 4.2 Endpoints de autenticação

| Rota | Comportamento |
|---|---|
| `POST /auth/signup` | Valida o corpo; em **uma transação** cria `users` e, se `profile=producer`, `producers` (com `user_id`) e `properties`; cria a sessão. Responde `201` com `{token, expires_at, user}`. E-mail repetido: `409`. Corpo inválido: `400` com mensagem por campo. |
| `POST /auth/login` | E-mail + senha. Mesma mensagem de erro (`401`) para e-mail inexistente e senha errada. Responde `200` com `{token, expires_at, user}`. |
| `POST /auth/logout` | Apaga a sessão do token informado. `204`. |
| `GET /auth/me` | Dados da conta (perfil, nome, organização, `concerns`, `onboarding`). `401` sem sessão. |

Corpo do `POST /auth/signup`:

```json
{
  "profile": "producer",
  "name": "José da Silva",
  "whatsapp_phone": "5519999999999",
  "email": "jose@exemplo.com",
  "password": "********",
  "organization": null,
  "concerns": ["frost", "severe_water_stress"],
  "onboarding": { "work_style": "family", "areas_bucket": "1" },
  "producer": { "city": "Campinas", "state": "SP" },
  "property": {
    "latitude": -22.9099, "longitude": -47.0626,
    "crop": "Milho", "soil_type": "argiloso", "crop_stage": "flowering"
  }
}
```

`producer` e `property` são obrigatórios para `producer` e proibidos para `technician`. O `producers.name` e `producers.whatsapp_phone` são copiados de `users.name` e `users.whatsapp_phone`.

### 4.3 Sessão e senha

- Token: 32 bytes aleatórios (`crypto/rand`), enviado em base64url. O banco guarda só o SHA-256. Validade de 30 dias; sessões expiradas são ignoradas na consulta e removidas no login.
- Senha: `bcrypt` (nova dependência `golang.org/x/crypto`), mínimo de 8 caracteres.
- Middleware `RequireSession` lê `Authorization: Bearer <token>`, carrega o usuário e o coloca no `context`. Sem token ou token inválido: `401`.

### 4.4 Isolamento de dados

`RequireSession` passa a proteger **todas** as rotas atuais; `/health` e `/auth/signup`, `/auth/login` permanecem públicas.

- `GET /producers`, `GET /properties`, `GET /alerts`: filtram por `producers.user_id = usuário`.
- `GET /producers/{id}`, `GET /properties/{id}`, `GET /properties/{id}/alerts`, `GET /properties/{id}/forecast`: retornam `404` se o recurso não for do usuário.
- `POST /producers`: grava `user_id` do usuário logado.
- `POST /properties`: retorna `404` se o `producer_id` não for do usuário.
- Produtores sem `user_id` (dados de teste anteriores) ficam invisíveis e continuam sendo varridos pelo ciclo de alertas.

### 4.5 Alertas personalizados

No ciclo de varredura (`internal/app/cycle.go`), antes do WhatsApp, o sistema consulta `concerns` do dono do produtor:

- Nível **crítico**: sempre notifica.
- Nível **médio**: notifica se `concerns` está vazio ou contém o tipo do alerta.
- Produtor sem `user_id`: notifica tudo (comportamento atual).
- O alerta é **sempre gravado**, independente de notificar, para histórico e dashboard.

## 5. Frontend (Next 16)

Segue `web/AGENTS.md`: `proxy.ts` (não `middleware`), `cookies()` assíncrono, server actions com `useActionState`.

### 5.1 Rotas

| Rota | Função |
|---|---|
| `/cadastro` | Quiz. |
| `/entrar` | Login. |
| `/bem-vindo` | Resumo personalizado pós-cadastro. |
| `/` e demais | Protegidas. `proxy.ts` redireciona para `/entrar` quando não há cookie `session` (verificação otimista); a verificação real fica em `lib/session.ts`, que chama `GET /auth/me`. |

### 5.2 Unidades

- `lib/quiz/steps.ts` — definição declarativa dos passos de cada perfil.
- `lib/quiz/state.ts` — reducer das respostas e persistência em `localStorage`.
- `lib/quiz/payload.ts` — respostas → corpo do `POST /auth/signup` (normaliza telefone, mapeia rótulos para valores).
- `components/quiz/*` — casca (progresso/voltar), botão de opção, um componente por tipo de passo, `LocationStep` (Leaflet carregado só no cliente, no padrão de `PropertyMapClient`).
- `app/actions/auth.ts` — server actions `signup`, `login`, `logout`. Gravam o cookie `session` (`httpOnly`, `sameSite=lax`, `secure` em produção, `path=/`, 30 dias) e redirecionam.
- `lib/session.ts` — acesso ao token e à conta (`requireSession`, `getUser`).
- `lib/api.ts` — roda no servidor e envia `Authorization: Bearer`. `NEXT_PUBLIC_API_URL` é substituída por `API_URL` (só servidor). Os formulários de produtor e propriedade mantêm a interface, mas chamam server actions.
- `lib/personalize.ts` — funções puras: ordenação de alertas por nível e depois por tipo priorizado; saudação; aviso de áreas faltando.

### 5.3 Localização (passo 5 do produtor)

Sugestões de cidade vêm da API de geocodificação do Open-Meteo (`geocoding-api.open-meteo.com`), chamada do navegador (sem chave). Falha na geocodificação: mapa abre no centro do Brasil e a pessoa aproxima manualmente, preenchendo cidade e UF à mão. GPS negado: aviso curto e uso do pino.

### 5.4 Dashboard personalizado

- **Produtor:** saudação pelo nome, chips das preocupações ("Geada · Seca"), alertas dos tipos priorizados primeiro. Se `areas_bucket` indica N áreas e há menos que N propriedades cadastradas (`2-5` → mínimo 2; `6+` → mínimo 6), mostra aviso "cadastre as outras áreas".
- **Técnico:** nome da organização, contadores baixo/médio/crítico em destaque e "+ Produtor" como ação principal enquanto não houver produtores.

## 6. Erros

| Situação | Comportamento |
|---|---|
| E-mail já cadastrado (`409`) | Mensagem no passo da conta, com link para `/entrar`. Respostas mantidas. |
| Falha de rede no envio | Mensagem e botão para tentar de novo; respostas mantidas. |
| Validação do servidor (`400`) | Mensagem por campo; se o campo é de um passo anterior, leva a pessoa até ele. |
| Credenciais inválidas no login | Mensagem genérica única. |
| Sessão expirada/inválida | `401` do Go → limpa cookie → `/entrar`. |

## 7. Testes

- **Go** (padrão atual, testes de handler e repositório): signup transacional (sucesso, e-mail duplicado, rollback se a propriedade falhar), login, middleware, isolamento entre dois usuários em cada rota, filtro de notificação (crítico sempre, médio filtrado, lista vazia, sem `user_id`).
- **Vitest:** reducer e persistência do quiz, `payload` (telefone, mapeamentos), validações, `personalize`.
- **Manual no navegador a 390 px:** os dois caminhos completos, voltar pelo botão do celular, recarregar no meio do quiz, GPS negado, e-mail repetido.

## 8. Fora de escopo

Importação em lote de produtores, verificação de e-mail ou de WhatsApp, "esqueci a senha", limite de tentativas de login, login social, tela para editar preferências depois do cadastro, métricas de abandono do funil e papéis/permissões de cooperativa. Cada um pode virar um trabalho próprio.
