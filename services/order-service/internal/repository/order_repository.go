package repository

import (
	"context"
	"fmt"
	"time"

	"pedidos-online/order-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// OrderRepository define a interface para operações de pedidos
type OrderRepository interface {
	// Create insere um novo pedido no banco de dados
	Create(ctx context.Context, order *model.Order) error

	// FindByID busca um pedido pelo ID
	FindByID(ctx context.Context, id string) (*model.Order, error)

	// FindByUserID busca pedidos de um usuário com paginação
	FindByUserID(ctx context.Context, userID string, limit, skip int) ([]*model.Order, error)

	// FindAll busca TODOS os pedidos (admin) com paginação
	FindAll(ctx context.Context, limit, skip int) ([]*model.Order, error)

	// Update atualiza um pedido completo
	Update(ctx context.Context, order *model.Order) error

	// UpdateStatus atualiza apenas o status de um pedido
	UpdateStatus(ctx context.Context, id, status string) error

	// Delete remove um pedido do banco de dados
	Delete(ctx context.Context, id string) error

	// Count retorna o total de pedidos de um usuário
	Count(ctx context.Context, userID string) (int64, error)

	// CountAll retorna o total de pedidos (admin)
	CountAll(ctx context.Context) (int64, error)

	// CreateIndexes cria os índices necessários
	CreateIndexes(ctx context.Context) error
}

// orderRepository implementa a interface OrderRepository
type orderRepository struct {
	collection *mongo.Collection
	timeout    time.Duration
}

// NewOrderRepository cria uma nova instância do repositório
func NewOrderRepository(db *mongo.Database) OrderRepository {
	return &orderRepository{
		collection: db.Collection("orders"),
		timeout:    10 * time.Second, // timeout padrão de 10 segundos
	}
}

// Create insere um novo pedido no banco de dados
func (r *orderRepository) Create(ctx context.Context, order *model.Order) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// Definir timestamps
	now := time.Now()
	order.CreatedAt = now
	order.UpdatedAt = now

	// Inserir no banco
	result, err := r.collection.InsertOne(ctx, order)
	if err != nil {
		return fmt.Errorf("erro ao criar pedido: %w", err)
	}

	// Atualizar o ID do pedido
	order.ID = result.InsertedID.(primitive.ObjectID)

	return nil
}

// FindByID busca um pedido pelo ID
func (r *orderRepository) FindByID(ctx context.Context, id string) (*model.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// Validar e converter ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("ID inválido: %w", err)
	}

	var order model.Order
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("pedido não encontrado")
		}
		return nil, fmt.Errorf("erro ao buscar pedido: %w", err)
	}

	return &order, nil
}

// FindByUserID busca pedidos de um usuário com paginação
func (r *orderRepository) FindByUserID(ctx context.Context, userID string, limit, skip int) ([]*model.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// Validar parâmetros de paginação
	if limit <= 0 {
		limit = 10 // limite padrão
	}
	if limit > 100 {
		limit = 100 // limite máximo
	}
	if skip < 0 {
		skip = 0
	}

	// Configurar opções de busca
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}). // mais recentes primeiro
		SetLimit(int64(limit)).
		SetSkip(int64(skip))

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar pedidos do usuário: %w", err)
	}
	defer cursor.Close(ctx)

	var orders []*model.Order
	if err = cursor.All(ctx, &orders); err != nil {
		return nil, fmt.Errorf("erro ao decodificar pedidos: %w", err)
	}

	// Retornar lista vazia ao invés de nil se não houver pedidos
	if orders == nil {
		orders = []*model.Order{}
	}

	return orders, nil
}

// Update atualiza um pedido completo
func (r *orderRepository) Update(ctx context.Context, order *model.Order) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// Atualizar timestamp
	order.UpdatedAt = time.Now()

	// Preparar update
	update := bson.M{
		"$set": bson.M{
			"items":        order.Items,
			"total_amount": order.TotalAmount,
			"status":       order.Status,
			"address":      order.Address,
			"updated_at":   order.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": order.ID}, update)
	if err != nil {
		return fmt.Errorf("erro ao atualizar pedido: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("pedido não encontrado")
	}

	return nil
}

// UpdateStatus atualiza apenas o status de um pedido
func (r *orderRepository) UpdateStatus(ctx context.Context, id, status string) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// Validar ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("ID inválido: %w", err)
	}

	// Validar status
	validStatuses := map[string]bool{
		model.OrderStatusPending:   true,
		model.OrderStatusConfirmed: true,
		model.OrderStatusPreparing: true,
		model.OrderStatusShipped:   true,
		model.OrderStatusDelivered: true,
		model.OrderStatusCancelled: true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("status inválido: %s", status)
	}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return fmt.Errorf("erro ao atualizar status do pedido: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("pedido não encontrado")
	}

	return nil
}

// Delete remove um pedido do banco de dados
func (r *orderRepository) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// Validar ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("ID inválido: %w", err)
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("erro ao deletar pedido: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("pedido não encontrado")
	}

	return nil
}

// Count retorna o total de pedidos de um usuário
func (r *orderRepository) Count(ctx context.Context, userID string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	count, err := r.collection.CountDocuments(ctx, bson.M{"user_id": userID})
	if err != nil {
		return 0, fmt.Errorf("erro ao contar pedidos: %w", err)
	}

	return count, nil
}

// FindAll busca TODOS os pedidos (admin) com paginação
func (r *orderRepository) FindAll(ctx context.Context, limit, skip int) ([]*model.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// Validar parâmetros de paginação
	if limit <= 0 {
		limit = 10 // limite padrão
	}
	if limit > 100 {
		limit = 100 // limite máximo
	}
	if skip < 0 {
		skip = 0
	}

	// Configurar opções de busca (SEM filtro de user_id)
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}). // mais recentes primeiro
		SetLimit(int64(limit)).
		SetSkip(int64(skip))

	cursor, err := r.collection.Find(ctx, bson.M{}, opts) // bson.M{} = buscar todos
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar todos os pedidos: %w", err)
	}
	defer cursor.Close(ctx)

	var orders []*model.Order
	if err = cursor.All(ctx, &orders); err != nil {
		return nil, fmt.Errorf("erro ao decodificar pedidos: %w", err)
	}

	// Retornar lista vazia ao invés de nil se não houver pedidos
	if orders == nil {
		orders = []*model.Order{}
	}

	return orders, nil
}

// CountAll retorna o total de pedidos (admin)
func (r *orderRepository) CountAll(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	count, err := r.collection.CountDocuments(ctx, bson.M{}) // bson.M{} = contar todos
	if err != nil {
		return 0, fmt.Errorf("erro ao contar todos os pedidos: %w", err)
	}

	return count, nil
}

// CreateIndexes cria índices necessários na collection
func (r *orderRepository) CreateIndexes(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Índice em user_id (para buscar pedidos do usuário)
	userIDIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: options.Index().SetName("idx_user_id"),
	}

	// Índice em created_at descendente (para ordenação por data)
	createdAtIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "created_at", Value: -1}},
		Options: options.Index().SetName("idx_created_at"),
	}

	// Índice em status (para filtrar por status)
	statusIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "status", Value: 1}},
		Options: options.Index().SetName("idx_status"),
	}

	// Índice composto user_id + created_at (otimiza busca paginada de pedidos do usuário)
	compositeIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "created_at", Value: -1},
		},
		Options: options.Index().SetName("idx_user_created"),
	}

	// Criar todos os índices
	indexNames, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		userIDIndex,
		createdAtIndex,
		statusIndex,
		compositeIndex,
	})

	if err != nil {
		return fmt.Errorf("erro ao criar índices: %w", err)
	}

	// Log dos índices criados (opcional)
	_ = indexNames // índices criados com sucesso

	return nil
}
