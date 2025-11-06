import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import orderService from '../../services/orderService';
import { Card, LoadingSpinner } from '../Layout';
import {
  formatCurrency,
  calculateTotal,
  validateOrderItem,
  validateAddress,
  formatCEP,
  BRAZILIAN_STATES,
} from '../../utils/orderHelpers';
import toast from 'react-hot-toast';

/**
 * OrderForm Component
 * Formulário para criar novo pedido
 */
const OrderForm = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [showItemForm, setShowItemForm] = useState(false);

  // Itens do pedido
  const [items, setItems] = useState([]);
  const [currentItem, setCurrentItem] = useState({
    product_name: '',
    quantity: 1,
    price: '',
  });
  const [itemErrors, setItemErrors] = useState({});

  // Endereço
  const [address, setAddress] = useState({
    street: '',
    number: '',
    complement: '',
    city: '',
    state: '',
    zip_code: '',
  });
  const [addressErrors, setAddressErrors] = useState({});
  const [loadingCEP, setLoadingCEP] = useState(false);

  // Adicionar item ao pedido
  const handleAddItem = () => {
    const validation = validateOrderItem(currentItem);
    
    if (!validation.isValid) {
      setItemErrors(validation.errors);
      return;
    }

    const newItem = {
      ...currentItem,
      price: parseFloat(currentItem.price),
      quantity: parseInt(currentItem.quantity),
      product_id: `product_${Date.now()}`, // ID temporário
    };

    setItems([...items, newItem]);
    setCurrentItem({ product_name: '', quantity: 1, price: '' });
    setItemErrors({});
    setShowItemForm(false);
    toast.success('Item adicionado ao pedido!');
  };

  // Remover item
  const handleRemoveItem = (index) => {
    const newItems = items.filter((_, i) => i !== index);
    setItems(newItems);
    toast.success('Item removido do pedido');
  };

  // Buscar CEP na API ViaCEP
  const handleCEPBlur = async () => {
    const cep = address.zip_code.replace(/\D/g, '');
    
    if (cep.length !== 8) return;

    try {
      setLoadingCEP(true);
      const response = await fetch(`https://viacep.com.br/ws/${cep}/json/`);
      const data = await response.json();

      if (data.erro) {
        toast.error('CEP não encontrado');
        return;
      }

      setAddress({
        ...address,
        street: data.logradouro || '',
        city: data.localidade || '',
        state: data.uf || '',
        complement: data.complemento || address.complement,
      });
      
      toast.success('Endereço encontrado!');
    } catch (error) {
      console.error('Erro ao buscar CEP:', error);
      toast.error('Erro ao buscar CEP');
    } finally {
      setLoadingCEP(false);
    }
  };

  // Submeter pedido
  const handleSubmit = async (e) => {
    e.preventDefault();

    // Validar itens
    if (items.length === 0) {
      toast.error('Adicione pelo menos um item ao pedido');
      return;
    }

    // Validar endereço
    const addressValidation = validateAddress(address);
    if (!addressValidation.isValid) {
      setAddressErrors(addressValidation.errors);
      toast.error('Preencha todos os campos obrigatórios do endereço');
      return;
    }

    try {
      setLoading(true);

      const orderData = {
        items: items.map(item => ({
          product_id: item.product_id,
          product_name: item.product_name,
          quantity: item.quantity,
          price: item.price,
        })),
        address: {
          street: address.street,
          number: address.number,
          complement: address.complement,
          city: address.city,
          state: address.state,
          zip_code: address.zip_code,
        },
        total_amount: calculateTotal(items),
      };

      const response = await orderService.createOrder(orderData);

      if (response.success) {
        toast.success('Pedido criado com sucesso!');
        navigate(`/orders/${response.data.id || response.data._id}`);
      } else {
        toast.error(response.error || 'Erro ao criar pedido');
      }
    } catch (error) {
      console.error('Erro ao criar pedido:', error);
      toast.error('Erro ao criar pedido. Tente novamente.');
    } finally {
      setLoading(false);
    }
  };

  const totalAmount = calculateTotal(items);

  return (
    <form onSubmit={handleSubmit} className="space-y-6 max-w-4xl mx-auto">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold text-gray-800">Novo Pedido</h1>
        <button
          type="button"
          onClick={() => navigate('/orders')}
          className="text-gray-600 hover:text-gray-800 transition-colors"
        >
          ← Voltar
        </button>
      </div>

      {/* Seção de Itens */}
      <Card variant="elevated" padding="large">
        <div className="flex justify-between items-center mb-6">
          <div>
            <h2 className="text-xl font-bold text-gray-800">Itens do Pedido</h2>
            <p className="text-sm text-gray-600 mt-1">
              {items.length} {items.length === 1 ? 'item adicionado' : 'itens adicionados'}
            </p>
          </div>
          <button
            type="button"
            onClick={() => setShowItemForm(!showItemForm)}
            className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors font-medium flex items-center space-x-2"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            <span>Adicionar Item</span>
          </button>
        </div>

        {/* Formulário de Item */}
        {showItemForm && (
          <Card variant="primary" padding="normal" className="mb-6">
            <h3 className="font-semibold text-gray-800 mb-4">Novo Item</h3>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Nome do Produto *
                </label>
                <input
                  type="text"
                  value={currentItem.product_name}
                  onChange={(e) => setCurrentItem({ ...currentItem, product_name: e.target.value })}
                  className={`w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent ${
                    itemErrors.product_name ? 'border-red-500' : 'border-gray-300'
                  }`}
                  placeholder="Ex: Pizza Margherita"
                />
                {itemErrors.product_name && (
                  <p className="text-red-600 text-sm mt-1">{itemErrors.product_name}</p>
                )}
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Quantidade *
                </label>
                <input
                  type="number"
                  min="1"
                  value={currentItem.quantity}
                  onChange={(e) => setCurrentItem({ ...currentItem, quantity: e.target.value })}
                  className={`w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent ${
                    itemErrors.quantity ? 'border-red-500' : 'border-gray-300'
                  }`}
                />
                {itemErrors.quantity && (
                  <p className="text-red-600 text-sm mt-1">{itemErrors.quantity}</p>
                )}
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Preço Unitário (R$) *
                </label>
                <input
                  type="number"
                  min="0"
                  step="0.01"
                  value={currentItem.price}
                  onChange={(e) => setCurrentItem({ ...currentItem, price: e.target.value })}
                  className={`w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent ${
                    itemErrors.price ? 'border-red-500' : 'border-gray-300'
                  }`}
                  placeholder="0.00"
                />
                {itemErrors.price && (
                  <p className="text-red-600 text-sm mt-1">{itemErrors.price}</p>
                )}
              </div>
            </div>
            <div className="flex justify-end space-x-3 mt-4">
              <button
                type="button"
                onClick={() => {
                  setShowItemForm(false);
                  setCurrentItem({ product_name: '', quantity: 1, price: '' });
                  setItemErrors({});
                }}
                className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
              >
                Cancelar
              </button>
              <button
                type="button"
                onClick={handleAddItem}
                className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
              >
                Adicionar ao Pedido
              </button>
            </div>
          </Card>
        )}

        {/* Lista de Itens */}
        {items.length > 0 ? (
          <div className="space-y-3">
            {items.map((item, index) => (
              <div
                key={index}
                className="flex justify-between items-center p-4 bg-gray-50 rounded-lg border border-gray-200"
              >
                <div className="flex-1">
                  <h4 className="font-semibold text-gray-800">{item.product_name}</h4>
                  <p className="text-sm text-gray-600">
                    {item.quantity}x {formatCurrency(item.price)} = {formatCurrency(item.quantity * item.price)}
                  </p>
                </div>
                <button
                  type="button"
                  onClick={() => handleRemoveItem(index)}
                  className="text-red-600 hover:text-red-800 transition-colors p-2"
                  title="Remover item"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            ))}
          </div>
        ) : (
          <div className="text-center py-12 text-gray-500">
            <svg className="w-16 h-16 mx-auto mb-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
            </svg>
            <p>Nenhum item adicionado ainda</p>
            <p className="text-sm mt-1">Clique em "Adicionar Item" para começar</p>
          </div>
        )}
      </Card>

      {/* Seção de Endereço */}
      <Card variant="elevated" padding="large">
        <h2 className="text-xl font-bold text-gray-800 mb-6">Endereço de Entrega</h2>
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* CEP */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              CEP *
            </label>
            <div className="relative">
              <input
                type="text"
                value={address.zip_code}
                onChange={(e) => setAddress({ ...address, zip_code: formatCEP(e.target.value) })}
                onBlur={handleCEPBlur}
                maxLength={9}
                className={`w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent ${
                  addressErrors.zip_code ? 'border-red-500' : 'border-gray-300'
                }`}
                placeholder="00000-000"
              />
              {loadingCEP && (
                <div className="absolute right-3 top-3">
                  <LoadingSpinner size="small" text="" />
                </div>
              )}
            </div>
            {addressErrors.zip_code && (
              <p className="text-red-600 text-sm mt-1">{addressErrors.zip_code}</p>
            )}
          </div>

          {/* Estado */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Estado *
            </label>
            <select
              value={address.state}
              onChange={(e) => setAddress({ ...address, state: e.target.value })}
              className={`w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent ${
                addressErrors.state ? 'border-red-500' : 'border-gray-300'
              }`}
            >
              <option value="">Selecione...</option>
              {BRAZILIAN_STATES.map((state) => (
                <option key={state.value} value={state.value}>
                  {state.label}
                </option>
              ))}
            </select>
            {addressErrors.state && (
              <p className="text-red-600 text-sm mt-1">{addressErrors.state}</p>
            )}
          </div>

          {/* Rua */}
          <div className="md:col-span-2">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Rua *
            </label>
            <input
              type="text"
              value={address.street}
              onChange={(e) => setAddress({ ...address, street: e.target.value })}
              className={`w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent ${
                addressErrors.street ? 'border-red-500' : 'border-gray-300'
              }`}
              placeholder="Nome da rua"
            />
            {addressErrors.street && (
              <p className="text-red-600 text-sm mt-1">{addressErrors.street}</p>
            )}
          </div>

          {/* Número */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Número *
            </label>
            <input
              type="text"
              value={address.number}
              onChange={(e) => setAddress({ ...address, number: e.target.value })}
              className={`w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent ${
                addressErrors.number ? 'border-red-500' : 'border-gray-300'
              }`}
              placeholder="123"
            />
            {addressErrors.number && (
              <p className="text-red-600 text-sm mt-1">{addressErrors.number}</p>
            )}
          </div>

          {/* Complemento */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Complemento
            </label>
            <input
              type="text"
              value={address.complement}
              onChange={(e) => setAddress({ ...address, complement: e.target.value })}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="Apto, bloco, etc (opcional)"
            />
          </div>

          {/* Cidade */}
          <div className="md:col-span-2">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Cidade *
            </label>
            <input
              type="text"
              value={address.city}
              onChange={(e) => setAddress({ ...address, city: e.target.value })}
              className={`w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent ${
                addressErrors.city ? 'border-red-500' : 'border-gray-300'
              }`}
              placeholder="Nome da cidade"
            />
            {addressErrors.city && (
              <p className="text-red-600 text-sm mt-1">{addressErrors.city}</p>
            )}
          </div>
        </div>
      </Card>

      {/* Resumo do Pedido */}
      <Card variant="primary" padding="large">
        <h2 className="text-xl font-bold text-gray-800 mb-4">Resumo do Pedido</h2>
        <div className="space-y-3">
          <div className="flex justify-between text-lg">
            <span className="text-gray-700">Total de Itens:</span>
            <span className="font-semibold">{items.length}</span>
          </div>
          <div className="flex justify-between text-2xl border-t pt-3">
            <span className="font-bold text-gray-800">Valor Total:</span>
            <span className="font-bold text-blue-600">{formatCurrency(totalAmount)}</span>
          </div>
        </div>
      </Card>

      {/* Botão de Submit */}
      <div className="flex justify-end space-x-4">
        <button
          type="button"
          onClick={() => navigate('/orders')}
          className="px-6 py-3 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors font-medium"
          disabled={loading}
        >
          Cancelar
        </button>
        <button
          type="submit"
          disabled={loading || items.length === 0}
          className="bg-blue-600 text-white px-8 py-3 rounded-lg hover:bg-blue-700 transition-colors font-medium disabled:bg-gray-400 disabled:cursor-not-allowed flex items-center space-x-2"
        >
          {loading ? (
            <>
              <LoadingSpinner size="small" text="" />
              <span>Criando Pedido...</span>
            </>
          ) : (
            <>
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
              <span>Finalizar Pedido</span>
            </>
          )}
        </button>
      </div>
    </form>
  );
};

export default OrderForm;
