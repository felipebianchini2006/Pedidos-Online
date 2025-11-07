#!/bin/bash

###############################################################################
# Script de Reset dos Bancos de Dados
###############################################################################
# Este script reseta completamente todos os bancos de dados:
# - DROP e RECREATE databases PostgreSQL
# - Limpa coleções MongoDB
# - Re-executa seeds
#
# ⚠️  ATENÇÃO: Esta operação é DESTRUTIVA e irá apagar TODOS os dados!
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
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Variáveis de ambiente (podem ser sobrescritas)
POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_USER="${POSTGRES_USER:-postgres}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-postgres}"
MONGO_URI="${MONGO_URI:-mongodb://localhost:27017}"
MONGO_DATABASE="${MONGO_DATABASE:-orders_db}"

###############################################################################
# Funções Auxiliares
###############################################################################

print_header() {
    echo -e "\n${BLUE}${BOLD}============================================================================${NC}"
    echo -e "${BLUE}${BOLD}$1${NC}"
    echo -e "${BLUE}${BOLD}============================================================================${NC}\n"
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
# Confirmação do Usuário
###############################################################################

confirm_reset() {
    print_header "⚠️  CONFIRMAÇÃO NECESSÁRIA"
    
    echo -e "${RED}${BOLD}ATENÇÃO: Esta operação irá APAGAR TODOS OS DADOS dos seguintes bancos:${NC}"
    echo -e "  ${YELLOW}• PostgreSQL: users_db${NC}"
    echo -e "  ${YELLOW}• MongoDB: orders_db (coleção orders)${NC}"
    echo ""
    echo -e "${RED}${BOLD}Esta ação NÃO PODE SER DESFEITA!${NC}"
    echo ""
    
    read -p "Tem certeza que deseja continuar? (Digite 'yes' para confirmar): " confirmation
    
    if [ "$confirmation" != "yes" ]; then
        print_warning "Reset cancelado pelo usuário"
        exit 0
    fi
    
    echo ""
    print_info "Confirmação recebida. Iniciando reset..."
    sleep 2
}

###############################################################################
# PostgreSQL - Drop e Recreate Databases
###############################################################################

reset_postgres() {
    print_header "Resetando PostgreSQL"
    
    export PGPASSWORD="$POSTGRES_PASSWORD"
    
    # Drop database users_db se existir
    print_info "Dropando database users_db..."
    psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -c "DROP DATABASE IF EXISTS users_db;" 2>/dev/null || true
    print_success "Database users_db dropado"
    
    # Recreate database users_db
    print_info "Recriando database users_db..."
    psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -c "CREATE DATABASE users_db;"
    print_success "Database users_db recriado"
    
    unset PGPASSWORD
}

###############################################################################
# PostgreSQL - Rodar Migrations
###############################################################################

run_migrations() {
    print_header "Aplicando Migrations"
    
    export PGPASSWORD="$POSTGRES_PASSWORD"
    
    MIGRATIONS_DIR="$PROJECT_ROOT/services/user-service/migrations"
    
    if [ -d "$MIGRATIONS_DIR" ]; then
        print_info "Aplicando migrations..."
        
        # Rodar migrations .up.sql
        for migration in "$MIGRATIONS_DIR"/*.up.sql; do
            if [ -f "$migration" ]; then
                print_info "Aplicando: $(basename "$migration")"
                psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d users_db -f "$migration"
            fi
        done
        
        print_success "Migrations aplicadas"
    else
        print_warning "Diretório de migrations não encontrado"
    fi
    
    unset PGPASSWORD
}

###############################################################################
# MongoDB - Limpar Coleções
###############################################################################

reset_mongodb() {
    print_header "Resetando MongoDB"
    
    print_info "Limpando coleção orders..."
    
    # Verificar se mongosh está disponível (novo client)
    if command -v mongosh &> /dev/null; then
        mongosh "$MONGO_URI/$MONGO_DATABASE" --quiet --eval "db.orders.deleteMany({}); print('Coleção orders limpa');"
    # Fallback para mongo (client legado)
    elif command -v mongo &> /dev/null; then
        mongo "$MONGO_URI/$MONGO_DATABASE" --quiet --eval "db.orders.deleteMany({}); print('Coleção orders limpa');"
    else
        print_warning "mongosh/mongo não encontrado, tentando via Node.js..."
        
        # Criar script temporário Node.js
        cat > /tmp/reset-mongo.js << 'EOF'
const { MongoClient } = require('mongodb');

async function reset() {
    const client = new MongoClient(process.env.MONGO_URI);
    try {
        await client.connect();
        const db = client.db(process.env.MONGO_DATABASE);
        const result = await db.collection('orders').deleteMany({});
        console.log(`Removidos ${result.deletedCount} documentos da coleção orders`);
    } finally {
        await client.close();
    }
}

reset().catch(console.error);
EOF
        
        MONGO_URI="$MONGO_URI" MONGO_DATABASE="$MONGO_DATABASE" node /tmp/reset-mongo.js
        rm /tmp/reset-mongo.js
    fi
    
    print_success "Coleção orders limpa (índices preservados)"
}

###############################################################################
# Re-executar Seeds
###############################################################################

run_seeds() {
    print_header "Re-populando Dados de Teste"
    
    # Seed PostgreSQL
    print_info "Executando seed do PostgreSQL..."
    export PGPASSWORD="$POSTGRES_PASSWORD"
    
    SEED_FILE="$SCRIPT_DIR/seed-data.sql"
    if [ -f "$SEED_FILE" ]; then
        psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d users_db -f "$SEED_FILE"
        print_success "Seed do PostgreSQL concluído"
    else
        print_error "Arquivo seed-data.sql não encontrado"
    fi
    
    unset PGPASSWORD
    
    # Seed MongoDB
    print_info "Executando seed do MongoDB..."
    SEED_SCRIPT="$SCRIPT_DIR/seed-orders.js"
    
    if [ -f "$SEED_SCRIPT" ]; then
        cd "$SCRIPT_DIR"
        
        # Instalar dependência se necessário
        if [ ! -d "node_modules/mongodb" ]; then
            print_info "Instalando dependência mongodb..."
            npm install mongodb --silent
        fi
        
        MONGO_URI="$MONGO_URI" MONGO_DATABASE="$MONGO_DATABASE" node "$SEED_SCRIPT"
        print_success "Seed do MongoDB concluído"
    else
        print_error "Arquivo seed-orders.js não encontrado"
    fi
}

###############################################################################
# Verificação Final
###############################################################################

verify_reset() {
    print_header "Verificando Reset"
    
    export PGPASSWORD="$POSTGRES_PASSWORD"
    
    # Contar usuários no PostgreSQL
    print_info "Verificando PostgreSQL..."
    USER_COUNT=$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d users_db -t -c "SELECT COUNT(*) FROM users;")
    echo "  Usuários criados: $USER_COUNT"
    
    unset PGPASSWORD
    
    # Contar pedidos no MongoDB
    print_info "Verificando MongoDB..."
    if command -v mongosh &> /dev/null; then
        ORDER_COUNT=$(mongosh "$MONGO_URI/$MONGO_DATABASE" --quiet --eval "db.orders.countDocuments()")
        echo "  Pedidos criados: $ORDER_COUNT"
    elif command -v mongo &> /dev/null; then
        ORDER_COUNT=$(mongo "$MONGO_URI/$MONGO_DATABASE" --quiet --eval "db.orders.countDocuments()")
        echo "  Pedidos criados: $ORDER_COUNT"
    else
        print_warning "Não foi possível verificar contagem do MongoDB"
    fi
    
    print_success "Verificação concluída"
}

###############################################################################
# Função Principal
###############################################################################

main() {
    print_header "🔄 Reset dos Bancos de Dados - Sistema de Pedidos Online"
    
    # Solicitar confirmação
    confirm_reset
    
    # Executar operações de reset
    reset_postgres
    run_migrations
    reset_mongodb
    run_seeds
    verify_reset
    
    print_header "✅ Reset Concluído com Sucesso!"
    
    print_info "Próximos passos:"
    echo "  1. Use 'make check' para verificar a saúde dos serviços"
    echo "  2. Use 'make test-api' para testar a API"
    echo ""
}

###############################################################################
# Execução
###############################################################################

main
