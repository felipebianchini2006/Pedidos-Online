import api from './api';

/**
 * Serviço de gerenciamento de pedidos
 */
const orderService = {
  /**
   * Criar novo pedido
   * @param {Array} items - Array de itens do pedido
   * @param {Object} address - Endereço de entrega
   * @param {number} totalAmount - Valor total do pedido
   * @returns {Promise<{success: boolean, data?: object, error?: string}>}
   */
  async createOrder(items, address, totalAmount) {
    try {
      const response = await api.post('/api/orders', {
        items,
        address,
        total_amount: totalAmount,
      });

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
   * @param {number} page - Número da página (padrão: 1)
   * @param {number} pageSize - Itens por página (padrão: 10)
   * @returns {Promise<{success: boolean, data?: object, error?: string}>}
   */
  async getOrders(page = 1, pageSize = 10) {
    try {
      const response = await api.get('/api/orders', {
        params: {
          page,
          page_size: pageSize,
        },
      });

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
  async getOrder(id) {
    try {
      const response = await api.get(`/api/orders/${id}`);

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
   * Cancelar um pedido
   * @param {string} id - ID do pedido
   * @returns {Promise<{success: boolean, data?: object, error?: string}>}
   */
  async cancelOrder(id) {
    try {
      const response = await api.put(`/api/orders/${id}/status`, {
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
      const response = await api.put(`/api/orders/${id}/status`, {
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
