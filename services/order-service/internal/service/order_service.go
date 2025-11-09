package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"pedidos-online/order-service/internal/model"
	"pedidos-online/order-service/internal/queue"
	"pedidos-online/order-service/internal/repository"
)

// OrderService define a interface para a camada de serviço de pedidos
type OrderService interface {
	// CreateOrder cria um novo pedido
	CreateOrder(ctx context.Context, userID string, items []model.OrderItem, address model.Address) (*model.Order, error)

	// GetOrder busca um pedido por ID
	GetOrder(ctx context.Context, orderID, userID string) (*model.Order, error)

	// ListOrders lista pedidos de um usuário com paginação
	ListOrders(ctx context.Context, userID string, page, pageSize int) ([]*model.Order, int64, error)

	// ListAllOrders lista TODOS os pedidos (admin) com paginação
	ListAllOrders(ctx context.Context, page, pageSize int) ([]*model.Order, int64, error)

	// UpdateOrderStatus atualiza o status de um pedido
	UpdateOrderStatus(ctx context.Context, orderID, newStatus string) (*model.Order, error)

	// CancelOrder cancela um pedido
	CancelOrder(ctx context.Context, orderID, userID string) error
}

// orderService implementa a interface OrderService
type orderService struct {
	repo      repository.OrderRepository
	publisher queue.EventPublisher
}

// NewOrderService cria uma nova instância do serviço
func NewOrderService(repo repository.OrderRepository, publisher queue.EventPublisher) OrderService {
	return &orderService{
		repo:      repo,
		publisher: publisher,
	}
}

// CreateOrder cria um novo pedido com validações de negócio
func (s *orderService) CreateOrder(ctx context.Context, userID string, items []model.OrderItem, address model.Address) (*model.Order, error) {
	log.Printf("📝 Criando pedido para usuário %s", userID)

	// Validar userID
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("user_id é obrigatório")
	}

	// Validar items
	if err := s.validateItems(items); err != nil {
		return nil, fmt.Errorf("itens inválidos: %w", err)
	}

	// Validar address
	if err := s.validateAddress(address); err != nil {
		return nil, fmt.Errorf("endereço inválido: %w", err)
	}

	// Criar pedido
	order := &model.Order{
		UserID:  userID,
		Items:   items,
		Status:  model.OrderStatusPending,
		Address: address,
	}

	// Calcular total
	order.CalculateTotal()

	// Validar total mínimo
	if order.TotalAmount <= 0 {
		return nil, fmt.Errorf("total do pedido deve ser maior que zero")
	}

	// Salvar no banco
	if err := s.repo.Create(ctx, order); err != nil {
		log.Printf("❌ Erro ao salvar pedido: %v", err)
		return nil, fmt.Errorf("erro ao criar pedido: %w", err)
	}

	log.Printf("✅ Pedido %s criado com sucesso (total: R$ %.2f)", order.ID.Hex(), order.TotalAmount)

	// Publicar evento order.created
	if err := s.publisher.PublishOrderCreated(order); err != nil {
		log.Printf("⚠️  Erro ao publicar evento order.created: %v", err)
		// Não retornar erro, pois o pedido já foi criado
	}

	return order, nil
}

// GetOrder busca um pedido por ID com verificação de ownership
func (s *orderService) GetOrder(ctx context.Context, orderID, userID string) (*model.Order, error) {
	log.Printf("🔍 Buscando pedido %s para usuário %s", orderID, userID)

	// Validar parâmetros
	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("order_id é obrigatório")
	}
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("user_id é obrigatório")
	}

	// Buscar pedido
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		log.Printf("❌ Pedido %s não encontrado: %v", orderID, err)
		return nil, fmt.Errorf("pedido não encontrado")
	}

	// Verificar ownership (segurança)
	if order.UserID != userID {
		log.Printf("🚫 Usuário %s tentou acessar pedido %s de outro usuário", userID, orderID)
		return nil, fmt.Errorf("acesso negado: este pedido não pertence ao usuário")
	}

	log.Printf("✅ Pedido %s encontrado (status: %s)", orderID, order.Status)
	return order, nil
}

// ListOrders lista pedidos de um usuário com paginação
func (s *orderService) ListOrders(ctx context.Context, userID string, page, pageSize int) ([]*model.Order, int64, error) {
	log.Printf("📋 Listando pedidos do usuário %s (página %d, tamanho %d)", userID, page, pageSize)

	// Validar userID
	if strings.TrimSpace(userID) == "" {
		return nil, 0, fmt.Errorf("user_id é obrigatório")
	}

	// Validar e ajustar paginação
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Calcular skip
	skip := (page - 1) * pageSize

	// Buscar pedidos
	orders, err := s.repo.FindByUserID(ctx, userID, pageSize, skip)
	if err != nil {
		log.Printf("❌ Erro ao buscar pedidos: %v", err)
		return nil, 0, fmt.Errorf("erro ao buscar pedidos: %w", err)
	}

	// Buscar total
	total, err := s.repo.Count(ctx, userID)
	if err != nil {
		log.Printf("⚠️  Erro ao contar pedidos: %v", err)
		total = 0 // Não falhar por isso
	}

	log.Printf("✅ Encontrados %d pedidos (total: %d)", len(orders), total)
	return orders, total, nil
}

// ListAllOrders lista TODOS os pedidos (admin) com paginação
func (s *orderService) ListAllOrders(ctx context.Context, page, pageSize int) ([]*model.Order, int64, error) {
	log.Printf("📋 [ADMIN] Listando TODOS os pedidos (página %d, tamanho %d)", page, pageSize)

	// Validar e ajustar paginação
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Calcular skip
	skip := (page - 1) * pageSize

	// Buscar TODOS os pedidos
	orders, err := s.repo.FindAll(ctx, pageSize, skip)
	if err != nil {
		log.Printf("❌ Erro ao buscar pedidos: %v", err)
		return nil, 0, fmt.Errorf("erro ao buscar pedidos: %w", err)
	}

	// Buscar total de pedidos
	total, err := s.repo.CountAll(ctx)
	if err != nil {
		log.Printf("⚠️  Erro ao contar pedidos: %v", err)
		total = 0 // Não falhar por isso
	}

	log.Printf("✅ [ADMIN] Encontrados %d pedidos (total: %d)", len(orders), total)
	return orders, total, nil
}

// UpdateOrderStatus atualiza o status de um pedido com validações
func (s *orderService) UpdateOrderStatus(ctx context.Context, orderID, newStatus string) (*model.Order, error) {
	log.Printf("🔄 Atualizando status do pedido %s para %s", orderID, newStatus)

	// Validar parâmetros
	if strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("order_id é obrigatório")
	}
	if strings.TrimSpace(newStatus) == "" {
		return nil, fmt.Errorf("novo status é obrigatório")
	}

	// Validar se status é válido
	if !model.IsValidStatus(newStatus) {
		return nil, fmt.Errorf("status inválido: %s", newStatus)
	}

	// Buscar pedido atual
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		log.Printf("❌ Pedido %s não encontrado: %v", orderID, err)
		return nil, fmt.Errorf("pedido não encontrado")
	}

	// Armazenar status antigo para o evento
	oldStatus := order.Status

	// Validar transição de status
	if !order.CanTransitionTo(newStatus) {
		log.Printf("🚫 Transição inválida: %s → %s", oldStatus, newStatus)
		return nil, fmt.Errorf("transição de status inválida: não é possível mudar de '%s' para '%s'", oldStatus, newStatus)
	}

	// Atualizar status no banco
	if err := s.repo.UpdateStatus(ctx, orderID, newStatus); err != nil {
		log.Printf("❌ Erro ao atualizar status: %v", err)
		return nil, fmt.Errorf("erro ao atualizar status: %w", err)
	}

	// Atualizar order local
	order.Status = newStatus

	log.Printf("✅ Status atualizado: %s → %s", oldStatus, newStatus)

	// Publicar evento order.updated
	if err := s.publisher.PublishOrderUpdated(orderID, oldStatus, newStatus); err != nil {
		log.Printf("⚠️  Erro ao publicar evento order.updated: %v", err)
		// Não retornar erro, pois o status já foi atualizado
	}

	return order, nil
}

// CancelOrder cancela um pedido se permitido
func (s *orderService) CancelOrder(ctx context.Context, orderID, userID string) error {
	log.Printf("❌ Cancelando pedido %s para usuário %s", orderID, userID)

	// Validar parâmetros
	if strings.TrimSpace(orderID) == "" {
		return fmt.Errorf("order_id é obrigatório")
	}
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("user_id é obrigatório")
	}

	// Buscar pedido
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		log.Printf("❌ Pedido %s não encontrado: %v", orderID, err)
		return fmt.Errorf("pedido não encontrado")
	}

	// Verificar ownership
	if order.UserID != userID {
		log.Printf("🚫 Usuário %s tentou cancelar pedido %s de outro usuário", userID, orderID)
		return fmt.Errorf("acesso negado: este pedido não pertence ao usuário")
	}

	// Verificar se pode ser cancelado
	if !s.canBeCancelled(order.Status) {
		log.Printf("🚫 Pedido %s não pode ser cancelado (status: %s)", orderID, order.Status)
		return fmt.Errorf("pedido não pode ser cancelado no status '%s'", order.Status)
	}

	// Armazenar status antigo
	oldStatus := order.Status

	// Atualizar status para cancelled
	if err := s.repo.UpdateStatus(ctx, orderID, model.OrderStatusCancelled); err != nil {
		log.Printf("❌ Erro ao cancelar pedido: %v", err)
		return fmt.Errorf("erro ao cancelar pedido: %w", err)
	}

	log.Printf("✅ Pedido %s cancelado com sucesso", orderID)

	// Publicar evento order.updated
	if err := s.publisher.PublishOrderUpdated(orderID, oldStatus, model.OrderStatusCancelled); err != nil {
		log.Printf("⚠️  Erro ao publicar evento de cancelamento: %v", err)
		// Não retornar erro, pois o pedido já foi cancelado
	}

	return nil
}

// validateItems valida os itens do pedido
func (s *orderService) validateItems(items []model.OrderItem) error {
	if len(items) == 0 {
		return fmt.Errorf("pedido deve conter pelo menos um item")
	}

	if len(items) > 50 {
		return fmt.Errorf("pedido não pode conter mais de 50 itens")
	}

	for i, item := range items {
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
			return fmt.Errorf("item %d: quantidade deve ser maior que zero", i+1)
		}

		if item.Quantity > 1000 {
			return fmt.Errorf("item %d: quantidade máxima é 1000 unidades", i+1)
		}

		// Validar price
		if item.Price <= 0 {
			return fmt.Errorf("item %d: preço deve ser maior que zero", i+1)
		}

		if item.Price > 1000000 {
			return fmt.Errorf("item %d: preço máximo é R$ 1.000.000,00", i+1)
		}
	}

	return nil
}

// validateAddress valida o endereço de entrega
func (s *orderService) validateAddress(address model.Address) error {
	// Validar street
	if strings.TrimSpace(address.Street) == "" {
		return fmt.Errorf("rua é obrigatória")
	}

	if len(address.Street) > 200 {
		return fmt.Errorf("rua não pode ter mais de 200 caracteres")
	}

	// Validar number
	if strings.TrimSpace(address.Number) == "" {
		return fmt.Errorf("número é obrigatório")
	}

	if len(address.Number) > 20 {
		return fmt.Errorf("número não pode ter mais de 20 caracteres")
	}

	// Validar city
	if strings.TrimSpace(address.City) == "" {
		return fmt.Errorf("cidade é obrigatória")
	}

	if len(address.City) > 100 {
		return fmt.Errorf("cidade não pode ter mais de 100 caracteres")
	}

	// Validar state
	if strings.TrimSpace(address.State) == "" {
		return fmt.Errorf("estado é obrigatório")
	}

	if len(address.State) != 2 {
		return fmt.Errorf("estado deve ter 2 caracteres (ex: SP, RJ)")
	}

	// Validar zip_code
	if strings.TrimSpace(address.ZipCode) == "" {
		return fmt.Errorf("CEP é obrigatório")
	}

	// Remover caracteres não numéricos do CEP
	zipCode := strings.ReplaceAll(address.ZipCode, "-", "")
	zipCode = strings.ReplaceAll(zipCode, ".", "")
	if len(zipCode) != 8 {
		return fmt.Errorf("CEP deve ter 8 dígitos")
	}

	// Validar complement (opcional, mas com limite de tamanho)
	if len(address.Complement) > 200 {
		return fmt.Errorf("complemento não pode ter mais de 200 caracteres")
	}

	return nil
}

// canBeCancelled verifica se um pedido pode ser cancelado baseado no status
func (s *orderService) canBeCancelled(status string) bool {
	// Pedidos podem ser cancelados apenas nos status: pending, confirmed
	cancellableStatuses := []string{
		model.OrderStatusPending,
		model.OrderStatusConfirmed,
	}

	for _, s := range cancellableStatuses {
		if status == s {
			return true
		}
	}

	return false
}

// GetOrderStatistics retorna estatísticas dos pedidos de um usuário (método extra)
func (s *orderService) GetOrderStatistics(ctx context.Context, userID string) (map[string]interface{}, error) {
	log.Printf("📊 Buscando estatísticas de pedidos do usuário %s", userID)

	// Buscar todos os pedidos do usuário
	orders, err := s.repo.FindByUserID(ctx, userID, 1000, 0) // limite alto para estatísticas
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar pedidos: %w", err)
	}

	// Calcular estatísticas
	stats := map[string]interface{}{
		"total_orders":     len(orders),
		"total_spent":      0.0,
		"orders_by_status": make(map[string]int),
	}

	totalSpent := 0.0
	statusCount := make(map[string]int)

	for _, order := range orders {
		totalSpent += order.TotalAmount
		statusCount[order.Status]++
	}

	stats["total_spent"] = totalSpent
	stats["orders_by_status"] = statusCount

	log.Printf("✅ Estatísticas calculadas: %d pedidos, R$ %.2f gastos", len(orders), totalSpent)
	return stats, nil
}
