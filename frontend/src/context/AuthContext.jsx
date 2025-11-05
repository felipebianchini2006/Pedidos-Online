import { createContext, useContext, useState, useEffect } from 'react';
import authService from '../services/authService';
import toast from 'react-hot-toast';

// Criar contexto de autenticação
const AuthContext = createContext(null);

/**
 * Provider do contexto de autenticação
 * Envolve toda a aplicação para fornecer estado de autenticação
 */
export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  /**
   * Carregar dados do usuário ao montar o componente
   * Verifica se existe token e busca perfil do usuário
   */
  useEffect(() => {
    const initAuth = async () => {
      try {
        // Verificar se existe token
        const token = authService.getToken();
        
        if (!token) {
          setLoading(false);
          return;
        }

        // Buscar perfil do usuário
        const result = await authService.getProfile();
        
        if (result.success) {
          setUser(result.data);
          setIsAuthenticated(true);
        } else {
          // Token inválido, limpar
          authService.logout();
        }
      } catch (error) {
        console.error('Erro ao inicializar autenticação:', error);
        authService.logout();
      } finally {
        setLoading(false);
      }
    };

    initAuth();
  }, []);

  /**
   * Fazer login do usuário
   */
  const login = async (email, password) => {
    try {
      const result = await authService.login(email, password);
      
      if (result.success) {
        setUser(result.data.user);
        setIsAuthenticated(true);
        toast.success('Login realizado com sucesso!');
        return { success: true };
      } else {
        toast.error(result.error);
        return { success: false, error: result.error };
      }
    } catch (error) {
      const errorMessage = 'Erro ao fazer login. Tente novamente.';
      toast.error(errorMessage);
      return { success: false, error: errorMessage };
    }
  };

  /**
   * Registrar novo usuário
   */
  const register = async (email, password, name, phone) => {
    try {
      const result = await authService.register(email, password, name, phone);
      
      if (result.success) {
        toast.success('Cadastro realizado com sucesso! Faça login para continuar.');
        return { success: true };
      } else {
        toast.error(result.error);
        return { success: false, error: result.error };
      }
    } catch (error) {
      const errorMessage = 'Erro ao registrar. Tente novamente.';
      toast.error(errorMessage);
      return { success: false, error: errorMessage };
    }
  };

  /**
   * Fazer logout do usuário
   */
  const logout = () => {
    setUser(null);
    setIsAuthenticated(false);
    authService.logout();
    toast.success('Logout realizado com sucesso!');
  };

  /**
   * Atualizar perfil do usuário
   */
  const updateProfile = async (name, phone) => {
    try {
      const result = await authService.updateProfile(name, phone);
      
      if (result.success) {
        setUser(result.data);
        toast.success('Perfil atualizado com sucesso!');
        return { success: true };
      } else {
        toast.error(result.error);
        return { success: false, error: result.error };
      }
    } catch (error) {
      const errorMessage = 'Erro ao atualizar perfil. Tente novamente.';
      toast.error(errorMessage);
      return { success: false, error: errorMessage };
    }
  };

  /**
   * Recarregar dados do usuário
   */
  const refreshUser = async () => {
    try {
      const result = await authService.getProfile();
      
      if (result.success) {
        setUser(result.data);
        return { success: true };
      } else {
        return { success: false, error: result.error };
      }
    } catch (error) {
      return { success: false, error: 'Erro ao atualizar dados do usuário' };
    }
  };

  // Valor do contexto
  const value = {
    user,
    loading,
    isAuthenticated,
    login,
    logout,
    register,
    updateProfile,
    refreshUser,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

/**
 * Hook customizado para usar o contexto de autenticação
 * @returns {object} Contexto de autenticação
 */
export function useAuth() {
  const context = useContext(AuthContext);
  
  if (!context) {
    throw new Error('useAuth deve ser usado dentro de um AuthProvider');
  }
  
  return context;
}

export default AuthContext;
