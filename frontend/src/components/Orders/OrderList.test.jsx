import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { BrowserRouter } from 'react-router-dom';
import OrderList from './OrderList';
import orderService from '../../services/orderService';

// Mock do orderService
vi.mock('../../services/orderService', () => ({
  default: {
    getOrders: vi.fn(),
  },
}));

// Mock dos helpers
vi.mock('../../utils/orderHelpers', () => ({
  formatDate: vi.fn((date) => '15/01/2024 10:00'),
  formatCurrency: vi.fn((value) => `R$ ${value.toFixed(2)}`),
  ORDER_STATUSES: [
    { value: '', label: 'Todos' },
    { value: 'pending', label: 'Pendente' },
    { value: 'confirmed', label: 'Confirmado' },
    { value: 'delivered', label: 'Entregue' },
  ],
}));

// Componente wrapper
function Wrapper({ children }) {
  return <BrowserRouter>{children}</BrowserRouter>;
}

describe('OrderList Component', () => {
  const mockOrders = [
    {
      id: 'order-1',
      user_id: 'user-123',
      items: [
        { product_id: 'prod-1', product_name: 'Product 1', quantity: 2, price: 50 },
      ],
      total_amount: 100,
      status: 'pending',
      created_at: '2024-01-15T10:00:00Z',
    },
    {
      id: 'order-2',
      user_id: 'user-123',
      items: [
        { product_id: 'prod-2', product_name: 'Product 2', quantity: 1, price: 75 },
      ],
      total_amount: 75,
      status: 'delivered',
      created_at: '2024-01-14T10:00:00Z',
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Loading state', () => {
    it('deve mostrar loading state inicialmente', () => {
      orderService.getOrders.mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      render(<OrderList />, { wrapper: Wrapper });

      // Skeleton loading deve estar visível
      const skeletons = screen.getAllByRole('generic');
      expect(skeletons.length).toBeGreaterThan(0);
    });
  });

  describe('Renderização de pedidos', () => {
    it('deve renderizar lista de pedidos após carregamento', async () => {
      orderService.getOrders.mockResolvedValue({
        success: true,
        data: mockOrders,
      });

      render(<OrderList />, { wrapper: Wrapper });

      await waitFor(() => {
        expect(screen.getByText(/meus pedidos/i)).toBeInTheDocument();
      });

      // Verificar se os pedidos foram renderizados
      expect(screen.getByText(/order-1/i)).toBeInTheDocument();
      expect(screen.getByText(/order-2/i)).toBeInTheDocument();
    });

    it('deve mostrar quantidade correta de pedidos', async () => {
      orderService.getOrders.mockResolvedValue({
        success: true,
        data: mockOrders,
      });

      render(<OrderList />, { wrapper: Wrapper });

      await waitFor(() => {
        expect(screen.getByText(/2 pedidos encontrados/i)).toBeInTheDocument();
      });
    });

    it('deve exibir informações corretas de cada pedido', async () => {
      orderService.getOrders.mockResolvedValue({
        success: true,
        data: mockOrders,
      });

      render(<OrderList />, { wrapper: Wrapper });

      await waitFor(() => {
        expect(screen.getByText(/R\$ 100\.00/)).toBeInTheDocument();
        expect(screen.getByText(/R\$ 75\.00/)).toBeInTheDocument();
      });

      // Verificar quantidade de itens
      const itemTexts = screen.getAllByText(/\d+ ite(m|ns)/i);
      expect(itemTexts.length).toBeGreaterThan(0);
    });

    it('deve renderizar links "Ver Detalhes" para cada pedido', async () => {
      orderService.getOrders.mockResolvedValue({
        success: true,
        data: mockOrders,
      });

      render(<OrderList />, { wrapper: Wrapper });

      await waitFor(() => {
        const detailLinks = screen.getAllByRole('link', { name: /ver detalhes/i });
        expect(detailLinks).toHaveLength(2);
        expect(detailLinks[0]).toHaveAttribute('href', '/orders/order-1');
        expect(detailLinks[1]).toHaveAttribute('href', '/orders/order-2');
      });
    });
  });

  describe('Empty state', () => {
    it('deve mostrar empty state quando não há pedidos', async () => {
      orderService.getOrders.mockResolvedValue({
        success: true,
        data: [],
      });

      render(<OrderList />, { wrapper: Wrapper });

      await waitFor(() => {
        expect(
          screen.getByText(/nenhum pedido encontrado/i)
        ).toBeInTheDocument();
        expect(
          screen.getByText(/você ainda não fez nenhum pedido/i)
        ).toBeInTheDocument();
      });
    });

    it('deve mostrar link para fazer primeiro pedido no empty state', async () => {
      orderService.getOrders.mockResolvedValue({
        success: true,
        data: [],
      });

      render(<OrderList />, { wrapper: Wrapper });

      await waitFor(() => {
        const newOrderLink = screen.getByRole('link', {
          name: /fazer primeiro pedido/i,
        });
        expect(newOrderLink).toBeInTheDocument();
        expect(newOrderLink).toHaveAttribute('href', '/orders/new');
      });
    });
  });

  describe('Error state', () => {
    it('deve mostrar erro quando falha ao carregar pedidos', async () => {
      orderService.getOrders.mockResolvedValue({
        success: false,
        error: 'Erro ao buscar pedidos',
      });

      render(<OrderList />, { wrapper: Wrapper });

      await waitFor(() => {
        expect(
          screen.getByText(/erro ao carregar pedidos/i)
        ).toBeInTheDocument();
        expect(screen.getByText(/erro ao buscar pedidos/i)).toBeInTheDocument();
      });
    });

    it('deve ter botão "Tentar Novamente" no estado de erro', async () => {
      orderService.getOrders.mockResolvedValue({
        success: false,
        error: 'Erro de conexão',
      });

      render(<OrderList />, { wrapper: Wrapper });

      await waitFor(() => {
        expect(
          screen.getByRole('button', { name: /tentar novamente/i })
        ).toBeInTheDocument();
      });
    });

    it('deve recarregar pedidos ao clicar em "Tentar Novamente"', async () => {
      const user = userEvent.setup();

      // Primeira chamada falha
      orderService.getOrders
        .mockResolvedValueOnce({
          success: false,
          error: 'Erro de conexão',
        })
        // Segunda chamada (retry) funciona
        .mockResolvedValueOnce({
          success: true,
          data: mockOrders,
        });

      render(<OrderList />, { wrapper: Wrapper });

      await waitFor(() => {
        expect(screen.getByText(/erro de conexão/i)).toBeInTheDocument();
      });

      const retryButton = screen.getByRole('button', { name: /tentar novamente/i });
      await user.click(retryButton);

      await waitFor(() => {
        expect(screen.getByText(/meus pedidos/i)).toBeInTheDocument();
      });
    });
  });

  describe('Filtro de status', () => {
    it('deve renderizar select de filtro de status', async () => {
      orderService.getOrders.mockResolvedValue({
        success: true,
        data: mockOrders,
      });

      render(<OrderList />, { wrapper: Wrapper });

      await waitFor(() => {
        expect(screen.getByLabelText(/filtrar por/i)).toBeInTheDocument();
      });
    });

    it('deve chamar getOrders com filtro ao mudar status', async () => {
      const user = userEvent.setup();

      orderService.getOrders.mockResolvedValue({
        success: true,
        data: mockOrders,
      });

      render(<OrderList />, { wrapper: Wrapper });

      await waitFor(() => {
        expect(screen.getByLabelText(/filtrar por/i)).toBeInTheDocument();
      });

      const select = screen.getByLabelText(/filtrar por/i);
      await user.selectOptions(select, 'pending');

      await waitFor(() => {
        expect(orderService.getOrders).toHaveBeenCalledWith(
          expect.objectContaining({
            status: 'pending',
            page: 1,
          })
        );
      });
    });

    it('deve resetar para página 1 ao aplicar filtro', async () => {
      const user = userEvent.setup();

      orderService.getOrders.mockResolvedValue({
        success: true,
        data: mockOrders,
      });

      render(<OrderList />, { wrapper: Wrapper });

      await waitFor(() => {
        expect(screen.getByLabelText(/filtrar por/i)).toBeInTheDocument();
      });

      const select = screen.getByLabelText(/filtrar por/i);
      await user.selectOptions(select, 'delivered');

      await waitFor(() => {
        expect(orderService.getOrders).toHaveBeenLastCalledWith(
          expect.objectContaining({
            page: 1,
          })
        );
      });
    });
  });

  describe('Paginação', () => {
    it('deve mostrar controles de paginação quando totalPages > 1', async () => {
      orderService.getOrders.mockResolvedValue({
        success: true,
        data: {
          orders: mockOrders,
          totalPages: 3,
        },
      });

      render(<OrderList />, { wrapper: Wrapper });

      await waitFor(() => {
        expect(screen.getByText(/página 1 de 3/i)).toBeInTheDocument();
        expect(
          screen.getByRole('button', { name: /anterior/i })
        ).toBeInTheDocument();
        expect(
          screen.getByRole('button', { name: /próximo/i })
        ).toBeInTheDocument();
      });
    });

    it('botão "Anterior" deve estar desabilitado na primeira página', async () => {
      orderService.getOrders.mockResolvedValue({
        success: true,
        data: {
          orders: mockOrders,
          totalPages: 3,
        },
      });

      render(<OrderList />, { wrapper: Wrapper });

      await waitFor(() => {
        const prevButton = screen.getByRole('button', { name: /anterior/i });
        expect(prevButton).toBeDisabled();
      });
    });

    it('botão "Próximo" deve estar desabilitado na última página', async () => {
      orderService.getOrders.mockResolvedValue({
        success: true,
        data: {
          orders: mockOrders,
          totalPages: 1,
        },
      });

      render(<OrderList />, { wrapper: Wrapper });

      await waitFor(() => {
        const nextButton = screen.getByRole('button', { name: /próximo/i });
        expect(nextButton).toBeDisabled();
      });
    });

    it('deve avançar para próxima página ao clicar em "Próximo"', async () => {
      const user = userEvent.setup();

      orderService.getOrders.mockResolvedValue({
        success: true,
        data: {
          orders: mockOrders,
          totalPages: 3,
        },
      });

      render(<OrderList />, { wrapper: Wrapper });

      await waitFor(() => {
        expect(screen.getByText(/página 1 de 3/i)).toBeInTheDocument();
      });

      const nextButton = screen.getByRole('button', { name: /próximo/i });
      await user.click(nextButton);

      await waitFor(() => {
        expect(orderService.getOrders).toHaveBeenCalledWith(
          expect.objectContaining({
            page: 2,
          })
        );
      });
    });

    it('deve voltar para página anterior ao clicar em "Anterior"', async () => {
      const user = userEvent.setup();

      orderService.getOrders.mockResolvedValue({
        success: true,
        data: {
          orders: mockOrders,
          totalPages: 3,
        },
      });

      render(<OrderList />, { wrapper: Wrapper });

      await waitFor(() => {
        expect(screen.getByText(/página 1 de 3/i)).toBeInTheDocument();
      });

      // Avançar para página 2
      const nextButton = screen.getByRole('button', { name: /próximo/i });
      await user.click(nextButton);

      await waitFor(() => {
        expect(screen.getByText(/página 2 de 3/i)).toBeInTheDocument();
      });

      // Voltar para página 1
      const prevButton = screen.getByRole('button', { name: /anterior/i });
      await user.click(prevButton);

      await waitFor(() => {
        expect(orderService.getOrders).toHaveBeenCalledWith(
          expect.objectContaining({
            page: 1,
          })
        );
      });
    });

    it('não deve mostrar paginação quando totalPages <= 1', async () => {
      orderService.getOrders.mockResolvedValue({
        success: true,
        data: mockOrders,
      });

      render(<OrderList />, { wrapper: Wrapper });

      await waitFor(() => {
        expect(screen.getByText(/meus pedidos/i)).toBeInTheDocument();
      });

      expect(screen.queryByText(/página/i)).not.toBeInTheDocument();
    });
  });

  describe('Chamadas da API', () => {
    it('deve chamar getOrders ao montar o componente', async () => {
      orderService.getOrders.mockResolvedValue({
        success: true,
        data: mockOrders,
      });

      render(<OrderList />, { wrapper: Wrapper });

      await waitFor(() => {
        expect(orderService.getOrders).toHaveBeenCalledWith({
          page: 1,
          limit: 9,
        });
      });
    });

    it('deve recarregar pedidos ao mudar página', async () => {
      const user = userEvent.setup();

      orderService.getOrders.mockResolvedValue({
        success: true,
        data: {
          orders: mockOrders,
          totalPages: 2,
        },
      });

      render(<OrderList />, { wrapper: Wrapper });

      await waitFor(() => {
        expect(orderService.getOrders).toHaveBeenCalledTimes(1);
      });

      const nextButton = screen.getByRole('button', { name: /próximo/i });
      await user.click(nextButton);

      await waitFor(() => {
        expect(orderService.getOrders).toHaveBeenCalledTimes(2);
      });
    });
  });
});
