/**
 * Aplicar máscara de telefone brasileiro
 * Formato: (XX) XXXXX-XXXX ou (XX) XXXX-XXXX
 * @param {string} value - Valor do input
 * @returns {string} Valor formatado
 */
export function phoneMask(value) {
  if (!value) return '';
  
  // Remove tudo que não é número
  const numbers = value.replace(/\D/g, '');
  
  // Aplica a máscara baseado no tamanho
  if (numbers.length <= 2) {
    return `(${numbers}`;
  } else if (numbers.length <= 6) {
    return `(${numbers.slice(0, 2)}) ${numbers.slice(2)}`;
  } else if (numbers.length <= 10) {
    return `(${numbers.slice(0, 2)}) ${numbers.slice(2, 6)}-${numbers.slice(6)}`;
  } else {
    // Celular com 9 dígitos
    return `(${numbers.slice(0, 2)}) ${numbers.slice(2, 7)}-${numbers.slice(7, 11)}`;
  }
}

/**
 * Remover máscara do telefone
 * @param {string} value - Valor formatado
 * @returns {string} Apenas números
 */
export function removePhoneMask(value) {
  return value.replace(/\D/g, '');
}

/**
 * Calcular força da senha
 * @param {string} password - Senha a ser avaliada
 * @returns {object} { strength: number (0-4), label: string, color: string }
 */
export function calculatePasswordStrength(password) {
  if (!password) {
    return { strength: 0, label: '', color: '', percentage: 0 };
  }

  let strength = 0;
  
  // Critérios de força
  const criteria = {
    length: password.length >= 8,
    lowercase: /[a-z]/.test(password),
    uppercase: /[A-Z]/.test(password),
    numbers: /\d/.test(password),
    special: /[!@#$%^&*(),.?":{}|<>]/.test(password),
  };

  // Contar critérios atendidos
  strength += criteria.length ? 1 : 0;
  strength += criteria.lowercase ? 1 : 0;
  strength += criteria.uppercase ? 1 : 0;
  strength += criteria.numbers ? 1 : 0;
  strength += criteria.special ? 1 : 0;

  // Determinar label e cor baseado na força
  let label = '';
  let color = '';
  let percentage = 0;

  if (strength <= 2) {
    label = 'Fraca';
    color = 'bg-red-500';
    percentage = 33;
  } else if (strength === 3) {
    label = 'Média';
    color = 'bg-yellow-500';
    percentage = 66;
  } else {
    label = 'Forte';
    color = 'bg-green-500';
    percentage = 100;
  }

  return { strength, label, color, percentage, criteria };
}

/**
 * Validar formulário de login
 * @param {object} values - Valores do formulário
 * @returns {object} Erros encontrados
 */
export function validateLoginForm(values) {
  const errors = {};

  if (!values.email) {
    errors.email = 'Email é obrigatório';
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(values.email)) {
    errors.email = 'Email inválido';
  }

  if (!values.password) {
    errors.password = 'Senha é obrigatória';
  }

  return errors;
}

/**
 * Validar formulário de registro
 * @param {object} values - Valores do formulário
 * @returns {object} Erros encontrados
 */
export function validateRegisterForm(values) {
  const errors = {};

  // Nome
  if (!values.name) {
    errors.name = 'Nome é obrigatório';
  } else if (values.name.length < 3) {
    errors.name = 'Nome deve ter pelo menos 3 caracteres';
  }

  // Email
  if (!values.email) {
    errors.email = 'Email é obrigatório';
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(values.email)) {
    errors.email = 'Email inválido';
  }

  // Telefone
  if (!values.phone) {
    errors.phone = 'Telefone é obrigatório';
  } else {
    const numbers = removePhoneMask(values.phone);
    if (numbers.length < 10 || numbers.length > 11) {
      errors.phone = 'Telefone inválido';
    }
  }

  // Senha
  if (!values.password) {
    errors.password = 'Senha é obrigatória';
  } else if (values.password.length < 8) {
    errors.password = 'Senha deve ter pelo menos 8 caracteres';
  }

  // Confirmar senha
  if (!values.confirmPassword) {
    errors.confirmPassword = 'Confirmação de senha é obrigatória';
  } else if (values.password !== values.confirmPassword) {
    errors.confirmPassword = 'As senhas não coincidem';
  }

  return errors;
}
