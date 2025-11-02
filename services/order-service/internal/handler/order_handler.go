package handler

import (
	"log"

	"pedidos-online/order-service/internal/middleware"
	"pedidos-online/order-service/internal/model"
	"pedidos-online/order-service/internal/service"

	"github.com/gofiber/fiber/v2"
)

// OrderHandler gerencia as requisições HTTP para pedidos
type OrderHandler struct {
	service service.OrderService
}

// NewOrderHandler cria uma nova instância do handler
func NewOrderHandler(svc service.OrderService) *OrderHandler {
	return &OrderHandler{
		service: svc,
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
			"error":   "Dados inválidos: " + err.Error(),
		})
	}

	// Criar pedido usando service layer
	order, err := h.service.CreateOrder(c.Context(), userID, req.Items, req.Address)
	if err != nil {
		log.Printf("❌ Erro ao criar pedido: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    order,
		"message": "Pedido criado com sucesso",
	})
}

// GetOrders retorna todos os pedidos do usuário autenticado
// GET /api/v1/orders?page=1&page_size=10
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
	pageSize := c.QueryInt("page_size", 10)

	// Listar pedidos usando service layer
	orders, total, err := h.service.ListOrders(c.Context(), userID, page, pageSize)
	if err != nil {
		log.Printf("❌ Erro ao listar pedidos: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    orders,
		"pagination": fiber.Map{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
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

	// Buscar pedido usando service layer
	order, err := h.service.GetOrder(c.Context(), orderID, userID)
	if err != nil {
		statusCode := fiber.StatusBadRequest
		if err.Error() == "pedido não encontrado" {
			statusCode = fiber.StatusNotFound
		} else if err.Error() == "acesso negado a este pedido" {
			statusCode = fiber.StatusForbidden
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
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
			"error":   "Dados inválidos: " + err.Error(),
		})
	}

	// Atualizar status usando service layer
	_, err := h.service.UpdateOrderStatus(c.Context(), orderID, req.Status)
	if err != nil {
		log.Printf("❌ Erro ao atualizar status: %v", err)
		statusCode := fiber.StatusBadRequest
		if err.Error() == "pedido não encontrado" {
			statusCode = fiber.StatusNotFound
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Status atualizado com sucesso",
	})
}

// CancelOrder cancela um pedido
// PUT /api/v1/orders/:id/cancel
func (h *OrderHandler) CancelOrder(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Usuário não autenticado",
		})
	}

	// Parse do ID
	orderID := c.Params("id")

	// Cancelar pedido usando service layer
	if err := h.service.CancelOrder(c.Context(), orderID, userID); err != nil {
		log.Printf("❌ Erro ao cancelar pedido: %v", err)
		statusCode := fiber.StatusBadRequest
		if err.Error() == "pedido não encontrado" {
			statusCode = fiber.StatusNotFound
		} else if err.Error() == "acesso negado a este pedido" {
			statusCode = fiber.StatusForbidden
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Pedido cancelado com sucesso",
	})
}
