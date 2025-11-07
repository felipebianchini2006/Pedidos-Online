#!/bin/bash

################################################################################
# Setup Completo do Projeto - Sistema de Pedidos Online
################################################################################
# Este script automatiza o setup completo do projeto:
# 1. Verifica requisitos
# 2. Instala dependências
# 3. Sobe containers
# 4. Inicializa bancos de dados
# 5. Executa testes
#
# Uso: bash scripts/full-setup.sh
################################################################################

set -e  # Exit on error

# Cores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

################################################################################
# Funções Auxiliares
################################################################################

print_header() {
    echo -e "\n${BLUE}${BOLD}============================================================================${NC}"
    echo -e "${BLUE}${BOLD}$1${NC}"
    echo -e "${BLUE}${BOLD}============================================================================${NC}\n"
}

print_step() {
    echo -e "\n${GREEN}${BOLD}▶️  $1${NC}\n"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

################################################################################
# Verificação de Requisitos
################################################################################

check_requirements() {
    print_step "Verificando Requisitos"
    
    local missing_tools=0
    
    # Docker
    if command -v docker &> /dev/null; then
        print_success "Docker instalado: $(docker --version)"
    else
        print_error "Docker não instalado"
        ((missing_tools++))
    fi
    
    # Docker Compose
    if command -v docker-compose &> /dev/null; then
        print_success "Docker Compose instalado: $(docker-compose --version)"
    else
        print_error "Docker Compose não instalado"
        ((missing_tools++))
    fi
    
    # Node.js
    if command -v node &> /dev/null; then
        print_success "Node.js instalado: $(node --version)"
    else
        print_error "Node.js não instalado"
        ((missing_tools++))
    fi
    
    # psql (opcional)
    if command -v psql &> /dev/null; then
        print_success "PostgreSQL client instalado: $(psql --version)"
    else
        print_warning "PostgreSQL client não instalado (opcional)"
    fi
    
    # mongosh (opcional)
    if command -v mongosh &> /dev/null; then
        print_success "MongoDB Shell instalado"
    elif command -v mongo &> /dev/null; then
        print_warning "MongoDB Shell (legacy) instalado"
    else
        print_warning "MongoDB Shell não instalado (opcional)"
    fi
    
    if [ $missing_tools -gt 0 ]; then
        print_error "$missing_tools ferramenta(s) essencial(is) faltando"
        echo ""
        print_info "Por favor, instale as ferramentas faltantes:"
        echo "  - Docker: https://docs.docker.com/get-docker/"
        echo "  - Docker Compose: https://docs.docker.com/compose/install/"
        echo "  - Node.js: https://nodejs.org/"
        exit 1
    fi
}

################################################################################
# Instalação de Dependências
################################################################################

install_dependencies() {
    print_step "Instalando Dependências"
    
    # Frontend
    print_info "Instalando dependências do Frontend..."
    cd "$PROJECT_ROOT/frontend"
    if [ -f "package.json" ]; then
        npm install
        print_success "Dependências do Frontend instaladas"
    else
        print_warning "package.json do frontend não encontrado"
    fi
    
    # Scripts
    print_info "Instalando dependências dos Scripts..."
    cd "$PROJECT_ROOT/scripts"
    if [ -f "package.json" ]; then
        npm install
        print_success "Dependências dos Scripts instaladas"
    else
        print_warning "package.json dos scripts não encontrado"
    fi
    
    cd "$PROJECT_ROOT"
}

################################################################################
# Build de Containers
################################################################################

build_containers() {
    print_step "Building Containers Docker"
    
    print_info "Isso pode levar alguns minutos..."
    cd "$PROJECT_ROOT"
    
    if docker-compose build --no-cache; then
        print_success "Containers buildados com sucesso"
    else
        print_error "Falha ao buildar containers"
        exit 1
    fi
}

################################################################################
# Iniciar Containers
################################################################################

start_containers() {
    print_step "Iniciando Containers"
    
    cd "$PROJECT_ROOT"
    
    if docker-compose up -d; then
        print_success "Containers iniciados"
    else
        print_error "Falha ao iniciar containers"
        exit 1
    fi
}

################################################################################
# Aguardar Serviços
################################################################################

wait_for_services() {
    print_step "Aguardando Serviços Ficarem Prontos"
    
    print_info "Aguardando 30 segundos para serviços iniciarem..."
    sleep 30
    
    print_info "Verificando saúde dos serviços..."
    if bash "$SCRIPT_DIR/check-services.sh"; then
        print_success "Todos os serviços estão prontos"
    else
        print_warning "Alguns serviços podem não estar prontos ainda"
        print_info "Continuando mesmo assim..."
    fi
}

################################################################################
# Inicializar Bancos de Dados
################################################################################

initialize_databases() {
    print_step "Inicializando Bancos de Dados"
    
    if bash "$SCRIPT_DIR/init-db.sh"; then
        print_success "Bancos de dados inicializados"
    else
        print_error "Falha ao inicializar bancos de dados"
        print_info "Você pode tentar manualmente: make init"
        exit 1
    fi
}

################################################################################
# Executar Testes
################################################################################

run_tests() {
    print_step "Executando Testes de API"
    
    print_info "Executando smoke tests..."
    if bash "$SCRIPT_DIR/test-api.sh"; then
        print_success "Todos os testes passaram"
    else
        print_warning "Alguns testes falharam"
        print_info "Isso pode ser normal em primeira execução"
    fi
}

################################################################################
# Exibir Informações Finais
################################################################################

show_final_info() {
    print_header "✅ Setup Concluído com Sucesso!"
    
    echo -e "${GREEN}${BOLD}🎉 Projeto pronto para uso!${NC}\n"
    
    echo -e "${YELLOW}📍 Serviços Disponíveis:${NC}"
    echo -e "   ${BLUE}Frontend:${NC}            http://localhost:3000"
    echo -e "   ${BLUE}API Gateway:${NC}         http://localhost:8000"
    echo -e "   ${BLUE}User Service:${NC}        http://localhost:8001"
    echo -e "   ${BLUE}Order Service:${NC}       http://localhost:8002"
    echo -e "   ${BLUE}Notification Service:${NC} http://localhost:8003"
    echo -e "   ${BLUE}RabbitMQ Management:${NC} http://localhost:15672 (guest/guest)"
    echo ""
    
    echo -e "${YELLOW}👤 Usuários de Teste:${NC}"
    echo -e "   ${BLUE}admin@example.com${NC} / ${GREEN}admin123${NC}"
    echo -e "   ${BLUE}user1@example.com${NC} / ${GREEN}user123${NC}"
    echo -e "   ${BLUE}user2@example.com${NC} / ${GREEN}user123${NC}"
    echo ""
    
    echo -e "${YELLOW}📚 Comandos Úteis:${NC}"
    echo -e "   ${BLUE}make check${NC}        - Verificar saúde dos serviços"
    echo -e "   ${BLUE}make test-api${NC}     - Testar API"
    echo -e "   ${BLUE}make logs${NC}         - Ver logs de todos os serviços"
    echo -e "   ${BLUE}make ps${NC}           - Ver status dos containers"
    echo -e "   ${BLUE}make down${NC}         - Parar todos os containers"
    echo -e "   ${BLUE}make help${NC}         - Ver todos os comandos"
    echo ""
    
    echo -e "${YELLOW}📖 Documentação:${NC}"
    echo -e "   ${BLUE}README.md${NC}              - Visão geral do projeto"
    echo -e "   ${BLUE}QUICK_REFERENCE.md${NC}     - Referência rápida de comandos"
    echo -e "   ${BLUE}scripts/README.md${NC}      - Documentação dos scripts"
    echo ""
    
    echo -e "${GREEN}Bom desenvolvimento! 🚀${NC}\n"
}

################################################################################
# Função Principal
################################################################################

main() {
    print_header "🚀 Setup Completo do Projeto - Sistema de Pedidos Online"
    
    # Executar etapas
    check_requirements
    install_dependencies
    build_containers
    start_containers
    wait_for_services
    initialize_databases
    run_tests
    show_final_info
}

################################################################################
# Execução
################################################################################

# Trap para capturar erros
trap 'echo -e "\n${RED}❌ Setup interrompido. Verifique os logs acima.${NC}\n"; exit 1' ERR

main
