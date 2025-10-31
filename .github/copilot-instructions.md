# Instruções do Projeto - Sistema de Pedidos Online

## Visão Geral do Projeto

Sistema de pedidos online construído com arquitetura de microserviços, permitindo usuários criar contas, fazer pedidos e acompanhar entregas em tempo real.

## Arquitetura

### Microserviços

#### 1. User Service (Serviço de Usuários)
- **Responsabilidade**: Autenticação, autorização e gerenciamento de usuários
- **Banco de Dados**: PostgreSQL
- **Porta**: 8001
- **Endpoints**:
  - `POST /api/v1/register` - Cadastro de usuário
  - `POST /api/v1/login` - Login e geração de JWT
  - `GET /api/v1/profile` - Obter perfil do usuário (autenticado)
  - `PUT /api/v1/profile` - Atualizar perfil (autenticado)
- **Autenticação**: JWT (token válido por 24 horas)
- **Modelo de Dados**:
  ```go
  type User struct {
      ID        uuid.UUID `json:"id"`
      Email     string    `json:"email"`
      Password  string    `json:"-"` // nunca retornar no JSON
      Name      string    `json:"name"`
      Phone     string    `json:"phone"`
      CreatedAt time.Time `json:"created_at"`
      UpdatedAt time.Time `json:"updated_at"`
  }
  ```

#### 2. Order Service (Serviço de Pedidos)
- **Responsabilidade**: Gerenciamento de pedidos e carrinho
- **Banco de Dados**: MongoDB
- **Porta**: 8002
- **Endpoints**:
  - `POST /api/v1/orders` - Criar novo pedido (autenticado)
  - `GET /api/v1/orders` - Listar pedidos do usuário (autenticado)
  - `GET /api/v1/orders/:id` - Obter detalhes do pedido (autenticado)
  - `PUT /api/v1/orders/:id/status` - Atualizar status do pedido
- **Status de Pedido**: pending, confirmed, preparing, shipped, delivered, cancelled
- **Eventos RabbitMQ**: Publica eventos `order.created` e `order.updated`
- **Modelo de Dados**:
  ```go
  type Order struct {
      ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
      UserID      string             `bson:"user_id" json:"user_id"`
      Items       []OrderItem        `bson:"items" json:"items"`
      TotalAmount float64            `bson:"total_amount" json:"total_amount"`
      Status      string             `bson:"status" json:"status"`
      Address     Address            `bson:"address" json:"address"`
      CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
      UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
  }

  type OrderItem struct {
      ProductID   string  `bson:"product_id" json:"product_id"`
      ProductName string  `bson:"product_name" json:"product_name"`
      Quantity    int     `bson:"quantity" json:"quantity"`
      Price       float64 `bson:"price" json:"price"`
  }

  type Address struct {
      Street     string `bson:"street" json:"street"`
      Number     string `bson:"number" json:"number"`
      City       string `bson:"city" json:"city"`
      State      string `bson:"state" json:"state"`
      ZipCode    string `bson:"zip_code" json:"zip_code"`
      Complement string `bson:"complement" json:"complement"`
  }
  ```

#### 3. Notification Service (Serviço de Notificações)
- **Responsabilidade**: Envio de notificações por e-mail
- **Porta**: 8003
- **Fila**: Consome eventos do RabbitMQ (exchange: `orders`, queues: `order.created`, `order.updated`)
- **Provedor de E-mail**: SMTP (configurável para SendGrid, Mailgun, etc.)
- **Templates de E-mail**:
  - Pedido criado: Confirmação com número do pedido e itens
  - Pedido atualizado: Notificação de mudança de status

### API Gateway
- **Porta**: 8000
- **Responsabilidade**: Roteamento centralizado, CORS, rate limiting
- **Rotas**:
  - `/api/users/*` → User Service (8001)
  - `/api/orders/*` → Order Service (8002)

### Frontend (React)
- **Framework**: React 18 + Vite
- **Roteamento**: React Router v6
- **Estado Global**: Context API
- **HTTP Client**: Axios com interceptors
- **Estilização**: Tailwind CSS
- **Páginas**:
  - `/login` - Login
  - `/register` - Cadastro
  - `/` - Lista de pedidos (protegida)
  - `/orders/new` - Criar novo pedido (protegida)
  - `/orders/:id` - Detalhes do pedido (protegida)
  - `/profile` - Perfil do usuário (protegida)

## Padrões de Código

### Go (Backend)

#### Estrutura de Pastas
- Seguir **Clean Architecture** com separação clara de camadas
- `cmd/`: Entry point da aplicação
- `internal/`: Código privado do serviço
- `pkg/`: Código reutilizável entre serviços

#### Convenções
- **Nomenclatura**: camelCase para variáveis privadas, PascalCase para exportadas
- **Errors**: Sempre retornar erros descritivos usando `fmt.Errorf` ou pacotes como `errors`
- **Context**: Sempre passar `context.Context` como primeiro parâmetro em funções que fazem I/O
- **Logging**: Usar logger estruturado (logrus ou zap)
- **Validação**: Usar biblioteca `validator` para validar structs
- **Configuração**: Usar variáveis de ambiente com pacote `godotenv` ou `viper`

#### HTTP Framework
- **Framework**: Fiber (preferencial) ou Gin
- **Response Format**:
  ```go
  type Response struct {
      Success bool        `json:"success"`
      Data    interface{} `json:"data,omitempty"`
      Error   string      `json:"error,omitempty"`
      Message string      `json:"message,omitempty"`
  }
  ```

#### Autenticação JWT
- **Secret**: Armazenar em variável de ambiente `JWT_SECRET`
- **Claims**:
  ```go
  type JWTClaims struct {
      UserID string `json:"user_id"`
      Email  string `json:"email"`
      jwt.RegisteredClaims
  }
  ```
- **Middleware**: Validar token no header `Authorization: Bearer <token>`

#### Banco de Dados

**PostgreSQL (User Service)**
- **Driver**: `github.com/lib/pq` ou `github.com/jackc/pgx`
- **Migration**: Usar `golang-migrate/migrate`
- **Connection Pool**: Configurar max open connections (25) e max idle connections (5)

**MongoDB (Order Service)**
- **Driver**: `go.mongodb.org/mongo-driver`
- **Collections**: `orders`
- **Indexes**: Criar index em `user_id` e `created_at`

#### RabbitMQ
- **Library**: `github.com/rabbitmq/amqp091-go`
- **Exchange**: `orders` (tipo: topic)
- **Routing Keys**: `order.created`, `order.updated`
- **Message Format**: JSON
- **Durabilidade**: Mensagens e filas duráveis
- **Acknowledgment**: Manual ACK após processamento bem-sucedido

### React (Frontend)

#### Estrutura de Componentes
- **Componentes funcionais** com hooks
- **Props TypeScript/PropTypes** quando possível
- **Composição** sobre herança

#### Convenções
- **Nomenclatura**: PascalCase para componentes, camelCase para funções/variáveis
- **Arquivos**: Um componente por arquivo, nome do arquivo igual ao componente
- **Import Order**: 
  1. React imports
  2. Third-party libraries
  3. Internal components
  4. Services/utils
  5. Styles

#### Estado e Efeitos
- **useState**: Para estado local do componente
- **useEffect**: Para side effects (API calls, subscriptions)
- **useContext**: Para consumir AuthContext
- **Custom Hooks**: Criar hooks reutilizáveis (ex: `useAuth`, `useOrders`)

#### API Calls
- **Axios Instance**: Configurar base URL e interceptors
- **Interceptors**:
  - Request: Adicionar token JWT automaticamente
  - Response: Tratar erros 401 (logout automático) e outros erros
- **Error Handling**: Exibir mensagens de erro amigáveis ao usuário

#### Roteamento
- **Rotas Protegidas**: Criar componente `ProtectedRoute` que verifica autenticação
- **Redirect**: Redirecionar para login se não autenticado

#### Formulários
- **Validação**: Validação client-side antes de enviar
- **Loading States**: Exibir indicadores de carregamento durante requisições
- **Feedback**: Toast/Snackbar para sucesso/erro

## Variáveis de Ambiente

### User Service
```env
PORT=8001
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=users_db
JWT_SECRET=your_super_secret_key_change_this
JWT_EXPIRATION=24h
```

### Order Service
```env
PORT=8002
MONGO_URI=mongodb://mongodb:27017
MONGO_DATABASE=orders_db
JWT_SECRET=your_super_secret_key_change_this
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
USER_SERVICE_URL=http://user-service:8001
```

### Notification Service
```env
PORT=8003
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASSWORD=your_app_password
EMAIL_FROM=noreply@pedidosonline.com
```

### API Gateway
```env
PORT=8000
USER_SERVICE_URL=http://user-service:8001
ORDER_SERVICE_URL=http://order-service:8002
```

### Frontend
```env
VITE_API_URL=http://localhost:8000
```

## Docker

### Dockerfile (Go Services)
```dockerfile
# Multi-stage build para otimizar tamanho
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8001
CMD ["./main"]
```

### Docker Compose
- **Networks**: Criar rede `pedidos-network` para comunicação entre serviços
- **Volumes**: Persistir dados do PostgreSQL e MongoDB
- **Health Checks**: Configurar health checks para garantir que serviços estejam prontos

## Fluxo de Desenvolvimento

### 1. Cadastro e Login
1. Usuário acessa `/register` no frontend
2. Frontend envia POST para API Gateway `/api/users/register`
3. API Gateway roteia para User Service
4. User Service valida dados, criptografa senha (bcrypt), salva no PostgreSQL
5. Retorna sucesso
6. Usuário faz login em `/login`
7. User Service valida credenciais, gera JWT
8. Frontend armazena token no localStorage/sessionStorage
9. Redireciona para página principal

### 2. Criar Pedido
1. Usuário autenticado preenche formulário de pedido
2. Frontend envia POST para `/api/orders` com token JWT
3. API Gateway valida e roteia para Order Service
4. Order Service valida token JWT
5. Order Service salva pedido no MongoDB
6. Order Service publica evento `order.created` no RabbitMQ
7. Notification Service consome evento
8. Notification Service envia e-mail de confirmação
9. Order Service retorna sucesso para frontend

### 3. Acompanhar Pedido
1. Usuário acessa `/orders/:id`
2. Frontend busca detalhes do pedido
3. Order Service retorna dados do pedido
4. Frontend exibe status e timeline do pedido

## Testes

### Backend (Go)
- **Unit Tests**: Testar handlers, services e repositories isoladamente
- **Integration Tests**: Testar fluxos completos com banco de dados de teste
- **Mocks**: Usar `gomock` ou `testify/mock`
- **Cobertura**: Mínimo de 70%

### Frontend (React)
- **Jest + React Testing Library**: Testes de componentes
- **User Flow Tests**: Simular interações do usuário
- **Mock API**: Usar MSW para mockar chamadas API

## Segurança

### Backend
- **HTTPS**: Usar HTTPS em produção
- **Rate Limiting**: Implementar no API Gateway
- **Input Validation**: Validar todos os inputs
- **SQL Injection**: Usar prepared statements
- **Password**: Hash com bcrypt (cost 10)
- **CORS**: Configurar origens permitidas
- **Headers**: Adicionar security headers (Helmet)

### Frontend
- **XSS**: Sanitizar inputs do usuário
- **CSRF**: Token CSRF se usar cookies
- **Token Storage**: Avaliar localStorage vs httpOnly cookies

## Monitoramento

- **Logs**: Logs estruturados em JSON
- **Metrics**: Expor métricas Prometheus
- **Tracing**: Implementar distributed tracing (Jaeger)
- **Health Checks**: Endpoints `/health` e `/ready`

## CI/CD

- **GitHub Actions**: Pipeline de build, test e deploy
- **Linting**: golangci-lint para Go, ESLint para React
- **Docker**: Build e push de imagens
- **Deploy**: Kubernetes ou Docker Swarm

## Observações Importantes

1. **Comunicação entre serviços**: User Service e Order Service NÃO se comunicam diretamente. Order Service apenas valida o JWT localmente.
2. **Consistência eventual**: Notificações podem ter delay devido à fila assíncrona.
3. **Idempotência**: Consumidores RabbitMQ devem ser idempotentes (processar mesma mensagem múltiplas vezes sem efeito colateral).
4. **Retry Logic**: Implementar retry com backoff exponencial para falhas temporárias.
5. **Dead Letter Queue**: Configurar DLQ no RabbitMQ para mensagens que falharam múltiplas vezes.

## Recursos Úteis

- **Go**: https://go.dev/doc/
- **Fiber**: https://docs.gofiber.io/
- **React**: https://react.dev/
- **RabbitMQ**: https://www.rabbitmq.com/tutorials/tutorial-one-go.html
- **MongoDB Go Driver**: https://www.mongodb.com/docs/drivers/go/current/
- **PostgreSQL**: https://www.postgresql.org/docs/

## Comandos Úteis

### Desenvolvimento Local
```bash
# Subir todos os serviços
docker-compose up -d

# Ver logs de um serviço específico
docker-compose logs -f user-service

# Rebuild de um serviço
docker-compose up -d --build user-service

# Executar migrations
make migrate-up

# Rodar testes
make test
```

### Go
```bash
# Instalar dependências
go mod download

# Rodar aplicação
go run cmd/main.go

# Rodar testes
go test ./...

# Gerar coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### React
```bash
# Instalar dependências
npm install

# Rodar dev server
npm run dev

# Build para produção
npm run build

# Rodar testes
npm test
```
