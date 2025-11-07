#!/bin/bash

################################################################################
# Script para dar permissões de execução aos scripts
################################################################################

echo "🔧 Configurando permissões de execução dos scripts..."

# Diretório dos scripts
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Dar permissão de execução a todos os scripts .sh
chmod +x "$SCRIPT_DIR"/*.sh

echo "✅ Permissões configuradas com sucesso!"
echo ""
echo "Scripts disponíveis:"
ls -lh "$SCRIPT_DIR"/*.sh
