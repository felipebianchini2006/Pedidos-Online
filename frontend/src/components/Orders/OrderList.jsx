import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import orderService from '../../services/orderService';
import { Card } from '../Layout';
import StatusBadge from './StatusBadge';
import { formatDate, formatCurrency, ORDER_STATUSES } from '../../utils/orderHelpers';

/**
 * OrderList Component
 * Lista todos os pedidos do usuário com paginação e filtros
 */
const OrderList = () => {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [statusFilter, setStatusFilter] = useState('');

  const fetchOrders = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const params = {
        page,
        limit: 9,
        ...(statusFilter && { status: statusFilter }),
      };
      
      const response = await orderService.getOrders(params);
      
      if (response.success) {
        setOrders(response.data.orders || response.data);
        setTotalPages(response.data.totalPages || 1);
      } else {
        setError(response.error || 'Erro ao carregar pedidos');
      }
    } catch (err) {
      console.error('Erro ao buscar pedidos:', err);
      setError('Erro ao carregar pedidos. Tente novamente.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchOrders();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, statusFilter]);

  const handleRetry = () => {
    fetchOrders();
  };

  const handlePreviousPage = () => {
    if (page > 1) {
      setPage(page - 1);
    }
  };

  const handleNextPage = () => {
    if (page < totalPages) {
      setPage(page + 1);
    }
  };

  const handleStatusFilterChange = (e) => {
    setStatusFilter(e.target.value);
    setPage(1); // Reset para primeira página ao filtrar
  };

  // Loading State com Skeleton
  if (loading) {
    return (
      <div className="space-y-6">
        <div className="flex justify-between items-center">
          <div className="h-8 bg-gray-200 rounded w-48 animate-pulse" />
          <div className="h-10 bg-gray-200 rounded w-40 animate-pulse" />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <Card key={i} variant="outlined" padding="normal">
              <div className="space-y-4 animate-pulse">
                <div className="h-6 bg-gray-200 rounded w-3/4" />
                <div className="h-4 bg-gray-200 rounded w-1/2" />
                <div className="h-4 bg-gray-200 rounded w-2/3" />
                <div className="h-10 bg-gray-200 rounded" />
              </div>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  // Error State
  if (error) {
    return (
      <Card variant="danger" padding="large" className="text-center">
        <div className="space-y-4">
          <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto">
            <svg className="w-8 h-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <h3 className="text-lg font-semibold text-red-800">Erro ao Carregar Pedidos</h3>
          <p className="text-red-600">{error}</p>
          <button
            onClick={handleRetry}
            className="bg-red-600 text-white px-6 py-2 rounded-lg hover:bg-red-700 transition-colors font-medium"
          >
            Tentar Novamente
          </button>
        </div>
      </Card>
    );
  }

  // Empty State
  if (!loading && orders.length === 0) {
    return (
      <Card variant="elevated" padding="large" className="text-center">
        <div className="space-y-4">
          <div className="w-24 h-24 bg-blue-100 rounded-full flex items-center justify-center mx-auto">
            <svg className="w-12 h-12 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
            </svg>
          </div>
          <h3 className="text-2xl font-bold text-gray-800">Nenhum Pedido Encontrado</h3>
          <p className="text-gray-600">
            {statusFilter
              ? `Você não tem pedidos com status "${ORDER_STATUSES.find(s => s.value === statusFilter)?.label}"`
              : 'Você ainda não fez nenhum pedido'}
          </p>
          <Link
            to="/orders/new"
            className="inline-block bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700 transition-colors font-medium"
          >
            Fazer Primeiro Pedido
          </Link>
        </div>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header com Título e Filtro */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-3xl font-bold text-gray-800">Meus Pedidos</h1>
          <p className="text-gray-600 mt-1">
            {orders.length} {orders.length === 1 ? 'pedido encontrado' : 'pedidos encontrados'}
          </p>
        </div>
        
        {/* Filtro de Status */}
        <div className="flex items-center space-x-3">
          <label htmlFor="status-filter" className="text-sm font-medium text-gray-700">
            Filtrar por:
          </label>
          <select
            id="status-filter"
            value={statusFilter}
            onChange={handleStatusFilterChange}
            className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            {ORDER_STATUSES.map((status) => (
              <option key={status.value} value={status.value}>
                {status.label}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Grid de Pedidos */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {orders.map((order) => (
          <Card
            key={order.id || order._id}
            variant="elevated"
            padding="normal"
            hover
            className="transition-all duration-300"
          >
            <div className="space-y-4">
              {/* Header do Card */}
              <div className="flex justify-between items-start">
                <div>
                  <h3 className="text-lg font-bold text-gray-800">
                    Pedido #{order.id || order._id?.slice(-6)}
                  </h3>
                  <p className="text-sm text-gray-500 mt-1">
                    {formatDate(order.created_at || order.createdAt)}
                  </p>
                </div>
                <StatusBadge status={order.status} />
              </div>

              {/* Informações do Pedido */}
              <div className="space-y-2 border-t border-gray-100 pt-4">
                <div className="flex justify-between text-sm">
                  <span className="text-gray-600">Itens:</span>
                  <span className="font-medium text-gray-800">
                    {order.items?.length || 0} {order.items?.length === 1 ? 'item' : 'itens'}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600 font-medium">Total:</span>
                  <span className="text-lg font-bold text-blue-600">
                    {formatCurrency(order.total_amount || order.totalAmount)}
                  </span>
                </div>
              </div>

              {/* Botão Ver Detalhes */}
              <Link
                to={`/orders/${order.id || order._id}`}
                className="block w-full bg-blue-600 text-white text-center py-2 rounded-lg hover:bg-blue-700 transition-colors font-medium"
              >
                Ver Detalhes
              </Link>
            </div>
          </Card>
        ))}
      </div>

      {/* Paginação */}
      {totalPages > 1 && (
        <div className="flex justify-center items-center space-x-4 pt-6">
          <button
            onClick={handlePreviousPage}
            disabled={page === 1}
            className={`
              px-4 py-2 rounded-lg font-medium transition-colors
              ${page === 1
                ? 'bg-gray-200 text-gray-400 cursor-not-allowed'
                : 'bg-blue-600 text-white hover:bg-blue-700'
              }
            `}
          >
            ← Anterior
          </button>
          
          <span className="text-gray-700 font-medium">
            Página {page} de {totalPages}
          </span>
          
          <button
            onClick={handleNextPage}
            disabled={page === totalPages}
            className={`
              px-4 py-2 rounded-lg font-medium transition-colors
              ${page === totalPages
                ? 'bg-gray-200 text-gray-400 cursor-not-allowed'
                : 'bg-blue-600 text-white hover:bg-blue-700'
              }
            `}
          >
            Próximo →
          </button>
        </div>
      )}

      {/* Botão Novo Pedido (fixo no canto inferior direito em mobile) */}
      <Link
        to="/orders/new"
        className="fixed bottom-6 right-6 md:hidden bg-blue-600 text-white p-4 rounded-full shadow-lg hover:bg-blue-700 transition-all hover:scale-110"
        title="Novo Pedido"
      >
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
        </svg>
      </Link>
    </div>
  );
};

export default OrderList;
