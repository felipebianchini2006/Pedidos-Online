import api from './api';

/**
 * Serviço de gerenciamento de pedidos
 */
const orderService = {
  /**
   * Criar novo pedido
   * @param {Object} orderData - Dados do pedido (items, address, total_amount)
   * @returns {Promise<{success: boolean, data?: object, error?: string}>}
   */
  async createOrder(orderData) {
    try {
      const response = await api.post('/api/orders/v1/orders', orderData);

      return {
        success: true,
        data: response.data.data || response.data,
      };
    } catch (error) {
      return {
        success: false,
        error: error.response?.data?.error || 'Erro ao criar pedido',
      };
    }
  },

  /**
   * Buscar lista de pedidos do usuário
   * @param {Object} params - Parâmetros de busca (page, limit, status)
   * @returns {Promise<{success: boolean, data?: object, error?: string}>}
   */
  async getOrders(params = {}) {
    try {
      const response = await api.get('/api/orders/v1/orders', { params });

      return {
        success: true,
        data: response.data.data || response.data,
      };
    } catch (error) {
      return {
        success: false,
        error: error.response?.data?.error || 'Erro ao buscar pedidos',
      };
    }
  },

  /**
   * Buscar detalhes de um pedido específico
   * @param {string} id - ID do pedido
   * @returns {Promise<{success: boolean, data?: object, error?: string}>}
   */
  async getOrderById(id) {
    try {
      const response = await api.get(`/api/orders/v1/orders/${id}`);

      return {
        success: true,
        data: response.data.data || response.data,
      };
    } catch (error) {
      return {
        success: false,
        error: error.response?.data?.error || 'Erro ao buscar pedido',
      };
    }
  },

  /**
   * Buscar detalhes de um pedido específico (alias)
   * @param {string} id - ID do pedido
   * @returns {Promise<{success: boolean, data?: object, error?: string}>}
   */
  async getOrder(id) {
    return this.getOrderById(id);
  },

  /**
   * Cancelar um pedido
   * @param {string} id - ID do pedido
   * @returns {Promise<{success: boolean, data?: object, error?: string}>}
   */
  async cancelOrder(id) {
    try {
      const response = await api.put(`/api/orders/v1/orders/${id}/status`, {
        status: 'cancelled',
      });

      return {
        success: true,
        data: response.data.data || response.data,
      };
    } catch (error) {
      return {
        success: false,
        error: error.response?.data?.error || 'Erro ao cancelar pedido',
      };
    }
  },

  /**
   * Atualizar status de um pedido (admin/sistema)
   * @param {string} id - ID do pedido
   * @param {string} status - Novo status do pedido
   * @returns {Promise<{success: boolean, data?: object, error?: string}>}
   */
  async updateOrderStatus(id, status) {
    try {
      const response = await api.put(`/api/orders/v1/orders/${id}/status`, {
        status,
      });

      return {
        success: true,
        data: response.data.data || response.data,
      };
    } catch (error) {
      return {
        success: false,
        error: error.response?.data?.error || 'Erro ao atualizar status do pedido',
      };
    }
  },
};

export default orderService;
