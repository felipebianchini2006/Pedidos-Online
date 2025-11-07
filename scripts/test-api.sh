#!/bin/bash

###############################################################################
# Script de Teste de API (Smoke Test)
###############################################################################
# Este script executa um smoke test completo na API:
# 1. Registrar novo usuário
# 2. Fazer login e obter token JWT
# 3. Criar novo pedido (autenticado)
# 4. Listar pedidos do usuário (autenticado)
# 5. Obter detalhes do pedido (autenticado)
###############################################################################

set -e  # Exit on error

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Configurações
API_GATEWAY_URL="${API_GATEWAY_URL:-http://localhost:8000}"
TEST_USER_EMAIL="test_$(date +%s)@example.com"
TEST_USER_PASSWORD="test123"
TEST_USER_NAME="Usuário Teste"
TEST_USER_PHONE="+55 11 99999-9999"

# Variáveis globais
JWT_TOKEN=""
ORDER_ID=""
FAILURES=0

###############################################################################
# Funções Auxiliares
###############################################################################

print_header() {
    echo -e "\n${BLUE}${BOLD}============================================================================${NC}"
    echo -e "${BLUE}${BOLD}$1${NC}"
    echo -e "${BLUE}${BOLD}============================================================================${NC}\n"
}

print_step() {
    echo -e "\n${BLUE}▶️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
    ((FAILURES++))
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_json() {
    echo -e "${YELLOW}$1${NC}" | jq . 2>/dev/null || echo -e "${YELLOW}$1${NC}"
}

check_http_status() {
    local expected=$1
    local actual=$2
    local step=$3
    
    if [ "$actual" -eq "$expected" ]; then
        print_success "$step (Status: $actual)"
        return 0
    else
        print_error "$step - Esperado: $expected, Recebido: $actual"
        return 1
    fi
}

###############################################################################
# Testes de API
###############################################################################

test_register() {
    print_step "Teste 1: Registrar Novo Usuário"
    print_info "Email: $TEST_USER_EMAIL"
    
    local response=$(curl -s -w "\n%{http_code}" -X POST \
        "$API_GATEWAY_URL/api/users/register" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$TEST_USER_EMAIL\",
            \"password\": \"$TEST_USER_PASSWORD\",
            \"name\": \"$TEST_USER_NAME\",
            \"phone\": \"$TEST_USER_PHONE\"
        }")
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    echo "Response Body:"
    print_json "$body"
    
    if check_http_status 201 "$http_code" "Registro de usuário"; then
        # Verificar se o response contém os dados esperados
        if echo "$body" | jq -e '.data.email' &>/dev/null; then
            local email=$(echo "$body" | jq -r '.data.email')
            if [ "$email" == "$TEST_USER_EMAIL" ]; then
                print_success "Email do usuário confirmado: $email"
            else
                print_error "Email divergente no response"
            fi
        fi
    fi
}

test_login() {
    print_step "Teste 2: Login e Obtenção de Token JWT"
    print_info "Email: $TEST_USER_EMAIL"
    
    local response=$(curl -s -w "\n%{http_code}" -X POST \
        "$API_GATEWAY_URL/api/users/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$TEST_USER_EMAIL\",
            \"password\": \"$TEST_USER_PASSWORD\"
        }")
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    echo "Response Body:"
    print_json "$body"
    
    if check_http_status 200 "$http_code" "Login"; then
        # Extrair token JWT
        JWT_TOKEN=$(echo "$body" | jq -r '.data.token // .token // empty')
        
        if [ -n "$JWT_TOKEN" ] && [ "$JWT_TOKEN" != "null" ]; then
            print_success "Token JWT obtido: ${JWT_TOKEN:0:20}..."
        else
            print_error "Token JWT não encontrado no response"
            exit 1
        fi
    else
        exit 1
    fi
}

test_create_order() {
    print_step "Teste 3: Criar Novo Pedido (Autenticado)"
    
    if [ -z "$JWT_TOKEN" ]; then
        print_error "Token JWT não disponível. Execute o teste de login primeiro."
        exit 1
    fi
    
    local response=$(curl -s -w "\n%{http_code}" -X POST \
        "$API_GATEWAY_URL/api/orders" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -d '{
            "items": [
                {
                    "product_id": "PROD-TEST-001",
                    "product_name": "Produto Teste 1",
                    "quantity": 2,
                    "price": 100.00
                },
                {
                    "product_id": "PROD-TEST-002",
                    "product_name": "Produto Teste 2",
                    "quantity": 1,
                    "price": 50.00
                }
            ],
            "total_amount": 250.00,
            "address": {
                "street": "Rua Teste",
                "number": "123",
                "city": "São Paulo",
                "state": "SP",
                "zip_code": "01234-567",
                "complement": "Apto 42"
            }
        }')
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    echo "Response Body:"
    print_json "$body"
    
    if check_http_status 201 "$http_code" "Criação de pedido"; then
        # Extrair ID do pedido
        ORDER_ID=$(echo "$body" | jq -r '.data.id // .id // empty')
        
        if [ -n "$ORDER_ID" ] && [ "$ORDER_ID" != "null" ]; then
            print_success "Pedido criado com ID: $ORDER_ID"
        else
            print_warning "ID do pedido não encontrado no response"
        fi
        
        # Verificar status do pedido
        local status=$(echo "$body" | jq -r '.data.status // .status // empty')
        if [ "$status" == "PENDING" ] || [ "$status" == "pending" ]; then
            print_success "Status do pedido: $status"
        else
            print_warning "Status do pedido inesperado: $status"
        fi
    fi
}

test_list_orders() {
    print_step "Teste 4: Listar Pedidos do Usuário (Autenticado)"
    
    if [ -z "$JWT_TOKEN" ]; then
        print_error "Token JWT não disponível"
        exit 1
    fi
    
    local response=$(curl -s -w "\n%{http_code}" -X GET \
        "$API_GATEWAY_URL/api/orders" \
        -H "Authorization: Bearer $JWT_TOKEN")
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    echo "Response Body:"
    print_json "$body"
    
    if check_http_status 200 "$http_code" "Listagem de pedidos"; then
        # Verificar se há pelo menos um pedido
        local order_count=$(echo "$body" | jq -r '.data | length' 2>/dev/null || echo "0")
        
        if [ "$order_count" -gt 0 ]; then
            print_success "Encontrados $order_count pedido(s)"
        else
            print_warning "Nenhum pedido encontrado (esperado pelo menos 1)"
        fi
    fi
}

test_get_order_details() {
    print_step "Teste 5: Obter Detalhes do Pedido (Autenticado)"
    
    if [ -z "$JWT_TOKEN" ]; then
        print_error "Token JWT não disponível"
        exit 1
    fi
    
    if [ -z "$ORDER_ID" ]; then
        print_warning "ID do pedido não disponível. Pulando teste de detalhes."
        return
    fi
    
    local response=$(curl -s -w "\n%{http_code}" -X GET \
        "$API_GATEWAY_URL/api/orders/$ORDER_ID" \
        -H "Authorization: Bearer $JWT_TOKEN")
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    echo "Response Body:"
    print_json "$body"
    
    if check_http_status 200 "$http_code" "Detalhes do pedido"; then
        # Verificar se o ID do pedido corresponde
        local returned_id=$(echo "$body" | jq -r '.data.id // .id // empty')
        
        if [ "$returned_id" == "$ORDER_ID" ]; then
            print_success "ID do pedido confirmado: $ORDER_ID"
        else
            print_warning "ID do pedido divergente: esperado $ORDER_ID, recebido $returned_id"
        fi
        
        # Verificar se contém items
        local items_count=$(echo "$body" | jq -r '.data.items | length' 2>/dev/null || echo "0")
        if [ "$items_count" -gt 0 ]; then
            print_success "Pedido contém $items_count item(s)"
        fi
    fi
}

test_unauthorized_access() {
    print_step "Teste 6: Verificar Proteção de Rota (Sem Token)"
    
    local response=$(curl -s -w "\n%{http_code}" -X GET \
        "$API_GATEWAY_URL/api/orders")
    
    local http_code=$(echo "$response" | tail -n1)
    
    if check_http_status 401 "$http_code" "Acesso sem autenticação bloqueado"; then
        print_success "Proteção de rota funcionando corretamente"
    fi
}

###############################################################################
# Função Principal
###############################################################################

main() {
    print_header "🧪 Teste de API (Smoke Test) - Sistema de Pedidos Online"
    
    print_info "API Gateway: $API_GATEWAY_URL"
    print_info "Usuário de teste: $TEST_USER_EMAIL"
    
    # Verificar se curl está instalado
    if ! command -v curl &> /dev/null; then
        print_error "curl não está instalado"
        exit 1
    fi
    
    # Verificar se jq está instalado (opcional, mas recomendado)
    if ! command -v jq &> /dev/null; then
        print_warning "jq não está instalado (formatação JSON será limitada)"
    fi
    
    # Executar testes
    test_register
    test_login
    test_create_order
    test_list_orders
    test_get_order_details
    test_unauthorized_access
    
    # Resumo
    print_header "Resumo dos Testes"
    
    if [ $FAILURES -eq 0 ]; then
        echo -e "${GREEN}${BOLD}🎉 Todos os testes passaram com sucesso!${NC}\n"
        exit 0
    else
        echo -e "${RED}${BOLD}⚠️  $FAILURES teste(s) falharam${NC}"
        echo -e "${YELLOW}Verifique os logs acima para mais detalhes${NC}\n"
        exit 1
    fi
}

###############################################################################
# Execução
###############################################################################

main
