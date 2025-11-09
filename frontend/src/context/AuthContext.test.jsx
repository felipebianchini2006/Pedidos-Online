import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { AuthProvider, useAuth } from './AuthContext';
import authService from '../services/authService';
import toast from 'react-hot-toast';

// Mock do authService
vi.mock('../services/authService', () => ({
  default: {
    getToken: vi.fn(),
    getProfile: vi.fn(),
    login: vi.fn(),
    register: vi.fn(),
    logout: vi.fn(),
    updateProfile: vi.fn(),
  },
}));

// Mock do react-hot-toast
vi.mock('react-hot-toast', () => ({
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

describe('AuthContext', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  describe('AuthProvider - Inicialização', () => {
    it('deve fornecer valores iniciais corretos', async () => {
      authService.getToken.mockReturnValue(null);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      expect(result.current.user).toBeNull();
      expect(result.current.isAuthenticated).toBe(false);
    });

    it('deve carregar usuário ao montar se token existe', async () => {
      const mockUser = {
        id: 'user-123',
        email: 'test@example.com',
        name: 'Test User',
      };

      authService.getToken.mockReturnValue('mock-token');
      authService.getProfile.mockResolvedValue({
        success: true,
        data: mockUser,
      });

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      expect(result.current.user).toEqual(mockUser);
      expect(result.current.isAuthenticated).toBe(true);
    });

    it('deve fazer logout se token for inválido', async () => {
      authService.getToken.mockReturnValue('invalid-token');
      authService.getProfile.mockResolvedValue({
        success: false,
        error: 'Token inválido',
      });

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      expect(authService.logout).toHaveBeenCalled();
      expect(result.current.user).toBeNull();
      expect(result.current.isAuthenticated).toBe(false);
    });

    it('deve tratar erro ao buscar perfil', async () => {
      authService.getToken.mockReturnValue('mock-token');
      authService.getProfile.mockRejectedValue(new Error('Network error'));

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      expect(authService.logout).toHaveBeenCalled();
    });
  });

  describe('login', () => {
    it('deve atualizar estado após login bem-sucedido', async () => {
      authService.getToken.mockReturnValue(null);

      const mockUser = {
        id: 'user-123',
        email: 'test@example.com',
        name: 'Test User',
      };

      authService.login.mockResolvedValue({
        success: true,
        data: {
          token: 'mock-token',
          user: mockUser,
        },
      });

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      let loginResult;
      await act(async () => {
        loginResult = await result.current.login(
          'test@example.com',
          'password123'
        );
      });

      expect(loginResult.success).toBe(true);
      expect(result.current.user).toEqual(mockUser);
      expect(result.current.isAuthenticated).toBe(true);
      expect(toast.success).toHaveBeenCalledWith('Login realizado com sucesso!');
    });

    it('deve mostrar erro se login falhar', async () => {
      authService.getToken.mockReturnValue(null);
      authService.login.mockResolvedValue({
        success: false,
        error: 'Credenciais inválidas',
      });

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      let loginResult;
      await act(async () => {
        loginResult = await result.current.login(
          'wrong@example.com',
          'wrongpassword'
        );
      });

      expect(loginResult.success).toBe(false);
      expect(loginResult.error).toBe('Credenciais inválidas');
      expect(toast.error).toHaveBeenCalledWith('Credenciais inválidas');
      expect(result.current.user).toBeNull();
      expect(result.current.isAuthenticated).toBe(false);
    });

    it('deve tratar exceção durante login', async () => {
      authService.getToken.mockReturnValue(null);
      authService.login.mockRejectedValue(new Error('Network error'));

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      let loginResult;
      await act(async () => {
        loginResult = await result.current.login('test@example.com', 'password');
      });

      expect(loginResult.success).toBe(false);
      expect(toast.error).toHaveBeenCalled();
    });
  });

  describe('register', () => {
    it('deve registrar usuário com sucesso', async () => {
      authService.getToken.mockReturnValue(null);
      authService.register.mockResolvedValue({
        success: true,
        data: { id: 'new-user-123' },
      });

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      let registerResult;
      await act(async () => {
        registerResult = await result.current.register(
          'newuser@example.com',
          'password123',
          'New User',
          '11999999999'
        );
      });

      expect(registerResult.success).toBe(true);
      expect(toast.success).toHaveBeenCalledWith(
        'Cadastro realizado com sucesso! Faça login para continuar.'
      );
    });

    it('deve mostrar erro se registro falhar', async () => {
      authService.getToken.mockReturnValue(null);
      authService.register.mockResolvedValue({
        success: false,
        error: 'Email já cadastrado',
      });

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      let registerResult;
      await act(async () => {
        registerResult = await result.current.register(
          'existing@example.com',
          'password123',
          'User',
          '11999999999'
        );
      });

      expect(registerResult.success).toBe(false);
      expect(toast.error).toHaveBeenCalledWith('Email já cadastrado');
    });
  });

  describe('logout', () => {
    it('deve limpar estado ao fazer logout', async () => {
      const mockUser = {
        id: 'user-123',
        email: 'test@example.com',
      };

      authService.getToken.mockReturnValue('mock-token');
      authService.getProfile.mockResolvedValue({
        success: true,
        data: mockUser,
      });

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.user).toEqual(mockUser);
      });

      act(() => {
        result.current.logout();
      });

      expect(result.current.user).toBeNull();
      expect(result.current.isAuthenticated).toBe(false);
      expect(authService.logout).toHaveBeenCalled();
      expect(toast.success).toHaveBeenCalledWith('Logout realizado com sucesso!');
    });
  });

  describe('updateProfile', () => {
    it('deve atualizar perfil com sucesso', async () => {
      const mockUser = {
        id: 'user-123',
        email: 'test@example.com',
        name: 'Test User',
      };

      const updatedUser = {
        ...mockUser,
        name: 'Updated Name',
        phone: '11999999999',
      };

      authService.getToken.mockReturnValue('mock-token');
      authService.getProfile.mockResolvedValue({
        success: true,
        data: mockUser,
      });
      authService.updateProfile.mockResolvedValue({
        success: true,
        data: updatedUser,
      });

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.user).toEqual(mockUser);
      });

      let updateResult;
      await act(async () => {
        updateResult = await result.current.updateProfile(
          'Updated Name',
          '11999999999'
        );
      });

      expect(updateResult.success).toBe(true);
      expect(result.current.user).toEqual(updatedUser);
      expect(toast.success).toHaveBeenCalledWith('Perfil atualizado com sucesso!');
    });

    it('deve mostrar erro se atualização falhar', async () => {
      authService.getToken.mockReturnValue('mock-token');
      authService.getProfile.mockResolvedValue({
        success: true,
        data: { id: 'user-123' },
      });
      authService.updateProfile.mockResolvedValue({
        success: false,
        error: 'Erro ao atualizar',
      });

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      let updateResult;
      await act(async () => {
        updateResult = await result.current.updateProfile('Name', 'Phone');
      });

      expect(updateResult.success).toBe(false);
      expect(toast.error).toHaveBeenCalledWith('Erro ao atualizar');
    });
  });

  describe('refreshUser', () => {
    it('deve recarregar dados do usuário', async () => {
      const initialUser = {
        id: 'user-123',
        name: 'Initial Name',
      };

      const refreshedUser = {
        id: 'user-123',
        name: 'Refreshed Name',
      };

      authService.getToken.mockReturnValue('mock-token');
      authService.getProfile
        .mockResolvedValueOnce({
          success: true,
          data: initialUser,
        })
        .mockResolvedValueOnce({
          success: true,
          data: refreshedUser,
        });

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.user).toEqual(initialUser);
      });

      let refreshResult;
      await act(async () => {
        refreshResult = await result.current.refreshUser();
      });

      expect(refreshResult.success).toBe(true);
      expect(result.current.user).toEqual(refreshedUser);
    });
  });

  describe('useAuth Hook', () => {
    it('deve lançar erro se usado fora do AuthProvider', () => {
      // Suprimir console.error para este teste
      const consoleError = vi
        .spyOn(console, 'error')
        .mockImplementation(() => {});

      expect(() => {
        renderHook(() => useAuth());
      }).toThrow('useAuth deve ser usado dentro de um AuthProvider');

      consoleError.mockRestore();
    });
  });
});
