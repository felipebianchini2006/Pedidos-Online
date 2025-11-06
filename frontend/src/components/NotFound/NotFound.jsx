import { Link } from 'react-router-dom';
import { Card } from '../Layout';

/**
 * NotFound Component
 * Página 404 amigável para rotas não encontradas
 */
const NotFound = () => {
  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-blue-50 to-indigo-50 flex items-center justify-center p-4">
      <Card variant="elevated" padding="large" className="max-w-lg w-full text-center">
        {/* Ilustração 404 */}
        <div className="mb-8">
          <div className="relative inline-block">
            <div className="text-9xl font-bold text-blue-600 opacity-20">404</div>
            <div className="absolute inset-0 flex items-center justify-center">
              <svg 
                className="w-32 h-32 text-blue-600 animate-bounce" 
                fill="none" 
                stroke="currentColor" 
                viewBox="0 0 24 24"
              >
                <path 
                  strokeLinecap="round" 
                  strokeLinejoin="round" 
                  strokeWidth={2} 
                  d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" 
                />
              </svg>
            </div>
          </div>
        </div>

        {/* Mensagem */}
        <h1 className="text-3xl font-bold text-gray-800 mb-4">
          Oops! Página Não Encontrada
        </h1>
        <p className="text-gray-600 mb-8">
          A página que você está procurando não existe ou foi movida para outro lugar.
          Que tal voltar para a página inicial?
        </p>

        {/* Botões de Ação */}
        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <Link
            to="/"
            className="inline-flex items-center justify-center space-x-2 bg-blue-600 text-white px-8 py-3 rounded-lg hover:bg-blue-700 transition-colors font-medium"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
            </svg>
            <span>Voltar para Home</span>
          </Link>
          
          <Link
            to="/orders/new"
            className="inline-flex items-center justify-center space-x-2 bg-gray-200 text-gray-700 px-8 py-3 rounded-lg hover:bg-gray-300 transition-colors font-medium"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            <span>Fazer um Pedido</span>
          </Link>
        </div>

        {/* Links Úteis */}
        <div className="mt-8 pt-6 border-t border-gray-200">
          <p className="text-sm text-gray-600 mb-3">Links úteis:</p>
          <div className="flex flex-wrap justify-center gap-4 text-sm">
            <Link to="/orders" className="text-blue-600 hover:text-blue-700 hover:underline">
              Meus Pedidos
            </Link>
            <Link to="/profile" className="text-blue-600 hover:text-blue-700 hover:underline">
              Meu Perfil
            </Link>
            <a href="mailto:contato@pedidosonline.com" className="text-blue-600 hover:text-blue-700 hover:underline">
              Suporte
            </a>
          </div>
        </div>
      </Card>
    </div>
  );
};

export default NotFound;
