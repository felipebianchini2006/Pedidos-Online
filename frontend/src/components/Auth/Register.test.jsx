import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { BrowserRouter } from 'react-router-dom';
import Register from './Register';
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

describe('Register Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    authService.getToken.mockReturnValue(null);
  });

  describe('Renderização', () => {
    it('deve renderizar todos os campos do formulário', () => {
      render(<Register />, { wrapper: Wrapper });

      expect(screen.getByLabelText(/nome completo/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/^email$/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/telefone/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/^senha$/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/confirmar senha/i)).toBeInTheDocument();
      expect(
        screen.getByRole('button', { name: /cadastrar/i })
      ).toBeInTheDocument();
    });

    it('deve renderizar título e descrição', () => {
      render(<Register />, { wrapper: Wrapper });

      expect(screen.getByText(/criar conta/i)).toBeInTheDocument();
      expect(
        screen.getByText(/preencha seus dados para começar/i)
      ).toBeInTheDocument();
    });

    it('deve renderizar link para página de login', () => {
      render(<Register />, { wrapper: Wrapper });

      const loginLink = screen.getByRole('link', { name: /faça login aqui/i });
      expect(loginLink).toBeInTheDocument();
      expect(loginLink).toHaveAttribute('href', '/login');
    });
  });

  describe('Validação de todos os campos', () => {
    it('deve validar nome vazio', async () => {
      const user = userEvent.setup();
      render(<Register />, { wrapper: Wrapper });

      const submitButton = screen.getByRole('button', { name: /cadastrar/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText(/nome é obrigatório/i)).toBeInTheDocument();
      });
    });

    it('deve validar nome com menos de 3 caracteres', async () => {
      const user = userEvent.setup();
      render(<Register />, { wrapper: Wrapper });

      const nameInput = screen.getByLabelText(/nome completo/i);
      const submitButton = screen.getByRole('button', { name: /cadastrar/i });

      await user.type(nameInput, 'Jo');
      await user.click(submitButton);

      await waitFor(() => {
        expect(
          screen.getByText(/nome deve ter pelo menos 3 caracteres/i)
        ).toBeInTheDocument();
      });
    });

    it('deve validar email vazio', async () => {
      const user = userEvent.setup();
      render(<Register />, { wrapper: Wrapper });

      const submitButton = screen.getByRole('button', { name: /cadastrar/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText(/email é obrigatório/i)).toBeInTheDocument();
      });
    });

    it('deve validar email inválido', async () => {
      const user = userEvent.setup();
      render(<Register />, { wrapper: Wrapper });

      const emailInput = screen.getByLabelText(/^email$/i);
      const submitButton = screen.getByRole('button', { name: /cadastrar/i });

      await user.type(emailInput, 'invalid-email');
      await user.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText(/email inválido/i)).toBeInTheDocument();
      });
    });

    it('deve validar telefone vazio', async () => {
      const user = userEvent.setup();
      render(<Register />, { wrapper: Wrapper });

      const submitButton = screen.getByRole('button', { name: /cadastrar/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText(/telefone é obrigatório/i)).toBeInTheDocument();
      });
    });

    it('deve validar senha com menos de 8 caracteres', async () => {
      const user = userEvent.setup();
      render(<Register />, { wrapper: Wrapper });

      const passwordInput = screen.getByLabelText(/^senha$/i);
      const submitButton = screen.getByRole('button', { name: /cadastrar/i });

      await user.type(passwordInput, 'pass');
      await user.click(submitButton);

      await waitFor(() => {
        expect(
          screen.getByText(/senha deve ter pelo menos 8 caracteres/i)
        ).toBeInTheDocument();
      });
    });

    it('deve validar que senhas devem ser iguais', async () => {
      const user = userEvent.setup();
      render(<Register />, { wrapper: Wrapper });

      const passwordInput = screen.getByLabelText(/^senha$/i);
      const confirmPasswordInput = screen.getByLabelText(/confirmar senha/i);
      const submitButton = screen.getByRole('button', { name: /cadastrar/i });

      await user.type(passwordInput, 'password123');
      await user.type(confirmPasswordInput, 'password456');
      await user.click(submitButton);

      await waitFor(() => {
        expect(
          screen.getByText(/as senhas não coincidem/i)
        ).toBeInTheDocument();
      });
    });
  });

  describe('Máscara de telefone', () => {
    it('deve aplicar máscara de telefone ao digitar', async () => {
      const user = userEvent.setup();
      render(<Register />, { wrapper: Wrapper });

      const phoneInput = screen.getByLabelText(/telefone/i);

      await user.type(phoneInput, '11999999999');

      expect(phoneInput).toHaveValue('(11) 99999-9999');
    });

    it('deve limitar o tamanho do telefone', () => {
      render(<Register />, { wrapper: Wrapper });

      const phoneInput = screen.getByLabelText(/telefone/i);

      expect(phoneInput).toHaveAttribute('maxLength', '15');
    });
  });

  describe('Indicador de força da senha', () => {
    it('deve mostrar indicador de força quando senha é digitada', async () => {
      const user = userEvent.setup();
      render(<Register />, { wrapper: Wrapper });

      const passwordInput = screen.getByLabelText(/^senha$/i);

      await user.type(passwordInput, 'password');

      await waitFor(() => {
        expect(screen.getByText(/fraca|média|forte/i)).toBeInTheDocument();
      });
    });

    it('deve mostrar requisitos de senha', async () => {
      const user = userEvent.setup();
      render(<Register />, { wrapper: Wrapper });

      const passwordInput = screen.getByLabelText(/^senha$/i);

      await user.type(passwordInput, 'Pass123!');

      await waitFor(() => {
        expect(screen.getByText(/mínimo 8 caracteres/i)).toBeInTheDocument();
        expect(screen.getByText(/letra maiúscula/i)).toBeInTheDocument();
        expect(screen.getByText(/letra minúscula/i)).toBeInTheDocument();
        expect(screen.getByText(/número/i)).toBeInTheDocument();
        expect(screen.getByText(/caractere especial/i)).toBeInTheDocument();
      });
    });
  });

  describe('Submit do formulário', () => {
    it('deve chamar authService.register ao submeter com dados válidos', async () => {
      const user = userEvent.setup();
      authService.register.mockResolvedValue({
        success: true,
        data: { id: 'new-user-123' },
      });

      render(<Register />, { wrapper: Wrapper });

      const nameInput = screen.getByLabelText(/nome completo/i);
      const emailInput = screen.getByLabelText(/^email$/i);
      const phoneInput = screen.getByLabelText(/telefone/i);
      const passwordInput = screen.getByLabelText(/^senha$/i);
      const confirmPasswordInput = screen.getByLabelText(/confirmar senha/i);
      const submitButton = screen.getByRole('button', { name: /cadastrar/i });

      await user.type(nameInput, 'Test User');
      await user.type(emailInput, 'test@example.com');
      await user.type(phoneInput, '11999999999');
      await user.type(passwordInput, 'password123');
      await user.type(confirmPasswordInput, 'password123');
      await user.click(submitButton);

      await waitFor(() => {
        expect(authService.register).toHaveBeenCalledWith(
          'test@example.com',
          'password123',
          'Test User',
          '11999999999' // Telefone sem máscara
        );
      });
    });

    it('deve redirecionar para /login após registro bem-sucedido', async () => {
      const user = userEvent.setup();
      authService.register.mockResolvedValue({
        success: true,
        data: { id: 'new-user-123' },
      });

      render(<Register />, { wrapper: Wrapper });

      const nameInput = screen.getByLabelText(/nome completo/i);
      const emailInput = screen.getByLabelText(/^email$/i);
      const phoneInput = screen.getByLabelText(/telefone/i);
      const passwordInput = screen.getByLabelText(/^senha$/i);
      const confirmPasswordInput = screen.getByLabelText(/confirmar senha/i);
      const submitButton = screen.getByRole('button', { name: /cadastrar/i });

      await user.type(nameInput, 'Test User');
      await user.type(emailInput, 'test@example.com');
      await user.type(phoneInput, '11999999999');
      await user.type(passwordInput, 'password123');
      await user.type(confirmPasswordInput, 'password123');
      await user.click(submitButton);

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/login');
      });
    });

    it('deve mostrar estado de loading durante submit', async () => {
      const user = userEvent.setup();

      authService.register.mockImplementation(
        () =>
          new Promise((resolve) =>
            setTimeout(
              () =>
                resolve({
                  success: true,
                  data: { id: 'new-user' },
                }),
              100
            )
          )
      );

      render(<Register />, { wrapper: Wrapper });

      const nameInput = screen.getByLabelText(/nome completo/i);
      const emailInput = screen.getByLabelText(/^email$/i);
      const phoneInput = screen.getByLabelText(/telefone/i);
      const passwordInput = screen.getByLabelText(/^senha$/i);
      const confirmPasswordInput = screen.getByLabelText(/confirmar senha/i);
      const submitButton = screen.getByRole('button', { name: /cadastrar/i });

      await user.type(nameInput, 'Test User');
      await user.type(emailInput, 'test@example.com');
      await user.type(phoneInput, '11999999999');
      await user.type(passwordInput, 'password123');
      await user.type(confirmPasswordInput, 'password123');
      await user.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText(/cadastrando.../i)).toBeInTheDocument();
      });

      expect(submitButton).toBeDisabled();
    });

    it('não deve redirecionar se registro falhar', async () => {
      const user = userEvent.setup();
      authService.register.mockResolvedValue({
        success: false,
        error: 'Email já cadastrado',
      });

      render(<Register />, { wrapper: Wrapper });

      const nameInput = screen.getByLabelText(/nome completo/i);
      const emailInput = screen.getByLabelText(/^email$/i);
      const phoneInput = screen.getByLabelText(/telefone/i);
      const passwordInput = screen.getByLabelText(/^senha$/i);
      const confirmPasswordInput = screen.getByLabelText(/confirmar senha/i);
      const submitButton = screen.getByRole('button', { name: /cadastrar/i });

      await user.type(nameInput, 'Test User');
      await user.type(emailInput, 'existing@example.com');
      await user.type(phoneInput, '11999999999');
      await user.type(passwordInput, 'password123');
      await user.type(confirmPasswordInput, 'password123');
      await user.click(submitButton);

      await waitFor(() => {
        expect(authService.register).toHaveBeenCalled();
      });

      expect(mockNavigate).not.toHaveBeenCalled();
    });
  });

  describe('Botão de submit', () => {
    it('deve estar desabilitado quando formulário é inválido', async () => {
      render(<Register />, { wrapper: Wrapper });

      const submitButton = screen.getByRole('button', { name: /cadastrar/i });

      // Formulário vazio é inválido
      expect(submitButton).toBeDisabled();
    });

    it('deve estar habilitado quando formulário é válido', async () => {
      const user = userEvent.setup();
      render(<Register />, { wrapper: Wrapper });

      const nameInput = screen.getByLabelText(/nome completo/i);
      const emailInput = screen.getByLabelText(/^email$/i);
      const phoneInput = screen.getByLabelText(/telefone/i);
      const passwordInput = screen.getByLabelText(/^senha$/i);
      const confirmPasswordInput = screen.getByLabelText(/confirmar senha/i);
      const submitButton = screen.getByRole('button', { name: /cadastrar/i });

      await user.type(nameInput, 'Test User');
      await user.type(emailInput, 'test@example.com');
      await user.type(phoneInput, '11999999999');
      await user.type(passwordInput, 'password123');
      await user.type(confirmPasswordInput, 'password123');

      await waitFor(() => {
        expect(submitButton).not.toBeDisabled();
      });
    });
  });

  describe('Campos do formulário', () => {
    it('deve ter autocomplete correto nos campos', () => {
      render(<Register />, { wrapper: Wrapper });

      const nameInput = screen.getByLabelText(/nome completo/i);
      const emailInput = screen.getByLabelText(/^email$/i);
      const phoneInput = screen.getByLabelText(/telefone/i);
      const passwordInput = screen.getByLabelText(/^senha$/i);
      const confirmPasswordInput = screen.getByLabelText(/confirmar senha/i);

      expect(nameInput).toHaveAttribute('autocomplete', 'name');
      expect(emailInput).toHaveAttribute('autocomplete', 'email');
      expect(phoneInput).toHaveAttribute('autocomplete', 'tel');
      expect(passwordInput).toHaveAttribute('autocomplete', 'new-password');
      expect(confirmPasswordInput).toHaveAttribute('autocomplete', 'new-password');
    });

    it('deve limpar erro do campo ao digitar', async () => {
      const user = userEvent.setup();
      render(<Register />, { wrapper: Wrapper });

      const emailInput = screen.getByLabelText(/^email$/i);
      const submitButton = screen.getByRole('button', { name: /cadastrar/i });

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
});
