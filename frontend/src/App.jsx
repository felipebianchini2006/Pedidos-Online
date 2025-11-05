import { Toaster } from 'react-hot-toast';
import { AuthProvider } from './context/AuthContext';

function App() {
  return (
    <AuthProvider>
      <div className="min-h-screen bg-gray-50">
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

        {/* Main content */}
        <div className="container-center py-8">
          <div className="card text-center">
            <h1 className="text-4xl font-bold text-primary-600 mb-4">
              🛒 Pedidos Online
            </h1>
            <p className="text-gray-600 mb-6">
              Sistema de pedidos online construído com React + Vite + Tailwind CSS
            </p>
            <div className="space-y-4">
              <div className="p-4 bg-green-50 border border-green-200 rounded-lg">
                <p className="text-green-800 font-medium">
                  ✅ Frontend configurado com sucesso!
                </p>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-left">
                <div className="p-4 bg-blue-50 border border-blue-200 rounded-lg">
                  <h3 className="font-semibold text-blue-900 mb-2">Serviços Criados:</h3>
                  <ul className="text-sm text-blue-800 space-y-1">
                    <li>• API Service (Axios)</li>
                    <li>• Auth Service</li>
                    <li>• Order Service</li>
                  </ul>
                </div>
                <div className="p-4 bg-purple-50 border border-purple-200 rounded-lg">
                  <h3 className="font-semibold text-purple-900 mb-2">Contextos:</h3>
                  <ul className="text-sm text-purple-800 space-y-1">
                    <li>• AuthContext (Provider)</li>
                    <li>• useAuth hook</li>
                  </ul>
                </div>
                <div className="p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
                  <h3 className="font-semibold text-yellow-900 mb-2">Utilitários:</h3>
                  <ul className="text-sm text-yellow-800 space-y-1">
                    <li>• Formatação de moeda</li>
                    <li>• Formatação de data</li>
                    <li>• Validações (email, telefone, CPF, CEP)</li>
                  </ul>
                </div>
                <div className="p-4 bg-pink-50 border border-pink-200 rounded-lg">
                  <h3 className="font-semibold text-pink-900 mb-2">Configurações:</h3>
                  <ul className="text-sm text-pink-800 space-y-1">
                    <li>• Tailwind CSS customizado</li>
                    <li>• Proxy Vite (/api)</li>
                    <li>• Variáveis de ambiente</li>
                  </ul>
                </div>
              </div>
            </div>
            <div className="mt-6 text-sm text-gray-500">
              <p>Próximo passo: Criar componentes de UI (páginas, formulários, etc.)</p>
            </div>
          </div>
        </div>
      </div>
    </AuthProvider>
  );
}

export default App;
