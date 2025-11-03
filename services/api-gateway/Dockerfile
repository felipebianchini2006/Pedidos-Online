# Dockerfile - API Gateway
# Multi-stage build para otimizar o tamanho da imagem

# Stage 1: Builder
FROM golang:alpine AS builder

# Instalar dependências necessárias
RUN apk add --no-cache git

# Definir diretório de trabalho
WORKDIR /app

# Copiar arquivos de dependências
COPY go.mod go.sum ./

# Baixar dependências
RUN go mod download

# Copiar código fonte
COPY . .

# Compilar a aplicação
# CGO_ENABLED=0 para criar binário estático
# -ldflags="-s -w" para reduzir tamanho do binário
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o api-gateway \
    cmd/main.go

# Stage 2: Runtime
FROM alpine:latest

# Instalar certificados CA para HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Criar usuário não-root para segurança
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Definir diretório de trabalho
WORKDIR /app

# Copiar binário do stage builder
COPY --from=builder /app/api-gateway .

# Definir propriedade do binário
RUN chown -R appuser:appgroup /app

# Usar usuário não-root
USER appuser

# Expor porta
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:8000/health || exit 1

# Comando para executar a aplicação
CMD ["./api-gateway"]
