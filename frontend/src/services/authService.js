import api from './api';

/**
 * Serviço de autenticação de usuários
 */
const authService = {
  /**
   * Registrar novo usuário
   * @param {string} email - Email do usuário
   * @param {string} password - Senha do usuário
   * @param {string} name - Nome completo do usuário
   * @param {string} phone - Telefone do usuário (opcional)
   * @returns {Promise<{success: boolean, data?: object, error?: string}>}
   */
  async register(email, password, name, phone) {
    try {
      const response = await api.post('/api/users/v1/register', {
        email,
        password,
        name,
        phone,
      });

      return {
        success: true,
        data: response.data,
      };
    } catch (error) {
      return {
        success: false,
        error: error.response?.data?.error || 'Erro ao registrar usuário',
      };
    }
  },

  /**
   * Fazer login do usuário
   * @param {string} email - Email do usuário
   * @param {string} password - Senha do usuário
   * @returns {Promise<{success: boolean, data?: object, error?: string}>}
   */
  async login(email, password) {
    try {
      const response = await api.post('/api/users/v1/login', {
        email,
        password,
      });

      // Extrair token e dados do usuário
      const { token, user } = response.data.data || response.data;

      // Salvar token no localStorage
      if (token) {
        this.setToken(token);
      }

      // Salvar dados do usuário
      if (user) {
        localStorage.setItem('user', JSON.stringify(user));
      }

      return {
        success: true,
        data: { token, user },
      };
    } catch (error) {
      return {
        success: false,
        error: error.response?.data?.error || 'Erro ao fazer login',
      };
    }
  },

  /**
   * Fazer logout do usuário
   * Limpa token e dados do localStorage e redireciona para login
   */
  logout() {
    // Limpar localStorage
    localStorage.removeItem('token');
    localStorage.removeItem('user');

    // Redirecionar para login
    window.location.href = '/login';
  },

  /**
   * Obter perfil do usuário autenticado
   * @returns {Promise<{success: boolean, data?: object, error?: string}>}
   */
  async getProfile() {
    try {
      const response = await api.get('/api/users/v1/profile');

      // Atualizar dados do usuário no localStorage
      if (response.data.data) {
        localStorage.setItem('user', JSON.stringify(response.data.data));
      }

      return {
        success: true,
        data: response.data.data || response.data,
      };
    } catch (error) {
      return {
        success: false,
        error: error.response?.data?.error || 'Erro ao buscar perfil',
      };
    }
  },

  /**
   * Atualizar perfil do usuário autenticado
   * @param {string} name - Novo nome do usuário
   * @param {string} phone - Novo telefone do usuário
   * @returns {Promise<{success: boolean, data?: object, error?: string}>}
   */
  async updateProfile(name, phone) {
    try {
      const response = await api.put('/api/users/v1/profile', {
        name,
        phone,
      });

      // Atualizar dados do usuário no localStorage
      if (response.data.data) {
        localStorage.setItem('user', JSON.stringify(response.data.data));
      }

      return {
        success: true,
        data: response.data.data || response.data,
      };
    } catch (error) {
      return {
        success: false,
        error: error.response?.data?.error || 'Erro ao atualizar perfil',
      };
    }
  },

  /**
   * Obter token JWT do localStorage
   * @returns {string|null} Token JWT ou null se não existir
   */
  getToken() {
    return localStorage.getItem('token');
  },

  /**
   * Salvar token JWT no localStorage
   * @param {string} token - Token JWT
   */
  setToken(token) {
    localStorage.setItem('token', token);
  },

  /**
   * Verificar se usuário está autenticado
   * @returns {boolean} True se token existe, false caso contrário
   */
  isAuthenticated() {
    const token = this.getToken();
    return token !== null && token !== undefined && token !== '';
  },

  /**
   * Obter dados do usuário do localStorage
   * @returns {object|null} Dados do usuário ou null
   */
  getCurrentUser() {
    const userStr = localStorage.getItem('user');
    if (!userStr) return null;

    try {
      return JSON.parse(userStr);
    } catch (error) {
      console.error('Erro ao parsear dados do usuário:', error);
      return null;
    }
  },
};

export default authService;
