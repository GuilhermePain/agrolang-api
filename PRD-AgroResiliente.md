# Documento de Requisitos de Produto (PRD) — AgroResiliente

## 1. Visão Geral do Produto

O **AgroResiliente** é uma plataforma de inteligência e adaptação climática focada em pequenos e médios agricultores. O sistema monitora dados meteorológicos locais via APIs abertas, processa regras de risco agrícola em um backend de alta performance em **Go**, e entrega alertas preventivos e acionáveis via **WhatsApp**, além de oferecer um dashboard web para cooperativas e produtores.

## 2. Objetivos e Métricas de Sucesso

### Objetivos Principais

- Reduzir perdas de safra causadas por eventos climáticos extremos (secas, geadas, ondas de calor e tempestades).
- Democratizar o acesso à informação meteorológica tratada para produtores sem letramento tecnológico avançado.
- Oferecer uma ferramenta de baixo custo e alta escala para cooperativas e prefeituras.

### Métricas de Sucesso (KPIs)

| KPI | Meta |
|---|---|
| Taxa de Entregabilidade dos Alertas | > 98% das notificações enviadas em menos de 30 segundos após o disparo do evento |
| Aderência dos Usuários | > 70% de leitura e engajamento com as recomendações enviadas via WhatsApp |
| Eficiência Computacional | Tempo de processamento da varredura de risco < 2 segundos para cada 1.000 propriedades |

## 3. Personas de Usuário

| Persona | Perfil | Dor Principal | Solução no AgroResiliente |
|---|---|---|---|
| **Seu Zé** (Produtor Familiar) | 52 anos, cultiva milho e hortaliças, usa apenas WhatsApp no celular | Não sabe interpretar gráficos meteorológicos e perde safras por falta de manejo preventivo | Recebe mensagens curtas e diretas no WhatsApp orientando a ação prática do dia |
| **Ana** (Agrônoma de Cooperativa) | 34 anos, atende 40 propriedades locais simultaneamente | Dificuldade em monitorar o risco climático de todos os cooperados em tempo real | Utiliza o Dashboard Web com mapa de calor para priorizar visitas e orientações técnicas |

## 4. Requisitos Funcionais (RF)

### RF-01: Gestão de Produtores e Propriedades

- **RF-01.1**: O sistema deve permitir o cadastro de produtores (Nome, Telefone/WhatsApp, Cidade, Estado).
- **RF-01.2**: O sistema deve armazenar a localização exata da propriedade (coordenadas GPS de latitude e longitude) e a cultura agrícola ativa (ex: Milho, Tomate, Café).
- **RF-01.3**: O sistema deve permitir cadastrar o tipo de solo e a fase atual do cultivo (ex: Germinação, Floração, Colheita).

### RF-02: Ingestão de Dados Meteorológicos

- **RF-02.1**: O backend em Go deve realizar chamadas periódicas (Cron Job) para APIs meteorológicas públicas (Open-Meteo / INMET / NASA POWER).
- **RF-02.2**: O sistema deve capturar previsões de curto prazo (24h a 72h) contendo: Temperatura máxima/mínima, Precipitação (mm), Umidade relativa do ar e Velocidade do vento.

### RF-03: Motor de Regras e Matriz de Risco (Core Engine)

- **RF-03.1**: O backend deve cruzar os dados climáticos coletados com os limites críticos de cada cultura cadastrada.
- **RF-03.2**: O sistema deve categorizar o nível de risco em: **Baixo** (Verde), **Médio** (Amarelo) e **Crítico** (Vermelho).
- **RF-03.3**: Exemplo de Regra Interna:
  - **Gatilho**: Temperatura Média > 35°C E Umidade do Ar < 30% por 48 horas seguidas na fase de Floração.
  - **Ação Gerada**: Disparo de Alerta de Estresse Hídrico Severo.

### RF-04: Notificação e Automação (WhatsApp)

- **RF-04.1**: Ao detectar um evento de risco Médio ou Crítico, o backend em Go deve acionar o webhook do serviço de mensageria (Evolution API / n8n).
- **RF-04.2**: A mensagem enviada ao agricultor deve conter obrigatoriamente:
  - Identificação do Risco.
  - Período do Evento.
  - Recomendação de Ação Prática (ex: "Aumente o turno de irrigação para a noite", "Antecipe a colheita das áreas baixas").

### RF-05: Dashboard Web Visual

- **RF-05.1**: Exibir mapa interativo contendo a localização dos talhões/propriedades e seus respectivos status de risco.
- **RF-05.2**: Exibir histórico de alertas disparados e previsão meteorológica para os próximos 7 dias.

## 5. Requisitos Não-Funcionais (RNF)

- **RNF-01 (Desempenho e Concorrência)**: O backend em Go deve utilizar Goroutines para efetuar as requisições de API meteorológica de forma paralela e assíncrona.
- **RNF-02 (Geoprocessamento)**: O banco de dados PostgreSQL deve utilizar a extensão PostGIS para armazenamento e consultas espaciais de alta performance.
- **RNF-03 (Conteinerização e Deploy)**: A aplicação deve ser totalmente conteinerizada via Docker e Docker Compose para execução em ambiente local ou nuvem.
- **RNF-04 (Usabilidade da Mensagem)**: As mensagens no WhatsApp não devem conter termos técnicos complexos (como "pressão atmosférica em hPa" ou "evapotranspiração potencial").

## 6. Arquitetura da Solução

```
[ APIs Meteorológicas ]                    [ WhatsApp Produtor ]
  (Open-Meteo / INMET)                            ▲
           │                                       │ (Notificação HTTP)
           ▼                                       │
┌───────────────────────────────────────────────────┐
│                BACKEND EM GOLANG                   │
│  ┌──────────────┐  ┌────────────┐  ┌────────────┐  │
│  │  Worker Cron │  │ Risk Engine│  │  HTTP API  │  │
│  └──────────────┘  └────────────┘  └────────────┘  │
└───────────────────────┬────────────────────────────┘
                         │
                         ▼
            [ PostgreSQL + PostGIS ]
                         ▲
                         │
              [ Dashboard Frontend ]
                 (Next.js / React)
```

## 7. Escopo do MVP (Plano de Entregas do Semestre)

- **Fase 1 (Semanas 1-3)**: Modelagem do banco PostGIS e desenvolvimento da API REST em Go (CRUD de Produtores e Propriedades).
- **Fase 2 (Semanas 4-6)**: Implementação do Ingestor de Clima em Go (Worker + Goroutines) e do Motor de Regras básico.
- **Fase 3 (Semanas 7-9)**: Integração do backend com a Evolution API/n8n para disparo automático de mensagens no WhatsApp.
- **Fase 4 (Semanas 10-12)**: Construção do Dashboard Web em Next.js com mapa interativo e validação final com dados simulados.
