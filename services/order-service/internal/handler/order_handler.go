package handler

import (
	"fmt"
	"log"
	"strings"

	"pedidos-online/order-service/internal/middleware"
	"pedidos-online/order-service/internal/model"
	"pedidos-online/order-service/internal/service"

	"github.com/gofiber/fiber/v2"
)

// ============================================================================
// STRUCTS
// ============================================================================

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

// CreateOrderRequest representa o payload para criar um novo pedido
type CreateOrderRequest struct {
	Items   []OrderItem `json:"items" validate:"required,min=1"`
	Address Address     `json:"address" validate:"required"`
}

// OrderItem representa um item no pedido
type OrderItem struct {
	ProductID   string  `json:"product_id" validate:"required"`
	ProductName string  `json:"product_name" validate:"required"`
	Quantity    int     `json:"quantity" validate:"required,min=1"`
	Price       float64 `json:"price" validate:"required,gt=0"`
}

// Address representa o endereço de entrega
type Address struct {
	Street     string `json:"street" validate:"required"`
	Number     string `json:"number" validate:"required"`
	City       string `json:"city" validate:"required"`
	State      string `json:"state" validate:"required,len=2"`
	ZipCode    string `json:"zip_code" validate:"required"`
	Complement string `json:"complement"`
}

// UpdateStatusRequest representa o payload para atualizar status
type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending confirmed preparing shipped delivered cancelled"`
}

// OrderListResponse representa a resposta de listagem de pedidos
type OrderListResponse struct {
	Success    bool           `json:"success"`
	Data       []*model.Order `json:"data"`
	Pagination Pagination     `json:"pagination"`
}

// Pagination representa metadados de paginação
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

// ErrorResponse representa uma resposta de erro padronizada
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse representa uma resposta de sucesso padronizada
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ============================================================================
// HANDLERS
// ============================================================================

// CreateOrder cria um novo pedido
// POST /api/v1/orders
func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {
	// 1. Extrair userID do contexto (injetado pelo AuthMiddleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		log.Printf("❌ Tentativa de criar pedido sem autenticação")
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Success: false,
			Error:   "Usuário não autenticado",
			Message: "Token JWT ausente ou inválido",
		})
	}

	// 2. Parse do body
	var req CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("❌ Erro ao fazer parse do request: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   "Formato de dados inválido",
			Message: "Verifique o JSON enviado",
		})
	}

	// 3. Validações client-side
	if err := validateCreateOrderRequest(&req); err != nil {
		log.Printf("⚠️  Validação client-side falhou: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	// 4. Converter para model.OrderItem e model.Address
	modelItems := make([]model.OrderItem, len(req.Items))
	for i, item := range req.Items {
		modelItems[i] = model.OrderItem{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
		}
	}

	modelAddress := model.Address{
		Street:     req.Address.Street,
		Number:     req.Address.Number,
		City:       req.Address.City,
		State:      strings.ToUpper(req.Address.State),
		ZipCode:    req.Address.ZipCode,
		Complement: req.Address.Complement,
	}

	// 5. Chamar service layer
	order, err := h.service.CreateOrder(c.Context(), userID, modelItems, modelAddress)
	if err != nil {
		log.Printf("❌ Erro ao criar pedido: %v", err)
		// Determinar código de status baseado no erro
		statusCode := fiber.StatusInternalServerError
		if strings.Contains(err.Error(), "inválido") || strings.Contains(err.Error(), "obrigatório") {
			statusCode = fiber.StatusBadRequest
		}
		return c.Status(statusCode).JSON(ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	log.Printf("✅ Pedido %s criado com sucesso para usuário %s", order.ID.Hex(), userID)

	// 6. Retornar 201 Created
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Data:    order,
		Message: "Pedido criado com sucesso",
	})
}

// ListOrders retorna todos os pedidos do usuário autenticado com paginação
// GET /api/v1/orders?page=1&page_size=10
func (h *OrderHandler) ListOrders(c *fiber.Ctx) error {
	// 1. Extrair userID do contexto
	userID := middleware.GetUserID(c)
	if userID == "" {
		log.Printf("❌ Tentativa de listar pedidos sem autenticação")
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Success: false,
			Error:   "Usuário não autenticado",
			Message: "Token JWT ausente ou inválido",
		})
	}

	// 2. Parse query params com defaults
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 10)

	// 3. Validar paginação
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // Máximo 100 itens por página
	}

	log.Printf("📋 Listando pedidos do usuário %s (page=%d, pageSize=%d)", userID, page, pageSize)

	// 4. Chamar service layer
	orders, total, err := h.service.ListOrders(c.Context(), userID, page, pageSize)
	if err != nil {
		log.Printf("❌ Erro ao listar pedidos: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Success: false,
			Error:   "Erro ao buscar pedidos",
			Message: err.Error(),
		})
	}

	// 5. Calcular total de páginas
	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + int64(pageSize) - 1) / int64(pageSize)
	}

	log.Printf("✅ %d pedidos encontrados (total: %d)", len(orders), total)

	// 6. Retornar 200 OK com lista paginada
	return c.Status(fiber.StatusOK).JSON(OrderListResponse{
		Success: true,
		Data:    orders,
		Pagination: Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// GetOrder retorna um pedido específico pelo ID
// GET /api/v1/orders/:id
func (h *OrderHandler) GetOrder(c *fiber.Ctx) error {
	// 1. Extrair userID do contexto
	userID := middleware.GetUserID(c)
	if userID == "" {
		log.Printf("❌ Tentativa de buscar pedido sem autenticação")
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Success: false,
			Error:   "Usuário não autenticado",
			Message: "Token JWT ausente ou inválido",
		})
	}

	// 2. Extrair orderID da URL
	orderID := c.Params("id")
	if orderID == "" {
		log.Printf("⚠️  ID do pedido não fornecido")
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   "ID do pedido é obrigatório",
		})
	}

	log.Printf("🔍 Buscando pedido %s para usuário %s", orderID, userID)

	// 3. Chamar service layer
	order, err := h.service.GetOrder(c.Context(), orderID, userID)
	if err != nil {
		// Determinar código de status baseado no erro
		statusCode := fiber.StatusInternalServerError
		errorMsg := err.Error()

		if strings.Contains(errorMsg, "não encontrado") {
			statusCode = fiber.StatusNotFound
			log.Printf("⚠️  Pedido %s não encontrado", orderID)
		} else if strings.Contains(errorMsg, "acesso negado") {
			statusCode = fiber.StatusForbidden
			log.Printf("🚫 Acesso negado ao pedido %s para usuário %s", orderID, userID)
		} else {
			log.Printf("❌ Erro ao buscar pedido %s: %v", orderID, err)
		}

		return c.Status(statusCode).JSON(ErrorResponse{
			Success: false,
			Error:   errorMsg,
		})
	}

	log.Printf("✅ Pedido %s encontrado", orderID)

	// 4. Retornar 200 OK com pedido
	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Success: true,
		Data:    order,
	})
}

// UpdateStatus atualiza o status de um pedido
// PUT /api/v1/orders/:id/status
// NOTA: Este endpoint pode ser protegido apenas para admins
func (h *OrderHandler) UpdateStatus(c *fiber.Ctx) error {
	// 1. Extrair orderID da URL
	orderID := c.Params("id")
	if orderID == "" {
		log.Printf("⚠️  ID do pedido não fornecido")
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   "ID do pedido é obrigatório",
		})
	}

	// 2. Parse do body
	var req UpdateStatusRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("❌ Erro ao fazer parse do request: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   "Formato de dados inválido",
			Message: "Verifique o JSON enviado",
		})
	}

	// 3. Validar novo status
	req.Status = strings.ToLower(strings.TrimSpace(req.Status))
	if req.Status == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   "Status é obrigatório",
		})
	}

	// Validar se status é um dos permitidos
	validStatuses := []string{"pending", "confirmed", "preparing", "shipped", "delivered", "cancelled"}
	isValid := false
	for _, validStatus := range validStatuses {
		if req.Status == validStatus {
			isValid = true
			break
		}
	}
	if !isValid {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("Status inválido: %s", req.Status),
			Message: "Status permitidos: pending, confirmed, preparing, shipped, delivered, cancelled",
		})
	}

	log.Printf("🔄 Atualizando status do pedido %s para %s", orderID, req.Status)

	// 4. Chamar service layer
	updatedOrder, err := h.service.UpdateOrderStatus(c.Context(), orderID, req.Status)
	if err != nil {
		// Determinar código de status baseado no erro
		statusCode := fiber.StatusInternalServerError
		errorMsg := err.Error()

		if strings.Contains(errorMsg, "não encontrado") {
			statusCode = fiber.StatusNotFound
			log.Printf("⚠️  Pedido %s não encontrado", orderID)
		} else if strings.Contains(errorMsg, "transição") || strings.Contains(errorMsg, "inválida") {
			statusCode = fiber.StatusBadRequest
			log.Printf("🚫 Transição de status inválida para pedido %s: %v", orderID, err)
		} else {
			log.Printf("❌ Erro ao atualizar status do pedido %s: %v", orderID, err)
		}

		return c.Status(statusCode).JSON(ErrorResponse{
			Success: false,
			Error:   errorMsg,
		})
	}

	log.Printf("✅ Status do pedido %s atualizado para %s", orderID, req.Status)

	// 5. Retornar 200 OK com pedido atualizado
	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Success: true,
		Data:    updatedOrder,
		Message: "Status atualizado com sucesso",
	})
}

// CancelOrder cancela um pedido (apenas o dono pode cancelar)
// DELETE /api/v1/orders/:id
func (h *OrderHandler) CancelOrder(c *fiber.Ctx) error {
	// 1. Extrair userID do contexto
	userID := middleware.GetUserID(c)
	if userID == "" {
		log.Printf("❌ Tentativa de cancelar pedido sem autenticação")
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Success: false,
			Error:   "Usuário não autenticado",
			Message: "Token JWT ausente ou inválido",
		})
	}

	// 2. Extrair orderID da URL
	orderID := c.Params("id")
	if orderID == "" {
		log.Printf("⚠️  ID do pedido não fornecido")
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   "ID do pedido é obrigatório",
		})
	}

	log.Printf("🚫 Cancelando pedido %s para usuário %s", orderID, userID)

	// 3. Chamar service layer
	err := h.service.CancelOrder(c.Context(), orderID, userID)
	if err != nil {
		// Determinar código de status baseado no erro
		statusCode := fiber.StatusInternalServerError
		errorMsg := err.Error()

		if strings.Contains(errorMsg, "não encontrado") {
			statusCode = fiber.StatusNotFound
			log.Printf("⚠️  Pedido %s não encontrado", orderID)
		} else if strings.Contains(errorMsg, "acesso negado") {
			statusCode = fiber.StatusForbidden
			log.Printf("🚫 Acesso negado ao cancelar pedido %s para usuário %s", orderID, userID)
		} else if strings.Contains(errorMsg, "não é possível cancelar") || strings.Contains(errorMsg, "status") {
			statusCode = fiber.StatusConflict // 409 - não pode ser cancelado
			log.Printf("⚠️  Pedido %s não pode ser cancelado: %v", orderID, err)
		} else {
			log.Printf("❌ Erro ao cancelar pedido %s: %v", orderID, err)
		}

		return c.Status(statusCode).JSON(ErrorResponse{
			Success: false,
			Error:   errorMsg,
		})
	}

	log.Printf("✅ Pedido %s cancelado com sucesso", orderID)

	// 4. Retornar 204 No Content (pedido cancelado com sucesso)
	return c.SendStatus(fiber.StatusNoContent)
}

// ============================================================================
// VALIDAÇÕES CLIENT-SIDE
// ============================================================================

// validateCreateOrderRequest valida o request de criação de pedido
func validateCreateOrderRequest(req *CreateOrderRequest) error {
	// Validar items
	if len(req.Items) == 0 {
		return fmt.Errorf("ao menos um item é obrigatório")
	}

	if len(req.Items) > 50 {
		return fmt.Errorf("máximo de 50 itens por pedido (enviado: %d)", len(req.Items))
	}

	// Validar cada item
	for i, item := range req.Items {
		// Validar product_id
		if strings.TrimSpace(item.ProductID) == "" {
			return fmt.Errorf("item %d: product_id é obrigatório", i+1)
		}

		// Validar product_name
		if strings.TrimSpace(item.ProductName) == "" {
			return fmt.Errorf("item %d: product_name é obrigatório", i+1)
		}

		// Validar quantity
		if item.Quantity <= 0 {
			return fmt.Errorf("item %d (%s): quantidade deve ser maior que 0 (recebido: %d)", i+1, item.ProductName, item.Quantity)
		}

		if item.Quantity > 1000 {
			return fmt.Errorf("item %d (%s): quantidade máxima é 1000 (recebido: %d)", i+1, item.ProductName, item.Quantity)
		}

		// Validar price
		if item.Price <= 0 {
			return fmt.Errorf("item %d (%s): preço deve ser maior que 0 (recebido: %.2f)", i+1, item.ProductName, item.Price)
		}

		if item.Price > 1000000 {
			return fmt.Errorf("item %d (%s): preço máximo é R$ 1.000.000,00 (recebido: R$ %.2f)", i+1, item.ProductName, item.Price)
		}
	}

	// Validar address
	if err := validateAddress(&req.Address); err != nil {
		return fmt.Errorf("endereço inválido: %w", err)
	}

	return nil
}

// validateAddress valida os campos obrigatórios do endereço
func validateAddress(addr *Address) error {
	if strings.TrimSpace(addr.Street) == "" {
		return fmt.Errorf("rua é obrigatória")
	}

	if strings.TrimSpace(addr.Number) == "" {
		return fmt.Errorf("número é obrigatório")
	}

	if strings.TrimSpace(addr.City) == "" {
		return fmt.Errorf("cidade é obrigatória")
	}

	if strings.TrimSpace(addr.State) == "" {
		return fmt.Errorf("estado é obrigatório")
	}

	// Validar formato do estado (2 caracteres)
	addr.State = strings.ToUpper(strings.TrimSpace(addr.State))
	if len(addr.State) != 2 {
		return fmt.Errorf("estado deve ter 2 caracteres (ex: SP, RJ)")
	}

	if strings.TrimSpace(addr.ZipCode) == "" {
		return fmt.Errorf("CEP é obrigatório")
	}

	// Remover caracteres não numéricos do CEP
	addr.ZipCode = strings.ReplaceAll(addr.ZipCode, "-", "")
	addr.ZipCode = strings.ReplaceAll(addr.ZipCode, ".", "")
	addr.ZipCode = strings.TrimSpace(addr.ZipCode)

	// Validar formato do CEP (8 dígitos)
	if len(addr.ZipCode) != 8 {
		return fmt.Errorf("CEP deve ter 8 dígitos (ex: 01234567)")
	}

	// Validar se CEP contém apenas números
	for _, char := range addr.ZipCode {
		if char < '0' || char > '9' {
			return fmt.Errorf("CEP deve conter apenas números")
		}
	}

	return nil
}

// ============================================================================
// REGISTRO DE ROTAS
// ============================================================================

// RegisterRoutes registra todas as rotas do Order Service
func RegisterRoutes(app *fiber.App, handler *OrderHandler, authMiddleware fiber.Handler) {
	// Criar grupo de rotas /api/v1/orders
	orders := app.Group("/api/v1/orders")

	// Aplicar middleware de autenticação em todas as rotas
	orders.Use(authMiddleware)

	// Rotas
	orders.Post("/", handler.CreateOrder)           // Criar pedido
	orders.Get("/", handler.ListOrders)             // Listar pedidos do usuário
	orders.Get("/:id", handler.GetOrder)            // Obter pedido específico
	orders.Put("/:id/status", handler.UpdateStatus) // Atualizar status (admin)
	orders.Delete("/:id", handler.CancelOrder)      // Cancelar pedido (usuário)

	log.Println("✅ Rotas do Order Service registradas:")
	log.Println("   POST   /api/v1/orders           - Criar pedido")
	log.Println("   GET    /api/v1/orders           - Listar pedidos")
	log.Println("   GET    /api/v1/orders/:id       - Obter pedido")
	log.Println("   PUT    /api/v1/orders/:id/status - Atualizar status")
	log.Println("   DELETE /api/v1/orders/:id       - Cancelar pedido")
}
