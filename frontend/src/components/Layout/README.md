# Componentes de Layout

Este diretório contém os componentes de layout reutilizáveis da aplicação.

## Componentes

### 1. Header.jsx
Navbar responsivo com menu desktop e mobile.

**Características:**
- Logo/Título do app
- Menu desktop com links (Meus Pedidos, Novo Pedido)
- Dropdown do usuário com avatar (iniciais do nome)
- Menu mobile (hamburger) com drawer/sidebar
- Só é exibido quando o usuário está autenticado
- Design responsivo com Tailwind CSS

**Uso:**
```jsx
import Header from './components/Layout/Header';

<Header />
```

### 2. Footer.jsx
Footer fixo na parte inferior da página.

**Características:**
- Copyright e informações da empresa
- Links rápidos (Meus Pedidos, Novo Pedido, Perfil)
- Links de contato e redes sociais
- Background escuro com texto claro
- Sticky bottom quando conteúdo é pequeno

**Uso:**
```jsx
import Footer from './components/Layout/Footer';

<Footer />
```

### 3. Layout.jsx
Componente wrapper que combina Header + children + Footer.

**Características:**
- Min-height 100vh (página completa)
- Background com gradiente sutil
- Padding apropriado para o conteúdo
- Flex column layout (header, main, footer)

**Uso:**
```jsx
import Layout from './components/Layout/Layout';

<Layout>
  <YourPageContent />
</Layout>
```

### 4. ProtectedRoute.jsx
Componente que protege rotas privadas verificando autenticação.

**Características:**
- Verifica se usuário está autenticado
- Redireciona para /login se não autenticado
- Mostra loading durante verificação
- Usa Navigate do react-router-dom

**Uso:**
```jsx
import ProtectedRoute from './components/Layout/ProtectedRoute';

<Route
  path="/protected"
  element={
    <ProtectedRoute>
      <ProtectedPage />
    </ProtectedRoute>
  }
/>
```

### 5. LoadingSpinner.jsx
Spinner de carregamento reutilizável.

**Características:**
- Animação de rotação com círculos
- 3 tamanhos disponíveis (small, medium, large)
- Texto de carregamento opcional
- Centralizado por padrão

**Props:**
- `text` (string, opcional): Texto a ser exibido abaixo do spinner. Padrão: "Carregando..."
- `size` (string, opcional): Tamanho do spinner ("small", "medium", "large"). Padrão: "medium"

**Uso:**
```jsx
import LoadingSpinner from './components/Layout/LoadingSpinner';

// Com texto padrão
<LoadingSpinner />

// Com texto customizado
<LoadingSpinner text="Salvando..." />

// Com tamanho personalizado
<LoadingSpinner size="large" text="Processando pedido..." />

// Sem texto
<LoadingSpinner text="" />
```

### 6. ErrorBoundary.jsx
Componente de classe que captura erros de componentes filhos.

**Características:**
- Captura erros de renderização dos componentes filhos
- Exibe UI amigável de erro
- Botão para recarregar página
- Botão para voltar à página inicial
- Mostra detalhes do erro em desenvolvimento
- Logging de erros no console

**Uso:**
```jsx
import ErrorBoundary from './components/Layout/ErrorBoundary';

<ErrorBoundary>
  <App />
</ErrorBoundary>
```

## Importação Facilitada

Você pode importar todos os componentes de uma vez usando o arquivo de índice:

```jsx
import { 
  Layout, 
  Header, 
  Footer, 
  ProtectedRoute, 
  LoadingSpinner, 
  ErrorBoundary 
} from './components/Layout';
```

## Exemplo de Uso Completo

```jsx
import { Routes, Route } from 'react-router-dom';
import { Layout, ProtectedRoute, ErrorBoundary } from './components/Layout';

function App() {
  return (
    <ErrorBoundary>
      <Routes>
        {/* Rota pública */}
        <Route path="/login" element={<Login />} />
        
        {/* Rota protegida com layout */}
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <Layout>
                <HomePage />
              </Layout>
            </ProtectedRoute>
          }
        />
      </Routes>
    </ErrorBoundary>
  );
}
```

## Design System

Todos os componentes utilizam:
- **Framework CSS**: Tailwind CSS
- **Paleta de cores**: Azul/Índigo como cor primária
- **Responsividade**: Mobile-first approach
- **Animações**: Transições suaves com CSS
- **Ícones**: SVG inline (sem dependências externas)

## Tailwind Classes Principais

- `bg-gradient-to-br from-gray-50 via-blue-50 to-indigo-50`: Background gradiente do Layout
- `bg-gradient-to-r from-blue-600 to-indigo-600`: Gradiente azul/índigo para elementos de destaque
- `shadow-lg`, `shadow-md`: Sombras para profundidade
- `hover:*`: Estados hover com transições suaves
- `transition-colors`, `transition-transform`: Animações

## Responsividade

- **Mobile**: < 768px - Menu hamburger, layout empilhado
- **Tablet**: 768px - 1024px - Layout adaptado
- **Desktop**: > 1024px - Menu horizontal completo

## Acessibilidade

- Uso de tags semânticas HTML5
- Atributos `aria-label` em links de ícones
- Navegação via teclado
- Cores com contraste adequado

## Próximos Passos

Para completar o sistema de layout, você pode adicionar:

1. **Breadcrumbs**: Navegação hierárquica
2. **Sidebar**: Menu lateral para navegação secundária
3. **Modal**: Componente de modal reutilizável
4. **Toast**: Sistema de notificações (já implementado via react-hot-toast)
5. **Skeleton Loading**: Placeholders de carregamento
6. **Empty State**: Componente para estados vazios
