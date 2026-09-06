# CapitalHub - Plataforma de Finanças Pessoais & Gestão de Investimentos

Plataforma completa e moderna de controle financeiro pessoal e consolidação de investimentos, desenvolvida com **Golang (Gin + GORM)** no backend em **Arquitetura em Camadas (Clean Architecture)**, **React (Vite + TypeScript + Tailwind CSS)** no frontend e orquestração total com **Docker & Docker Compose**.

---

## 🚀 Arquitetura & Stack Tecnológica

### Backend (Golang)
- **Framework Web**: `gin-gonic/gin`
- **Banco de Dados & ORM**: PostgreSQL 16 + `GORM` com `AutoMigrate` e seeds automáticas
- **Cache & Filas**: Redis 7
- **Segurança**: Autenticação **JWT (Access Token + Refresh Token)** com hash **Bcrypt**
- **Precisão Monetária**: `shopspring/decimal` (ponto fixo para evitar erros de floating point)
- **Cotações de Mercado**: Integração com APIs financeiras (Brapi para B3/FIIs e CoinGecko para Cripto) com cache em Redis

### Frontend (React)
- **Engine**: React 18 + Vite + TypeScript
- **Estilização & UI**: Tailwind CSS + Lucide Icons (Design moderno e responsivo)
- **Gráficos & Visualização**: Recharts (Fluxo de caixa, distribuição de carteira e evolução)
- **Estado Global**: Zustand (Auth State e Modo Privacidade para ocultar valores)
- **Comunicação HTTP**: Axios com interceptors para envio de Bearer Token e renovação automática de Refresh Token

### Orquestração (Docker)
- `docker-compose.yml` orquestrando:
  - `postgres`: Banco de dados relacional (porta 5433)
  - `redis`: Cache de cotações e sessões (porta 6380)
  - `backend`: API REST Golang compilada em imagem multi-stage leve (porta 8081)
  - `frontend`: Aplicação React servida via Nginx otimizado (porta 3001)

---

## 🧪 Test-Driven Development (TDD)

O projeto foi desenvolvido seguindo a metodologia **TDD**:

### Testes do Backend (Go)
Para rodar a suíte completa de testes unitários:
```bash
make test-backend
# ou
cd backend && go test -v -race ./...
```

**Cobertura de Testes**:
- **Cálculo de Preço Médio Ponderado Móvel**: Compras sucessivas com taxas, vendas parciais (sem alterar PM) e apuração de lucro/prejuízo.
- **Cartões de Crédito**: Projeção de faturas mensais, melhor dia de compra (fechamento vs vencimento) e parcelamentos futuros com ajuste exato de centavos.
- **Orçamentos**: Monitoramento de consumo, alertas em 80% e aviso de estouro.
- **Multi-Moeda**: Valoração de ativos internacionais (USD/Cripto) com conversão cambial para BRL.
- **Segurança**: Geração, expiração e validação de tokens JWT (Access/Refresh) e senhas Bcrypt.
- **Use Cases**: Testes unitários com mocks para fluxos de autenticação, transações e ordens.

### Testes do Frontend (Vitest)
Para rodar os testes do frontend:
```bash
make test-frontend
# ou
cd frontend && npm test
```

---

## 🐳 Como Executar com Docker Compose

### 1. Clonar e configurar variáveis de ambiente
```bash
cp .env.example .env
```

### 2. Subir todos os serviços
```bash
make up
# ou
docker compose up -d --build
```

### 3. Acessar a aplicação
- **Frontend Web**: [http://localhost:3001](http://localhost:3001)
- **API Backend**: [http://localhost:8081/api/v1/health](http://localhost:8081/api/v1/health)

---

## 📂 Estrutura de Pastas

```
.
├── backend/
│   ├── cmd/api/main.go             # Entrypoint da API REST
│   ├── internal/
│   │   ├── domain/                 # Entidades puras e regras de domínio (Account, Transaction, Position, etc.)
│   │   ├── usecase/                # Casos de uso (Auth, Finance, Cards, Investments) e Interfaces
│   │   ├── adapter/
│   │   │   ├── http/               # Gin Routers, Handlers, Middlewares e DTOs
│   │   │   ├── repository/         # Implementação GORM Postgres e Redis
│   │   │   └── external/           # Clientes Brapi e CoinGecko
│   │   └── config/                 # Carregador de configurações
│   ├── pkg/                        # Pacotes utilitários (JWT Maker, Bcrypt Hasher)
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── features/               # Módulos por Domínio (auth, dashboard, finances, credit-cards, investments)
│   │   ├── components/layout/      # Sidebar, Header e Layout principal
│   │   ├── services/               # Cliente Axios com interceptors
│   │   ├── stores/                 # Zustand (Auth, Privacy Mode)
│   │   ├── utils/                  # Formatadores de moeda (BRL, USD) e percentuais
│   │   ├── routes.tsx              # Rotas públicas e privadas
│   │   └── App.tsx
│   ├── nginx.conf
│   └── Dockerfile
├── docker-compose.yml
├── Makefile
└── README.md
```
# inv
