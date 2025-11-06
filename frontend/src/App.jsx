import { Routes, Route, Navigate } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import { AuthProvider } from './context/AuthContext';
import Login from './components/Auth/Login';
import Register from './components/Auth/Register';
import { useAuth } from './hooks/useAuth';
import { Layout, ProtectedRoute, ErrorBoundary, Card } from './components/Layout';

// Componente temporário para Home (será substituído depois)
function Home() {
  const { user } = useAuth();

  return (
    <div className="max-w-4xl mx-auto">
      <Card 
        variant="elevated"
        padding="large"
        className="text-center"
      >
        <h1 className="text-4xl font-bold text-blue-600 mb-4">
          🛒 Bem-vindo ao Pedidos Online
        </h1>
        <p className="text-gray-600 text-lg mb-6">
          Olá, {user?.name || 'Usuário'}!
        </p>
        
        <div className="space-y-4">
          <Card variant="success" padding="normal">
            <p className="text-green-800 font-medium">
              ✅ Você está autenticado e pronto para fazer pedidos!
            </p>
          </Card>
          
          <Card variant="outlined" padding="normal">
            <h2 className="font-semibold text-gray-800 mb-3 text-left">
              Seus dados:
            </h2>
            <div className="space-y-2 text-gray-600 text-left">
              <p><strong>Nome:</strong> {user?.name}</p>
              <p><strong>Email:</strong> {user?.email}</p>
              {user?.phone && <p><strong>Telefone:</strong> {user?.phone}</p>}
            </div>
          </Card>
          
          <Card variant="primary" padding="normal">
            <p className="font-medium text-blue-800 mb-2">📋 Próximos passos:</p>
            <ul className="text-left space-y-1 text-gray-700">
              <li>• Use o menu para navegar entre as páginas</li>
              <li>• Crie um novo pedido clicando em "Novo Pedido"</li>
              <li>• Visualize seus pedidos em "Meus Pedidos"</li>
              <li>• Atualize seu perfil em "Perfil"</li>
            </ul>
          </Card>
        </div>
      </Card>
    </div>
  );
}

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

        {/* Rotas */}
        <Routes>
          {/* Rotas públicas (sem layout) */}
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          
          {/* Rotas protegidas (com layout) */}
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <Layout>
                  <Home />
                </Layout>
              </ProtectedRoute>
            }
          />
          
          {/* Rota placeholder para profile */}
          <Route
            path="/profile"
            element={
              <ProtectedRoute>
                <Layout>
                  <div className="max-w-2xl mx-auto">
                    <Card variant="elevated" padding="large" className="text-center">
                      <div className="mb-6">
                        <div className="w-24 h-24 mx-auto bg-gradient-to-r from-blue-500 to-indigo-500 rounded-full flex items-center justify-center text-white text-3xl font-bold mb-4">
                          👤
                        </div>
                        <h1 className="text-3xl font-bold text-gray-800 mb-2">
                          Página de Perfil
                        </h1>
                        <p className="text-gray-600">
                          Em construção... Esta página será implementada em breve.
                        </p>
                      </div>
                      <Card variant="primary" padding="normal">
                        <p className="text-sm text-gray-700">
                          💡 <strong>Funcionalidades planejadas:</strong> Edição de dados pessoais, alteração de senha, gerenciamento de endereços e muito mais.
                        </p>
                      </Card>
                    </Card>
                  </div>
                </Layout>
              </ProtectedRoute>
            }
          />
          
          {/* Rota placeholder para novo pedido */}
          <Route
            path="/orders/new"
            element={
              <ProtectedRoute>
                <Layout>
                  <div className="max-w-2xl mx-auto">
                    <Card variant="elevated" padding="large" className="text-center">
                      <div className="mb-6">
                        <div className="w-24 h-24 mx-auto bg-gradient-to-r from-green-500 to-emerald-500 rounded-full flex items-center justify-center text-white text-3xl font-bold mb-4">
                          ➕
                        </div>
                        <h1 className="text-3xl font-bold text-gray-800 mb-2">
                          Novo Pedido
                        </h1>
                        <p className="text-gray-600">
                          Em construção... Esta página será implementada em breve.
                        </p>
                      </div>
                      <Card variant="primary" padding="normal">
                        <p className="text-sm text-gray-700">
                          💡 <strong>Funcionalidades planejadas:</strong> Seleção de produtos, carrinho de compras, endereço de entrega e finalização do pedido.
                        </p>
                      </Card>
                    </Card>
                  </div>
                </Layout>
              </ProtectedRoute>
            }
          />
          
          {/* Redirect para home */}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AuthProvider>
    </ErrorBoundary>
  );
}

export default App;
