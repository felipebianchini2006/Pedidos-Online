import { Routes, Route } from 'react-router-dom';
import { lazy, Suspense } from 'react';
import { Toaster } from 'react-hot-toast';
import { AuthProvider } from './context/AuthContext';
import { Layout, ProtectedRoute, AdminRoute, ErrorBoundary, LoadingSpinner } from './components/Layout';

// Lazy load components for code splitting
const Login = lazy(() => import('./components/Auth/Login'));
const Register = lazy(() => import('./components/Auth/Register'));
const OrderList = lazy(() => import('./components/Orders/OrderList'));
const OrderForm = lazy(() => import('./components/Orders/OrderForm'));
const OrderDetails = lazy(() => import('./components/Orders/OrderDetails'));
const Profile = lazy(() => import('./components/Profile/Profile'));
const AdminOrders = lazy(() => import('./components/Admin/AdminOrders'));
const NotFound = lazy(() => import('./components/NotFound/NotFound'));

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

        {/* Rotas da Aplicação com Suspense para lazy loading */}
        <Suspense fallback={
          <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-50">
            <LoadingSpinner size="large" text="Carregando..." />
          </div>
        }>
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
          
          {/* Admin - Gerenciar Pedidos */}
          <Route
            path="/admin/orders"
            element={
              <AdminRoute>
                <Layout>
                  <AdminOrders />
                </Layout>
              </AdminRoute>
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
        </Suspense>
      </AuthProvider>
    </ErrorBoundary>
  );
}

export default App;
