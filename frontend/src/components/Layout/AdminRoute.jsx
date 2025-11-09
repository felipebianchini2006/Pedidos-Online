import { Navigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';

/**
 * AdminRoute Component
 * Protege rotas que só devem ser acessadas por administradores
 * Verifica se o usuário está autenticado E se é admin
 */
const AdminRoute = ({ children }) => {
  const { user, isAuthenticated } = useAuth();

  // Se não está autenticado, redireciona para login
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  // Se está autenticado mas não é admin, redireciona para home com mensagem
  const isAdmin = user?.email === 'admin@gmail.com';
  
  if (!isAdmin) {
    // Poderia usar toast aqui para mostrar mensagem de erro
    alert('⚠️ Acesso negado! Apenas administradores podem acessar esta página.');
    return <Navigate to="/" replace />;
  }

  // Se é admin, renderiza o conteúdo
  return children;
};

export default AdminRoute;
