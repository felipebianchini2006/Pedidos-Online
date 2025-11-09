/**
 * Helpers para manipulação de pedidos
 */

/**
 * Formatar valor monetário para Real Brasileiro
 */
export const formatCurrency = (value) => {
  return new Intl.NumberFormat('pt-BR', {
    style: 'currency',
    currency: 'BRL',
  }).format(value);
};

/**
 * Formatar data para formato brasileiro
 */
export const formatDate = (dateString) => {
  const date = new Date(dateString);
  return new Intl.DateTimeFormat('pt-BR', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date);
};

/**
 * Formatar data curta (apenas dia/mês/ano)
 */
export const formatDateShort = (dateString) => {
  const date = new Date(dateString);
  return new Intl.DateTimeFormat('pt-BR', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  }).format(date);
};

/**
 * Obter cor do status do pedido
 */
export const getStatusColor = (status) => {
  const colors = {
    pending: 'bg-yellow-100 text-yellow-800 border-yellow-300',
    confirmed: 'bg-blue-100 text-blue-800 border-blue-300',
    preparing: 'bg-purple-100 text-purple-800 border-purple-300',
    shipped: 'bg-indigo-100 text-indigo-800 border-indigo-300',
    delivered: 'bg-green-100 text-green-800 border-green-300',
    cancelled: 'bg-red-100 text-red-800 border-red-300',
  };
  return colors[status] || 'bg-gray-100 text-gray-800 border-gray-300';
};

/**
 * Obter label em português do status
 */
export const getStatusLabel = (status) => {
  const labels = {
    pending: 'Pendente',
    confirmed: 'Confirmado',
    preparing: 'Em Preparação',
    shipped: 'Enviado',
    delivered: 'Entregue',
    cancelled: 'Cancelado',
  };
  return labels[status] || status;
};

/**
 * Obter ícone SVG do status (retorna path string)
 */
export const getStatusIcon = (status) => {
  const icons = {
    pending: "M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z",
    confirmed: "M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z",
    preparing: "M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4",
    shipped: "M13 16V6a1 1 0 00-1-1H4a1 1 0 00-1 1v10a1 1 0 001 1h1m8-1a1 1 0 01-1 1H9m4-1V8a1 1 0 011-1h2.586a1 1 0 01.707.293l3.414 3.414a1 1 0 01.293.707V16a1 1 0 01-1 1h-1m-6-1a1 1 0 001 1h1M5 17a2 2 0 104 0m-4 0a2 2 0 114 0m6 0a2 2 0 104 0m-4 0a2 2 0 114 0",
    delivered: "M5 13l4 4L19 7",
    cancelled: "M6 18L18 6M6 6l12 12",
  };
  return icons[status] || icons.pending;
};

/**
 * Calcular total do pedido
 */
export const calculateTotal = (items) => {
  return items.reduce((total, item) => {
    return total + (item.price * item.quantity);
  }, 0);
};

/**
 * Validar CEP (formato brasileiro)
 */
export const isValidCEP = (cep) => {
  const cepRegex = /^\d{5}-?\d{3}$/;
  return cepRegex.test(cep);
};

/**
 * Formatar CEP (adicionar hífen)
 */
export const formatCEP = (cep) => {
  const cleaned = cep.replace(/\D/g, '');
  if (cleaned.length === 8) {
    return `${cleaned.slice(0, 5)}-${cleaned.slice(5)}`;
  }
  return cep;
};

/**
 * Validar item do pedido
 */
export const validateOrderItem = (item) => {
  const errors = {};
  
  if (!item.product_name || item.product_name.trim() === '') {
    errors.product_name = 'Nome do produto é obrigatório';
  }
  
  if (!item.quantity || item.quantity <= 0) {
    errors.quantity = 'Quantidade deve ser maior que zero';
  }
  
  if (!item.price || item.price <= 0) {
    errors.price = 'Preço deve ser maior que zero';
  }
  
  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
};

/**
 * Validar endereço
 */
export const validateAddress = (address) => {
  const errors = {};
  
  if (!address.street || address.street.trim() === '') {
    errors.street = 'Rua é obrigatória';
  }
  
  if (!address.number || address.number.trim() === '') {
    errors.number = 'Número é obrigatório';
  }
  
  if (!address.city || address.city.trim() === '') {
    errors.city = 'Cidade é obrigatória';
  }
  
  if (!address.state || address.state.trim() === '') {
    errors.state = 'Estado é obrigatório';
  }
  
  if (!address.zip_code || address.zip_code.trim() === '') {
    errors.zip_code = 'CEP é obrigatório';
  } else if (!isValidCEP(address.zip_code)) {
    errors.zip_code = 'CEP inválido';
  }
  
  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
};

/**
 * Estados brasileiros
 */
export const BRAZILIAN_STATES = [
  { value: 'AC', label: 'Acre' },
  { value: 'AL', label: 'Alagoas' },
  { value: 'AP', label: 'Amapá' },
  { value: 'AM', label: 'Amazonas' },
  { value: 'BA', label: 'Bahia' },
  { value: 'CE', label: 'Ceará' },
  { value: 'DF', label: 'Distrito Federal' },
  { value: 'ES', label: 'Espírito Santo' },
  { value: 'GO', label: 'Goiás' },
  { value: 'MA', label: 'Maranhão' },
  { value: 'MT', label: 'Mato Grosso' },
  { value: 'MS', label: 'Mato Grosso do Sul' },
  { value: 'MG', label: 'Minas Gerais' },
  { value: 'PA', label: 'Pará' },
  { value: 'PB', label: 'Paraíba' },
  { value: 'PR', label: 'Paraná' },
  { value: 'PE', label: 'Pernambuco' },
  { value: 'PI', label: 'Piauí' },
  { value: 'RJ', label: 'Rio de Janeiro' },
  { value: 'RN', label: 'Rio Grande do Norte' },
  { value: 'RS', label: 'Rio Grande do Sul' },
  { value: 'RO', label: 'Rondônia' },
  { value: 'RR', label: 'Roraima' },
  { value: 'SC', label: 'Santa Catarina' },
  { value: 'SP', label: 'São Paulo' },
  { value: 'SE', label: 'Sergipe' },
  { value: 'TO', label: 'Tocantins' },
];

/**
 * Status do pedido disponíveis
 */
export const ORDER_STATUSES = [
  { value: '', label: 'Todos os Status' },
  { value: 'pending', label: 'Pendente' },
  { value: 'confirmed', label: 'Confirmado' },
  { value: 'preparing', label: 'Em Preparação' },
  { value: 'shipped', label: 'Enviado' },
  { value: 'delivered', label: 'Entregue' },
  { value: 'cancelled', label: 'Cancelado' },
];
