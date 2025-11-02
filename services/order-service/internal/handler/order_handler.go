package handler

import (
	"log"

	"pedidos-online/order-service/internal/middleware"
	"pedidos-online/order-service/internal/model"
	"pedidos-online/order-service/internal/queue"
	"pedidos-online/order-service/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// OrderHandler gerencia as requisições HTTP para pedidos
type OrderHandler struct {
	repo      repository.OrderRepository
	validator *validator.Validate
	publisher *queue.Publisher
}

// NewOrderHandler cria uma nova instância do handler
func NewOrderHandler(repo repository.OrderRepository, publisher *queue.Publisher) *OrderHandler {
	return &OrderHandler{
		repo:      repo,
		validator: validator.New(),
		publisher: publisher,
	}
}

// CreateOrder cria um novo pedido
// POST /api/v1/orders
func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {
	// Extrair user_id do contexto (injetado pelo AuthMiddleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Usuário não autenticado",
		})
	}

	// Parse do body
	var req model.CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Dados inválidos",
		})
	}

	// Validar request
	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	// Criar ordem
	order := &model.Order{
		UserID:  userID,
		Items:   req.Items,
		Status:  model.OrderStatusPending,
		Address: req.Address,
	}

	// Calcular total
	order.CalculateTotal()

	// Salvar no banco
	if err := h.repo.Create(c.Context(), order); err != nil {
		log.Printf("❌ Erro ao criar pedido: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Erro ao criar pedido",
		})
	}

	// Publicar evento order.created no RabbitMQ
	if err := h.publisher.PublishOrderCreated(order); err != nil {
		log.Printf("⚠️  Erro ao publicar evento order.created: %v", err)
		// Não retornar erro, pois o pedido foi criado com sucesso
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    order,
		"message": "Pedido criado com sucesso",
	})
}

// GetOrders retorna todos os pedidos do usuário autenticado
// GET /api/v1/orders?page=1&limit=10
func (h *OrderHandler) GetOrders(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Usuário não autenticado",
		})
	}

	// Parâmetros de paginação
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	// Calcular skip
	skip := (page - 1) * limit

	// Buscar pedidos com paginação
	orders, err := h.repo.FindByUserID(c.Context(), userID, limit, skip)
	if err != nil {
		log.Printf("❌ Erro ao buscar pedidos: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Erro ao buscar pedidos",
		})
	}

	// Contar total de pedidos
	total, err := h.repo.Count(c.Context(), userID)
	if err != nil {
		log.Printf("⚠️  Erro ao contar pedidos: %v", err)
		total = 0 // não falhar por isso
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    orders,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetOrderByID retorna um pedido específico pelo ID
// GET /api/v1/orders/:id
func (h *OrderHandler) GetOrderByID(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Usuário não autenticado",
		})
	}

	// Parse do ID
	orderID := c.Params("id")

	// Buscar pedido
	order, err := h.repo.FindByID(c.Context(), orderID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Pedido não encontrado",
		})
	}

	// Verificar se o pedido pertence ao usuário
	if order.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "Acesso negado a este pedido",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    order,
	})
}

// UpdateOrderStatus atualiza o status de um pedido
// PUT /api/v1/orders/:id/status
func (h *OrderHandler) UpdateOrderStatus(c *fiber.Ctx) error {
	// Parse do ID
	orderID := c.Params("id")

	// Parse do body
	var req model.UpdateOrderStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Dados inválidos",
		})
	}

	// Validar request
	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	// Buscar pedido atual
	order, err := h.repo.FindByID(c.Context(), orderID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Pedido não encontrado",
		})
	}

	// Validar transição de status
	if !order.CanTransitionTo(req.Status) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Transição de status inválida",
			"message": "Não é possível mudar de '" + order.Status + "' para '" + req.Status + "'",
		})
	}

	// Atualizar status
	if err := h.repo.UpdateStatus(c.Context(), orderID, req.Status); err != nil {
		log.Printf("❌ Erro ao atualizar status: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Erro ao atualizar status",
		})
	}

	// Publicar evento order.updated no RabbitMQ
	if err := h.publisher.PublishOrderUpdated(orderID, order.Status, req.Status); err != nil {
		log.Printf("⚠️  Erro ao publicar evento order.updated: %v", err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Status atualizado com sucesso",
	})
}
