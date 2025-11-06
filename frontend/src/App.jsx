import { Routes, Route } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import { AuthProvider } from './context/AuthContext';
import Login from './components/Auth/Login';
import Register from './components/Auth/Register';
import { Layout, ProtectedRoute, ErrorBoundary } from './components/Layout';
import { OrderList, OrderForm, OrderDetails } from './components/Orders';
import Profile from './components/Profile/Profile';
import NotFound from './components/NotFound/NotFound';

/**
 * App Component
 * Componente raiz da aplicação com configuração de rotas
 */
function App() {
  return (
    <ErrorBoundary>
      <AuthProvider>
        {/* Toast notifications */}
        <Toaster
          position="top-right"
          toastOptions={{
            duration: 4000,
            style: {
              background: '#363636',
              color: '#fff',
            },
            success: {
              duration: 3000,
              iconTheme: {
                primary: '#22c55e',
                secondary: '#fff',
              },
            },
            error: {
              duration: 4000,
              iconTheme: {
                primary: '#ef4444',
                secondary: '#fff',
              },
            },
          }}
        />

        {/* Rotas da Aplicação */}
        <Routes>
          {/* ==================== ROTAS PÚBLICAS ==================== */}
          {/* Sem layout - apenas o componente */}
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          
          {/* ==================== ROTAS PROTEGIDAS ==================== */}
          {/* Com layout (Header + Footer) e proteção de autenticação */}
          
          {/* Home - Lista de Pedidos */}
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <Layout>
                  <OrderList />
                </Layout>
              </ProtectedRoute>
            }
          />
          
          {/* Lista de Pedidos (rota alternativa) */}
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
          
          {/* Novo Pedido */}
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
          
          {/* Detalhes do Pedido */}
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
          
          {/* Perfil do Usuário */}
          <Route
            path="/profile"
            element={
              <ProtectedRoute>
                <Layout>
                  <Profile />
                </Layout>
              </ProtectedRoute>
            }
          />
          
          {/* Alterar Senha (placeholder) */}
          <Route
            path="/change-password"
            element={
              <ProtectedRoute>
                <Layout>
                  <div className="max-w-2xl mx-auto text-center py-12">
                    <h1 className="text-3xl font-bold text-gray-800 mb-4">
                      Alterar Senha
                    </h1>
                    <p className="text-gray-600">
                      Esta funcionalidade será implementada em breve.
                    </p>
                  </div>
                </Layout>
              </ProtectedRoute>
            }
          />
          
          {/* ==================== PÁGINA 404 ==================== */}
          {/* Rota catch-all para páginas não encontradas */}
          <Route path="*" element={<NotFound />} />
        </Routes>
      </AuthProvider>
    </ErrorBoundary>
  );
}

export default App;
