#!/bin/bash

###############################################################################
# Script de Verificação de Saúde dos Serviços
###############################################################################
# Este script verifica se todos os serviços do sistema estão rodando:
# - PostgreSQL
# - MongoDB
# - RabbitMQ
# - User Service
# - Order Service
# - Notification Service
# - API Gateway
# - Frontend
#
# Exit code: 0 se todos OK, 1 se algum falhar
###############################################################################

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Configurações (podem ser sobrescritas por variáveis de ambiente)
POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_USER="${POSTGRES_USER:-postgres}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-postgres}"
MONGO_URI="${MONGO_URI:-mongodb://localhost:27017}"
RABBITMQ_HOST="${RABBITMQ_HOST:-localhost}"
RABBITMQ_PORT="${RABBITMQ_PORT:-5672}"
RABBITMQ_MANAGEMENT_PORT="${RABBITMQ_MANAGEMENT_PORT:-15672}"
USER_SERVICE_URL="${USER_SERVICE_URL:-http://localhost:8001}"
ORDER_SERVICE_URL="${ORDER_SERVICE_URL:-http://localhost:8002}"
NOTIFICATION_SERVICE_URL="${NOTIFICATION_SERVICE_URL:-http://localhost:8003}"
API_GATEWAY_URL="${API_GATEWAY_URL:-http://localhost:8000}"
FRONTEND_URL="${FRONTEND_URL:-http://localhost:5173}"

# Contador de falhas
FAILURES=0

###############################################################################
# Funções Auxiliares
###############################################################################

print_header() {
    echo -e "\n${BLUE}${BOLD}============================================================================${NC}"
    echo -e "${BLUE}${BOLD}$1${NC}"
    echo -e "${BLUE}${BOLD}============================================================================${NC}\n"
}

print_checking() {
    echo -ne "${BLUE}🔍 Verificando $1...${NC}"
}

print_ok() {
    echo -e "\r${GREEN}✅ $1 está OK${NC}                    "
}

print_fail() {
    echo -e "\r${RED}❌ $1 está FALHANDO${NC}              "
    ((FAILURES++))
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

###############################################################################
# Verificações de Serviços
###############################################################################

check_postgresql() {
    print_checking "PostgreSQL"
    
    export PGPASSWORD="$POSTGRES_PASSWORD"
    
    if pg_isready -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" &> /dev/null; then
        # Verificar se consegue conectar
        if psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -c "SELECT 1;" &> /dev/null; then
            print_ok "PostgreSQL ($POSTGRES_HOST:$POSTGRES_PORT)"
        else
            print_fail "PostgreSQL (conexão falhou)"
        fi
    else
        print_fail "PostgreSQL (não está respondendo)"
    fi
    
    unset PGPASSWORD
}

check_mongodb() {
    print_checking "MongoDB"
    
    # Tentar com mongosh (novo client)
    if command -v mongosh &> /dev/null; then
        if mongosh "$MONGO_URI" --quiet --eval "db.adminCommand('ping')" &> /dev/null; then
            print_ok "MongoDB"
            return
        fi
    fi
    
    # Fallback para mongo (client legado)
    if command -v mongo &> /dev/null; then
        if mongo "$MONGO_URI" --quiet --eval "db.adminCommand('ping')" &> /dev/null; then
            print_ok "MongoDB"
            return
        fi
    fi
    
    # Tentar via telnet/nc como último recurso
    MONGO_HOST=$(echo "$MONGO_URI" | sed -e 's/mongodb:\/\///' -e 's/:.*$//')
    MONGO_PORT=$(echo "$MONGO_URI" | grep -oP ':\K[0-9]+' || echo "27017")
    
    if nc -z "$MONGO_HOST" "$MONGO_PORT" 2>/dev/null || telnet "$MONGO_HOST" "$MONGO_PORT" </dev/null 2>/dev/null | grep -q "Connected"; then
        print_ok "MongoDB (porta aberta)"
    else
        print_fail "MongoDB"
    fi
}

check_rabbitmq() {
    print_checking "RabbitMQ"
    
    # Verificar porta AMQP
    if nc -z "$RABBITMQ_HOST" "$RABBITMQ_PORT" 2>/dev/null; then
        # Tentar acessar API de management
        if curl -s -f "http://$RABBITMQ_HOST:$RABBITMQ_MANAGEMENT_PORT/api/healthchecks/node" &> /dev/null; then
            print_ok "RabbitMQ ($RABBITMQ_HOST:$RABBITMQ_PORT)"
        else
            print_ok "RabbitMQ (AMQP OK, Management API não acessível)"
        fi
    else
        print_fail "RabbitMQ"
    fi
}

check_http_service() {
    local service_name=$1
    local service_url=$2
    local health_endpoint="${3:-/health}"
    
    print_checking "$service_name"
    
    # Tentar acessar endpoint de health
    if curl -s -f --max-time 5 "$service_url$health_endpoint" &> /dev/null; then
        print_ok "$service_name ($service_url)"
    else
        # Tentar apenas verificar se a porta está aberta
        local host=$(echo "$service_url" | sed -e 's|http://||' -e 's|https://||' -e 's|/.*||' -e 's|:.*||')
        local port=$(echo "$service_url" | grep -oP ':\K[0-9]+' || echo "80")
        
        if nc -z "$host" "$port" 2>/dev/null; then
            print_warning "$service_name (porta aberta, mas endpoint $health_endpoint não responde)"
            ((FAILURES++))
        else
            print_fail "$service_name"
        fi
    fi
}

###############################################################################
# Função Principal
###############################################################################

main() {
    print_header "🏥 Verificação de Saúde dos Serviços - Sistema de Pedidos Online"
    
    echo -e "${BLUE}Verificando infraestrutura...${NC}"
    check_postgresql
    check_mongodb
    check_rabbitmq
    
    echo -e "\n${BLUE}Verificando microserviços...${NC}"
    check_http_service "User Service" "$USER_SERVICE_URL"
    check_http_service "Order Service" "$ORDER_SERVICE_URL"
    check_http_service "Notification Service" "$NOTIFICATION_SERVICE_URL"
    
    echo -e "\n${BLUE}Verificando gateway e frontend...${NC}"
    check_http_service "API Gateway" "$API_GATEWAY_URL"
    check_http_service "Frontend" "$FRONTEND_URL" "/"
    
    # Resumo
    print_header "Resumo da Verificação"
    
    if [ $FAILURES -eq 0 ]; then
        echo -e "${GREEN}${BOLD}🎉 Todos os serviços estão funcionando corretamente!${NC}\n"
        exit 0
    else
        echo -e "${RED}${BOLD}⚠️  $FAILURES serviço(s) com problemas detectado(s)${NC}"
        echo -e "${YELLOW}Verifique os logs acima para mais detalhes${NC}\n"
        exit 1
    fi
}

###############################################################################
# Execução
###############################################################################

main
