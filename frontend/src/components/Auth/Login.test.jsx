import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { BrowserRouter } from 'react-router-dom';
import Login from './Login';
import { AuthProvider } from '../../context/AuthContext';
import authService from '../../services/authService';

// Mock do useNavigate
const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

// Mock do authService
vi.mock('../../services/authService');

// Mock do react-hot-toast
vi.mock('react-hot-toast', () => ({
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

// Componente wrapper com providers necessários
function Wrapper({ children }) {
  return (
    <BrowserRouter>
      <AuthProvider>{children}</AuthProvider>
    </BrowserRouter>
  );
}

describe('Login Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    authService.getToken.mockReturnValue(null);
  });

  describe('Renderização', () => {
    it('deve renderizar form com campos email e password', () => {
      render(<Login />, { wrapper: Wrapper });

      expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/senha/i)).toBeInTheDocument();
      expect(
        screen.getByRole('button', { name: /entrar/i })
      ).toBeInTheDocument();
    });

    it('deve renderizar título e descrição', () => {
      render(<Login />, { wrapper: Wrapper });

      expect(screen.getByText(/bem-vindo de volta!/i)).toBeInTheDocument();
      expect(
        screen.getByText(/faça login para acessar sua conta/i)
      ).toBeInTheDocument();
    });

    it('deve renderizar link para página de registro', () => {
      render(<Login />, { wrapper: Wrapper });

      const registerLink = screen.getByRole('link', {
        name: /cadastre-se aqui/i,
      });
      expect(registerLink).toBeInTheDocument();
      expect(registerLink).toHaveAttribute('href', '/register');
    });
  });

  describe('Validação de formulário', () => {
    it('deve validar que botão fica habilitado quando campos estão preenchidos', async () => {
      const user = userEvent.setup();
      render(<Login />, { wrapper: Wrapper });

      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/senha/i);
      const submitButton = screen.getByRole('button', { name: /entrar/i });

      // Inicialmente habilitado (mas validação ocorre no submit)
      expect(submitButton).not.toBeDisabled();

      await user.type(emailInput, 'test@example.com');
      await user.type(passwordInput, 'password123');

      expect(submitButton).not.toBeDisabled();
    });

    it('deve mostrar erro para email vazio', async () => {
      const user = userEvent.setup();
      render(<Login />, { wrapper: Wrapper });

      const submitButton = screen.getByRole('button', { name: /entrar/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText(/email é obrigatório/i)).toBeInTheDocument();
      });
    });

    it('deve mostrar erro para email inválido', async () => {
      const user = userEvent.setup();
      render(<Login />, { wrapper: Wrapper });

      const emailInput = screen.getByLabelText(/email/i);
      const submitButton = screen.getByRole('button', { name: /entrar/i });

      await user.type(emailInput, 'invalid-email');
      await user.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText(/email inválido/i)).toBeInTheDocument();
      });
    });

    it('deve mostrar erro para senha vazia', async () => {
      const user = userEvent.setup();
      render(<Login />, { wrapper: Wrapper });

      const emailInput = screen.getByLabelText(/email/i);
      const submitButton = screen.getByRole('button', { name: /entrar/i });

      await user.type(emailInput, 'test@example.com');
      await user.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText(/senha é obrigatória/i)).toBeInTheDocument();
      });
    });

    it('deve limpar erro do campo ao digitar', async () => {
      const user = userEvent.setup();
      render(<Login />, { wrapper: Wrapper });

      const emailInput = screen.getByLabelText(/email/i);
      const submitButton = screen.getByRole('button', { name: /entrar/i });

      // Trigger validation error
      await user.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText(/email é obrigatório/i)).toBeInTheDocument();
      });

      // Digitar no campo deve limpar o erro
      await user.type(emailInput, 'test@example.com');

      await waitFor(() => {
        expect(
          screen.queryByText(/email é obrigatório/i)
        ).not.toBeInTheDocument();
      });
    });
  });

  describe('Submit do formulário', () => {
    it('deve chamar authService.login ao submeter com dados válidos', async () => {
      const user = userEvent.setup();
      authService.login.mockResolvedValue({
        success: true,
        data: {
          token: 'mock-token',
          user: { id: '123', email: 'test@example.com' },
        },
      });

      render(<Login />, { wrapper: Wrapper });

      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/senha/i);
      const submitButton = screen.getByRole('button', { name: /entrar/i });

      await user.type(emailInput, 'test@example.com');
      await user.type(passwordInput, 'password123');
      await user.click(submitButton);

      await waitFor(() => {
        expect(authService.login).toHaveBeenCalledWith(
          'test@example.com',
          'password123'
        );
      });
    });

    it('deve redirecionar para / após login bem-sucedido', async () => {
      const user = userEvent.setup();
      authService.login.mockResolvedValue({
        success: true,
        data: {
          token: 'mock-token',
          user: { id: '123', email: 'test@example.com' },
        },
      });

      render(<Login />, { wrapper: Wrapper });

      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/senha/i);
      const submitButton = screen.getByRole('button', { name: /entrar/i });

      await user.type(emailInput, 'test@example.com');
      await user.type(passwordInput, 'password123');
      await user.click(submitButton);

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/');
      });
    });

    it('deve mostrar estado de loading durante submit', async () => {
      const user = userEvent.setup();

      // Simular login lento
      authService.login.mockImplementation(
        () =>
          new Promise((resolve) =>
            setTimeout(
              () =>
                resolve({
                  success: true,
                  data: { token: 'mock-token', user: {} },
                }),
              100
            )
          )
      );

      render(<Login />, { wrapper: Wrapper });

      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/senha/i);
      const submitButton = screen.getByRole('button', { name: /entrar/i });

      await user.type(emailInput, 'test@example.com');
      await user.type(passwordInput, 'password123');
      await user.click(submitButton);

      // Durante loading
      await waitFor(() => {
        expect(screen.getByText(/entrando.../i)).toBeInTheDocument();
      });

      // Botão deve estar desabilitado durante loading
      expect(submitButton).toBeDisabled();

      // Após concluir
      await waitFor(
        () => {
          expect(screen.queryByText(/entrando.../i)).not.toBeInTheDocument();
        },
        { timeout: 200 }
      );
    });

    it('não deve redirecionar se login falhar', async () => {
      const user = userEvent.setup();
      authService.login.mockResolvedValue({
        success: false,
        error: 'Credenciais inválidas',
      });

      render(<Login />, { wrapper: Wrapper });

      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/senha/i);
      const submitButton = screen.getByRole('button', { name: /entrar/i });

      await user.type(emailInput, 'wrong@example.com');
      await user.type(passwordInput, 'wrongpassword');
      await user.click(submitButton);

      await waitFor(() => {
        expect(authService.login).toHaveBeenCalled();
      });

      // Não deve navegar
      expect(mockNavigate).not.toHaveBeenCalled();
    });
  });

  describe('Campos do formulário', () => {
    it('deve ter autocomplete correto nos campos', () => {
      render(<Login />, { wrapper: Wrapper });

      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/senha/i);

      expect(emailInput).toHaveAttribute('autocomplete', 'email');
      expect(passwordInput).toHaveAttribute('autocomplete', 'current-password');
    });

    it('deve ter tipo correto nos inputs', () => {
      render(<Login />, { wrapper: Wrapper });

      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/senha/i);

      expect(emailInput).toHaveAttribute('type', 'email');
      expect(passwordInput).toHaveAttribute('type', 'password');
    });

    it('deve ter placeholders nos campos', () => {
      render(<Login />, { wrapper: Wrapper });

      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/senha/i);

      expect(emailInput).toHaveAttribute('placeholder', 'seu@email.com');
      expect(passwordInput).toHaveAttribute('placeholder', '••••••••');
    });

    it('deve desabilitar inputs durante loading', async () => {
      const user = userEvent.setup();

      authService.login.mockImplementation(
        () =>
          new Promise((resolve) =>
            setTimeout(
              () =>
                resolve({
                  success: true,
                  data: { token: 'mock-token', user: {} },
                }),
              100
            )
          )
      );

      render(<Login />, { wrapper: Wrapper });

      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/senha/i);
      const submitButton = screen.getByRole('button', { name: /entrar/i });

      await user.type(emailInput, 'test@example.com');
      await user.type(passwordInput, 'password123');
      await user.click(submitButton);

      await waitFor(() => {
        expect(emailInput).toBeDisabled();
        expect(passwordInput).toBeDisabled();
      });
    });
  });
});
