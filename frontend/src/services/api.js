import axios from 'axios';
import toast from 'react-hot-toast';

// Criar instância do axios com configurações base
const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8000',
  timeout: 30000, // 30 segundos
  headers: {
    'Content-Type': 'application/json',
  },
});

// Interceptor de Request - Adicionar token JWT automaticamente
api.interceptors.request.use(
  (config) => {
    // Obter token do localStorage
    const token = localStorage.getItem('token');
    
    // Se token existe, adicionar ao header Authorization
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Interceptor de Response - Tratamento de erros
api.interceptors.response.use(
  (response) => {
    // Resposta bem-sucedida, retornar normalmente
    return response;
  },
  (error) => {
    // Extrair informações do erro
    const status = error.response?.status;
    const errorMessage = error.response?.data?.error || 
                        error.response?.data?.message || 
                        'Erro ao processar requisição';

    // Tratamento específico por código de status
    switch (status) {
      case 401:
        // Não autorizado - Token inválido ou expirado
        toast.error('Sessão expirada. Faça login novamente.');
        
        // Limpar token do localStorage
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        
        // Redirecionar para login
        if (window.location.pathname !== '/login') {
          window.location.href = '/login';
        }
        break;

      case 403:
        // Permissão negada
        toast.error('Você não tem permissão para acessar este recurso.');
        break;

      case 404:
        // Recurso não encontrado
        toast.error('Recurso não encontrado.');
        break;

      case 429:
        // Rate limit excedido
        toast.error('Muitas requisições. Por favor, aguarde alguns instantes.');
        break;

      case 500:
      case 502:
      case 503:
      case 504:
        // Erro do servidor
        toast.error('Erro no servidor. Tente novamente mais tarde.');
        break;

      default:
        // Outros erros
        if (errorMessage) {
          toast.error(errorMessage);
        }
    }

    // Rejeitar promise com o erro
    return Promise.reject(error);
  }
);

export default api;
