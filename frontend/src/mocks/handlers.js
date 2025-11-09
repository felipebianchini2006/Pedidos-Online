import { http, HttpResponse } from 'msw';

const BASE_URL = '/api';

export const handlers = [
  // ============ USER SERVICE HANDLERS ============

  // POST /api/users/register
  http.post(`${BASE_URL}/users/register`, async ({ request }) => {
    const body = await request.json();
    const { email, password, name, phone } = body;

    // Simular erro de validação
    if (!email || !password || !name) {
      return HttpResponse.json(
        { success: false, error: 'Todos os campos são obrigatórios' },
        { status: 400 }
      );
    }

    // Simular email já cadastrado
    if (email === 'existing@example.com') {
      return HttpResponse.json(
        { success: false, error: 'Email já cadastrado' },
        { status: 409 }
      );
    }

    // Sucesso
    return HttpResponse.json(
      {
        success: true,
        data: {
          id: 'user-123',
          email,
          name,
          phone,
          created_at: new Date().toISOString(),
        },
      },
      { status: 201 }
    );
  }),

  // POST /api/users/login
  http.post(`${BASE_URL}/users/login`, async ({ request }) => {
    const body = await request.json();
    const { email, password } = body;

    // Simular credenciais inválidas
    if (email !== 'test@example.com' || password !== 'password123') {
      return HttpResponse.json(
        { success: false, error: 'Credenciais inválidas' },
        { status: 401 }
      );
    }

    // Sucesso
    return HttpResponse.json({
      success: true,
      data: {
        token: 'mock-jwt-token-123',
        user: {
          id: 'user-123',
          email: 'test@example.com',
          name: 'Test User',
          phone: '(11) 99999-9999',
        },
      },
    });
  }),

  // GET /api/users/profile
  http.get(`${BASE_URL}/users/profile`, ({ request }) => {
    const authHeader = request.headers.get('Authorization');

    // Verificar token
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return HttpResponse.json(
        { success: false, error: 'Token não fornecido' },
        { status: 401 }
      );
    }

    const token = authHeader.replace('Bearer ', '');
    if (token !== 'mock-jwt-token-123') {
      return HttpResponse.json(
        { success: false, error: 'Token inválido' },
        { status: 401 }
      );
    }

    // Sucesso
    return HttpResponse.json({
      success: true,
      data: {
        id: 'user-123',
        email: 'test@example.com',
        name: 'Test User',
        phone: '(11) 99999-9999',
        created_at: '2024-01-15T10:00:00Z',
      },
    });
  }),

  // PUT /api/users/profile
  http.put(`${BASE_URL}/users/profile`, async ({ request }) => {
    const authHeader = request.headers.get('Authorization');

    // Verificar token
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return HttpResponse.json(
        { success: false, error: 'Token não fornecido' },
        { status: 401 }
      );
    }

    const body = await request.json();
    const { name, phone } = body;

    // Sucesso
    return HttpResponse.json({
      success: true,
      data: {
        id: 'user-123',
        email: 'test@example.com',
        name: name || 'Test User',
        phone: phone || '(11) 99999-9999',
        updated_at: new Date().toISOString(),
      },
    });
  }),

  // ============ ORDER SERVICE HANDLERS ============

  // GET /api/orders
  http.get(`${BASE_URL}/orders`, ({ request }) => {
    const authHeader = request.headers.get('Authorization');

    // Verificar token
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return HttpResponse.json(
        { success: false, error: 'Token não fornecido' },
        { status: 401 }
      );
    }

    // Mock de pedidos
    const orders = [
      {
        id: 'order-1',
        user_id: 'user-123',
        items: [
          {
            product_id: 'prod-1',
            product_name: 'Product 1',
            quantity: 2,
            price: 50.0,
          },
        ],
        total_amount: 100.0,
        status: 'pending',
        address: {
          street: 'Rua Teste',
          number: '123',
          city: 'São Paulo',
          state: 'SP',
          zip_code: '01234-567',
        },
        created_at: '2024-01-15T10:00:00Z',
      },
      {
        id: 'order-2',
        user_id: 'user-123',
        items: [
          {
            product_id: 'prod-2',
            product_name: 'Product 2',
            quantity: 1,
            price: 75.0,
          },
        ],
        total_amount: 75.0,
        status: 'delivered',
        address: {
          street: 'Rua Teste',
          number: '123',
          city: 'São Paulo',
          state: 'SP',
          zip_code: '01234-567',
        },
        created_at: '2024-01-14T10:00:00Z',
      },
    ];

    return HttpResponse.json({
      success: true,
      data: orders,
    });
  }),

  // GET /api/orders/:id
  http.get(`${BASE_URL}/orders/:id`, ({ params, request }) => {
    const authHeader = request.headers.get('Authorization');

    // Verificar token
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return HttpResponse.json(
        { success: false, error: 'Token não fornecido' },
        { status: 401 }
      );
    }

    const { id } = params;

    // Simular pedido não encontrado
    if (id === 'invalid-id') {
      return HttpResponse.json(
        { success: false, error: 'Pedido não encontrado' },
        { status: 404 }
      );
    }

    // Mock de pedido
    const order = {
      id: id,
      user_id: 'user-123',
      items: [
        {
          product_id: 'prod-1',
          product_name: 'Product 1',
          quantity: 2,
          price: 50.0,
        },
      ],
      total_amount: 100.0,
      status: 'pending',
      address: {
        street: 'Rua Teste',
        number: '123',
        city: 'São Paulo',
        state: 'SP',
        zip_code: '01234-567',
        complement: 'Apt 101',
      },
      created_at: '2024-01-15T10:00:00Z',
      updated_at: '2024-01-15T10:00:00Z',
    };

    return HttpResponse.json({
      success: true,
      data: order,
    });
  }),

  // POST /api/orders
  http.post(`${BASE_URL}/orders`, async ({ request }) => {
    const authHeader = request.headers.get('Authorization');

    // Verificar token
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return HttpResponse.json(
        { success: false, error: 'Token não fornecido' },
        { status: 401 }
      );
    }

    const body = await request.json();
    const { items, address } = body;

    // Validar items
    if (!items || items.length === 0) {
      return HttpResponse.json(
        { success: false, error: 'O pedido deve conter pelo menos um item' },
        { status: 400 }
      );
    }

    // Validar endereço
    if (!address || !address.street || !address.city) {
      return HttpResponse.json(
        { success: false, error: 'Endereço incompleto' },
        { status: 400 }
      );
    }

    // Calcular total
    const total_amount = items.reduce(
      (sum, item) => sum + item.price * item.quantity,
      0
    );

    // Criar pedido
    const newOrder = {
      id: 'order-' + Date.now(),
      user_id: 'user-123',
      items,
      total_amount,
      status: 'pending',
      address,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    return HttpResponse.json(
      {
        success: true,
        data: newOrder,
        message: 'Pedido criado com sucesso',
      },
      { status: 201 }
    );
  }),

  // PUT /api/orders/:id/status
  http.put(`${BASE_URL}/orders/:id/status`, async ({ params, request }) => {
    const authHeader = request.headers.get('Authorization');

    // Verificar token
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return HttpResponse.json(
        { success: false, error: 'Token não fornecido' },
        { status: 401 }
      );
    }

    const { id } = params;
    const body = await request.json();
    const { status } = body;

    const validStatuses = [
      'pending',
      'confirmed',
      'preparing',
      'shipped',
      'delivered',
      'cancelled',
    ];

    if (!validStatuses.includes(status)) {
      return HttpResponse.json(
        { success: false, error: 'Status inválido' },
        { status: 400 }
      );
    }

    return HttpResponse.json({
      success: true,
      data: {
        id,
        status,
        updated_at: new Date().toISOString(),
      },
      message: 'Status do pedido atualizado com sucesso',
    });
  }),
];
