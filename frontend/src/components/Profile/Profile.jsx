import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import orderService from '../../services/orderService';
import { Card, LoadingSpinner } from '../Layout';
import { formatCurrency } from '../../utils/orderHelpers';
import toast from 'react-hot-toast';

/**
 * Profile Component
 * Página de perfil do usuário com edição de dados e estatísticas
 */
const Profile = () => {
  const { user, updateProfile } = useAuth();
  const [loading, setLoading] = useState(false);
  const [loadingStats, setLoadingStats] = useState(true);
  const [formData, setFormData] = useState({
    name: '',
    phone: '',
  });
  const [stats, setStats] = useState({
    totalOrders: 0,
    totalSpent: 0,
    lastOrder: null,
  });

  useEffect(() => {
    if (user) {
      setFormData({
        name: user.name || '',
        phone: user.phone || '',
      });
    }
  }, [user]);

  useEffect(() => {
    fetchStatistics();
  }, []);

  const fetchStatistics = async () => {
    try {
      setLoadingStats(true);
      const response = await orderService.getOrders({ limit: 100 });
      
      if (response.success) {
        const orders = response.data.orders || response.data || [];
        const totalSpent = orders.reduce((sum, order) => {
          return sum + (order.total_amount || order.totalAmount || 0);
        }, 0);
        
        const sortedOrders = [...orders].sort((a, b) => {
          const dateA = new Date(a.created_at || a.createdAt);
          const dateB = new Date(b.created_at || b.createdAt);
          return dateB - dateA;
        });

        setStats({
          totalOrders: orders.length,
          totalSpent,
          lastOrder: sortedOrders[0] || null,
        });
      }
    } catch (error) {
      console.error('Erro ao buscar estatísticas:', error);
    } finally {
      setLoadingStats(false);
    }
  };

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData({
      ...formData,
      [name]: value,
    });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!formData.name.trim()) {
      toast.error('Nome é obrigatório');
      return;
    }

    try {
      setLoading(true);
      const result = await updateProfile(formData.name, formData.phone);
      
      if (result.success) {
        toast.success('Perfil atualizado com sucesso!');
      }
    } catch (error) {
      console.error('Erro ao atualizar perfil:', error);
      toast.error('Erro ao atualizar perfil');
    } finally {
      setLoading(false);
    }
  };

  const getLastOrderDate = () => {
    if (!stats.lastOrder) return 'Nenhum pedido ainda';
    
    const date = new Date(stats.lastOrder.created_at || stats.lastOrder.createdAt);
    return date.toLocaleDateString('pt-BR', {
      day: '2-digit',
      month: 'long',
      year: 'numeric',
    });
  };

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold text-gray-800">Meu Perfil</h1>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Coluna Principal - Formulário */}
        <div className="lg:col-span-2 space-y-6">
          <Card variant="elevated" padding="large">
            <div className="flex items-center space-x-4 mb-6 pb-6 border-b border-gray-200">
              <div className="w-20 h-20 bg-gradient-to-r from-blue-500 to-indigo-500 rounded-full flex items-center justify-center text-white text-2xl font-bold">
                {user?.name?.split(' ').map(n => n[0]).join('').substring(0, 2).toUpperCase() || 'U'}
              </div>
              <div>
                <h2 className="text-2xl font-bold text-gray-800">{user?.name}</h2>
                <p className="text-gray-600">{user?.email}</p>
              </div>
            </div>

            <form onSubmit={handleSubmit} className="space-y-6">
              <h3 className="text-lg font-semibold text-gray-800">Informações Pessoais</h3>
              
              {/* Nome */}
              <div>
                <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
                  Nome Completo *
                </label>
                <input
                  type="text"
                  id="name"
                  name="name"
                  value={formData.name}
                  onChange={handleChange}
                  required
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="Seu nome completo"
                />
              </div>

              {/* Email (readonly) */}
              <div>
                <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
                  Email
                </label>
                <input
                  type="email"
                  id="email"
                  value={user?.email || ''}
                  readOnly
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-600 cursor-not-allowed"
                  title="Email não pode ser alterado"
                />
                <p className="text-xs text-gray-500 mt-1">
                  O email não pode ser alterado
                </p>
              </div>

              {/* Telefone */}
              <div>
                <label htmlFor="phone" className="block text-sm font-medium text-gray-700 mb-1">
                  Telefone
                </label>
                <input
                  type="tel"
                  id="phone"
                  name="phone"
                  value={formData.phone}
                  onChange={handleChange}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="(00) 00000-0000"
                />
              </div>

              {/* Botões */}
              <div className="flex flex-col sm:flex-row gap-3 pt-4">
                <button
                  type="submit"
                  disabled={loading}
                  className="flex-1 bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700 transition-colors font-medium disabled:bg-gray-400 disabled:cursor-not-allowed flex items-center justify-center space-x-2"
                >
                  {loading ? (
                    <>
                      <LoadingSpinner size="small" text="" />
                      <span>Salvando...</span>
                    </>
                  ) : (
                    <>
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                      <span>Salvar Alterações</span>
                    </>
                  )}
                </button>
                
                <Link
                  to="/change-password"
                  className="flex-1 bg-gray-600 text-white px-6 py-3 rounded-lg hover:bg-gray-700 transition-colors font-medium text-center flex items-center justify-center space-x-2"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                  </svg>
                  <span>Alterar Senha</span>
                </Link>
              </div>
            </form>
          </Card>
        </div>

        {/* Coluna Lateral - Estatísticas */}
        <div className="space-y-6">
          <Card variant="elevated" padding="large">
            <h3 className="text-lg font-semibold text-gray-800 mb-4">Estatísticas</h3>
            
            {loadingStats ? (
              <div className="space-y-4">
                {[1, 2, 3].map((i) => (
                  <div key={i} className="animate-pulse">
                    <div className="h-4 bg-gray-200 rounded w-2/3 mb-2"></div>
                    <div className="h-6 bg-gray-200 rounded w-1/2"></div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="space-y-6">
                {/* Total de Pedidos */}
                <div>
                  <p className="text-sm text-gray-600 mb-1">Total de Pedidos</p>
                  <p className="text-3xl font-bold text-blue-600">{stats.totalOrders}</p>
                </div>

                {/* Total Gasto */}
                <div>
                  <p className="text-sm text-gray-600 mb-1">Total Gasto</p>
                  <p className="text-2xl font-bold text-green-600">{formatCurrency(stats.totalSpent)}</p>
                </div>

                {/* Último Pedido */}
                <div>
                  <p className="text-sm text-gray-600 mb-1">Último Pedido</p>
                  <p className="text-sm font-medium text-gray-800">{getLastOrderDate()}</p>
                  {stats.lastOrder && (
                    <Link
                      to={`/orders/${stats.lastOrder.id || stats.lastOrder._id}`}
                      className="text-blue-600 hover:text-blue-700 text-sm mt-2 inline-block"
                    >
                      Ver detalhes →
                    </Link>
                  )}
                </div>
              </div>
            )}
          </Card>

          {/* Card de Ações Rápidas */}
          <Card variant="primary" padding="normal">
            <h3 className="font-semibold text-gray-800 mb-3">Ações Rápidas</h3>
            <div className="space-y-2">
              <Link
                to="/orders/new"
                className="block w-full bg-white text-blue-600 px-4 py-2 rounded-lg hover:bg-blue-50 transition-colors text-center font-medium"
              >
                Fazer Novo Pedido
              </Link>
              <Link
                to="/orders"
                className="block w-full bg-white text-blue-600 px-4 py-2 rounded-lg hover:bg-blue-50 transition-colors text-center font-medium"
              >
                Ver Meus Pedidos
              </Link>
            </div>
          </Card>
        </div>
      </div>
    </div>
  );
};

export default Profile;
