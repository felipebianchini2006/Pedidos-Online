#!/usr/bin/env node

/**
 * ============================================================================
 * Script de Seed para MongoDB - Order Service
 * ============================================================================
 * Este script popula a coleção 'orders' com dados de teste e cria índices
 * necessários para otimizar as consultas.
 * 
 * Uso:
 *   node scripts/seed-orders.js
 * 
 * Requisitos:
 *   npm install mongodb
 * ============================================================================
 */

const { MongoClient, ObjectId } = require('mongodb');

// ============================================================================
// Configuração da Conexão
// ============================================================================
const MONGO_URI = process.env.MONGO_URI || 'mongodb://localhost:27017';
const DB_NAME = process.env.MONGO_DATABASE || 'orders_db';
const COLLECTION_NAME = 'orders';

// ============================================================================
// IDs de Usuários (devem corresponder aos usuários do PostgreSQL)
// ============================================================================
// NOTA: Em produção, você precisaria buscar os IDs reais do PostgreSQL
// Para este seed, usamos IDs fictícios que devem ser substituídos pelos reais
const USER_IDS = {
  admin: 'replace-with-real-admin-uuid',
  user1: 'replace-with-real-user1-uuid',
  user2: 'replace-with-real-user2-uuid'
};

// ============================================================================
// Dados de Seed - Pedidos de Teste
// ============================================================================
const seedOrders = [
  {
    user_id: USER_IDS.admin,
    items: [
      {
        product_id: 'PROD-001',
        product_name: 'Notebook Dell Inspiron 15',
        quantity: 1,
        price: 3500.00
      },
      {
        product_id: 'PROD-002',
        product_name: 'Mouse Logitech MX Master 3',
        quantity: 1,
        price: 450.00
      }
    ],
    total_amount: 3950.00,
    status: 'DELIVERED',
    address: {
      street: 'Av. Paulista',
      number: '1000',
      city: 'São Paulo',
      state: 'SP',
      zip_code: '01310-100',
      complement: 'Apto 101'
    },
    created_at: new Date('2024-01-15T10:30:00Z'),
    updated_at: new Date('2024-01-20T14:00:00Z')
  },
  {
    user_id: USER_IDS.user1,
    items: [
      {
        product_id: 'PROD-003',
        product_name: 'Teclado Mecânico Keychron K2',
        quantity: 1,
        price: 650.00
      },
      {
        product_id: 'PROD-004',
        product_name: 'Webcam Logitech C920',
        quantity: 1,
        price: 550.00
      },
      {
        product_id: 'PROD-005',
        product_name: 'Headset HyperX Cloud II',
        quantity: 1,
        price: 480.00
      }
    ],
    total_amount: 1680.00,
    status: 'SHIPPED',
    address: {
      street: 'Rua das Flores',
      number: '250',
      city: 'Rio de Janeiro',
      state: 'RJ',
      zip_code: '20040-020',
      complement: ''
    },
    created_at: new Date('2024-02-10T15:45:00Z'),
    updated_at: new Date('2024-02-12T09:20:00Z')
  },
  {
    user_id: USER_IDS.user1,
    items: [
      {
        product_id: 'PROD-006',
        product_name: 'Monitor LG UltraWide 34"',
        quantity: 1,
        price: 2200.00
      }
    ],
    total_amount: 2200.00,
    status: 'PENDING',
    address: {
      street: 'Rua das Flores',
      number: '250',
      city: 'Rio de Janeiro',
      state: 'RJ',
      zip_code: '20040-020',
      complement: ''
    },
    created_at: new Date('2024-02-25T11:00:00Z'),
    updated_at: new Date('2024-02-25T11:00:00Z')
  },
  {
    user_id: USER_IDS.user2,
    items: [
      {
        product_id: 'PROD-007',
        product_name: 'SSD Samsung 970 EVO 1TB',
        quantity: 2,
        price: 850.00
      },
      {
        product_id: 'PROD-008',
        product_name: 'Memória RAM Corsair 16GB',
        quantity: 2,
        price: 450.00
      }
    ],
    total_amount: 2600.00,
    status: 'CANCELLED',
    address: {
      street: 'Av. Brasil',
      number: '5000',
      city: 'Belo Horizonte',
      state: 'MG',
      zip_code: '30140-000',
      complement: 'Bloco B, Sala 305'
    },
    created_at: new Date('2024-03-01T08:15:00Z'),
    updated_at: new Date('2024-03-02T16:30:00Z')
  }
];

// ============================================================================
// Funções Principais
// ============================================================================

/**
 * Cria índices na coleção de pedidos para otimizar consultas
 */
async function createIndexes(collection) {
  console.log('\n📊 Criando índices...');
  
  try {
    // Índice para buscar pedidos por usuário (consulta mais comum)
    await collection.createIndex(
      { user_id: 1 },
      { name: 'idx_user_id' }
    );
    console.log('✅ Índice criado: idx_user_id');

    // Índice para buscar pedidos por status
    await collection.createIndex(
      { status: 1 },
      { name: 'idx_status' }
    );
    console.log('✅ Índice criado: idx_status');

    // Índice composto para buscar pedidos por usuário e status
    await collection.createIndex(
      { user_id: 1, status: 1 },
      { name: 'idx_user_status' }
    );
    console.log('✅ Índice criado: idx_user_status');

    // Índice para ordenar por data de criação (mais recentes primeiro)
    await collection.createIndex(
      { created_at: -1 },
      { name: 'idx_created_at_desc' }
    );
    console.log('✅ Índice criado: idx_created_at_desc');

    // Índice composto para consultas complexas (usuário + data)
    await collection.createIndex(
      { user_id: 1, created_at: -1 },
      { name: 'idx_user_created_at' }
    );
    console.log('✅ Índice criado: idx_user_created_at');

  } catch (error) {
    console.error('❌ Erro ao criar índices:', error.message);
    throw error;
  }
}

/**
 * Popula a coleção com dados de teste
 */
async function seedOrders(collection) {
  console.log('\n🌱 Populando pedidos de teste...');
  
  try {
    // Limpar dados existentes (opcional)
    const deleteResult = await collection.deleteMany({});
    console.log(`🗑️  ${deleteResult.deletedCount} pedidos existentes removidos`);

    // Inserir novos pedidos
    const result = await collection.insertMany(seedOrders);
    console.log(`✅ ${result.insertedCount} pedidos inseridos com sucesso`);

    // Exibir resumo dos pedidos inseridos
    console.log('\n📦 Resumo dos Pedidos Inseridos:');
    for (const [index, order] of seedOrders.entries()) {
      const insertedId = Object.values(result.insertedIds)[index];
      console.log(`  ${index + 1}. ID: ${insertedId} | Status: ${order.status} | Total: R$ ${order.total_amount.toFixed(2)}`);
    }

  } catch (error) {
    console.error('❌ Erro ao popular pedidos:', error.message);
    throw error;
  }
}

/**
 * Verifica os dados inseridos
 */
async function verifyData(collection) {
  console.log('\n🔍 Verificando dados inseridos...');
  
  try {
    // Contagem total
    const totalCount = await collection.countDocuments({});
    console.log(`📊 Total de pedidos: ${totalCount}`);

    // Contagem por status
    const statusCounts = await collection.aggregate([
      {
        $group: {
          _id: '$status',
          count: { $sum: 1 }
        }
      },
      {
        $sort: { count: -1 }
      }
    ]).toArray();

    console.log('\n📈 Pedidos por Status:');
    statusCounts.forEach(stat => {
      console.log(`  ${stat._id}: ${stat.count}`);
    });

    // Valor total de pedidos
    const totalValue = await collection.aggregate([
      {
        $group: {
          _id: null,
          total: { $sum: '$total_amount' }
        }
      }
    ]).toArray();

    if (totalValue.length > 0) {
      console.log(`\n💰 Valor Total de Pedidos: R$ ${totalValue[0].total.toFixed(2)}`);
    }

  } catch (error) {
    console.error('❌ Erro ao verificar dados:', error.message);
    throw error;
  }
}

/**
 * Lista os índices criados
 */
async function listIndexes(collection) {
  console.log('\n📋 Índices Criados:');
  
  try {
    const indexes = await collection.listIndexes().toArray();
    indexes.forEach(index => {
      console.log(`  - ${index.name}: ${JSON.stringify(index.key)}`);
    });
  } catch (error) {
    console.error('❌ Erro ao listar índices:', error.message);
    throw error;
  }
}

/**
 * Função principal
 */
async function main() {
  console.log('============================================================================');
  console.log('🚀 Iniciando Seed do MongoDB - Order Service');
  console.log('============================================================================');
  console.log(`📍 URI: ${MONGO_URI}`);
  console.log(`📍 Database: ${DB_NAME}`);
  console.log(`📍 Collection: ${COLLECTION_NAME}`);

  let client;

  try {
    // Conectar ao MongoDB
    console.log('\n🔌 Conectando ao MongoDB...');
    client = new MongoClient(MONGO_URI);
    await client.connect();
    console.log('✅ Conectado com sucesso!');

    // Obter referências do database e collection
    const db = client.db(DB_NAME);
    const collection = db.collection(COLLECTION_NAME);

    // Executar operações de seed
    await createIndexes(collection);
    await seedOrders(collection);
    await verifyData(collection);
    await listIndexes(collection);

    console.log('\n============================================================================');
    console.log('✅ Seed concluído com sucesso!');
    console.log('============================================================================\n');

  } catch (error) {
    console.error('\n============================================================================');
    console.error('❌ Erro durante o seed:', error.message);
    console.error('============================================================================\n');
    process.exit(1);

  } finally {
    // Fechar conexão
    if (client) {
      await client.close();
      console.log('🔌 Conexão fechada.\n');
    }
  }
}

// ============================================================================
// Execução
// ============================================================================
if (require.main === module) {
  main();
}

module.exports = { seedOrders, createIndexes };
