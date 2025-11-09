import { describe, it, expect } from 'vitest';
import {
  formatCurrency,
  formatDate,
  formatDateOnly,
  formatRelativeDate,
  getStatusColor,
  getStatusLabel,
  validateEmail,
  validatePhone,
  formatPhone,
  validateCPF,
  formatCPF,
  validateCEP,
  formatCEP,
  truncateText,
  getInitials,
  calculateOrderTotal,
} from './helpers';

describe('formatCurrency', () => {
  it('deve formatar valores numéricos corretamente', () => {
    expect(formatCurrency(1234.56)).toBe('R$ 1.234,56');
    expect(formatCurrency(0)).toBe('R$ 0,00');
    expect(formatCurrency(100)).toBe('R$ 100,00');
  });

  it('deve retornar R$ 0,00 para valores inválidos', () => {
    expect(formatCurrency(null)).toBe('R$ 0,00');
    expect(formatCurrency(undefined)).toBe('R$ 0,00');
    expect(formatCurrency(NaN)).toBe('R$ 0,00');
  });
});

describe('formatDate', () => {
  it('deve formatar datas ISO corretamente', () => {
    const result = formatDate('2024-01-15T10:30:00Z');
    expect(result).toMatch(/15\/01\/2024/);
  });

  it('deve formatar objetos Date', () => {
    const date = new Date('2024-01-15T10:30:00Z');
    const result = formatDate(date);
    expect(result).toMatch(/15\/01\/2024/);
  });

  it('deve retornar "-" para valores inválidos', () => {
    expect(formatDate(null)).toBe('-');
    expect(formatDate(undefined)).toBe('-');
    expect(formatDate('')).toBe('-');
  });
});

describe('formatDateOnly', () => {
  it('deve formatar data sem hora', () => {
    const result = formatDateOnly('2024-01-15T10:30:00Z');
    expect(result).toBe('15/01/2024');
  });

  it('deve retornar "-" para valores inválidos', () => {
    expect(formatDateOnly(null)).toBe('-');
  });
});

describe('formatRelativeDate', () => {
  it('deve formatar data relativa', () => {
    const recentDate = new Date();
    const result = formatRelativeDate(recentDate);
    expect(result).toContain('há');
  });

  it('deve retornar "-" para valores inválidos', () => {
    expect(formatRelativeDate(null)).toBe('-');
  });
});

describe('getStatusColor', () => {
  it('deve retornar classes corretas para cada status', () => {
    expect(getStatusColor('pending')).toContain('yellow');
    expect(getStatusColor('confirmed')).toContain('blue');
    expect(getStatusColor('preparing')).toContain('purple');
    expect(getStatusColor('shipped')).toContain('indigo');
    expect(getStatusColor('delivered')).toContain('green');
    expect(getStatusColor('cancelled')).toContain('red');
  });

  it('deve retornar classe padrão para status desconhecido', () => {
    expect(getStatusColor('unknown')).toContain('gray');
  });
});

describe('getStatusLabel', () => {
  it('deve retornar labels em português', () => {
    expect(getStatusLabel('pending')).toBe('Pendente');
    expect(getStatusLabel('confirmed')).toBe('Confirmado');
    expect(getStatusLabel('preparing')).toBe('Em Preparação');
    expect(getStatusLabel('shipped')).toBe('Enviado');
    expect(getStatusLabel('delivered')).toBe('Entregue');
    expect(getStatusLabel('cancelled')).toBe('Cancelado');
  });

  it('deve retornar o status original se não mapeado', () => {
    expect(getStatusLabel('unknown')).toBe('unknown');
  });
});

describe('validateEmail', () => {
  it('deve validar emails corretos', () => {
    expect(validateEmail('test@example.com')).toBe(true);
    expect(validateEmail('user.name@domain.co.uk')).toBe(true);
    expect(validateEmail('user+tag@example.com')).toBe(true);
  });

  it('deve rejeitar emails inválidos', () => {
    expect(validateEmail('invalid')).toBe(false);
    expect(validateEmail('invalid@')).toBe(false);
    expect(validateEmail('@example.com')).toBe(false);
    expect(validateEmail('test@')).toBe(false);
    expect(validateEmail('')).toBe(false);
    expect(validateEmail(null)).toBe(false);
  });
});

describe('validatePhone', () => {
  it('deve validar telefones celulares válidos', () => {
    expect(validatePhone('11999999999')).toBe(true);
    expect(validatePhone('(11) 99999-9999')).toBe(true);
    expect(validatePhone('+5511999999999')).toBe(true);
  });

  it('deve validar telefones fixos válidos', () => {
    expect(validatePhone('1133334444')).toBe(true);
    expect(validatePhone('(11) 3333-4444')).toBe(true);
  });

  it('deve rejeitar telefones inválidos', () => {
    expect(validatePhone('123')).toBe(false);
    expect(validatePhone('00000000000')).toBe(false);
    expect(validatePhone('')).toBe(false);
    expect(validatePhone(null)).toBe(false);
  });
});

describe('formatPhone', () => {
  it('deve formatar celular de 11 dígitos', () => {
    expect(formatPhone('11999999999')).toBe('(11) 99999-9999');
  });

  it('deve formatar fixo de 10 dígitos', () => {
    expect(formatPhone('1133334444')).toBe('(11) 3333-4444');
  });

  it('deve formatar celular com código do país', () => {
    expect(formatPhone('5511999999999')).toBe('+55 (11) 99999-9999');
  });

  it('deve retornar string vazia para valores inválidos', () => {
    expect(formatPhone('')).toBe('');
    expect(formatPhone(null)).toBe('');
  });
});

describe('validateCPF', () => {
  it('deve validar CPFs válidos', () => {
    expect(validateCPF('12345678909')).toBe(true);
    expect(validateCPF('111.444.777-35')).toBe(true);
  });

  it('deve rejeitar CPFs inválidos', () => {
    expect(validateCPF('00000000000')).toBe(false);
    expect(validateCPF('11111111111')).toBe(false);
    expect(validateCPF('12345678901')).toBe(false);
    expect(validateCPF('123')).toBe(false);
    expect(validateCPF('')).toBe(false);
    expect(validateCPF(null)).toBe(false);
  });
});

describe('formatCPF', () => {
  it('deve formatar CPF de 11 dígitos', () => {
    expect(formatCPF('12345678909')).toBe('123.456.789-09');
  });

  it('deve retornar string original se não tiver 11 dígitos', () => {
    expect(formatCPF('123')).toBe('123');
  });

  it('deve retornar string vazia para valores inválidos', () => {
    expect(formatCPF('')).toBe('');
    expect(formatCPF(null)).toBe('');
  });
});

describe('validateCEP', () => {
  it('deve validar CEPs válidos', () => {
    expect(validateCEP('12345678')).toBe(true);
    expect(validateCEP('12345-678')).toBe(true);
  });

  it('deve rejeitar CEPs inválidos', () => {
    expect(validateCEP('123')).toBe(false);
    expect(validateCEP('123456789')).toBe(false);
    expect(validateCEP('')).toBe(false);
    expect(validateCEP(null)).toBe(false);
  });
});

describe('formatCEP', () => {
  it('deve formatar CEP de 8 dígitos', () => {
    expect(formatCEP('12345678')).toBe('12345-678');
  });

  it('deve retornar string original se não tiver 8 dígitos', () => {
    expect(formatCEP('123')).toBe('123');
  });

  it('deve retornar string vazia para valores inválidos', () => {
    expect(formatCEP('')).toBe('');
    expect(formatCEP(null)).toBe('');
  });
});

describe('truncateText', () => {
  it('deve truncar textos longos', () => {
    const longText = 'a'.repeat(100);
    expect(truncateText(longText, 50)).toBe('a'.repeat(50) + '...');
  });

  it('deve manter textos curtos intactos', () => {
    expect(truncateText('short', 50)).toBe('short');
  });

  it('deve retornar string vazia para valores inválidos', () => {
    expect(truncateText('')).toBe('');
    expect(truncateText(null)).toBe('');
  });
});

describe('getInitials', () => {
  it('deve gerar iniciais de nome completo', () => {
    expect(getInitials('João Silva')).toBe('JS');
    expect(getInitials('Maria da Silva Santos')).toBe('MS');
  });

  it('deve gerar duas letras para nome único', () => {
    expect(getInitials('João')).toBe('JO');
  });

  it('deve retornar string vazia para valores inválidos', () => {
    expect(getInitials('')).toBe('');
    expect(getInitials(null)).toBe('');
  });
});

describe('calculateOrderTotal', () => {
  it('deve calcular total de itens corretamente', () => {
    const items = [
      { price: 50, quantity: 2 },
      { price: 30, quantity: 1 },
    ];
    expect(calculateOrderTotal(items)).toBe(130);
  });

  it('deve retornar 0 para array vazio', () => {
    expect(calculateOrderTotal([])).toBe(0);
  });

  it('deve retornar 0 para valores inválidos', () => {
    expect(calculateOrderTotal(null)).toBe(0);
    expect(calculateOrderTotal(undefined)).toBe(0);
    expect(calculateOrderTotal('invalid')).toBe(0);
  });
});
