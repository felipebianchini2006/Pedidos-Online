import { setupServer } from 'msw/node';
import { handlers } from './handlers';

// Configurar servidor MSW para testes Node
export const server = setupServer(...handlers);
