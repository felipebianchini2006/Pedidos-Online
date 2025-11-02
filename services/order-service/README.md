# 📦 Order Service

Microserviço de gerenciamento de pedidos do sistema Pedidos Online.

## 🎯 Funcionalidades

- ✅ Criar pedidos com múltiplos itens
- ✅ Listar pedidos do usuário autenticado
- ✅ Obter detalhes de um pedido específico
- ✅ Atualizar status do pedido com validação de transições
- ✅ Cálculo automático de total
- ✅ Publicação de eventos no RabbitMQ (order.created, order.updated)
- ✅ Autenticação via JWT (compatível com User Service)
- ✅ Armazenamento no MongoDB

## 🏗️ Arquitetura

```
order-service/
├── cmd/
│   └── main.go                    # Entry point da aplicação
├── internal/
│   ├── config/
│   │   └── config.go              # Configurações e env vars
│   ├── model/
│   │   └── order.go               # Modelos de dados
│   ├── repository/
│   │   └── order_repository.go   # Camada de dados (MongoDB)
│   ├── handler/
│   │   └── order_handler.go      # Handlers HTTP
│   └── middleware/
│       └── auth.go                # Middleware de autenticação JWT
├── pkg/
│   ├── jwt/
│   │   └── jwt.go                 # Utilitário JWT
│   └── rabbitmq/
│       └── publisher.go           # Publisher RabbitMQ
├── .env.example                   # Template de variáveis de ambiente
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## 🚀 Como Executar

### Pré-requisitos

- Go 1.24+
- MongoDB (local ou Docker)
- RabbitMQ (local ou Docker)
- User Service rodando (para autenticação)

### 1. Configurar Variáveis de Ambiente

```bash
# Copiar template
cp .env.example .env

# Editar valores
nano .env
```

**Importante**: `JWT_SECRET` deve ser o **mesmo** do User Service!

### 2. Instalar Dependências

```bash
go mod download
```

### 3. Executar

```bash
# Desenvolvimento
go run cmd/main.go

# Ou compilar e executar
go build -o main cmd/main.go
./main
```

O serviço estará disponível em `http://localhost:8002`

## 📡 Endpoints da API

### Públicos

```http
GET /               # Info do serviço
GET /health         # Health check
```

### Protegidos (requer JWT token)

```http
POST   /api/v1/orders           # Criar pedido
GET    /api/v1/orders           # Listar pedidos do usuário
GET    /api/v1/orders/:id       # Obter pedido específico
PUT    /api/v1/orders/:id/status # Atualizar status do pedido
```

## 📝 Exemplos de Uso

### 1. Criar Pedido

```bash
curl -X POST http://localhost:8002/api/v1/orders \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {
        "product_id": "prod123",
        "product_name": "Pizza Margherita",
        "quantity": 2,
        "price": 35.90
      },
      {
        "product_id": "prod456",
        "product_name": "Refrigerante 2L",
        "quantity": 1,
        "price": 8.50
      }
    ],
    "address": {
      "street": "Rua das Flores",
      "number": "123",
      "city": "São Paulo",
      "state": "SP",
      "zip_code": "01234567",
      "complement": "Apto 101"
    }
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "674123abc456def789012345",
    "user_id": "user-uuid",
    "items": [...],
    "total_amount": 80.30,
    "status": "pending",
    "address": {...},
    "created_at": "2025-11-02T10:30:00Z",
    "updated_at": "2025-11-02T10:30:00Z"
  },
  "message": "Pedido criado com sucesso"
}
```

### 2. Listar Pedidos

```bash
curl -X GET http://localhost:8002/api/v1/orders \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 3. Obter Pedido Específico

```bash
curl -X GET http://localhost:8002/api/v1/orders/674123abc456def789012345 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 4. Atualizar Status

```bash
curl -X PUT http://localhost:8002/api/v1/orders/674123abc456def789012345/status \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "confirmed"
  }'
```

## 📊 Status de Pedidos

O sistema gerencia os seguintes status com validação de transições:

```
pending → confirmed → preparing → shipped → delivered
    ↓         ↓           ↓
cancelled  cancelled  cancelled
```

### Status Disponíveis

- `pending`: Pedido criado, aguardando confirmação
- `confirmed`: Pedido confirmado pelo sistema
- `preparing`: Pedido em preparação
- `shipped`: Pedido enviado para entrega
- `delivered`: Pedido entregue (estado final)
- `cancelled`: Pedido cancelado (estado final)

### Validação de Transições

O sistema valida automaticamente se uma transição de status é válida:

✅ `pending → confirmed` (válido)
✅ `confirmed → preparing` (válido)
❌ `pending → delivered` (inválido)
❌ `delivered → confirmed` (inválido)

## 🐰 Eventos RabbitMQ

O serviço publica eventos no exchange `orders`:

### order.created
Publicado quando um novo pedido é criado.

```json
{
  "event_type": "order.created",
  "order_id": "674123abc456def789012345",
  "user_id": "user-uuid",
  "status": "pending",
  "total": 80.30,
  "timestamp": "2025-11-02T10:30:00Z"
}
```

### order.updated
Publicado quando o status de um pedido é atualizado.

```json
{
  "event_type": "order.updated",
  "order_id": "674123abc456def789012345",
  "user_id": "user-uuid",
  "status": "confirmed",
  "total": 80.30,
  "timestamp": "2025-11-02T10:35:00Z"
}
```

## 🍃 MongoDB

### Database
- Nome: `orders_db` (configurável via `MONGO_DATABASE`)

### Collections
- `orders`: Armazena todos os pedidos

### Índices Criados Automaticamente

1. `user_id` - Para buscar pedidos por usuário
2. `created_at` (DESC) - Para ordenação
3. `status` - Para filtrar por status
4. `user_id + created_at` (composto) - Otimiza busca de pedidos do usuário

## 🔐 Autenticação

O Order Service **não gera** tokens JWT, apenas **valida** tokens gerados pelo User Service.

### Como obter um token:

1. Faça login no User Service:
```bash
curl -X POST http://localhost:8001/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "yourpassword"
  }'
```

2. Use o token retornado no header `Authorization`:
```bash
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## 🛠️ Desenvolvimento

### Executar com hot reload (Air)

```bash
# Instalar Air
go install github.com/cosmtrek/air@latest

# Executar
air
```

### Executar testes

```bash
go test ./...
```

### Build de produção

```bash
CGO_ENABLED=0 GOOS=linux go build -o order-service cmd/main.go
```

## 🐳 Docker

```bash
# Build
docker build -t order-service:latest .

# Run
docker run -d \
  --name order-service \
  -p 8002:8002 \
  --env-file .env \
  order-service:latest
```

## 📦 Dependências

- **Fiber**: Framework web HTTP
- **MongoDB Driver**: Cliente MongoDB oficial
- **RabbitMQ**: Cliente AMQP
- **JWT**: Validação de tokens
- **Validator**: Validação de structs
- **Godotenv**: Carregamento de .env

## 🔧 Troubleshooting

### Erro: "JWT_SECRET é obrigatório"
Certifique-se de definir `JWT_SECRET` no `.env` com o **mesmo valor** do User Service.

### Erro: "Erro ao conectar ao MongoDB"
Verifique se MongoDB está rodando e a `MONGO_URI` está correta.

### Erro: "Erro ao conectar ao RabbitMQ"
Verifique se RabbitMQ está rodando e a `RABBITMQ_URL` está correta.

### Health check retorna "unhealthy"
Execute `GET /health` para ver qual componente está com problema (MongoDB ou RabbitMQ).

## 📚 Referências

- [Fiber Documentation](https://docs.gofiber.io/)
- [MongoDB Go Driver](https://www.mongodb.com/docs/drivers/go/current/)
- [RabbitMQ Tutorials](https://www.rabbitmq.com/tutorials/)
- [JWT Best Practices](https://jwt.io/introduction)

## 📄 Licença

MIT

---

**Order Service** - Parte do sistema Pedidos Online  
Desenvolvido com ❤️ usando Go
