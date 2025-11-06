# Componentes de Pedidos

Este diretório contém todos os componentes relacionados ao gerenciamento de pedidos.

## Componentes Criados

### 1. OrderList.jsx
Lista todos os pedidos do usuário com paginação e filtros.

**Características:**
- Carrega pedidos automaticamente ao montar
- Estados: orders, loading, error, page, totalPages
- Paginação com botões Anterior/Próximo
- Cada pedido exibido em card com:
  - Número do pedido (ID curto)
  - Data de criação formatada
  - Status com badge colorido
  - Total do pedido formatado em R$
  - Botão "Ver Detalhes" para navegação
- Empty state: "Nenhum Pedido Encontrado"
- Loading state: Skeleton screens animados
- Error state: Mensagem com botão "Tentar novamente"
- Filtro por status: Dropdown com todos os status
- Design: Grid responsivo (1 col mobile, 2-3 cols desktop)
- Botão flutuante "Novo Pedido" no mobile

**Uso:**
```jsx
import { OrderList } from './components/Orders';

<OrderList />
```

### 2. OrderForm.jsx
Formulário completo para criar novo pedido.

**Características:**

**Seção "Itens do Pedido":**
- Lista de itens adicionados com opção de remover
- Botão "Adicionar Item" que abre formulário inline
- Cada item mostra: nome, quantidade, preço, subtotal
- Validação completa de campos

**Formulário de Item:**
- Campos: product_name, quantity, price
- Validação: campos obrigatórios, quantity > 0, price > 0
- Feedback visual de erros

**Seção "Endereço de Entrega":**
- Campos: CEP, Estado, Rua, Número, Complemento, Cidade
- Busca automática de CEP via API ViaCEP
- Preenche automaticamente rua, cidade e estado
- Validação: todos obrigatórios exceto complemento
- Dropdown de estados brasileiros

**Resumo do Pedido:**
- Total de itens adicionados
- Valor total formatado em R$
- Card com destaque visual

**Funcionalidades:**
- Validação que há pelo menos 1 item
- Loading durante criação do pedido
- Redireciona para /orders/:id após sucesso
- Toast de sucesso/erro
- Botão "Cancelar" para voltar

**Uso:**
```jsx
import { OrderForm } from './components/Orders';

<OrderForm />
```

### 3. OrderDetails.jsx
Exibe detalhes completos de um pedido específico.

**Características:**

**Header:**
- Número do pedido
- Status badge colorido
- Data de criação formatada
- Botão "Voltar" para lista

**Seção "Itens":**
- Tabela completa com colunas:
  - Produto
  - Quantidade
  - Preço Unitário
  - Subtotal
- Total geral em destaque

**Seção "Endereço de Entrega":**
- Endereço completo formatado
- Rua, número, complemento
- Cidade, estado, CEP

**Timeline de Status:**
- Timeline vertical visual
- Ícones para cada status
- Status atual destacado
- Steps completados em verde
- Steps futuros em cinza
- Tratamento especial para cancelado

**Botão "Cancelar Pedido":**
- Disponível se status for pending ou confirmed
- Modal de confirmação antes de cancelar
- Loading durante cancelamento
- Atualização em tempo real após cancelamento

**Estados:**
- Loading: Spinner centralizado
- Error: Mensagem com botões de ação
- Success: Exibição completa dos dados

**Uso:**
```jsx
import { OrderDetails } from './components/Orders';

<Route path="/orders/:id" element={<OrderDetails />} />
```

### 4. StatusBadge.jsx
Badge reutilizável para exibir status de pedidos.

**Características:**
- Componente simples e reutilizável
- Props: status, className
- Cores dinâmicas baseadas no status:
  - pending: Amarelo
  - confirmed: Azul
  - preparing: Roxo
  - shipped: Índigo
  - delivered: Verde
  - cancelled: Vermelho
- Labels em português
- Design: Pills arredondados com borda

**Uso:**
```jsx
import { StatusBadge } from './components/Orders';

<StatusBadge status="pending" />
<StatusBadge status="delivered" />
```

### 5. OrderTimeline.jsx
Timeline vertical mostrando o progresso do pedido.

**Características:**
- Timeline visual com ícones SVG
- Steps do pedido:
  1. Pedido Criado (pending)
  2. Confirmado (confirmed)
  3. Em Preparação (preparing)
  4. Enviado (shipped)
  5. Entregue (delivered)
- Status atual com animação pulse
- Steps completados: Verde
- Step atual: Azul com pulse
- Steps futuros: Cinza
- Tratamento especial para cancelado: Card vermelho
- Ícones diferentes para cada status

**Uso:**
```jsx
import { OrderTimeline } from './components/Orders';

<OrderTimeline currentStatus="shipped" />
```

## Helpers Utilitários

### orderHelpers.js
Funções auxiliares para manipulação de dados de pedidos.

**Funções de Formatação:**
- `formatCurrency(value)` - Formata para R$
- `formatDate(dateString)` - Data e hora em PT-BR
- `formatDateShort(dateString)` - Apenas data
- `formatCEP(cep)` - Adiciona hífen no CEP

**Funções de Status:**
- `getStatusColor(status)` - Retorna classes Tailwind para cor
- `getStatusLabel(status)` - Retorna label em português
- `getStatusIcon(status)` - Retorna SVG path do ícone

**Funções de Cálculo:**
- `calculateTotal(items)` - Calcula total do pedido

**Funções de Validação:**
- `isValidCEP(cep)` - Valida formato de CEP
- `validateOrderItem(item)` - Valida item do pedido
- `validateAddress(address)` - Valida endereço completo

**Constantes:**
- `BRAZILIAN_STATES` - Array de estados brasileiros
- `ORDER_STATUSES` - Array de status disponíveis

## Fluxo de Navegação

```
/orders (ou /)
  ├─ Lista de pedidos
  ├─ Filtrar por status
  ├─ Paginação
  └─ Clicar em "Ver Detalhes" → /orders/:id

/orders/new
  ├─ Adicionar itens
  ├─ Preencher endereço
  ├─ Finalizar pedido
  └─ Redireciona para /orders/:id

/orders/:id
  ├─ Ver detalhes completos
  ├─ Acompanhar status (timeline)
  └─ Cancelar pedido (se aplicável)
```

## Integração com API

Todos os componentes utilizam o `orderService` que se comunica com o backend:

```javascript
import orderService from '../../services/orderService';

// Criar pedido
await orderService.createOrder(orderData);

// Listar pedidos
await orderService.getOrders({ page: 1, limit: 9, status: 'pending' });

// Buscar pedido por ID
await orderService.getOrderById(id);

// Atualizar status
await orderService.updateOrderStatus(id, 'cancelled');
```

## Estados de UI

### Loading States
- **OrderList**: Skeleton cards animados (6 placeholders)
- **OrderForm**: Spinner no botão durante submit
- **OrderDetails**: Spinner centralizado
- **CEP**: Spinner no input durante busca

### Empty States
- **OrderList**: Ícone de sacola + mensagem + botão "Fazer Primeiro Pedido"
- **OrderForm (items)**: Ícone + mensagem "Nenhum item adicionado"

### Error States
- Card vermelho com ícone de alerta
- Mensagem de erro descritiva
- Botão "Tentar Novamente"
- Botão secundário para voltar

## Design System

### Cores por Status
- **pending**: bg-yellow-100, text-yellow-800, border-yellow-300
- **confirmed**: bg-blue-100, text-blue-800, border-blue-300
- **preparing**: bg-purple-100, text-purple-800, border-purple-300
- **shipped**: bg-indigo-100, text-indigo-800, border-indigo-300
- **delivered**: bg-green-100, text-green-800, border-green-300
- **cancelled**: bg-red-100, text-red-800, border-red-300

### Componentes de Layout Utilizados
- **Card**: Para containers de conteúdo
- **LoadingSpinner**: Para estados de carregamento
- **Layout**: Wrapper com Header e Footer

### Responsividade
- **Mobile** (< 768px): 
  - Grid 1 coluna
  - Botão flutuante "Novo Pedido"
  - Formulário empilhado
- **Tablet** (768px - 1024px):
  - Grid 2 colunas
  - Formulário em 2 colunas
- **Desktop** (> 1024px):
  - Grid 3 colunas
  - Timeline lateral
  - Formulário otimizado

## Validações

### Item do Pedido
```javascript
{
  product_name: 'Obrigatório',
  quantity: 'Maior que 0',
  price: 'Maior que 0'
}
```

### Endereço
```javascript
{
  street: 'Obrigatório',
  number: 'Obrigatório',
  city: 'Obrigatório',
  state: 'Obrigatório',
  zip_code: 'Obrigatório e formato válido (00000-000)',
  complement: 'Opcional'
}
```

### Pedido
- Mínimo de 1 item adicionado
- Todos os campos de endereço válidos

## APIs Externas

### ViaCEP
Integração para busca automática de endereço por CEP:

```javascript
fetch(`https://viacep.com.br/ws/${cep}/json/`)
```

Preenche automaticamente:
- Rua (logradouro)
- Cidade (localidade)
- Estado (uf)
- Complemento (se disponível)

## Animações

### CSS Transitions
- Hover em cards: transform + shadow
- Botões: colors
- Status badge: smooth transitions

### Animações Específicas
- Timeline: `animate-pulse` no status atual
- Skeleton: `animate-pulse` durante loading
- Spinner: `animate-spin` rotação contínua

## Notificações (Toast)

Usando `react-hot-toast`:
- ✅ **Sucesso**: Pedido criado, item adicionado, pedido cancelado
- ❌ **Erro**: Validação, API errors, busca de CEP
- 📍 **Info**: Endereço encontrado

## Exemplo de Uso Completo

```jsx
import { Routes, Route } from 'react-router-dom';
import { Layout, ProtectedRoute } from './components/Layout';
import { OrderList, OrderForm, OrderDetails } from './components/Orders';

function App() {
  return (
    <Routes>
      {/* Lista de pedidos */}
      <Route
        path="/orders"
        element={
          <ProtectedRoute>
            <Layout>
              <OrderList />
            </Layout>
          </ProtectedRoute>
        }
      />
      
      {/* Novo pedido */}
      <Route
        path="/orders/new"
        element={
          <ProtectedRoute>
            <Layout>
              <OrderForm />
            </Layout>
          </ProtectedRoute>
        }
      />
      
      {/* Detalhes do pedido */}
      <Route
        path="/orders/:id"
        element={
          <ProtectedRoute>
            <Layout>
              <OrderDetails />
            </Layout>
          </ProtectedRoute>
        }
      />
    </Routes>
  );
}
```

## Melhorias Futuras

- [ ] Busca de pedidos por ID ou data
- [ ] Exportar pedidos para PDF
- [ ] Filtro por range de datas
- [ ] Ordenação personalizada
- [ ] Pedidos favoritos/recorrentes
- [ ] Rastreamento em tempo real
- [ ] Notificações push de status
- [ ] Chat com suporte
- [ ] Avaliação de pedidos
- [ ] Histórico de pedidos com gráficos

## Arquivos Criados

```
src/
├── components/
│   └── Orders/
│       ├── OrderList.jsx          ✅ Lista de pedidos
│       ├── OrderForm.jsx          ✅ Formulário de criação
│       ├── OrderDetails.jsx       ✅ Detalhes do pedido
│       ├── StatusBadge.jsx        ✅ Badge de status
│       ├── OrderTimeline.jsx      ✅ Timeline de progresso
│       ├── index.js               ✅ Exportações
│       └── README.md              ✅ Esta documentação
├── utils/
│   └── orderHelpers.js            ✅ Funções auxiliares
└── services/
    └── orderService.js            ✅ (Atualizado) API calls
```

## Dependências

- `react-router-dom` - Navegação
- `react-hot-toast` - Notificações
- Componentes de Layout (Card, LoadingSpinner, etc.)
- AuthContext para autenticação
- Tailwind CSS para estilização

---

Criado com ❤️ para o projeto Pedidos Online
