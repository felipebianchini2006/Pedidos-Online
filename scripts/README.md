# Scripts de Gerenciamento do Projeto

Este diretório contém scripts para facilitar o gerenciamento dos bancos de dados e testes do sistema.

## 📁 Estrutura dos Scripts

### 🌱 Seed Scripts (Popular Dados)

#### `seed-data.sql`
Script SQL para popular o PostgreSQL com usuários de teste.

**Usuários criados:**
- `admin@example.com` (senha: `admin123`)
- `user1@example.com` (senha: `user123`)
- `user2@example.com` (senha: `user123`)

**Uso:**
```bash
psql -h localhost -p 5432 -U postgres -d users_db -f scripts/seed-data.sql
```

#### `seed-orders.js`
Script Node.js para popular o MongoDB com pedidos de teste e criar índices.

**Funcionalidades:**
- Cria 4 pedidos de teste com diferentes status
- Cria índices otimizados para consultas
- Exibe resumo dos dados inseridos

**Requisitos:**
```bash
npm install mongodb
```

**Uso:**
```bash
node scripts/seed-orders.js
```

**Variáveis de Ambiente:**
- `MONGO_URI` - URI de conexão (default: `mongodb://localhost:27017`)
- `MONGO_DATABASE` - Nome do banco (default: `orders_db`)

---

### 🗄️ Database Management Scripts

#### `init-db.sh`
Inicializa todos os bancos de dados do sistema.

**O que faz:**
1. Verifica se os serviços estão rodando
2. Cria databases no PostgreSQL
3. Roda migrations
4. Popula dados de teste (PostgreSQL e MongoDB)
5. Cria índices no MongoDB

**Uso:**
```bash
bash scripts/init-db.sh
# ou
make init
```

**Variáveis de Ambiente:**
- `POSTGRES_HOST` (default: `localhost`)
- `POSTGRES_PORT` (default: `5432`)
- `POSTGRES_USER` (default: `postgres`)
- `POSTGRES_PASSWORD` (default: `postgres`)
- `MONGO_URI` (default: `mongodb://localhost:27017`)

#### `reset-db.sh`
Reseta completamente todos os bancos de dados.

**⚠️ ATENÇÃO:** Esta operação é DESTRUTIVA e irá apagar TODOS os dados!

**O que faz:**
1. Solicita confirmação do usuário
2. Dropa e recria databases PostgreSQL
3. Limpa coleções MongoDB (preservando índices)
4. Re-executa migrations
5. Re-popula dados de teste

**Uso:**
```bash
bash scripts/reset-db.sh
# ou
make reset-db
```

**Confirmação:**
Você precisará digitar `yes` para confirmar a operação.

---

### 🏥 Health Check Scripts

#### `check-services.sh`
Verifica a saúde de todos os serviços do sistema.

**Serviços verificados:**
- ✅ PostgreSQL
- ✅ MongoDB
- ✅ RabbitMQ
- ✅ User Service
- ✅ Order Service
- ✅ Notification Service
- ✅ API Gateway
- ✅ Frontend

**Uso:**
```bash
bash scripts/check-services.sh
# ou
make check
```

**Exit Codes:**
- `0` - Todos os serviços OK
- `1` - Um ou mais serviços com problemas

**Output:**
```
✅ PostgreSQL está OK
✅ MongoDB está OK
✅ RabbitMQ está OK
✅ User Service está OK
...
```

---

### 🧪 API Testing Scripts

#### `test-api.sh`
Executa smoke tests completos na API.

**Fluxo de Testes:**
1. ✅ Registrar novo usuário
2. ✅ Fazer login e obter token JWT
3. ✅ Criar novo pedido (autenticado)
4. ✅ Listar pedidos do usuário (autenticado)
5. ✅ Obter detalhes do pedido (autenticado)
6. ✅ Verificar proteção de rota (sem token)

**Uso:**
```bash
bash scripts/test-api.sh
# ou
make test-api
```

**Requisitos:**
- `curl` - Para fazer requisições HTTP
- `jq` (opcional) - Para formatação JSON

**Variáveis de Ambiente:**
- `API_GATEWAY_URL` (default: `http://localhost:8000`)

**Output:**
```
▶️  Teste 1: Registrar Novo Usuário
✅ Registro de usuário (Status: 201)
✅ Email do usuário confirmado: test_1699999999@example.com

▶️  Teste 2: Login e Obtenção de Token JWT
✅ Login (Status: 200)
✅ Token JWT obtido: eyJhbGciOiJIUzI1NiIs...

...

🎉 Todos os testes passaram com sucesso!
```

---

## 🚀 Uso via Makefile

Todos os scripts podem ser executados através do Makefile para maior conveniência:

### Comandos de Banco de Dados
```bash
# Inicializar bancos (criar, migrar e popular)
make init

# Popular apenas dados de teste
make seed

# Resetar todos os bancos (⚠️ apaga tudo!)
make reset-db
```

### Comandos de Verificação
```bash
# Verificar saúde dos serviços
make check

# Testar API com smoke tests
make test-api
```

### Setup Completo (First Time)
```bash
# Setup completo do projeto (build, start, init DB, check)
make quick-start
```

---

## 🔧 Troubleshooting

### Problema: "pg_isready: command not found"
**Solução:** Instale o PostgreSQL client tools
```bash
# Ubuntu/Debian
sudo apt-get install postgresql-client

# macOS
brew install postgresql

# Windows
# Baixe do site oficial: https://www.postgresql.org/download/windows/
```

### Problema: "mongosh: command not found"
**Solução:** Os scripts tentarão usar `mongo` (client legado) ou Node.js como fallback.

Para instalar mongosh:
```bash
# Ubuntu/Debian
wget -qO - https://www.mongodb.org/static/pgp/server-6.0.asc | sudo apt-key add -
sudo apt-get install -y mongodb-mongosh

# macOS
brew install mongosh

# Windows
# Baixe do site oficial: https://www.mongodb.com/try/download/shell
```

### Problema: "nc: command not found" ou "telnet: command not found"
**Solução:** Instale netcat ou telnet
```bash
# Ubuntu/Debian
sudo apt-get install netcat

# macOS
brew install netcat

# Windows (PowerShell)
# Use: Test-NetConnection -ComputerName localhost -Port 5432
```

### Problema: Scripts não executam no Windows
**Solução:** Use Git Bash ou WSL (Windows Subsystem for Linux)

**Alternativa:** Execute os comandos manualmente através do PowerShell:
```powershell
# PostgreSQL Seed
$env:PGPASSWORD="postgres"; psql -h localhost -p 5432 -U postgres -d users_db -f scripts\seed-data.sql

# MongoDB Seed
$env:MONGO_URI="mongodb://localhost:27017"; node scripts\seed-orders.js
```

---

## 📚 Referências

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [MongoDB Node.js Driver](https://www.mongodb.com/docs/drivers/node/current/)
- [Bash Scripting Guide](https://www.gnu.org/software/bash/manual/bash.html)
- [curl Documentation](https://curl.se/docs/)
- [jq Manual](https://stedolan.github.io/jq/manual/)

---

## 🤝 Contribuindo

Ao adicionar novos scripts:

1. Adicione comentários explicativos
2. Use cores para output (GREEN, RED, YELLOW, BLUE)
3. Trate erros adequadamente (`set -e`)
4. Adicione variáveis de ambiente configuráveis
5. Atualize este README com a documentação
6. Adicione o comando correspondente no Makefile

---

## 📝 Notas Importantes

1. **Senhas de Teste:** As senhas nos scripts de seed são exemplos. Em produção, sempre use hashes bcrypt reais.

2. **IDs de Usuário:** O script `seed-orders.js` usa IDs fictícios. Para dados reais, você precisaria buscar os IDs reais do PostgreSQL.

3. **Scripts Bash no Windows:** Recomenda-se usar Git Bash ou WSL para executar os scripts bash no Windows.

4. **Confirmação de Reset:** O script `reset-db.sh` sempre solicitará confirmação para evitar perdas acidentais de dados.

5. **Health Checks:** O script `check-services.sh` verifica apenas se os serviços estão respondendo, não valida a integridade dos dados.
