import { Routes, Route, Navigate } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import { AuthProvider } from './context/AuthContext';
import Login from './components/Auth/Login';
import Register from './components/Auth/Register';
import { useAuth } from './hooks/useAuth';

// Componente temporário para Home (será substituído depois)
function Home() {
  const { user, logout } = useAuth();

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="container-center py-8">
        <div className="card text-center">
          <h1 className="text-4xl font-bold text-primary-600 mb-4">
            🛒 Pedidos Online
          </h1>
          <p className="text-gray-600 mb-6">
            Bem-vindo, {user?.name || 'Usuário'}!
          </p>
          <div className="space-y-4">
            <div className="p-4 bg-green-50 border border-green-200 rounded-lg">
              <p className="text-green-800 font-medium">
                ✅ Você está autenticado!
              </p>
            </div>
            <div className="text-sm text-gray-600">
              <p>Email: {user?.email}</p>
              {user?.phone && <p>Telefone: {user?.phone}</p>}
            </div>
            <button
              onClick={logout}
              className="btn-danger"
            >
              Sair
            </button>
          </div>
          <div className="mt-6 text-sm text-gray-500">
            <p>Próximo passo: Criar páginas de pedidos</p>
          </div>
        </div>
      </div>
    </div>
  );
}

// Componente de rota protegida
function ProtectedRoute({ children }) {
  const { isAuthenticated, loading } = useAuth();

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Carregando...</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return children;
}

function App() {
  return (
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
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <Home />
            </ProtectedRoute>
          }
        />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </AuthProvider>
  );
}

export default App;
