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

// OrderRepository gerencia operações de banco de dados para pedidos
type OrderRepository struct {
	collection *mongo.Collection
	timeout    time.Duration
}

// NewOrderRepository cria uma nova instância do repositório
func NewOrderRepository(db *mongo.Database, timeout time.Duration) *OrderRepository {
	return &OrderRepository{
		collection: db.Collection("orders"),
		timeout:    timeout,
	}
}

// Create insere um novo pedido no banco de dados
func (r *OrderRepository) Create(ctx context.Context, order *model.Order) error {
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
func (r *OrderRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var order model.Order
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("pedido não encontrado")
		}
		return nil, fmt.Errorf("erro ao buscar pedido: %w", err)
	}

	return &order, nil
}

// FindByUserID busca todos os pedidos de um usuário
func (r *OrderRepository) FindByUserID(ctx context.Context, userID string) ([]*model.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// Ordenar por data de criação decrescente (mais recentes primeiro)
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar pedidos do usuário: %w", err)
	}
	defer cursor.Close(ctx)

	var orders []*model.Order
	if err = cursor.All(ctx, &orders); err != nil {
		return nil, fmt.Errorf("erro ao decodificar pedidos: %w", err)
	}

	return orders, nil
}

// UpdateStatus atualiza o status de um pedido
func (r *OrderRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("erro ao atualizar status do pedido: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("pedido não encontrado")
	}

	return nil
}

// Delete remove um pedido (soft delete seria melhor em produção)
func (r *OrderRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("erro ao deletar pedido: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("pedido não encontrado")
	}

	return nil
}

// CreateIndexes cria índices necessários na collection
func (r *OrderRepository) CreateIndexes(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Índice em user_id (para buscar pedidos do usuário)
	userIDIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	}

	// Índice em created_at descendente (para ordenação)
	createdAtIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "created_at", Value: -1}},
	}

	// Índice em status (para filtrar por status)
	statusIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "status", Value: 1}},
	}

	// Índice composto user_id + created_at (otimiza busca de pedidos do usuário)
	compositeIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "created_at", Value: -1},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		userIDIndex,
		createdAtIndex,
		statusIndex,
		compositeIndex,
	})

	if err != nil {
		return fmt.Errorf("erro ao criar índices: %w", err)
	}

	return nil
}
