import { useContext } from 'react';
import AuthContext from '../context/AuthContext';

/**
 * Custom hook para acessar o contexto de autenticação
 * @returns {object} Contexto de autenticação
 * @throws {Error} Se usado fora do AuthProvider
 */
export function useAuth() {
  const context = useContext(AuthContext);
  
  if (!context) {
    throw new Error('useAuth deve ser usado dentro de um AuthProvider');
  }
  
  return context;
}

export default useAuth;
