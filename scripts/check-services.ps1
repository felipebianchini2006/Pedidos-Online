################################################################################
# Script de Verificação de Saúde dos Serviços - PowerShell
################################################################################
# Este script verifica se todos os serviços do sistema estão rodando no Windows
# Uso: .\check-services.ps1
################################################################################

# Configurações
$POSTGRES_HOST = if ($env:POSTGRES_HOST) { $env:POSTGRES_HOST } else { "localhost" }
$POSTGRES_PORT = if ($env:POSTGRES_PORT) { $env:POSTGRES_PORT } else { 5432 }
$POSTGRES_USER = if ($env:POSTGRES_USER) { $env:POSTGRES_USER } else { "postgres" }
$POSTGRES_PASSWORD = if ($env:POSTGRES_PASSWORD) { $env:POSTGRES_PASSWORD } else { "postgres" }
$MONGO_URI = if ($env:MONGO_URI) { $env:MONGO_URI } else { "mongodb://localhost:27017" }
$USER_SERVICE_URL = if ($env:USER_SERVICE_URL) { $env:USER_SERVICE_URL } else { "http://localhost:8001" }
$ORDER_SERVICE_URL = if ($env:ORDER_SERVICE_URL) { $env:ORDER_SERVICE_URL } else { "http://localhost:8002" }
$NOTIFICATION_SERVICE_URL = if ($env:NOTIFICATION_SERVICE_URL) { $env:NOTIFICATION_SERVICE_URL } else { "http://localhost:8003" }
$API_GATEWAY_URL = if ($env:API_GATEWAY_URL) { $env:API_GATEWAY_URL } else { "http://localhost:8000" }
$FRONTEND_URL = if ($env:FRONTEND_URL) { $env:FRONTEND_URL } else { "http://localhost:5173" }

$FAILURES = 0

################################################################################
# Funções Auxiliares
################################################################################

function Write-Header {
    param([string]$Message)
    Write-Host ""
    Write-Host "============================================================================" -ForegroundColor Blue
    Write-Host $Message -ForegroundColor Blue
    Write-Host "============================================================================" -ForegroundColor Blue
    Write-Host ""
}

function Write-Checking {
    param([string]$Service)
    Write-Host "🔍 Verificando $Service..." -NoNewline -ForegroundColor Blue
}

function Write-OK {
    param([string]$Service)
    Write-Host "`r✅ $Service está OK                    " -ForegroundColor Green
}

function Write-Fail {
    param([string]$Service)
    Write-Host "`r❌ $Service está FALHANDO              " -ForegroundColor Red
    $script:FAILURES++
}

function Write-Warning {
    param([string]$Message)
    Write-Host "⚠️  $Message" -ForegroundColor Yellow
}

################################################################################
# Verificações de Serviços
################################################################################

function Test-PostgreSQL {
    Write-Checking "PostgreSQL"
    
    try {
        $result = Test-NetConnection -ComputerName $POSTGRES_HOST -Port $POSTGRES_PORT -WarningAction SilentlyContinue
        
        if ($result.TcpTestSucceeded) {
            Write-OK "PostgreSQL ($POSTGRES_HOST`:$POSTGRES_PORT)"
        } else {
            Write-Fail "PostgreSQL"
        }
    } catch {
        Write-Fail "PostgreSQL"
    }
}

function Test-MongoDB {
    Write-Checking "MongoDB"
    
    try {
        # Extrair host e porta do URI
        $mongoHost = "localhost"
        $mongoPort = 27017
        
        if ($MONGO_URI -match "mongodb://([^:]+):(\d+)") {
            $mongoHost = $Matches[1]
            $mongoPort = [int]$Matches[2]
        }
        
        $result = Test-NetConnection -ComputerName $mongoHost -Port $mongoPort -WarningAction SilentlyContinue
        
        if ($result.TcpTestSucceeded) {
            Write-OK "MongoDB"
        } else {
            Write-Fail "MongoDB"
        }
    } catch {
        Write-Fail "MongoDB"
    }
}

function Test-RabbitMQ {
    Write-Checking "RabbitMQ"
    
    try {
        $result = Test-NetConnection -ComputerName localhost -Port 5672 -WarningAction SilentlyContinue
        
        if ($result.TcpTestSucceeded) {
            Write-OK "RabbitMQ"
        } else {
            Write-Fail "RabbitMQ"
        }
    } catch {
        Write-Fail "RabbitMQ"
    }
}

function Test-HttpService {
    param(
        [string]$ServiceName,
        [string]$ServiceUrl,
        [string]$HealthEndpoint = "/health"
    )
    
    Write-Checking $ServiceName
    
    try {
        $response = Invoke-WebRequest -Uri "$ServiceUrl$HealthEndpoint" -TimeoutSec 5 -UseBasicParsing -ErrorAction Stop
        
        if ($response.StatusCode -eq 200) {
            Write-OK "$ServiceName ($ServiceUrl)"
        } else {
            Write-Warning "$ServiceName (resposta inesperada: $($response.StatusCode))"
            $script:FAILURES++
        }
    } catch {
        # Tentar verificar se a porta está aberta
        $uri = [System.Uri]$ServiceUrl
        $result = Test-NetConnection -ComputerName $uri.Host -Port $uri.Port -WarningAction SilentlyContinue
        
        if ($result.TcpTestSucceeded) {
            Write-Warning "$ServiceName (porta aberta, mas endpoint $HealthEndpoint não responde)"
            $script:FAILURES++
        } else {
            Write-Fail $ServiceName
        }
    }
}

################################################################################
# Função Principal
################################################################################

function Main {
    Write-Header "🏥 Verificação de Saúde dos Serviços - Sistema de Pedidos Online"
    
    Write-Host "Verificando infraestrutura..." -ForegroundColor Blue
    Test-PostgreSQL
    Test-MongoDB
    Test-RabbitMQ
    
    Write-Host ""
    Write-Host "Verificando microserviços..." -ForegroundColor Blue
    Test-HttpService "User Service" $USER_SERVICE_URL
    Test-HttpService "Order Service" $ORDER_SERVICE_URL
    Test-HttpService "Notification Service" $NOTIFICATION_SERVICE_URL
    
    Write-Host ""
    Write-Host "Verificando gateway e frontend..." -ForegroundColor Blue
    Test-HttpService "API Gateway" $API_GATEWAY_URL
    Test-HttpService "Frontend" $FRONTEND_URL "/"
    
    # Resumo
    Write-Header "Resumo da Verificação"
    
    if ($FAILURES -eq 0) {
        Write-Host "🎉 Todos os serviços estão funcionando corretamente!" -ForegroundColor Green
        Write-Host ""
        exit 0
    } else {
        Write-Host "⚠️  $FAILURES serviço(s) com problemas detectado(s)" -ForegroundColor Red
        Write-Host "Verifique os logs acima para mais detalhes" -ForegroundColor Yellow
        Write-Host ""
        exit 1
    }
}

################################################################################
# Execução
################################################################################

Main
