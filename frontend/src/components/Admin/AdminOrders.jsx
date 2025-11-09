import { useState, useEffect } from 'react';
import orderService from '../../services/orderService';
import { formatCurrency, formatDate } from '../../utils/orderHelpers';

/**
 * AdminOrders Component
 * Página de administração para gerenciar pedidos
 * Permite visualizar todos os pedidos e atualizar seus status
 */
const AdminOrders = () => {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(false);
  const [updating, setUpdating] = useState(null); // ID do pedido sendo atualizado

  const statusOptions = [
    { value: 'pending', label: 'Pendente', color: 'bg-yellow-100 text-yellow-800' },
    { value: 'confirmed', label: 'Confirmado', color: 'bg-blue-100 text-blue-800' },
    { value: 'preparing', label: 'Em Preparação', color: 'bg-purple-100 text-purple-800' },
    { value: 'shipped', label: 'Enviado', color: 'bg-indigo-100 text-indigo-800' },
    { value: 'delivered', label: 'Entregue', color: 'bg-green-100 text-green-800' },
    { value: 'cancelled', label: 'Cancelado', color: 'bg-red-100 text-red-800' }
  ];

  useEffect(() => {
    loadOrders();
  }, []);

  const loadOrders = async () => {
    try {
      setLoading(true);
      const response = await orderService.getAllOrders();
      
      // Se response.data é array direto
      if (Array.isArray(response.data)) {
        setOrders(response.data);
      } else {
        // Se está dentro de response.data.data ou response.data.orders
        setOrders(response.data || []);
      }
    } catch (error) {
      console.error('Erro ao carregar pedidos:', error);
      alert('Erro ao carregar pedidos. Verifique o console.');
    } finally {
      setLoading(false);
    }
  };

  const handleStatusChange = async (orderId, newStatus) => {
    if (!window.confirm(`Confirma a mudança de status para "${statusOptions.find(s => s.value === newStatus)?.label}"?`)) {
      return;
    }

    try {
      setUpdating(orderId);
      const response = await orderService.updateOrderStatus(orderId, newStatus);
      
      if (response.success) {
        // Atualizar o pedido localmente
        setOrders(prevOrders => 
          prevOrders.map(order => 
            order.id === orderId 
              ? { ...order, status: newStatus, updated_at: new Date().toISOString() }
              : order
          )
        );
        
        alert(`✅ Status atualizado para: ${statusOptions.find(s => s.value === newStatus)?.label}`);
      } else {
        alert(`❌ Erro: ${response.error}`);
      }
    } catch (error) {
      console.error('Erro ao atualizar status:', error);
      alert('❌ Erro ao atualizar status do pedido');
    } finally {
      setUpdating(null);
    }
  };

  const getStatusColor = (status) => {
    return statusOptions.find(s => s.value === status)?.color || 'bg-gray-100 text-gray-800';
  };

  const getStatusLabel = (status) => {
    return statusOptions.find(s => s.value === status)?.label || status;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Carregando pedidos...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">Gerenciar Pedidos</h1>
              <p className="mt-2 text-gray-600">
                Atualize o status dos pedidos dos clientes
              </p>
            </div>
            <button
              onClick={loadOrders}
              className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              <svg className="h-5 w-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              Atualizar
            </button>
          </div>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4 mb-8">
          {statusOptions.slice(0, 4).map((status) => {
            const count = orders.filter(o => o.status === status.value).length;
            return (
              <div key={status.value} className="bg-white overflow-hidden shadow rounded-lg">
                <div className="p-5">
                  <div className="flex items-center">
                    <div className="flex-1">
                      <dt className="text-sm font-medium text-gray-500 truncate">
                        {status.label}
                      </dt>
                      <dd className="mt-1 text-3xl font-semibold text-gray-900">
                        {count}
                      </dd>
                    </div>
                    <div className={`flex-shrink-0 p-3 rounded-full ${status.color}`}>
                      <span className="text-lg font-bold">{count}</span>
                    </div>
                  </div>
                </div>
              </div>
            );
          })}
        </div>

        {/* Orders Table */}
        <div className="bg-white shadow overflow-hidden sm:rounded-lg">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    ID do Pedido
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Cliente
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Itens
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Valor Total
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status Atual
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Atualizar Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Data
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {orders.map((order) => (
                  <tr key={order.id} className="hover:bg-gray-50 transition-colors">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm font-mono text-gray-900">
                        {order.id.slice(0, 8)}...
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm font-mono text-gray-500">
                        {order.user_id.slice(0, 8)}...
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {order.items?.length || 0} {order.items?.length === 1 ? 'item' : 'itens'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-semibold text-gray-900">
                      {formatCurrency(order.total_amount)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`px-3 py-1 inline-flex text-xs leading-5 font-semibold rounded-full ${getStatusColor(order.status)}`}>
                        {getStatusLabel(order.status)}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <select
                        value={order.status}
                        onChange={(e) => handleStatusChange(order.id, e.target.value)}
                        disabled={updating === order.id}
                        className={`block w-full px-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 ${
                          updating === order.id ? 'bg-gray-100 cursor-not-allowed' : 'bg-white cursor-pointer'
                        }`}
                      >
                        {statusOptions.map((status) => (
                          <option key={status.value} value={status.value}>
                            {status.label}
                          </option>
                        ))}
                      </select>
                      {updating === order.id && (
                        <p className="mt-1 text-xs text-gray-500">Atualizando...</p>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {formatDate(order.created_at)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Empty State */}
          {orders.length === 0 && !loading && (
            <div className="text-center py-12">
              <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              <h3 className="mt-2 text-sm font-medium text-gray-900">Nenhum pedido encontrado</h3>
              <p className="mt-1 text-sm text-gray-500">Não há pedidos cadastrados no momento.</p>
            </div>
          )}
        </div>

        {/* Legend */}
        <div className="mt-6 bg-white shadow sm:rounded-lg p-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Fluxo de Status</h3>
          <div className="flex flex-wrap gap-3">
            {statusOptions.map((status, index) => (
              <div key={status.value} className="flex items-center">
                <span className={`px-3 py-1 inline-flex text-xs leading-5 font-semibold rounded-full ${status.color}`}>
                  {status.label}
                </span>
                {index < statusOptions.length - 2 && (
                  <svg className="h-5 w-5 mx-2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                  </svg>
                )}
              </div>
            ))}
          </div>
          <p className="mt-4 text-sm text-gray-500">
            ℹ️ Cada mudança de status envia uma notificação automática por e-mail ao cliente
          </p>
        </div>
      </div>
    </div>
  );
};

export default AdminOrders;
