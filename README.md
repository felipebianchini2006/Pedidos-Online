# Sistema de Pedidos Online

> Sistema completo de gerenciamento de pedidos online construído com arquitetura de microserviços

[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/felipebianchini2006/Pedidos-Online)
[![Coverage](https://img.shields.io/badge/coverage-70%25-yellow)](https://github.com/felipebianchini2006/Pedidos-Online)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org)
[![React Version](https://img.shields.io/badge/React-18+-61DAFB?logo=react)](https://reactjs.org)

---

## Índice

- [Visão Geral](#visão-geral)
- [Arquitetura](#arquitetura)
- [Tecnologias Utilizadas](#tecnologias-utilizadas)
- [Pré-requisitos](#pré-requisitos)
- [Instalação e Execução](#instalação-e-execução)
- [Desenvolvimento Local](#desenvolvimento-local)
- [Estrutura do Projeto](#estrutura-do-projeto)
- [API Endpoints](#api-endpoints)
- [Testes](#testes)
- [Deploy](#deploy)
- [Troubleshooting](#troubleshooting)
- [Contribuindo](#contribuindo)
- [Licença](#licença)

---

## Visão Geral

O **Sistema de Pedidos Online** é uma plataforma completa que permite usuários criar contas, realizar pedidos e acompanhar entregas em tempo real. Construído com arquitetura de microserviços, o sistema oferece escalabilidade, manutenibilidade e separação clara de responsabilidades.

### Problema que Resolve

- Centraliza o gerenciamento de pedidos online
- Fornece autenticação e autorização seguras
- Notifica usuários sobre status de pedidos em tempo real
- Escala independentemente por serviço conforme demanda

### Principais Funcionalidades

- **Autenticação JWT**: Login e cadastro seguros com tokens JWT
- **Gerenciamento de Pedidos**: Criação, listagem e acompanhamento de pedidos
- **Notificações**: E-mails automáticos para eventos de pedidos
- **Perfil de Usuário**: Atualização de dados pessoais
- **Rastreamento em Tempo Real**: Acompanhamento do status do pedido
- **Interface Responsiva**: Frontend moderno construído com React e Tailwind CSS

---

## Arquitetura

O sistema utiliza uma arquitetura de microserviços com os seguintes componentes:

```
┌─────────────────┐
│   Frontend      │
│   (React)       │
│   Port: 3000    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  API Gateway    │
│   (Go/Fiber)    │
│   Port: 8000    │
└────────┬────────┘
         │
         ├──────────────┬──────────────┐
         ▼              ▼              ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│User Service │ │Order Service│ │Notification │
│   (Go)      │ │   (Go)      │ │Service (Go) │
│Port: 8001   │ │Port: 8002   │ │Port: 8003   │
└──────┬──────┘ └──────┬──────┘ └──────┬──────┘
       │               │               │
       ▼               ▼               ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│ PostgreSQL  │ │  MongoDB    │ │  RabbitMQ   │
│Port: 5432   │ │Port: 27017  │ │Port: 5672   │
└─────────────┘ └─────────────┘ └─────────────┘
```

### Descrição dos Serviços

#### API Gateway (Port: 8000)
- Ponto de entrada único para todas as requisições
- Roteamento para microserviços
- CORS e rate limiting
- Agregação de respostas

#### User Service (Port: 8001)
- Autenticação e autorização (JWT)
- Gerenciamento de usuários
- CRUD de perfis
- Banco: PostgreSQL

#### Order Service (Port: 8002)
- Gerenciamento de pedidos
- Carrinho de compras
- Atualização de status
- Banco: MongoDB
- Publica eventos no RabbitMQ

#### Notification Service (Port: 8003)
- Envio de e-mails transacionais
- Consome eventos do RabbitMQ
- Templates personalizados
- Integração SMTP

---

## Tecnologias Utilizadas

### Backend
- **Go 1.21+**: Linguagem principal dos microserviços
- **Fiber**: Framework HTTP de alta performance
- **PostgreSQL**: Banco relacional para dados de usuários
- **MongoDB**: Banco NoSQL para pedidos
- **RabbitMQ**: Message broker para comunicação assíncrona
- **JWT**: Autenticação stateless
- **Docker**: Containerização

### Frontend
- **React 18**: Biblioteca UI
- **Vite**: Build tool e dev server
- **React Router v6**: Roteamento
- **Axios**: Cliente HTTP
- **Tailwind CSS**: Framework CSS utility-first
- **Context API**: Gerenciamento de estado

### DevOps
- **Docker Compose**: Orquestração local
- **Make**: Automação de tarefas
- **GitHub Actions**: CI/CD
- **Nginx**: Servidor web para frontend

---

## Pré-requisitos

Antes de começar, certifique-se de ter instalado:

- **Docker** 20.10+
- **Docker Compose** 2.0+
- **Node.js** 18+ (para desenvolvimento local)
- **Go** 1.21+ (para desenvolvimento local)
- **Make** (opcional, mas recomendado)
- **Git**

### Verificar Instalações

```bash
docker --version
docker-compose --version
node --version
go version
make --version
```

---

## Instalação e Execução

### 1. Clone o Repositório

```bash
git clone https://github.com/felipebianchini2006/Pedidos-Online.git
cd Pedidos-Online
```

### 2. Configure as Variáveis de Ambiente

```bash
# Criar arquivo .env na raiz (se necessário)
# As variáveis já estão configuradas no docker-compose.yml
```

### 3. Inicie os Serviços

```bash
# Usando Make (recomendado)
make up

# Ou usando Docker Compose diretamente
docker-compose up -d

# Para ver os logs
docker-compose logs -f
```

### 4. Inicialize o Banco de Dados

```bash
# Executar migrations do PostgreSQL
make migrate-up

# Ou executar script manualmente
docker-compose exec user-service ./user-service migrate
```

### 5. Acesse a Aplicação

- **Frontend**: [http://localhost:3000](http://localhost:3000)
- **API Gateway**: [http://localhost:8000](http://localhost:8000)
- **RabbitMQ Management**: [http://localhost:15672](http://localhost:15672) (user: `guest`, password: `guest`)

### 6. Parar os Serviços

```bash
make down
# ou
docker-compose down
```

---

## Desenvolvimento Local

### Rodar Serviços Individualmente

#### User Service

```bash
cd services/user-service
go mod download
go run cmd/main.go
```

#### Order Service

```bash
cd services/order-service
go mod download
go run cmd/main.go
```

#### Notification Service

```bash
cd services/notification-service
go mod download
go run cmd/main.go
```

#### Frontend

```bash
cd frontend
npm install
npm run dev
```

### Hot Reload

Todos os serviços estão configurados com hot reload em modo de desenvolvimento:

```bash
# Subir com modo de desenvolvimento
docker-compose -f docker-compose.dev.yml up -d
```

### Comandos Úteis do Makefile

```bash
make help              # Mostrar todos os comandos disponíveis
make build             # Build de todos os serviços
make up                # Subir todos os serviços
make down              # Parar todos os serviços
make logs              # Ver logs de todos os serviços
make restart           # Reiniciar todos os serviços
make clean             # Limpar volumes e containers
make test              # Rodar todos os testes
make test-integration  # Rodar testes de integração
make migrate-up        # Executar migrations
make migrate-down      # Reverter migrations
make seed              # Popular banco com dados de teste
```

---

## Estrutura do Projeto

```
Pedidos-Online/
├── frontend/                  # Aplicação React
│   ├── src/
│   │   ├── components/       # Componentes React
│   │   ├── context/          # Context API (AuthContext)
│   │   ├── hooks/            # Custom hooks
│   │   ├── services/         # Chamadas API
│   │   └── utils/            # Funções auxiliares
│   ├── Dockerfile            # Dockerfile de produção
│   └── package.json          # Dependências npm
│
├── services/
│   ├── api-gateway/          # API Gateway (Go/Fiber)
│   │   ├── cmd/              # Entry point
│   │   └── internal/         # Código interno
│   │       ├── config/       # Configurações
│   │       ├── middleware/   # Middlewares HTTP
│   │       └── proxy/        # Proxy reverso
│   │
│   ├── user-service/         # Serviço de Usuários (Go)
│   │   ├── cmd/              # Entry point
│   │   ├── internal/         # Código interno
│   │   │   ├── handler/      # HTTP handlers
│   │   │   ├── model/        # Modelos de dados
│   │   │   ├── repository/   # Camada de dados
│   │   │   └── service/      # Lógica de negócio
│   │   ├── migrations/       # Migrations SQL
│   │   └── pkg/              # Código reutilizável
│   │
│   ├── order-service/        # Serviço de Pedidos (Go)
│   │   ├── cmd/              # Entry point
│   │   ├── internal/         # Código interno
│   │   │   ├── handler/      # HTTP handlers
│   │   │   ├── model/        # Modelos de dados
│   │   │   ├── repository/   # Camada de dados (MongoDB)
│   │   │   ├── queue/        # RabbitMQ publisher
│   │   │   └── service/      # Lógica de negócio
│   │   └── pkg/              # Código reutilizável
│   │
│   └── notification-service/ # Serviço de Notificações (Go)
│       ├── cmd/              # Entry point
│       └── internal/         # Código interno
│           ├── config/       # Configurações
│           ├── email/        # Templates e envio
│           ├── queue/        # RabbitMQ consumer
│           └── handler/      # Event handlers
│
├── tests/
│   └── integration/          # Testes de integração
│
├── scripts/                  # Scripts auxiliares
│   ├── init-db.sh           # Inicializar banco
│   ├── seed-data.ps1        # Popular dados de teste
│   └── check-services.sh    # Health check
│
├── docker-compose.yml        # Orquestração Docker
├── docker-compose.dev.yml    # Ambiente de desenvolvimento
├── docker-compose.test.yml   # Ambiente de testes
├── Makefile                  # Automação de tarefas
└── README.md                 # Este arquivo
```

---

## API Endpoints

### User Service (via API Gateway `/api/users`)

| Método | Endpoint            | Descrição                | Auth Necessário |
|--------|---------------------|--------------------------|-----------------|
| POST   | `/api/v1/register`  | Cadastrar novo usuário   | Não             |
| POST   | `/api/v1/login`     | Login e obter JWT        | Não             |
| GET    | `/api/v1/profile`   | Obter perfil do usuário  | Sim             |
| PUT    | `/api/v1/profile`   | Atualizar perfil         | Sim             |

### Order Service (via API Gateway `/api/orders`)

| Método | Endpoint                    | Descrição                | Auth Necessário |
|--------|-----------------------------|--------------------------|-----------------|
| POST   | `/api/v1/orders`            | Criar novo pedido        | Sim             |
| GET    | `/api/v1/orders`            | Listar pedidos do usuário| Sim             |
| GET    | `/api/v1/orders/:id`        | Obter detalhes do pedido | Sim             |
| PUT    | `/api/v1/orders/:id/status` | Atualizar status         | Sim             |

### Autenticação

Para endpoints autenticados, inclua o header:

```
Authorization: Bearer <seu_jwt_token>
```

### Exemplo de Requisição

```bash
# Cadastro
curl -X POST http://localhost:8000/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "usuario@example.com",
    "password": "senha123",
    "name": "João Silva",
    "phone": "+5511999999999"
  }'

# Login
curl -X POST http://localhost:8000/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "usuario@example.com",
    "password": "senha123"
  }'

# Criar Pedido (autenticado)
curl -X POST http://localhost:8000/api/v1/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "items": [
      {
        "product_id": "123",
        "product_name": "Produto A",
        "quantity": 2,
        "price": 50.00
      }
    ],
    "address": {
      "street": "Rua das Flores",
      "number": "100",
      "city": "São Paulo",
      "state": "SP",
      "zip_code": "01234-567"
    }
  }'
```

---

## Testes

### Testes Unitários

#### Backend (Go)

```bash
# Rodar todos os testes
make test

# Ou por serviço
cd services/user-service
go test ./...

# Com coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### Frontend (React)

```bash
cd frontend
npm test

# Com coverage
npm test -- --coverage

# Watch mode
npm test -- --watch
```

### Testes de Integração

```bash
# Rodar testes de integração
make test-integration

# Ou manualmente
cd tests/integration
go test -v ./...
```

### Testes E2E

```bash
# Subir ambiente de teste
docker-compose -f docker-compose.test.yml up -d

# Rodar testes
./tests/run-integration-tests.sh

# Derrubar ambiente
docker-compose -f docker-compose.test.yml down
```

### Cobertura de Testes

O projeto mantém uma cobertura mínima de **70%** de testes unitários.

```bash
# Verificar cobertura
make coverage
```

---

## Deploy

### Deploy em Produção

#### 1. Preparar Variáveis de Ambiente

Crie um arquivo `.env.production` com as variáveis de produção:

```env
# User Service
JWT_SECRET=your_super_secret_key_production
DB_HOST=postgres-production.example.com
DB_PASSWORD=secure_password

# Order Service
MONGO_URI=mongodb://mongo-production.example.com:27017
RABBITMQ_URL=amqp://user:password@rabbitmq-production.example.com:5672/

# Notification Service
SMTP_HOST=smtp.sendgrid.net
SMTP_USER=apikey
SMTP_PASSWORD=your_sendgrid_api_key

# Frontend
VITE_API_URL=https://api.seudominio.com
```

#### 2. Build das Imagens

```bash
# Build de produção
docker-compose -f docker-compose.yml build

# Tag das imagens
docker tag pedidos-online-user-service:latest registry.example.com/user-service:1.0.0
docker tag pedidos-online-order-service:latest registry.example.com/order-service:1.0.0
docker tag pedidos-online-notification-service:latest registry.example.com/notification-service:1.0.0
docker tag pedidos-online-frontend:latest registry.example.com/frontend:1.0.0
```

#### 3. Push para Registry

```bash
docker push registry.example.com/user-service:1.0.0
docker push registry.example.com/order-service:1.0.0
docker push registry.example.com/notification-service:1.0.0
docker push registry.example.com/frontend:1.0.0
```

#### 4. Deploy

**Usando Docker Swarm:**

```bash
docker stack deploy -c docker-compose.yml pedidos-online
```

**Usando Kubernetes:**

```bash
kubectl apply -f k8s/
```

### Checklist de Produção

- [ ] Configurar HTTPS/TLS
- [ ] Configurar rate limiting no API Gateway
- [ ] Configurar backups automáticos dos bancos de dados
- [ ] Configurar monitoramento (Prometheus + Grafana)
- [ ] Configurar logs centralizados (ELK Stack)
- [ ] Configurar health checks
- [ ] Revisar e aplicar security headers
- [ ] Configurar firewall e network policies
- [ ] Implementar disaster recovery plan
- [ ] Documentar runbooks para incidentes

---

## Troubleshooting

### Problema: Serviços não sobem

**Solução:**

```bash
# Verificar logs
docker-compose logs

# Verificar portas em uso
netstat -tulpn | grep LISTEN

# Limpar containers antigos
docker-compose down -v
docker system prune -a
```

### Problema: Erro de conexão com banco de dados

**Solução:**

```bash
# Verificar se PostgreSQL está rodando
docker-compose ps postgres

# Verificar logs do banco
docker-compose logs postgres

# Recriar volume do banco
docker-compose down -v
docker-compose up -d postgres
```

### Problema: Frontend não conecta com API

**Solução:**

- Verificar se `VITE_API_URL` está configurado corretamente
- Verificar CORS no API Gateway
- Verificar se API Gateway está rodando na porta 8000

```bash
# Testar API diretamente
curl http://localhost:8000/health
```

### Problema: Notificações não são enviadas

**Solução:**

```bash
# Verificar RabbitMQ
docker-compose logs rabbitmq

# Verificar se serviço de notificação está consumindo mensagens
docker-compose logs notification-service

# Acessar RabbitMQ Management
# http://localhost:15672
# Verificar filas e mensagens
```

### Problema: JWT token inválido

**Solução:**

- Verificar se `JWT_SECRET` é o mesmo em User Service e Order Service
- Verificar se token não expirou (válido por 24h)
- Limpar localStorage/sessionStorage no frontend

### Problema: Migrations falhando

**Solução:**

```bash
# Resetar banco de dados
make reset-db

# Rodar migrations novamente
make migrate-up
```
---

**Feito com Go e React** | **2025**