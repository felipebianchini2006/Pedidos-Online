# ==============================================================================
# Builder Stage - Compilar o binário Go
# ==============================================================================
FROM golang:1.23-alpine AS builder

# Instalar dependências necessárias para compilação
RUN apk add --no-cache git ca-certificates tzdata

# Definir diretório de trabalho
WORKDIR /app

# Copiar arquivos de dependências
COPY go.mod go.sum ./

# Baixar dependências (camada cacheada)
RUN go mod download && go mod verify

# Copiar código fonte
COPY . .

# Compilar o binário com otimizações
# CGO_ENABLED=0: Desabilitar CGO para binário estático
# -ldflags="-s -w": Remover símbolos de debug para reduzir tamanho
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /app/bin/notification-service \
    ./cmd/main.go

# Verificar se o binário foi criado
RUN ls -lh /app/bin/notification-service

# ==============================================================================
# Runtime Stage - Imagem mínima para execução
# ==============================================================================
FROM alpine:3.19

# Instalar certificados CA (necessário para HTTPS/TLS)
RUN apk --no-cache add ca-certificates tzdata

# Criar usuário não-root para execução
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Definir diretório de trabalho
WORKDIR /home/appuser

# Copiar binário do stage builder
COPY --from=builder --chown=appuser:appgroup /app/bin/notification-service ./notification-service

# Mudar para usuário não-root
USER appuser

# Expor porta do serviço
EXPOSE 8003

# Health check para verificar se o serviço está respondendo
# Intervalo: 30s, Timeout: 10s, Start period: 40s (tempo de inicialização), Retries: 3
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8003/health || exit 1

# Comando para executar o serviço
CMD ["./notification-service"]
