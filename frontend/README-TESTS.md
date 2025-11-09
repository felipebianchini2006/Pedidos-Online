# Testes Frontend - Pedidos Online

## Visão Geral

Este documento descreve a suite de testes implementada para o frontend React do sistema de Pedidos Online.

## Tecnologias Utilizadas

- **Vitest**: Framework de testes rápido e moderno para Vite
- **React Testing Library**: Para testes de componentes React
- **@testing-library/user-event**: Simulação de interações do usuário
- **MSW (Mock Service Worker)**: Mock de APIs HTTP
- **jsdom**: Ambiente DOM para testes

## Estrutura de Testes

### 1. Setup e Configuração

**Arquivo**: [`src/test/setup.js`](src/test/setup.js)

Configuração global dos testes:
- Extensão de matchers do jest-dom
- Cleanup automático após cada teste
- Configuração do MSW server
- Mocks globais do localStorage e window.location

### 2. Mock Service Worker (MSW)

**Arquivos**:
- [`src/mocks/handlers.js`](src/mocks/handlers.js)
- [`src/mocks/server.js`](src/mocks/server.js)

Handlers para simular as APIs:
- **User Service**: register, login, profile, updateProfile
- **Order Service**: getOrders, getOrderById, createOrder, updateOrderStatus

### 3. Testes de Utilitários

#### helpers.test.js
**Arquivo**: [`src/utils/helpers.test.js`](src/utils/helpers.test.js)

**Funções testadas**:
- `formatCurrency()` - Formatação de valores monetários
- `formatDate()` - Formatação de datas
- `validateEmail()` - Validação de emails
- `validatePhone()` - Validação de telefones
- `formatPhone()` - Formatação de telefones
- `validateCPF()` - Validação de CPF
- `formatCPF()` - Formatação de CPF
- `validateCEP()` - Validação de CEP
- `formatCEP()` - Formatação de CEP
- `getStatusColor()` - Classes CSS por status
- `getStatusLabel()` - Labels em português
- `truncateText()` - Truncamento de texto
- `getInitials()` - Iniciais de nomes
- `calculateOrderTotal()` - Cálculo de totais

**Total**: 41 testes

### 4. Testes de Serviços

#### authService.test.js
**Arquivo**: [`src/services/authService.test.js`](src/services/authService.test.js)

**Funcionalidades testadas**:
- **register**: Cadastro de novos usuários
- **login**: Autenticação e salvamento de token
- **logout**: Limpeza de dados e redirecionamento
- **getProfile**: Buscar perfil do usuário
- **updateProfile**: Atualizar dados do perfil
- **getToken/setToken**: Gerenciamento de tokens
- **isAuthenticated**: Verificação de autenticação
- **getCurrentUser**: Obtenção de dados do usuário

**Total**: 28 testes

### 5. Testes de Contexto

#### AuthContext.test.jsx
**Arquivo**: [`src/context/AuthContext.test.jsx`](src/context/AuthContext.test.jsx)

**Funcionalidades testadas**:
- Inicialização do contexto
- Carregamento automático do usuário
- Login/Logout
- Registro de usuários
- Atualização de perfil
- Refresh de dados
- Tratamento de erros
- Validação do hook useAuth

**Total**: 14 testes

### 6. Testes de Componentes

#### Login.test.jsx
**Arquivo**: [`src/components/Auth/Login.test.jsx`](src/components/Auth/Login.test.jsx)

**Cenários testados**:
- Renderização do formulário
- Validação de campos (email/senha)
- Estados de loading
- Submit do formulário
- Redirecionamento após login
- Tratamento de erros
- Acessibilidade (autocomplete, placeholders)

**Total**: 16 testes

#### Register.test.jsx
**Arquivo**: [`src/components/Auth/Register.test.jsx`](src/components/Auth/Register.test.jsx)

**Cenários testados**:
- Renderização de todos os campos
- Validação completa do formulário
- Máscara de telefone
- Indicador de força da senha
- Submit e redirecionamento
- Botão desabilitado quando inválido
- Limpeza de erros ao digitar

**Total**: 10+ testes

#### OrderList.test.jsx
**Arquivo**: [`src/components/Orders/OrderList.test.jsx`](src/components/Orders/OrderList.test.jsx)

**Cenários testados**:
- Loading state com skeleton
- Renderização de pedidos
- Empty state
- Error state e retry
- Filtro por status
- Paginação (anterior/próximo)
- Links para detalhes
- Chamadas da API

**Total**: 20+ testes

#### OrderForm.test.jsx
**Arquivo**: [`src/components/Orders/OrderForm.test.jsx`](src/components/Orders/OrderForm.test.jsx)

**Cenários testados**:
- Renderização do formulário
- Adicionar/remover itens
- Validação de itens
- Cálculo de total
- Validação de endereço
- Submit do formulário
- Redirecionamento
- Navegação (voltar)

**Total**: 15+ testes

## Comandos de Teste

```bash
# Executar testes em modo watch
npm test

# Executar todos os testes uma vez
npm run test:run

# Executar testes com UI
npm run test:ui

# Gerar relatório de cobertura
npm run test:coverage
```

## Cobertura de Código

O projeto está configurado com threshold mínimo de **70%** de cobertura:
- Lines: 70%
- Functions: 70%
- Branches: 70%
- Statements: 70%

## Estrutura de Arquivos de Teste

```
frontend/
├── src/
│   ├── test/
│   │   └── setup.js                 # Setup global dos testes
│   ├── mocks/
│   │   ├── handlers.js              # Handlers MSW
│   │   └── server.js                # Configuração MSW
│   ├── utils/
│   │   └── helpers.test.js          # Testes de utilitários
│   ├── services/
│   │   └── authService.test.js      # Testes de serviços
│   ├── context/
│   │   └── AuthContext.test.jsx     # Testes de contexto
│   └── components/
│       ├── Auth/
│       │   ├── Login.test.jsx       # Testes de Login
│       │   └── Register.test.jsx    # Testes de Register
│       └── Orders/
│           ├── OrderList.test.jsx   # Testes de OrderList
│           └── OrderForm.test.jsx   # Testes de OrderForm
└── vitest.config.js                 # Configuração do Vitest
```

## Boas Práticas Implementadas

1. **Isolamento**: Cada teste é independente e não afeta outros
2. **Cleanup**: Limpeza automática após cada teste
3. **Mocks**: Uso de MSW para mocks realistas de API
4. **User-Centric**: Testes focados no comportamento do usuário
5. **Acessibilidade**: Uso de queries por role e label
6. **Cobertura**: Threshold mínimo de 70%
7. **Performance**: Vitest para execução rápida

## Padrões de Teste

### Estrutura AAA (Arrange, Act, Assert)

```javascript
it('deve fazer algo', async () => {
  // Arrange - Preparar
  const user = userEvent.setup();
  render(<Component />, { wrapper: Wrapper });

  // Act - Agir
  const button = screen.getByRole('button');
  await user.click(button);

  // Assert - Verificar
  expect(screen.getByText(/sucesso/i)).toBeInTheDocument();
});
```

### Queries Preferidas

1. **getByRole**: Para elementos interativos (buttons, links)
2. **getByLabelText**: Para inputs de formulário
3. **getByText**: Para conteúdo de texto
4. **getByPlaceholderText**: Quando não há label

### Async/Await

Sempre usar `waitFor` para operações assíncronas:

```javascript
await waitFor(() => {
  expect(screen.getByText(/loaded/i)).toBeInTheDocument();
});
```

## Solução de Problemas

### Erro: "Not wrapped in act(...)"

Usar `waitFor` ou `act` para operações assíncronas:

```javascript
await waitFor(() => {
  // assertions
});
```

### MSW não está interceptando

Verificar que o setup está correto em `src/test/setup.js` e que os handlers estão definidos corretamente.

### Testes lentos

- Usar `userEvent.setup()` apenas uma vez por teste
- Evitar múltiplos `waitFor` aninhados
- Considerar usar `findBy*` ao invés de `getBy* + waitFor`

## Contribuindo

Ao adicionar novos testes:

1. Seguir a estrutura AAA
2. Usar queries acessíveis (role, label)
3. Testar comportamento, não implementação
4. Manter cobertura > 70%
5. Adicionar casos de erro e edge cases

## Referências

- [Vitest Docs](https://vitest.dev/)
- [React Testing Library](https://testing-library.com/react)
- [MSW Docs](https://mswjs.io/)
- [Testing Best Practices](https://kentcdodds.com/blog/common-mistakes-with-react-testing-library)
