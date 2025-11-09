import { describe, it, expect, beforeEach, vi } from 'vitest';
import authService from './authService';

// O MSW já está configurado no setup global, então os testes vão funcionar
// com os handlers mockados

describe('authService', () => {
  beforeEach(() => {
    // Limpar localStorage antes de cada teste
    localStorage.clear();
    vi.clearAllMocks();
  });

  describe('register', () => {
    it('deve chamar POST /api/users/register com dados corretos', async () => {
      const userData = {
        email: 'newuser@example.com',
        password: 'password123',
        name: 'New User',
        phone: '(11) 99999-9999',
      };

      const result = await authService.register(
        userData.email,
        userData.password,
        userData.name,
        userData.phone
      );

      expect(result.success).toBe(true);
      expect(result.data).toBeDefined();
      expect(result.data.email).toBe(userData.email);
    });

    it('deve retornar sucesso para resposta 201', async () => {
      const result = await authService.register(
        'test@example.com',
        'password123',
        'Test User',
        '11999999999'
      );

      expect(result.success).toBe(true);
      expect(result.data).toBeDefined();
    });

    it('deve retornar erro para resposta 400 (dados inválidos)', async () => {
      const result = await authService.register('', '', '', '');

      expect(result.success).toBe(false);
      expect(result.error).toBeDefined();
      expect(result.error).toContain('obrigatórios');
    });

    it('deve retornar erro para resposta 409 (email já existe)', async () => {
      const result = await authService.register(
        'existing@example.com',
        'password123',
        'Existing User',
        '11999999999'
      );

      expect(result.success).toBe(false);
      expect(result.error).toBeDefined();
      expect(result.error).toContain('já cadastrado');
    });
  });

  describe('login', () => {
    it('deve chamar POST /api/users/login com credenciais corretas', async () => {
      const result = await authService.login(
        'test@example.com',
        'password123'
      );

      expect(result.success).toBe(true);
      expect(result.data).toBeDefined();
      expect(result.data.token).toBeDefined();
      expect(result.data.user).toBeDefined();
    });

    it('deve salvar token no localStorage após login bem-sucedido', async () => {
      await authService.login('test@example.com', 'password123');

      expect(localStorage.setItem).toHaveBeenCalledWith(
        'token',
        'mock-jwt-token-123'
      );
    });

    it('deve salvar dados do usuário no localStorage', async () => {
      await authService.login('test@example.com', 'password123');

      // Verificar que setItem foi chamado com 'user'
      const userCalls = vi.mocked(localStorage.setItem).mock.calls.filter(
        call => call[0] === 'user'
      );

      expect(userCalls.length).toBeGreaterThan(0);

      const savedUser = JSON.parse(userCalls[0][1]);
      expect(savedUser.email).toBe('test@example.com');
    });

    it('deve retornar token em caso de sucesso', async () => {
      const result = await authService.login(
        'test@example.com',
        'password123'
      );

      expect(result.success).toBe(true);
      expect(result.data.token).toBe('mock-jwt-token-123');
    });

    it('deve retornar erro para credenciais inválidas', async () => {
      const result = await authService.login(
        'wrong@example.com',
        'wrongpassword'
      );

      expect(result.success).toBe(false);
      expect(result.error).toBeDefined();
    });

    it('deve retornar erro genérico se resposta não tiver erro específico', async () => {
      // Login com credenciais incorretas retorna erro genérico
      const result = await authService.login('wrong@test.com', 'wrong');

      expect(result.success).toBe(false);
      expect(result.error).toBeDefined();
    });
  });

  describe('logout', () => {
    it('deve limpar token do localStorage', () => {
      // Configurar localStorage com dados
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: '123' }));

      authService.logout();

      expect(localStorage.removeItem).toHaveBeenCalledWith('token');
      expect(localStorage.removeItem).toHaveBeenCalledWith('user');
    });

    it('deve verificar que token foi removido', () => {
      localStorage.setItem('token', 'mock-token');

      authService.logout();

      expect(localStorage.removeItem).toHaveBeenCalledWith('token');
    });

    it('deve redirecionar para /login', () => {
      authService.logout();

      expect(window.location.href).toBe('/login');
    });
  });

  describe('getProfile', () => {
    it('deve chamar GET /api/users/profile com token', async () => {
      // Simular token no localStorage
      localStorage.getItem.mockReturnValue('mock-jwt-token-123');

      const result = await authService.getProfile();

      expect(result.success).toBe(true);
      expect(result.data).toBeDefined();
      expect(result.data.email).toBe('test@example.com');
    });

    it('deve atualizar dados do usuário no localStorage', async () => {
      localStorage.getItem.mockReturnValue('mock-jwt-token-123');

      await authService.getProfile();

      // Verificar que setItem foi chamado com 'user'
      const userCalls = vi.mocked(localStorage.setItem).mock.calls.filter(
        call => call[0] === 'user'
      );

      expect(userCalls.length).toBeGreaterThan(0);
    });

    it('deve retornar erro se não houver token', async () => {
      localStorage.getItem.mockReturnValue(null);

      const result = await authService.getProfile();

      expect(result.success).toBe(false);
      expect(result.error).toBeDefined();
    });
  });

  describe('updateProfile', () => {
    it('deve chamar PUT /api/users/profile com novos dados', async () => {
      localStorage.getItem.mockReturnValue('mock-jwt-token-123');

      const result = await authService.updateProfile(
        'Updated Name',
        '(11) 88888-8888'
      );

      expect(result.success).toBe(true);
      expect(result.data).toBeDefined();
      expect(result.data.name).toBe('Updated Name');
    });

    it('deve atualizar dados do usuário no localStorage', async () => {
      localStorage.getItem.mockReturnValue('mock-jwt-token-123');

      await authService.updateProfile('New Name', '11999999999');

      const userCalls = vi.mocked(localStorage.setItem).mock.calls.filter(
        call => call[0] === 'user'
      );

      expect(userCalls.length).toBeGreaterThan(0);
    });

    it('deve retornar erro se não houver token', async () => {
      localStorage.getItem.mockReturnValue(null);

      const result = await authService.updateProfile('Name', 'Phone');

      expect(result.success).toBe(false);
      expect(result.error).toBeDefined();
    });
  });

  describe('getToken', () => {
    it('deve retornar token do localStorage', () => {
      const mockToken = 'mock-token-123';
      localStorage.getItem.mockReturnValue(mockToken);

      const token = authService.getToken();

      expect(localStorage.getItem).toHaveBeenCalledWith('token');
      expect(token).toBe(mockToken);
    });

    it('deve retornar null se não houver token', () => {
      localStorage.getItem.mockReturnValue(null);

      const token = authService.getToken();

      expect(token).toBeNull();
    });
  });

  describe('setToken', () => {
    it('deve salvar token no localStorage', () => {
      const token = 'new-token-123';

      authService.setToken(token);

      expect(localStorage.setItem).toHaveBeenCalledWith('token', token);
    });
  });

  describe('isAuthenticated', () => {
    it('deve retornar true se token existe', () => {
      localStorage.getItem.mockReturnValue('mock-token');

      const isAuth = authService.isAuthenticated();

      expect(isAuth).toBe(true);
    });

    it('deve retornar false se token não existe', () => {
      localStorage.getItem.mockReturnValue(null);

      const isAuth = authService.isAuthenticated();

      expect(isAuth).toBe(false);
    });

    it('deve retornar false se token é string vazia', () => {
      localStorage.getItem.mockReturnValue('');

      const isAuth = authService.isAuthenticated();

      expect(isAuth).toBe(false);
    });
  });

  describe('getCurrentUser', () => {
    it('deve retornar dados do usuário do localStorage', () => {
      const mockUser = { id: '123', email: 'test@example.com' };
      localStorage.getItem.mockReturnValue(JSON.stringify(mockUser));

      const user = authService.getCurrentUser();

      expect(user).toEqual(mockUser);
    });

    it('deve retornar null se não houver usuário', () => {
      localStorage.getItem.mockReturnValue(null);

      const user = authService.getCurrentUser();

      expect(user).toBeNull();
    });

    it('deve retornar null se JSON for inválido', () => {
      localStorage.getItem.mockReturnValue('invalid-json');

      const user = authService.getCurrentUser();

      expect(user).toBeNull();
    });
  });
});
