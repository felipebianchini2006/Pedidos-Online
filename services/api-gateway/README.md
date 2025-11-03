# API Gateway - Sistema de Pedidos Online

API Gateway centralizado para o sistema de pedidos online, fornecendo um ponto de entrada único para todos os microserviços.

## 📋 Funcionalidades

- **Proxy Reverso**: Roteia requisições para os microserviços apropriados
- **CORS**: Configuração flexível de Cross-Origin Resource Sharing
- **Rate Limiting**: Proteção contra abuso com limite de requisições por IP
- **Logging**: Logs estruturados em JSON para análise e monitoramento
- **Retry Logic**: Tentativas automáticas em caso de falha temporária
- **Health Checks**: Monitoramento da saúde de todos os serviços
- **Graceful Shutdown**: Desligamento gracioso garantindo conclusão de requisições

## 🏗️ Arquitetura

```
┌─────────────────┐
│   Frontend      │
│  (React/Vue)    │
└────────┬────────┘
         │
         │ HTTP/HTTPS
         ▼
┌─────────────────────────────────────┐
│         API Gateway (8000)          │
│  ┌──────────────────────────────┐  │
│  │  Middlewares                 │  │
│  │  • Recovery                  │  │
│  │  • CORS                      │  │
│  │  • Logger                    │  │
│  │  • Rate Limiter              │  │
│  └──────────────────────────────┘  │
│  ┌──────────────────────────────┐  │
│  │  Proxy Handler               │  │
│  │  • Header Forwarding         │  │
│  │  • Timeout Control           │  │
│  │  • Retry with Backoff        │  │
│  └──────────────────────────────┘  │
└───────────┬──────────┬──────────────┘
            │          │
    ┌───────┘          └────────┐
    │                           │
    ▼                           ▼
┌────────────┐          ┌────────────┐
│User Service│          │Order Service│
│   (8001)   │          │   (8002)   │
│ PostgreSQL │          │  MongoDB   │
└────────────┘          └────────────┘
```

## 🚀 Rotas

### Informações do Gateway
- `GET /` - Informações sobre o API Gateway

### Health Check
- `GET /health` - Status de saúde agregado de todos os serviços

### User Service
- `POST /api/users/register` - Cadastro de usuário
- `POST /api/users/login` - Login e geração de JWT
- `GET /api/users/profile` - Obter perfil (autenticado)
- `PUT /api/users/profile` - Atualizar perfil (autenticado)

### Order Service
- `POST /api/orders` - Criar novo pedido (autenticado)
- `GET /api/orders` - Listar pedidos do usuário (autenticado)
- `GET /api/orders/:id` - Obter detalhes do pedido (autenticado)
- `PUT /api/orders/:id/status` - Atualizar status do pedido

## 🛠️ Configuração

### Variáveis de Ambiente

Copie o arquivo `.env.example` para `.env` e ajuste conforme necessário:

```bash
cp .env.example .env
```

#### Principais Variáveis

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `PORT` | `8000` | Porta do API Gateway |
| `USER_SERVICE_URL` | `http://user-service:8001` | URL do User Service |
| `ORDER_SERVICE_URL` | `http://order-service:8002` | URL do Order Service |
| `ALLOWED_ORIGINS` | `http://localhost:3000` | Origens permitidas para CORS |
| `RATE_LIMIT_PER_MIN` | `100` | Requisições permitidas por minuto por IP |
| `REQUEST_TIMEOUT` | `30` | Timeout em segundos para requisições |
| `MAX_RETRIES` | `3` | Número de tentativas em caso de falha |
| `ENABLE_DETAILED_LOGS` | `true` | Ativar logs detalhados em JSON |

## 📦 Instalação

### Pré-requisitos

- Go 1.21 ou superior
- Docker e Docker Compose (opcional)

### Desenvolvimento Local

1. **Clone o repositório**
```bash
cd services/api-gateway
```

2. **Instale as dependências**
```bash
go mod download
```

3. **Configure as variáveis de ambiente**
```bash
cp .env.example .env
# Edite o arquivo .env conforme necessário
```

4. **Execute a aplicação**
```bash
go run cmd/main.go
```

A aplicação estará disponível em `http://localhost:8000`

### Build para Produção

```bash
go build -o api-gateway cmd/main.go
./api-gateway
```

## 🐳 Docker

### Build da Imagem

```bash
docker build -t pedidos-online/api-gateway:latest .
```

### Executar Container

```bash
docker run -d \
  --name api-gateway \
  -p 8000:8000 \
  --env-file .env \
  pedidos-online/api-gateway:latest
```

### Docker Compose

```yaml
version: '3.8'

services:
  api-gateway:
    build: ./services/api-gateway
    ports:
      - "8000:8000"
    environment:
      PORT: 8000
      USER_SERVICE_URL: http://user-service:8001
      ORDER_SERVICE_URL: http://order-service:8002
      ALLOWED_ORIGINS: http://localhost:3000
      RATE_LIMIT_PER_MIN: 100
      REQUEST_TIMEOUT: 30
      MAX_RETRIES: 3
      ENABLE_DETAILED_LOGS: true
    depends_on:
      - user-service
      - order-service
    networks:
      - pedidos-network
    restart: unless-stopped

networks:
  pedidos-network:
    driver: bridge
```

## 🔒 Segurança

### CORS
O API Gateway implementa CORS para proteger contra requisições não autorizadas de origens diferentes:
- Métodos permitidos: GET, POST, PUT, DELETE, PATCH, OPTIONS
- Headers permitidos: Origin, Content-Type, Accept, Authorization
- Credenciais: Permitidas (cookies, JWT)

### Rate Limiting
Proteção contra abuso com limite de requisições por IP:
- Padrão: 100 requisições por minuto
- Resposta: HTTP 429 Too Many Requests
- Headers informativos: X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset

### Headers de Segurança
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`

## 📊 Monitoramento

### Health Check

O endpoint `/health` verifica a saúde de todos os serviços:

```bash
curl http://localhost:8000/health
```

Resposta:
```json
{
  "status": "healthy",
  "services": {
    "user-service": {
      "status": "healthy",
      "latency_ms": 5
    },
    "order-service": {
      "status": "healthy",
      "latency_ms": 3
    }
  },
  "timestamp": "2024-01-20T15:04:05Z"
}
```

Status possíveis:
- `healthy`: Todos os serviços operacionais
- `degraded`: Algum serviço com problemas
- `unhealthy`: Múltiplos serviços indisponíveis

### Logs

Logs estruturados em JSON para facilitar análise:

```json
{
  "timestamp": "2024-01-20T15:04:05.123456Z",
  "method": "POST",
  "path": "/api/orders",
  "status_code": 201,
  "latency_ms": 245,
  "ip": "192.168.1.100",
  "user_agent": "Mozilla/5.0..."
}
```

## ⚙️ Middleware

### 1. Recovery
Recupera de panics e retorna erro 500 controlado

### 2. CORS
Configuração flexível de Cross-Origin Resource Sharing

### 3. Logger
Logging estruturado com informações detalhadas de cada requisição

### 4. Rate Limiter
Controle de taxa de requisições por IP com limpeza automática de cache

## 🔄 Retry Logic

O proxy implementa retry com exponential backoff:
- **Primeira tentativa**: Imediata
- **Segunda tentativa**: 100ms de delay
- **Terceira tentativa**: 200ms de delay
- **Quarta tentativa**: 400ms de delay

Códigos de status que acionam retry:
- 502 Bad Gateway
- 503 Service Unavailable
- 504 Gateway Timeout

## 🚦 Códigos de Status

| Código | Descrição |
|--------|-----------|
| 200 | Sucesso |
| 201 | Criado |
| 400 | Requisição inválida |
| 401 | Não autenticado |
| 403 | Não autorizado |
| 404 | Não encontrado |
| 429 | Muitas requisições (rate limit) |
| 500 | Erro interno |
| 502 | Bad Gateway |
| 503 | Serviço indisponível |
| 504 | Gateway timeout |

## 🧪 Testes

### Executar Testes
```bash
go test ./... -v
```

### Cobertura de Testes
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 📝 Exemplos de Uso

### Cadastro de Usuário
```bash
curl -X POST http://localhost:8000/api/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "senha123",
    "name": "João Silva",
    "phone": "11999999999"
  }'
```

### Login
```bash
curl -X POST http://localhost:8000/api/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "senha123"
  }'
```

### Criar Pedido (autenticado)
```bash
curl -X POST http://localhost:8000/api/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <seu-token-jwt>" \
  -d '{
    "items": [
      {
        "product_id": "123",
        "product_name": "Pizza Margherita",
        "quantity": 2,
        "price": 35.00
      }
    ],
    "address": {
      "street": "Rua das Flores",
      "number": "123",
      "city": "São Paulo",
      "state": "SP",
      "zip_code": "01234-567"
    }
  }'
```

## 🐛 Troubleshooting

### Gateway não inicia
- Verifique se a porta 8000 está disponível
- Confirme que as variáveis de ambiente estão configuradas
- Verifique os logs para mensagens de erro

### Erro 503 Service Unavailable
- Verifique se os microserviços (User e Order) estão rodando
- Confirme as URLs dos serviços nas variáveis de ambiente
- Execute `/health` para verificar o status dos serviços

### Erro 429 Too Many Requests
- Você atingiu o limite de rate limiting
- Aguarde o tempo especificado no header `Retry-After`
- Ajuste `RATE_LIMIT_PER_MIN` se necessário

## 📚 Tecnologias

- **[Fiber](https://gofiber.io/)**: Framework HTTP rápido e expressivo
- **Go 1.21**: Linguagem de programação
- **Docker**: Containerização
- **Alpine Linux**: Imagem base leve para produção

## 🤝 Contribuindo

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo [LICENSE](../../LICENSE) para mais detalhes.

## 👥 Autores

- Sistema de Pedidos Online Team

## 📞 Suporte

Para suporte, envie um email para support@pedidosonline.com ou abra uma issue no GitHub.
