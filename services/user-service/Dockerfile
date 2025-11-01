# ==============================================================================
# Multi-stage Dockerfile for User Service
# ==============================================================================
# Este Dockerfile utiliza multi-stage build para criar uma imagem Docker 
# otimizada e segura do User Service.
#
# Etapas:
# 1. Builder Stage: Compila o código Go em um binário estático
# 2. Runtime Stage: Cria uma imagem minimal com o binário compilado
#
# Tamanho esperado da imagem final: < 20MB
# ==============================================================================

# ==============================================================================
# STAGE 1: Builder
# ==============================================================================
# Usa imagem golang:1.23-alpine para compilar o código Go
# Alpine é escolhida por ser leve (~5MB base) e ter as ferramentas necessárias
# Versão 1.23 é necessária para suportar go.mod com go 1.24.0
FROM golang:1.23-alpine AS builder

# Metadados da imagem (visíveis com docker inspect)
LABEL stage=builder
LABEL description="Builder stage for User Service"

# Instalar dependências necessárias para o build
# - git: para baixar dependências Go que usam git
# - ca-certificates: certificados SSL/TLS necessários para HTTPS
# - tzdata: fusos horários para timestamps corretos
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata

# Definir diretório de trabalho
WORKDIR /app

# ==============================================================================
# CACHE DE DEPENDÊNCIAS GO
# ==============================================================================
# Copiar apenas go.mod e go.sum primeiro para aproveitar o cache do Docker
# Se essas files não mudarem, o Docker reutiliza o cache das dependências
# Isso acelera muito os builds subsequentes
COPY go.mod go.sum ./

# Configurar GOTOOLCHAIN para permitir download automático da versão necessária
# Isso permite que dependências com go 1.24+ sejam baixadas automaticamente
ENV GOTOOLCHAIN=auto

# Baixar todas as dependências
# Esta layer será cacheada se go.mod e go.sum não mudarem
RUN go mod download

# Verificar integridade das dependências
RUN go mod verify

# ==============================================================================
# BUILD DO BINÁRIO
# ==============================================================================
# Copiar todo o código fonte
COPY . .

# Compilar o binário com otimizações de tamanho e segurança
# Flags explicadas:
# - CGO_ENABLED=0: Desabilita CGO para criar binário 100% estático (sem dependências C)
# - GOOS=linux: Sistema operacional alvo (Linux)
# - -a: Força rebuild de todos os pacotes
# - -installsuffix cgo: Adiciona sufixo para diferenciar builds com/sem CGO
# - -ldflags="-w -s": Remove símbolos de debug e tabela de símbolos (reduz ~30% do tamanho)
#   * -w: Omit DWARF symbol table
#   * -s: Omit symbol table and debug info
# - -o main: Nome do binário de saída
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a \
    -installsuffix cgo \
    -ldflags="-w -s" \
    -o main \
    ./cmd/main.go

# Verificar que o binário foi criado e é executável
RUN chmod +x main

# ==============================================================================
# STAGE 2: Runtime (Imagem Final)
# ==============================================================================
# Usa Alpine Linux minimal para imagem final ultra-leve
# Esta é a imagem que será deployada em produção
FROM alpine:latest

# ==============================================================================
# METADADOS DA IMAGEM
# ==============================================================================
LABEL maintainer="felipe@pedidosonline.com"
LABEL version="1.0.0"
LABEL description="User Service - Microserviço de autenticação e gerenciamento de usuários"
LABEL service="user-service"
LABEL component="backend"
LABEL framework="go"

# ==============================================================================
# DEPENDÊNCIAS RUNTIME
# ==============================================================================
# Instalar apenas certificados SSL, timezone data e wget (para health check)
# Não instalar ferramentas de desenvolvimento ou compilação
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    wget

# ==============================================================================
# SEGURANÇA: USUÁRIO NÃO-ROOT
# ==============================================================================
# Criar um usuário não-root para executar a aplicação
# Isso segue as melhores práticas de segurança Docker
# - Usuário: appuser
# - UID: 10001 (fora do range de usuários do sistema)
# - Sem senha
# - Sem shell de login (mais seguro)
RUN adduser -D -u 10001 -s /sbin/nologin appuser

# Definir diretório de trabalho
WORKDIR /app

# Copiar binário compilado do stage builder
# --from=builder: copia do stage anterior
# --chown=appuser:appuser: define ownership para o usuário não-root
COPY --from=builder --chown=appuser:appuser /app/main .

# Copiar migrations (necessário para o User Service aplicar migrations)
# Cria diretório primeiro, depois copia se existir
RUN mkdir -p migrations
COPY --from=builder --chown=appuser:appuser /app/migrations/ ./migrations/

# Copiar certificados SSL do builder (garantir compatibilidade)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# ==============================================================================
# CONFIGURAÇÕES DO CONTAINER
# ==============================================================================
# Mudar para usuário não-root (todas as operações seguintes serão como appuser)
USER appuser

# Expor porta 8001 (porta padrão do User Service)
EXPOSE 8001

# ==============================================================================
# HEALTH CHECK
# ==============================================================================
# Configurar health check para monitoramento de saúde do container
# - interval: verificar a cada 30 segundos
# - timeout: timeout de 3 segundos para cada verificação
# - start-period: aguardar 5 segundos antes de começar (tempo de inicialização)
# - retries: marcar como unhealthy após 3 falhas consecutivas
# Usa wget ao invés de curl (mais leve e já vem no Alpine)
HEALTHCHECK --interval=30s \
            --timeout=3s \
            --start-period=5s \
            --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8001/health || exit 1

# ==============================================================================
# COMANDO DE INICIALIZAÇÃO
# ==============================================================================
# Executar o binário
# Usa forma exec (sem shell) para:
# - Receber sinais do Docker corretamente (SIGTERM, SIGKILL)
# - Permitir graceful shutdown
# - PID 1 será o processo Go (não o shell)
CMD ["./main"]

# ==============================================================================
# BUILD INSTRUCTIONS
# ==============================================================================
# Para buildar esta imagem:
#   docker build -t user-service:latest .
#
# Para rodar o container:
#   docker run -d \
#     --name user-service \
#     -p 8001:8001 \
#     --env-file .env \
#     user-service:latest
#
# Para ver os logs:
#   docker logs -f user-service
#
# Para verificar health:
#   docker inspect --format='{{.State.Health.Status}}' user-service
#
# Tamanho esperado: 15-20MB (comparado a ~800MB com imagem golang completa)
# ==============================================================================
