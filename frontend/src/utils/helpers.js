import { format, parseISO, formatDistanceToNow } from 'date-fns';
import { ptBR } from 'date-fns/locale';

/**
 * Formatar valor para moeda brasileira (R$)
 * @param {number} value - Valor numérico
 * @returns {string} Valor formatado (ex: R$ 1.234,56)
 */
export function formatCurrency(value) {
  if (value === null || value === undefined || isNaN(value)) {
    return 'R$ 0,00';
  }

  return new Intl.NumberFormat('pt-BR', {
    style: 'currency',
    currency: 'BRL',
  }).format(value);
}

/**
 * Formatar data no formato brasileiro com hora
 * @param {string|Date} date - Data a ser formatada
 * @returns {string} Data formatada (ex: 15/01/2024 10:30)
 */
export function formatDate(date) {
  if (!date) return '-';

  try {
    const parsedDate = typeof date === 'string' ? parseISO(date) : date;
    return format(parsedDate, "dd/MM/yyyy HH:mm", { locale: ptBR });
  } catch (error) {
    console.error('Erro ao formatar data:', error);
    return '-';
  }
}

/**
 * Formatar data no formato brasileiro sem hora
 * @param {string|Date} date - Data a ser formatada
 * @returns {string} Data formatada (ex: 15/01/2024)
 */
export function formatDateOnly(date) {
  if (!date) return '-';

  try {
    const parsedDate = typeof date === 'string' ? parseISO(date) : date;
    return format(parsedDate, "dd/MM/yyyy", { locale: ptBR });
  } catch (error) {
    console.error('Erro ao formatar data:', error);
    return '-';
  }
}

/**
 * Formatar data relativa (ex: "há 2 horas")
 * @param {string|Date} date - Data a ser formatada
 * @returns {string} Data relativa
 */
export function formatRelativeDate(date) {
  if (!date) return '-';

  try {
    const parsedDate = typeof date === 'string' ? parseISO(date) : date;
    return formatDistanceToNow(parsedDate, { 
      addSuffix: true, 
      locale: ptBR 
    });
  } catch (error) {
    console.error('Erro ao formatar data relativa:', error);
    return '-';
  }
}

/**
 * Obter classe de cor Tailwind baseada no status do pedido
 * @param {string} status - Status do pedido
 * @returns {string} Classes Tailwind CSS
 */
export function getStatusColor(status) {
  const statusColors = {
    pending: 'bg-yellow-100 text-yellow-800 border-yellow-200',
    confirmed: 'bg-blue-100 text-blue-800 border-blue-200',
    preparing: 'bg-purple-100 text-purple-800 border-purple-200',
    shipped: 'bg-indigo-100 text-indigo-800 border-indigo-200',
    delivered: 'bg-green-100 text-green-800 border-green-200',
    cancelled: 'bg-red-100 text-red-800 border-red-200',
  };

  return statusColors[status] || 'bg-gray-100 text-gray-800 border-gray-200';
}

/**
 * Obter label em português do status do pedido
 * @param {string} status - Status do pedido
 * @returns {string} Label em português
 */
export function getStatusLabel(status) {
  const statusLabels = {
    pending: 'Pendente',
    confirmed: 'Confirmado',
    preparing: 'Em Preparação',
    shipped: 'Enviado',
    delivered: 'Entregue',
    cancelled: 'Cancelado',
  };

  return statusLabels[status] || status;
}

/**
 * Validar formato de email
 * @param {string} email - Email a ser validado
 * @returns {boolean} True se válido, false caso contrário
 */
export function validateEmail(email) {
  if (!email) return false;

  // Regex para validação de email
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}

/**
 * Validar formato de telefone brasileiro
 * Aceita formatos: (11) 99999-9999, 11999999999, +5511999999999
 * @param {string} phone - Telefone a ser validado
 * @returns {boolean} True se válido, false caso contrário
 */
export function validatePhone(phone) {
  if (!phone) return false;

  // Remover caracteres não numéricos
  const cleanPhone = phone.replace(/\D/g, '');

  // Validar:
  // - 11 dígitos: (DDD + 9 + 8 dígitos) para celular
  // - 10 dígitos: (DDD + 8 dígitos) para fixo
  // - 13 dígitos: +55 + DDD + 9 + 8 dígitos (com código do país)
  // - 12 dígitos: +55 + DDD + 8 dígitos (fixo com código do país)
  const isValid = /^(\+?55)?([1-9]{2})(9?\d{8})$/.test(cleanPhone);

  return isValid;
}

/**
 * Formatar telefone brasileiro
 * @param {string} phone - Telefone a ser formatado
 * @returns {string} Telefone formatado
 */
export function formatPhone(phone) {
  if (!phone) return '';

  // Remover caracteres não numéricos
  const cleanPhone = phone.replace(/\D/g, '');

  // Formatar baseado no tamanho
  if (cleanPhone.length === 11) {
    // Celular: (11) 99999-9999
    return cleanPhone.replace(/(\d{2})(\d{5})(\d{4})/, '($1) $2-$3');
  } else if (cleanPhone.length === 10) {
    // Fixo: (11) 9999-9999
    return cleanPhone.replace(/(\d{2})(\d{4})(\d{4})/, '($1) $2-$3');
  } else if (cleanPhone.length === 13) {
    // Celular com código do país: +55 (11) 99999-9999
    return cleanPhone.replace(/(\d{2})(\d{2})(\d{5})(\d{4})/, '+$1 ($2) $3-$4');
  } else if (cleanPhone.length === 12) {
    // Fixo com código do país: +55 (11) 9999-9999
    return cleanPhone.replace(/(\d{2})(\d{2})(\d{4})(\d{4})/, '+$1 ($2) $3-$4');
  }

  return phone;
}

/**
 * Validar CPF brasileiro
 * @param {string} cpf - CPF a ser validado
 * @returns {boolean} True se válido, false caso contrário
 */
export function validateCPF(cpf) {
  if (!cpf) return false;

  // Remover caracteres não numéricos
  const cleanCPF = cpf.replace(/\D/g, '');

  // Validar tamanho
  if (cleanCPF.length !== 11) return false;

  // Validar se todos os dígitos são iguais (ex: 111.111.111-11)
  if (/^(\d)\1{10}$/.test(cleanCPF)) return false;

  // Validar dígitos verificadores
  let sum = 0;
  let remainder;

  for (let i = 1; i <= 9; i++) {
    sum += parseInt(cleanCPF.substring(i - 1, i)) * (11 - i);
  }

  remainder = (sum * 10) % 11;
  if (remainder === 10 || remainder === 11) remainder = 0;
  if (remainder !== parseInt(cleanCPF.substring(9, 10))) return false;

  sum = 0;
  for (let i = 1; i <= 10; i++) {
    sum += parseInt(cleanCPF.substring(i - 1, i)) * (12 - i);
  }

  remainder = (sum * 10) % 11;
  if (remainder === 10 || remainder === 11) remainder = 0;
  if (remainder !== parseInt(cleanCPF.substring(10, 11))) return false;

  return true;
}

/**
 * Formatar CPF brasileiro
 * @param {string} cpf - CPF a ser formatado
 * @returns {string} CPF formatado (ex: 123.456.789-00)
 */
export function formatCPF(cpf) {
  if (!cpf) return '';

  const cleanCPF = cpf.replace(/\D/g, '');

  if (cleanCPF.length === 11) {
    return cleanCPF.replace(/(\d{3})(\d{3})(\d{3})(\d{2})/, '$1.$2.$3-$4');
  }

  return cpf;
}

/**
 * Validar CEP brasileiro
 * @param {string} cep - CEP a ser validado
 * @returns {boolean} True se válido, false caso contrário
 */
export function validateCEP(cep) {
  if (!cep) return false;

  const cleanCEP = cep.replace(/\D/g, '');
  return /^\d{8}$/.test(cleanCEP);
}

/**
 * Formatar CEP brasileiro
 * @param {string} cep - CEP a ser formatado
 * @returns {string} CEP formatado (ex: 12345-678)
 */
export function formatCEP(cep) {
  if (!cep) return '';

  const cleanCEP = cep.replace(/\D/g, '');

  if (cleanCEP.length === 8) {
    return cleanCEP.replace(/(\d{5})(\d{3})/, '$1-$2');
  }

  return cep;
}

/**
 * Truncar texto longo
 * @param {string} text - Texto a ser truncado
 * @param {number} maxLength - Tamanho máximo
 * @returns {string} Texto truncado
 */
export function truncateText(text, maxLength = 50) {
  if (!text) return '';
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength) + '...';
}

/**
 * Gerar iniciais do nome
 * @param {string} name - Nome completo
 * @returns {string} Iniciais (ex: "João Silva" -> "JS")
 */
export function getInitials(name) {
  if (!name) return '';

  const words = name.trim().split(' ');
  if (words.length === 1) return words[0].substring(0, 2).toUpperCase();

  return (words[0][0] + words[words.length - 1][0]).toUpperCase();
}

/**
 * Calcular total de itens do pedido
 * @param {Array} items - Array de itens
 * @returns {number} Total calculado
 */
export function calculateOrderTotal(items) {
  if (!items || !Array.isArray(items)) return 0;

  return items.reduce((total, item) => {
    return total + (item.price * item.quantity);
  }, 0);
}
