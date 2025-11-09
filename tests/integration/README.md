# Integration Tests - Sistema de Pedidos Online

Testes end-to-end (E2E) que validam o funcionamento completo do sistema de pedidos online, incluindo todos os microserviços e suas integrações.

## Visão Geral

Os testes de integração verificam:

1. **Fluxo de Usuários** - Registro, login, autenticação e perfil
2. **Fluxo de Pedidos** - Criação, listagem, atualização e cancelamento
3. **Notificações** - RabbitMQ e processamento assíncrono de emails
4. **API Gateway** - Roteamento, CORS, rate limiting e propagação de headers

## Estrutura dos Testes

```
tests/integration/
├── go.mod                    # Dependências dos testes
├── helpers.go                # Funções auxiliares reutilizáveis
├── user_flow_test.go         # Testes do fluxo de usuários
├── order_flow_test.go        # Testes do fluxo de pedidos
├── notification_test.go      # Testes de notificações e RabbitMQ
├── gateway_test.go           # Testes do API Gateway
└── README.md                 # Esta documentação
```

## Pré-requisitos

- Docker e Docker Compose instalados
- Go 1.21 ou superior
- Portas disponíveis: 8000-8003, 5433, 27018, 5673, 15673, 1025, 8025

## Como Executar

### Opção 1: Usando Makefile (Recomendado)

```bash
# Rodar todos os testes de integração
make test-integration

# Rodar testes em modo curto (pula testes de carga/timeout)
make test-integration-short

# Rodar com saída verbosa
make test-integration-verbose

# Manter containers rodando após os testes (útil para debug)
make test-integration-keep

# Rodar sem rebuild das imagens Docker
make test-integration-skip-build

# Limpar ambiente de testes
make test-integration-clean

# Rodar TODOS os testes (unitários + integração)
make test-all
```

### Opção 2: Usando Script Diretamente

```bash
# Rodar com todas as opções padrão
./tests/run-integration-tests.sh

# Ver opções disponíveis
./tests/run-integration-tests.sh --help

# Exemplos
./tests/run-integration-tests.sh --verbose
./tests/run-integration-tests.sh --short --keep-running
./tests/run-integration-tests.sh --skip-build
```

### Opção 3: Manualmente

```bash
# 1. Subir ambiente de testes
docker-compose -f docker-compose.test.yml up -d

# 2. Aguardar serviços ficarem healthy (pode levar 1-2 minutos)
docker-compose -f docker-compose.test.yml ps

# 3. Rodar testes
cd tests/integration
go test -v ./...

# 4. Limpar ambiente
docker-compose -f docker-compose.test.yml down -v
```

## Testes Incluídos

### 1. User Flow Tests (`user_flow_test.go`)

#### TestUserRegistration
- ✅ Registra novo usuário
- ✅ Verifica estrutura da resposta
- ✅ Valida que senha não é retornada

#### TestUserRegistrationDuplicateEmail
- ✅ Tenta registrar email duplicado
- ✅ Verifica erro 409 Conflict

#### TestUserRegistrationInvalidData
- ✅ Valida email ausente
- ✅ Valida formato de email inválido
- ✅ Valida senha fraca
- ✅ Valida nome ausente

#### TestUserLogin
- ✅ Login com credenciais corretas
- ✅ Verifica retorno de token JWT
- ✅ Valida dados do usuário

#### TestUserLoginInvalidCredentials
- ✅ Senha incorreta
- ✅ Usuário inexistente
- ✅ Senha vazia

#### TestUserProfile
- ✅ Busca perfil autenticado
- ✅ Verifica todos os campos
- ✅ Confirma que senha não é retornada

#### TestUserProfileWithoutAuth / TestUserProfileWithInvalidToken
- ✅ Rejeita acesso sem token
- ✅ Rejeita token inválido

#### TestUpdateUserProfile
- ✅ Atualiza nome e telefone
- ✅ Verifica persistência das mudanças

#### TestCompleteUserFlow
- ✅ Registro → Login → Profile → Update → Verificação

### 2. Order Flow Tests (`order_flow_test.go`)

#### TestCreateOrder
- ✅ Cria pedido válido
- ✅ Verifica cálculo de total
- ✅ Valida itens e endereço

#### TestCreateOrderWithoutAuth
- ✅ Rejeita pedido sem autenticação

#### TestCreateOrderInvalidData
- ✅ Itens vazios
- ✅ Itens ausentes
- ✅ Endereço ausente
- ✅ Quantidade inválida
- ✅ Preço inválido

#### TestListOrders
- ✅ Lista pedidos do usuário
- ✅ Verifica presença de pedidos criados

#### TestListOrdersIsolation
- ✅ Usuário 1 não vê pedidos do Usuário 2
- ✅ Verifica isolamento de dados

#### TestGetOrderDetails
- ✅ Busca detalhes completos do pedido
- ✅ Verifica itens e endereço

#### TestGetOrderDetailsUnauthorized
- ✅ Usuário não pode ver pedido de outro

#### TestUpdateOrderStatus
- ✅ Transições: pending → confirmed → preparing → shipped → delivered
- ✅ Verifica cada mudança de status

#### TestCancelOrder
- ✅ Cancela pedido pendente

#### TestCompleteOrderFlow
- ✅ Criação → Listagem → Detalhes → Atualizações de Status

### 3. Notification Tests (`notification_test.go`)

#### TestOrderCreatedNotification
- ✅ Verifica publicação no RabbitMQ
- ✅ Valida estrutura da mensagem
- ✅ Confirma dados do pedido

#### TestOrderUpdatedNotification
- ✅ Mudança de status gera evento
- ✅ Verifica routing key correto

#### TestMultipleStatusUpdatesNotifications
- ✅ Múltiplas transições geram múltiplos eventos
- ✅ Cada evento contém status correto

#### TestNotificationServiceConsumption
- ✅ Filas estão configuradas
- ✅ Consumers estão conectados

#### TestNotificationIdempotency
- ✅ Mensagens não são duplicadas
- ✅ ACK está funcionando corretamente

#### TestNotificationServiceHealthDuringLoad
- ✅ Cria 10 pedidos rapidamente
- ✅ Verifica que todas notificações são processadas
- ✅ Service permanece healthy

#### TestNotificationMessageFormat
- ✅ Content-type é JSON
- ✅ Mensagem é persistente
- ✅ Estrutura contém campos obrigatórios

### 4. Gateway Tests (`gateway_test.go`)

#### TestGatewayRouting
- ✅ Roteia /api/v1/register para user-service
- ✅ Roteia /api/v1/orders para order-service
- ✅ Propaga status codes corretamente

#### TestGatewayRoutingWithValidRequests
- ✅ Register via gateway funciona
- ✅ Create order via gateway funciona

#### TestGatewayHeaderPropagation
- ✅ Authorization header é propagado
- ✅ Custom headers passam pelo gateway

#### TestGatewayCORS
- ✅ Responde a preflight OPTIONS
- ✅ Retorna headers CORS corretos

#### TestGatewayErrorHandling
- ✅ 404 do serviço → 404 no gateway
- ✅ 400 do serviço → 400 no gateway
- ✅ 401 do serviço → 401 no gateway

#### TestGatewayRateLimiting
- ✅ Faz 100 requisições concorrentes
- ✅ Verifica se rate limiting é aplicado (se configurado)

#### TestGatewayRateLimitHeaders
- ✅ Busca headers de rate limit na resposta

#### TestGatewayTimeout
- ✅ Requisições não excedem timeout
- ✅ Respondem em tempo razoável

#### TestGatewayHealthCheck
- ✅ /health retorna 200
- ✅ Resposta contém dados de saúde

#### TestGatewayRequestSize
- ✅ Pedido com 100 itens
- ✅ Não retorna erro 500 ou timeout

#### TestGatewayConcurrentRequests
- ✅ 20 registros simultâneos
- ✅ Maioria deve ter sucesso

## Configuração do Ambiente de Testes

O arquivo `docker-compose.test.yml` define:

### Serviços
- **postgres-test**: PostgreSQL na porta 5433
- **mongodb-test**: MongoDB na porta 27018
- **rabbitmq-test**: RabbitMQ nas portas 5673 (AMQP) e 15673 (Management)
- **mailhog**: Mock SMTP nas portas 1025 (SMTP) e 8025 (Web UI)
- **user-service-test**: User Service na porta 8001
- **order-service-test**: Order Service na porta 8002
- **notification-service-test**: Notification Service na porta 8003
- **api-gateway-test**: API Gateway na porta 8000

### Variáveis de Ambiente
- JWT_SECRET: `test_jwt_secret_key_for_integration_tests`
- GIN_MODE: `release`
- Timeouts e health checks configurados

### Volumes
- Dados são persistidos em volumes nomeados
- Logs salvos em `tests/logs/` em caso de falha

## Troubleshooting

### Serviços não ficam healthy

```bash
# Ver logs de um serviço específico
docker-compose -f docker-compose.test.yml logs user-service-test

# Ver status dos containers
docker-compose -f docker-compose.test.yml ps

# Verificar health checks
docker inspect pedidos-user-service-test | grep -A 10 Health
```

### Portas em uso

```bash
# Verificar quem está usando a porta
lsof -i :8000

# Parar ambiente de testes
make test-integration-clean

# Parar ambiente de desenvolvimento
make down
```

### Testes falhando com timeout

```bash
# Aumentar timeout (padrão: 600s)
TEST_TIMEOUT=1200 ./tests/run-integration-tests.sh

# Ou rodar em modo short (pula testes de carga)
make test-integration-short
```

### Debugging de testes

```bash
# Manter containers rodando após os testes
make test-integration-keep

# Acessar serviços para debug
curl http://localhost:8000/health
curl http://localhost:8001/health

# Ver emails enviados (MailHog UI)
open http://localhost:8025

# Ver RabbitMQ Management
open http://localhost:15673  # guest/guest
```

### Limpar completamente

```bash
# Parar containers e remover volumes
make test-integration-clean

# Limpar imagens Docker também
docker-compose -f docker-compose.test.yml down -v --rmi all
```

## Logs e Debugging

### Logs são salvos automaticamente em caso de falha

```
tests/logs/
├── all-services.log          # Logs de todos os serviços
├── user-service.log          # Logs específicos
├── order-service.log
├── notification-service.log
├── api-gateway.log
└── test-output.log           # Saída dos testes
```

### Ver logs em tempo real

```bash
# Durante os testes (com --keep-running)
docker-compose -f docker-compose.test.yml logs -f

# Serviço específico
docker-compose -f docker-compose.test.yml logs -f user-service-test
```

## CI/CD Integration

### GitHub Actions

```yaml
- name: Run Integration Tests
  run: |
    make test-integration
  timeout-minutes: 15
```

### GitLab CI

```yaml
integration-tests:
  script:
    - make test-integration
  timeout: 15m
```

## Melhores Práticas

1. **Isolamento**: Cada teste cria seus próprios dados (usuários únicos)
2. **Cleanup**: Script limpa automaticamente após execução
3. **Timeout**: Testes têm timeout de 10 minutos por padrão
4. **Idempotência**: Testes podem ser executados múltiplas vezes
5. **Verbosidade**: Use `--verbose` para debug detalhado

## Próximos Passos

- [ ] Adicionar testes de performance/load
- [ ] Testes de failover e retry
- [ ] Testes de segurança (SQL injection, XSS)
- [ ] Integração com ferramentas de coverage
- [ ] Paralelização de testes

## Suporte

Para problemas ou dúvidas:
1. Verifique logs em `tests/logs/`
2. Use `--verbose` para mais detalhes
3. Consulte documentação do projeto principal

---

**Última atualização**: 2025-11-09
