#!/bin/bash

###############################################################################
# Script de Inicialização dos Bancos de Dados
###############################################################################
# Este script inicializa todos os bancos de dados do sistema:
# - Verifica se os serviços estão rodando
# - Cria os databases no PostgreSQL
# - Roda migrations
# - Popula dados de teste (seed)
# - Cria índices no MongoDB
###############################################################################

set -e  # Exit on error

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configurações
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Variáveis de ambiente (podem ser sobrescritas)
POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_USER="${POSTGRES_USER:-postgres}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-postgres}"
MONGO_URI="${MONGO_URI:-mongodb://localhost:27017}"

###############################################################################
# Funções Auxiliares
###############################################################################

print_header() {
    echo -e "\n${BLUE}============================================================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}============================================================================${NC}\n"
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

###############################################################################
# Verificar Serviços
###############################################################################

check_services() {
    print_header "Verificando Serviços"
    
    # Verificar se o script check-services.sh existe
    if [ -f "$SCRIPT_DIR/check-services.sh" ]; then
        if bash "$SCRIPT_DIR/check-services.sh"; then
            print_success "Todos os serviços estão rodando"
            return 0
        else
            print_error "Alguns serviços não estão disponíveis"
            print_warning "Continuando mesmo assim... (alguns passos podem falhar)"
            return 1
        fi
    else
        print_warning "Script check-services.sh não encontrado, pulando verificação"
        return 0
    fi
}

###############################################################################
# PostgreSQL - Criar Databases
###############################################################################

create_postgres_databases() {
    print_header "Criando Databases no PostgreSQL"
    
    export PGPASSWORD="$POSTGRES_PASSWORD"
    
    # Criar database users_db se não existir
    print_info "Criando database users_db..."
    psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -tc "SELECT 1 FROM pg_database WHERE datname = 'users_db'" | grep -q 1 || \
        psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -c "CREATE DATABASE users_db;"
    
    print_success "Database users_db criado/verificado"
    
    unset PGPASSWORD
}

###############################################################################
# PostgreSQL - Rodar Migrations
###############################################################################

run_postgres_migrations() {
    print_header "Rodando Migrations do PostgreSQL"
    
    export PGPASSWORD="$POSTGRES_PASSWORD"
    
    # Verificar se existem migrations
    MIGRATIONS_DIR="$PROJECT_ROOT/services/user-service/migrations"
    
    if [ -d "$MIGRATIONS_DIR" ]; then
        print_info "Aplicando migrations..."
        
        # Rodar migrations .up.sql
        for migration in "$MIGRATIONS_DIR"/*.up.sql; do
            if [ -f "$migration" ]; then
                print_info "Aplicando migration: $(basename "$migration")"
                psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d users_db -f "$migration" || {
                    print_warning "Migration pode já ter sido aplicada: $(basename "$migration")"
                }
            fi
        done
        
        print_success "Migrations aplicadas"
    else
        print_warning "Diretório de migrations não encontrado: $MIGRATIONS_DIR"
    fi
    
    unset PGPASSWORD
}

###############################################################################
# PostgreSQL - Popular Dados de Teste
###############################################################################

seed_postgres_data() {
    print_header "Populando Dados de Teste no PostgreSQL"
    
    export PGPASSWORD="$POSTGRES_PASSWORD"
    
    SEED_FILE="$SCRIPT_DIR/seed-data.sql"
    
    if [ -f "$SEED_FILE" ]; then
        print_info "Executando seed-data.sql..."
        psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d users_db -f "$SEED_FILE"
        print_success "Dados de teste populados no PostgreSQL"
    else
        print_error "Arquivo seed-data.sql não encontrado: $SEED_FILE"
        return 1
    fi
    
    unset PGPASSWORD
}

###############################################################################
# MongoDB - Popular Dados de Teste e Criar Índices
###############################################################################

seed_mongodb_data() {
    print_header "Populando Dados de Teste no MongoDB"
    
    SEED_SCRIPT="$SCRIPT_DIR/seed-orders.js"
    
    if [ -f "$SEED_SCRIPT" ]; then
        print_info "Executando seed-orders.js..."
        
        # Verificar se Node.js está instalado
        if ! command -v node &> /dev/null; then
            print_error "Node.js não está instalado"
            return 1
        fi
        
        # Verificar se mongodb está instalado
        cd "$SCRIPT_DIR"
        if [ ! -d "node_modules/mongodb" ]; then
            print_info "Instalando dependência mongodb..."
            npm install mongodb
        fi
        
        # Executar script de seed
        MONGO_URI="$MONGO_URI" node "$SEED_SCRIPT"
        
        print_success "Dados de teste populados no MongoDB"
    else
        print_error "Arquivo seed-orders.js não encontrado: $SEED_SCRIPT"
        return 1
    fi
}

###############################################################################
# Função Principal
###############################################################################

main() {
    print_header "🚀 Inicializando Bancos de Dados - Sistema de Pedidos Online"
    
    echo -e "${BLUE}Configurações:${NC}"
    echo -e "  PostgreSQL Host: $POSTGRES_HOST:$POSTGRES_PORT"
    echo -e "  PostgreSQL User: $POSTGRES_USER"
    echo -e "  MongoDB URI: $MONGO_URI"
    echo ""
    
    # Verificar serviços (não bloqueia se falhar)
    check_services || true
    
    # PostgreSQL
    create_postgres_databases
    run_postgres_migrations
    seed_postgres_data
    
    # MongoDB
    seed_mongodb_data
    
    print_header "✅ Inicialização Concluída com Sucesso!"
    
    print_info "Próximos passos:"
    echo "  1. Verifique os logs acima para confirmar que tudo foi criado"
    echo "  2. Use 'make check' para verificar a saúde dos serviços"
    echo "  3. Use 'make test-api' para testar a API"
    echo ""
}

###############################################################################
# Execução
###############################################################################

main
