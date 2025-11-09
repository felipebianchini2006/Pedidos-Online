import { expect, afterEach, beforeAll, afterAll, vi } from 'vitest';
import { cleanup } from '@testing-library/react';
import * as matchers from '@testing-library/jest-dom/matchers';
import { server } from '../mocks/server';

// Extender matchers do Vitest com jest-dom
expect.extend(matchers);

// Iniciar MSW server antes de todos os testes
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));

// Resetar handlers após cada teste
afterEach(() => {
  server.resetHandlers();
  cleanup();
  localStorage.clear();
  sessionStorage.clear();
  vi.clearAllMocks();
});

// Fechar MSW server após todos os testes
afterAll(() => server.close());

// Mock do localStorage para testes
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
};
global.localStorage = localStorageMock;

// Mock do window.location
delete window.location;
window.location = { href: '', assign: vi.fn(), reload: vi.fn() };

// Suprimir console.error em testes (opcional)
// global.console.error = vi.fn();
