@echo off
REM ============================================================================
REM Script de Setup Completo para Windows
REM ============================================================================
REM Este script automatiza o setup do projeto no Windows
REM Uso: setup.bat
REM ============================================================================

echo ============================================================================
echo Setup Completo do Projeto - Sistema de Pedidos Online
echo ============================================================================
echo.

REM Verificar Docker
echo [1/8] Verificando Docker...
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERRO] Docker nao encontrado! Por favor, instale o Docker Desktop.
    echo Download: https://www.docker.com/products/docker-desktop
    pause
    exit /b 1
)
echo [OK] Docker instalado

REM Verificar Docker Compose
echo [2/8] Verificando Docker Compose...
docker-compose --version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERRO] Docker Compose nao encontrado!
    pause
    exit /b 1
)
echo [OK] Docker Compose instalado

REM Verificar Node.js
echo [3/8] Verificando Node.js...
node --version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERRO] Node.js nao encontrado! Por favor, instale o Node.js.
    echo Download: https://nodejs.org/
    pause
    exit /b 1
)
echo [OK] Node.js instalado

REM Instalar dependencias do Frontend
echo [4/8] Instalando dependencias do Frontend...
cd frontend
if exist package.json (
    call npm install
    echo [OK] Dependencias do Frontend instaladas
) else (
    echo [AVISO] package.json do frontend nao encontrado
)
cd ..

REM Instalar dependencias dos Scripts
echo [5/8] Instalando dependencias dos Scripts...
cd scripts
if exist package.json (
    call npm install
    echo [OK] Dependencias dos Scripts instaladas
) else (
    echo [AVISO] package.json dos scripts nao encontrado
)
cd ..

REM Build dos containers
echo [6/8] Building containers Docker...
echo Isso pode levar alguns minutos...
docker-compose build --no-cache
if %errorlevel% neq 0 (
    echo [ERRO] Falha ao buildar containers
    pause
    exit /b 1
)
echo [OK] Containers buildados

REM Iniciar containers
echo [7/8] Iniciando containers...
docker-compose up -d
if %errorlevel% neq 0 (
    echo [ERRO] Falha ao iniciar containers
    pause
    exit /b 1
)
echo [OK] Containers iniciados

REM Aguardar servicos
echo [8/8] Aguardando servicos iniciarem...
timeout /t 30 /nobreak

echo.
echo ============================================================================
echo Setup Concluido!
echo ============================================================================
echo.
echo Servicos Disponiveis:
echo   - Frontend:            http://localhost:3000
echo   - API Gateway:         http://localhost:8000
echo   - RabbitMQ Management: http://localhost:15672 (guest/guest)
echo.
echo Proximos passos:
echo   1. Aguarde mais alguns segundos para todos os servicos iniciarem
echo   2. Inicialize os bancos de dados: scripts\init-db.sh (via Git Bash)
echo      ou execute manualmente os seeds
echo   3. Verifique a saude: scripts\check-services.ps1
echo   4. Teste a API: scripts\test-api.sh (via Git Bash)
echo.
echo Comandos uteis:
echo   docker-compose ps      - Ver status dos containers
echo   docker-compose logs    - Ver logs
echo   docker-compose down    - Parar containers
echo.
echo Documentacao:
echo   README.md              - Visao geral do projeto
echo   QUICK_REFERENCE.md     - Referencia rapida de comandos
echo   scripts\README.md      - Documentacao dos scripts
echo.
echo Bom desenvolvimento!
echo.
pause
