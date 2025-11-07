################################################################################
# Script de Seed para PostgreSQL - PowerShell
################################################################################
# Este script popula a tabela 'users' com dados de teste no Windows
# Uso: .\seed-data.ps1
################################################################################

# Configurações
$POSTGRES_HOST = if ($env:POSTGRES_HOST) { $env:POSTGRES_HOST } else { "localhost" }
$POSTGRES_PORT = if ($env:POSTGRES_PORT) { $env:POSTGRES_PORT } else { 5432 }
$POSTGRES_USER = if ($env:POSTGRES_USER) { $env:POSTGRES_USER } else { "postgres" }
$POSTGRES_PASSWORD = if ($env:POSTGRES_PASSWORD) { $env:POSTGRES_PASSWORD } else { "postgres" }
$POSTGRES_DB = "users_db"

Write-Host "============================================================================" -ForegroundColor Blue
Write-Host "🌱 Populando PostgreSQL com dados de teste" -ForegroundColor Blue
Write-Host "============================================================================" -ForegroundColor Blue
Write-Host ""

# Verificar se psql está disponível
$psqlPath = Get-Command psql -ErrorAction SilentlyContinue

if (-not $psqlPath) {
    Write-Host "❌ psql não encontrado no PATH" -ForegroundColor Red
    Write-Host "Por favor, instale o PostgreSQL client tools" -ForegroundColor Yellow
    Write-Host "Download: https://www.postgresql.org/download/windows/" -ForegroundColor Yellow
    exit 1
}

# Definir password
$env:PGPASSWORD = $POSTGRES_PASSWORD

# Executar script SQL
$sqlFile = Join-Path $PSScriptRoot "seed-data.sql"

if (Test-Path $sqlFile) {
    Write-Host "📄 Executando seed-data.sql..." -ForegroundColor Blue
    
    try {
        psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -f $sqlFile
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host ""
            Write-Host "✅ Dados de teste populados com sucesso!" -ForegroundColor Green
            Write-Host ""
            Write-Host "💡 Credenciais de teste:" -ForegroundColor Yellow
            Write-Host "   - admin@example.com / admin123" -ForegroundColor White
            Write-Host "   - user1@example.com / user123" -ForegroundColor White
            Write-Host "   - user2@example.com / user123" -ForegroundColor White
            Write-Host ""
        } else {
            Write-Host ""
            Write-Host "❌ Erro ao executar seed" -ForegroundColor Red
            exit 1
        }
    } catch {
        Write-Host ""
        Write-Host "❌ Erro: $_" -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host "❌ Arquivo seed-data.sql não encontrado: $sqlFile" -ForegroundColor Red
    exit 1
}

# Limpar password
Remove-Item Env:\PGPASSWORD
