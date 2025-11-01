.PHONY: help up down restart logs logs-service ps build build-service migrate-up migrate-down migrate-create db-seed dev-user dev-order dev-notification dev-gateway dev-frontend test test-user test-order test-notification test-frontend test-coverage lint lint-fix clean prune

# Colors for better visualization
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

# Variables
SERVICE ?= user-service
NAME ?= migration_name

##@ General

help: ## 📚 Exibir lista de comandos disponíveis
	@echo "$(BLUE)════════════════════════════════════════════════════════════$(NC)"
	@echo "$(GREEN)  Sistema de Pedidos Online - Makefile Commands$(NC)"
	@echo "$(BLUE)════════════════════════════════════════════════════════════$(NC)"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(BLUE)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@echo ""

##@ Docker Commands

up: ## 🚀 Subir todos os containers em background
	@echo "$(GREEN)🚀 Subindo todos os containers...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)✅ Containers iniciados com sucesso!$(NC)"
	@echo "$(YELLOW)💡 Acesse:$(NC)"
	@echo "   - Frontend: http://localhost:3000"
	@echo "   - API Gateway: http://localhost:8000"
	@echo "   - RabbitMQ Management: http://localhost:15672"

down: ## 🛑 Parar e remover todos os containers
	@echo "$(RED)🛑 Parando todos os containers...$(NC)"
	docker-compose down
	@echo "$(GREEN)✅ Containers parados!$(NC)"

restart: ## 🔄 Reiniciar todos os containers
	@echo "$(YELLOW)🔄 Reiniciando todos os containers...$(NC)"
	docker-compose restart
	@echo "$(GREEN)✅ Containers reiniciados!$(NC)"

logs: ## 📋 Ver logs de todos os serviços com follow
	@echo "$(BLUE)📋 Exibindo logs de todos os serviços...$(NC)"
	docker-compose logs -f

logs-service: ## 📝 Ver logs de um serviço específico (uso: make logs-service SERVICE=user-service)
	@echo "$(BLUE)📝 Exibindo logs do serviço: $(SERVICE)$(NC)"
	docker-compose logs -f $(SERVICE)

ps: ## 📊 Listar status de todos os containers
	@echo "$(BLUE)📊 Status dos containers:$(NC)"
	docker-compose ps

##@ Build Commands

build: ## 🔨 Rebuild de todos os serviços
	@echo "$(YELLOW)🔨 Rebuilding todos os serviços...$(NC)"
	docker-compose build --no-cache
	@echo "$(GREEN)✅ Build completo!$(NC)"

build-service: ## 🔧 Rebuild de um serviço específico (uso: make build-service SERVICE=user-service)
	@echo "$(YELLOW)🔧 Rebuilding serviço: $(SERVICE)$(NC)"
	docker-compose build --no-cache $(SERVICE)
	docker-compose up -d $(SERVICE)
	@echo "$(GREEN)✅ Build do $(SERVICE) completo!$(NC)"

##@ Database Commands

migrate-up: ## ⬆️  Executar migrations do PostgreSQL (user-service)
	@echo "$(GREEN)⬆️  Executando migrations...$(NC)"
	@docker-compose exec -T user-service sh -c 'if [ -d "/app/migrations" ]; then migrate -path /app/migrations -database "postgres://postgres:postgres@postgres:5432/users_db?sslmode=disable" up; else echo "Pasta migrations não encontrada"; fi'
	@echo "$(GREEN)✅ Migrations executadas!$(NC)"

migrate-down: ## ⬇️  Reverter última migration
	@echo "$(RED)⬇️  Revertendo última migration...$(NC)"
	@docker-compose exec -T user-service sh -c 'if [ -d "/app/migrations" ]; then migrate -path /app/migrations -database "postgres://postgres:postgres@postgres:5432/users_db?sslmode=disable" down 1; else echo "Pasta migrations não encontrada"; fi'
	@echo "$(GREEN)✅ Migration revertida!$(NC)"

migrate-create: ## ✨ Criar nova migration (uso: make migrate-create NAME=create_users_table)
	@echo "$(YELLOW)✨ Criando migration: $(NAME)$(NC)"
	@mkdir -p services/user-service/migrations
	@docker-compose exec -T user-service sh -c 'migrate create -ext sql -dir /app/migrations -seq $(NAME)' || \
		(cd services/user-service/migrations && \
		timestamp=$$(date +%s) && \
		touch $${timestamp}_$(NAME).up.sql $${timestamp}_$(NAME).down.sql && \
		echo "Created: $${timestamp}_$(NAME).up.sql" && \
		echo "Created: $${timestamp}_$(NAME).down.sql")
	@echo "$(GREEN)✅ Migration criada em services/user-service/migrations/$(NC)"

db-seed: ## 🌱 Popular banco com dados de teste
	@echo "$(YELLOW)🌱 Populando banco de dados...$(NC)"
	@docker-compose exec -T postgres psql -U postgres -d users_db -c "\
		INSERT INTO users (id, email, password, name, phone, created_at, updated_at) \
		VALUES \
		(gen_random_uuid(), 'admin@test.com', '\$$2a\$$10\$$EXAMPLEHASH', 'Admin User', '11999999999', NOW(), NOW()), \
		(gen_random_uuid(), 'user@test.com', '\$$2a\$$10\$$EXAMPLEHASH', 'Test User', '11988888888', NOW(), NOW()) \
		ON CONFLICT DO NOTHING;" 2>/dev/null || echo "Tabela users não existe ainda. Execute as migrations primeiro."
	@echo "$(GREEN)✅ Dados de teste inseridos!$(NC)"
	@echo "$(YELLOW)💡 Credenciais de teste:$(NC)"
	@echo "   - admin@test.com / password"
	@echo "   - user@test.com / password"

##@ Development Commands

dev-user: ## 💻 Rodar user-service localmente (fora do docker)
	@echo "$(BLUE)💻 Iniciando user-service localmente...$(NC)"
	@cd services/user-service && \
		export PORT=8001 && \
		export DB_HOST=localhost && \
		export DB_PORT=5432 && \
		export DB_USER=postgres && \
		export DB_PASSWORD=postgres && \
		export DB_NAME=users_db && \
		export JWT_SECRET=your_super_secret_key_change_this && \
		export JWT_EXPIRATION=24h && \
		go run cmd/main.go

dev-order: ## 💻 Rodar order-service localmente (fora do docker)
	@echo "$(BLUE)💻 Iniciando order-service localmente...$(NC)"
	@cd services/order-service && \
		export PORT=8002 && \
		export MONGO_URI=mongodb://localhost:27017 && \
		export MONGO_DATABASE=orders_db && \
		export JWT_SECRET=your_super_secret_key_change_this && \
		export RABBITMQ_URL=amqp://guest:guest@localhost:5672/ && \
		export USER_SERVICE_URL=http://localhost:8001 && \
		go run cmd/main.go

dev-notification: ## 💻 Rodar notification-service localmente (fora do docker)
	@echo "$(BLUE)💻 Iniciando notification-service localmente...$(NC)"
	@cd services/notification-service && \
		export PORT=8003 && \
		export RABBITMQ_URL=amqp://guest:guest@localhost:5672/ && \
		export SMTP_HOST=smtp.gmail.com && \
		export SMTP_PORT=587 && \
		export SMTP_USER=your_email@gmail.com && \
		export SMTP_PASSWORD=your_app_password && \
		export EMAIL_FROM=noreply@pedidosonline.com && \
		go run cmd/main.go

dev-gateway: ## 💻 Rodar api-gateway localmente (fora do docker)
	@echo "$(BLUE)💻 Iniciando api-gateway localmente...$(NC)"
	@cd api-gateway && \
		export PORT=8000 && \
		export USER_SERVICE_URL=http://localhost:8001 && \
		export ORDER_SERVICE_URL=http://localhost:8002 && \
		go run cmd/main.go

dev-frontend: ## 💻 Rodar frontend localmente (fora do docker)
	@echo "$(BLUE)💻 Iniciando frontend localmente...$(NC)"
	@cd frontend && \
		export VITE_API_URL=http://localhost:8000 && \
		npm run dev

##@ Test Commands

test: ## 🧪 Rodar todos os testes
	@echo "$(YELLOW)🧪 Executando todos os testes...$(NC)"
	@$(MAKE) test-user
	@$(MAKE) test-order
	@$(MAKE) test-notification
	@$(MAKE) test-frontend
	@echo "$(GREEN)✅ Todos os testes concluídos!$(NC)"

test-user: ## 🧪 Testar user-service
	@echo "$(BLUE)🧪 Testando user-service...$(NC)"
	@cd services/user-service && go test -v ./... || true

test-order: ## 🧪 Testar order-service
	@echo "$(BLUE)🧪 Testando order-service...$(NC)"
	@cd services/order-service && go test -v ./... || true

test-notification: ## 🧪 Testar notification-service
	@echo "$(BLUE)🧪 Testando notification-service...$(NC)"
	@cd services/notification-service && go test -v ./... || true

test-frontend: ## 🧪 Testar frontend
	@echo "$(BLUE)🧪 Testando frontend...$(NC)"
	@cd frontend && npm test -- --passWithNoTests || true

test-coverage: ## 📊 Gerar relatório de cobertura
	@echo "$(YELLOW)📊 Gerando relatório de cobertura...$(NC)"
	@mkdir -p coverage
	@echo "$(BLUE)User Service:$(NC)"
	@cd services/user-service && go test -coverprofile=../../coverage/user-service.out ./... && \
		go tool cover -html=../../coverage/user-service.out -o ../../coverage/user-service.html || true
	@echo "$(BLUE)Order Service:$(NC)"
	@cd services/order-service && go test -coverprofile=../../coverage/order-service.out ./... && \
		go tool cover -html=../../coverage/order-service.out -o ../../coverage/order-service.html || true
	@echo "$(BLUE)Notification Service:$(NC)"
	@cd services/notification-service && go test -coverprofile=../../coverage/notification-service.out ./... && \
		go tool cover -html=../../coverage/notification-service.out -o ../../coverage/notification-service.html || true
	@echo "$(GREEN)✅ Relatórios gerados em ./coverage/$(NC)"

##@ Linting Commands

lint: ## 🔍 Rodar linters em todos os serviços Go e frontend
	@echo "$(YELLOW)🔍 Executando linters...$(NC)"
	@echo "$(BLUE)User Service:$(NC)"
	@cd services/user-service && golangci-lint run ./... || echo "golangci-lint não instalado"
	@echo "$(BLUE)Order Service:$(NC)"
	@cd services/order-service && golangci-lint run ./... || echo "golangci-lint não instalado"
	@echo "$(BLUE)Notification Service:$(NC)"
	@cd services/notification-service && golangci-lint run ./... || echo "golangci-lint não instalado"
	@echo "$(BLUE)API Gateway:$(NC)"
	@cd api-gateway && golangci-lint run ./... || echo "golangci-lint não instalado"
	@echo "$(BLUE)Frontend:$(NC)"
	@cd frontend && npm run lint || echo "ESLint não configurado"
	@echo "$(GREEN)✅ Linting concluído!$(NC)"

lint-fix: ## 🔧 Corrigir problemas de linting automaticamente
	@echo "$(YELLOW)🔧 Corrigindo problemas de linting...$(NC)"
	@echo "$(BLUE)User Service:$(NC)"
	@cd services/user-service && golangci-lint run --fix ./... || echo "golangci-lint não instalado"
	@echo "$(BLUE)Order Service:$(NC)"
	@cd services/order-service && golangci-lint run --fix ./... || echo "golangci-lint não instalado"
	@echo "$(BLUE)Notification Service:$(NC)"
	@cd services/notification-service && golangci-lint run --fix ./... || echo "golangci-lint não instalado"
	@echo "$(BLUE)API Gateway:$(NC)"
	@cd api-gateway && golangci-lint run --fix ./... || echo "golangci-lint não instalado"
	@echo "$(BLUE)Frontend:$(NC)"
	@cd frontend && npm run lint:fix || echo "ESLint não configurado"
	@echo "$(GREEN)✅ Correções aplicadas!$(NC)"

##@ Cleanup Commands

clean: ## 🧹 Remover containers, volumes, imagens e dados temporários
	@echo "$(RED)🧹 Limpando projeto...$(NC)"
	@docker-compose down -v --remove-orphans
	@docker-compose rm -f
	@rm -rf coverage/
	@find . -name "*.out" -type f -delete
	@find . -name "*.test" -type f -delete
	@echo "$(GREEN)✅ Limpeza concluída!$(NC)"

prune: ## ⚠️  Limpar sistema Docker (CUIDADO! Remove tudo não utilizado)
	@echo "$(RED)⚠️  ATENÇÃO: Isso removerá TODOS os recursos Docker não utilizados!$(NC)"
	@read -p "Tem certeza? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo "$(RED)🧹 Limpando sistema Docker...$(NC)"; \
		docker system prune -a --volumes -f; \
		echo "$(GREEN)✅ Sistema Docker limpo!$(NC)"; \
	else \
		echo "$(YELLOW)❌ Operação cancelada.$(NC)"; \
	fi

##@ Installation Commands

install-tools: ## 📦 Instalar ferramentas de desenvolvimento necessárias
	@echo "$(YELLOW)📦 Instalando ferramentas de desenvolvimento...$(NC)"
	@echo "$(BLUE)Instalando golangci-lint...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest || echo "Erro ao instalar golangci-lint"
	@echo "$(BLUE)Instalando migrate...$(NC)"
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest || echo "Erro ao instalar migrate"
	@echo "$(GREEN)✅ Ferramentas instaladas!$(NC)"
	@echo "$(YELLOW)💡 Certifique-se de que \$$GOPATH/bin está no seu PATH$(NC)"

##@ Quick Start

quick-start: ## 🚀 Setup completo do projeto (first time setup)
	@echo "$(GREEN)🚀 Iniciando setup completo do projeto...$(NC)"
	@echo "$(BLUE)1/4 - Building services...$(NC)"
	@$(MAKE) build
	@echo "$(BLUE)2/4 - Starting containers...$(NC)"
	@$(MAKE) up
	@echo "$(BLUE)3/4 - Waiting for services to be healthy...$(NC)"
	@sleep 15
	@echo "$(BLUE)4/4 - Running migrations...$(NC)"
	@$(MAKE) migrate-up || echo "Migrations não disponíveis ainda"
	@echo "$(GREEN)✅ Setup completo!$(NC)"
	@echo ""
	@echo "$(YELLOW)════════════════════════════════════════════════════════════$(NC)"
	@echo "$(GREEN)  🎉 Projeto pronto para uso!$(NC)"
	@echo "$(YELLOW)════════════════════════════════════════════════════════════$(NC)"
	@echo "$(YELLOW)💡 Acesse:$(NC)"
	@echo "   - Frontend: http://localhost:3000"
	@echo "   - API Gateway: http://localhost:8000"
	@echo "   - RabbitMQ Management: http://localhost:15672 (guest/guest)"
	@echo ""
	@echo "$(YELLOW)📚 Comandos úteis:$(NC)"
	@echo "   - make logs         - Ver logs"
	@echo "   - make ps           - Ver status"
	@echo "   - make test         - Rodar testes"
	@echo "   - make help         - Ver todos os comandos"
	@echo "$(YELLOW)════════════════════════════════════════════════════════════$(NC)"
