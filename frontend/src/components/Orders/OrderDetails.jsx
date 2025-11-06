import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import orderService from '../../services/orderService';
import { Card, LoadingSpinner } from '../Layout';
import StatusBadge from './StatusBadge';
import OrderTimeline from './OrderTimeline';
import { formatDate, formatCurrency } from '../../utils/orderHelpers';
import toast from 'react-hot-toast';

/**
 * OrderDetails Component
 * Exibe detalhes completos de um pedido
 */
const OrderDetails = () => {
  const { id } = useParams();
  const [order, setOrder] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showCancelModal, setShowCancelModal] = useState(false);
  const [cancelling, setCancelling] = useState(false);

  useEffect(() => {
    fetchOrderDetails();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id]);

  const fetchOrderDetails = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const response = await orderService.getOrderById(id);
      
      if (response.success) {
        setOrder(response.data);
      } else {
        setError(response.error || 'Erro ao carregar pedido');
      }
    } catch (err) {
      console.error('Erro ao buscar detalhes do pedido:', err);
      setError('Erro ao carregar pedido. Tente novamente.');
    } finally {
      setLoading(false);
    }
  };

  const handleCancelOrder = async () => {
    try {
      setCancelling(true);
      
      const response = await orderService.updateOrderStatus(id, 'cancelled');
      
      if (response.success) {
        toast.success('Pedido cancelado com sucesso!');
        setOrder({ ...order, status: 'cancelled' });
        setShowCancelModal(false);
      } else {
        toast.error(response.error || 'Erro ao cancelar pedido');
      }
    } catch (error) {
      console.error('Erro ao cancelar pedido:', error);
      toast.error('Erro ao cancelar pedido. Tente novamente.');
    } finally {
      setCancelling(false);
    }
  };

  const canCancelOrder = () => {
    return order && (order.status === 'pending' || order.status === 'confirmed');
  };

  // Loading State
  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-[400px]">
        <LoadingSpinner text="Carregando detalhes do pedido..." />
      </div>
    );
  }

  // Error State
  if (error) {
    return (
      <Card variant="danger" padding="large" className="text-center max-w-2xl mx-auto">
        <div className="space-y-4">
          <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto">
            <svg className="w-8 h-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <h3 className="text-lg font-semibold text-red-800">Erro ao Carregar Pedido</h3>
          <p className="text-red-600">{error}</p>
          <div className="flex justify-center space-x-4">
            <button
              onClick={fetchOrderDetails}
              className="bg-red-600 text-white px-6 py-2 rounded-lg hover:bg-red-700 transition-colors font-medium"
            >
              Tentar Novamente
            </button>
            <Link
              to="/orders"
              className="bg-gray-600 text-white px-6 py-2 rounded-lg hover:bg-gray-700 transition-colors font-medium"
            >
              Voltar para Pedidos
            </Link>
          </div>
        </div>
      </Card>
    );
  }

  if (!order) {
    return null;
  }

  return (
    <div className="space-y-6 max-w-5xl mx-auto">
      {/* Header */}
      <div className="flex justify-between items-start">
        <div>
          <div className="flex items-center space-x-3 mb-2">
            <h1 className="text-3xl font-bold text-gray-800">
              Pedido #{order.id || order._id?.slice(-6)}
            </h1>
            <StatusBadge status={order.status} />
          </div>
          <p className="text-gray-600">
            Criado em {formatDate(order.created_at || order.createdAt)}
          </p>
        </div>
        <Link
          to="/orders"
          className="text-gray-600 hover:text-gray-800 transition-colors flex items-center space-x-2"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
          <span>Voltar</span>
        </Link>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Coluna Principal */}
        <div className="lg:col-span-2 space-y-6">
          {/* Itens do Pedido */}
          <Card variant="elevated" padding="large">
            <h2 className="text-xl font-bold text-gray-800 mb-6">Itens do Pedido</h2>
            
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-gray-200">
                    <th className="text-left py-3 px-4 font-semibold text-gray-700">Produto</th>
                    <th className="text-center py-3 px-4 font-semibold text-gray-700">Qtd</th>
                    <th className="text-right py-3 px-4 font-semibold text-gray-700">Preço Unit.</th>
                    <th className="text-right py-3 px-4 font-semibold text-gray-700">Subtotal</th>
                  </tr>
                </thead>
                <tbody>
                  {order.items?.map((item, index) => (
                    <tr key={index} className="border-b border-gray-100 hover:bg-gray-50">
                      <td className="py-4 px-4">
                        <p className="font-medium text-gray-800">{item.product_name}</p>
                      </td>
                      <td className="py-4 px-4 text-center text-gray-700">
                        {item.quantity}
                      </td>
                      <td className="py-4 px-4 text-right text-gray-700">
                        {formatCurrency(item.price)}
                      </td>
                      <td className="py-4 px-4 text-right font-semibold text-gray-800">
                        {formatCurrency(item.price * item.quantity)}
                      </td>
                    </tr>
                  ))}
                </tbody>
                <tfoot>
                  <tr className="border-t-2 border-gray-300">
                    <td colSpan="3" className="py-4 px-4 text-right font-bold text-gray-800">
                      Total:
                    </td>
                    <td className="py-4 px-4 text-right font-bold text-blue-600 text-xl">
                      {formatCurrency(order.total_amount || order.totalAmount)}
                    </td>
                  </tr>
                </tfoot>
              </table>
            </div>
          </Card>

          {/* Endereço de Entrega */}
          <Card variant="elevated" padding="large">
            <h2 className="text-xl font-bold text-gray-800 mb-4">Endereço de Entrega</h2>
            <div className="bg-gray-50 rounded-lg p-4 space-y-2">
              <p className="text-gray-800">
                <span className="font-semibold">Rua:</span> {order.address?.street}, {order.address?.number}
              </p>
              {order.address?.complement && (
                <p className="text-gray-800">
                  <span className="font-semibold">Complemento:</span> {order.address.complement}
                </p>
              )}
              <p className="text-gray-800">
                <span className="font-semibold">Cidade:</span> {order.address?.city} - {order.address?.state}
              </p>
              <p className="text-gray-800">
                <span className="font-semibold">CEP:</span> {order.address?.zip_code}
              </p>
            </div>
          </Card>

          {/* Botão Cancelar */}
          {canCancelOrder() && (
            <button
              onClick={() => setShowCancelModal(true)}
              className="w-full bg-red-600 text-white py-3 rounded-lg hover:bg-red-700 transition-colors font-medium flex items-center justify-center space-x-2"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
              <span>Cancelar Pedido</span>
            </button>
          )}
        </div>

        {/* Timeline */}
        <div className="lg:col-span-1">
          <Card variant="elevated" padding="large">
            <h2 className="text-xl font-bold text-gray-800 mb-6">Status do Pedido</h2>
            <OrderTimeline currentStatus={order.status} />
          </Card>
        </div>
      </div>

      {/* Modal de Confirmação de Cancelamento */}
      {showCancelModal && (
        <>
          <div 
            className="fixed inset-0 bg-black bg-opacity-50 z-40"
            onClick={() => !cancelling && setShowCancelModal(false)}
          />
          <div className="fixed inset-0 flex items-center justify-center z-50 p-4">
            <Card variant="elevated" padding="large" className="max-w-md w-full">
              <div className="text-center space-y-4">
                <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto">
                  <svg className="w-8 h-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-gray-800">Cancelar Pedido?</h3>
                <p className="text-gray-600">
                  Tem certeza que deseja cancelar este pedido? Esta ação não pode ser desfeita.
                </p>
                <div className="flex space-x-3 pt-4">
                  <button
                    onClick={() => setShowCancelModal(false)}
                    disabled={cancelling}
                    className="flex-1 px-4 py-3 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors font-medium disabled:opacity-50"
                  >
                    Não, voltar
                  </button>
                  <button
                    onClick={handleCancelOrder}
                    disabled={cancelling}
                    className="flex-1 bg-red-600 text-white px-4 py-3 rounded-lg hover:bg-red-700 transition-colors font-medium disabled:opacity-50 flex items-center justify-center space-x-2"
                  >
                    {cancelling ? (
                      <>
                        <LoadingSpinner size="small" text="" />
                        <span>Cancelando...</span>
                      </>
                    ) : (
                      <span>Sim, cancelar</span>
                    )}
                  </button>
                </div>
              </div>
            </Card>
          </div>
        </>
      )}
    </div>
  );
};

export default OrderDetails;
