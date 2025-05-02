package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	agentflow "kunalkushwaha/agentflow/internal/core" // Import core types

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

// PgVectorMemory implements the agentflow.VectorMemory interface using PostgreSQL with pgvector.
type PgVectorMemory struct {
	pool      *pgxpool.Pool
	tableName string // The PostgreSQL table name to store vectors in
	// TODO: Add configuration for index type (HNSW, IVFFlat), list counts etc.
}

// PgVectorConfig holds configuration options for the PgVector client.
type PgVectorConfig struct {
	DSN        string // PostgreSQL Data Source Name (e.g., "postgres://user:password@host:port/dbname")
	TableName  string // The PostgreSQL table name to use (e.g., "memory_chunks")
	Dimensions int    // The expected dimension of the vectors
}

// NewPgVectorMemory creates a new PgVectorMemory instance and ensures the table and extension exist.
func NewPgVectorMemory(ctx context.Context, config PgVectorConfig) (*PgVectorMemory, error) {
	if config.TableName == "" {
		return nil, errors.New("PgVector table name cannot be empty")
	}
	if config.Dimensions <= 0 {
		return nil, errors.New("vector dimensions must be positive")
	}
	if config.DSN == "" {
		return nil, errors.New("PostgreSQL DSN cannot be empty")
	}

	poolConfig, err := pgxpool.ParseConfig(config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	// Register pgvector types with pgx
	// poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
	//  pgvector.RegisterTypes(conn.TypeMap())
	//  return nil
	// }
	// Note: pgxpool v5 might handle type registration differently or automatically.
	// We will register types on the pool's config directly if needed, or rely on inference.
	// Let's try without explicit registration first, pgx might infer correctly.

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Test connection
	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	mem := &PgVectorMemory{
		pool:      pool,
		tableName: config.TableName,
	}

	// Ensure the extension and table exist
	err = mem.ensureTableExists(ctx, config.Dimensions)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ensure PgVector table '%s': %w", config.TableName, err)
	}

	log.Printf("PgVectorMemory initialized for table '%s'", config.TableName)
	return mem, nil
}

// ensureTableExists checks if the pgvector extension and the configured table exist, creating them if not.
func (m *PgVectorMemory) ensureTableExists(ctx context.Context, dimensions int) error {
	// 1. Ensure pgvector extension exists
	_, err := m.pool.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS vector;")
	if err != nil {
		return fmt.Errorf("failed to create vector extension: %w", err)
	}
	log.Println("Ensured pgvector extension exists.")

	// 2. Ensure table exists
	// Use proper SQL identifier quoting for the table name
	createTableSQL := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            id TEXT PRIMARY KEY,
            embedding VECTOR(%d),
            metadata JSONB,
            created_at TIMESTAMPTZ DEFAULT NOW(),
            updated_at TIMESTAMPTZ DEFAULT NOW()
        );`, m.tableName, dimensions) // Use %s for table name, %d for dimensions

	_, err = m.pool.Exec(ctx, createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table '%s': %w", m.tableName, err)
	}
	log.Printf("Ensured table '%s' exists.", m.tableName)

	// 3. Ensure indexes exist (optional but highly recommended for performance)
	// Index on metadata (GIN index is good for JSONB)
	createMetaIndexSQL := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_metadata_gin ON %s USING GIN (metadata);`, m.sanitizeIdentifier(m.tableName), m.tableName)
	_, err = m.pool.Exec(ctx, createMetaIndexSQL)
	if err != nil {
		// Log warning, but don't fail initialization
		log.Printf("Warning: Failed to create GIN index on metadata for table '%s': %v", m.tableName, err)
	} else {
		log.Printf("Ensured GIN index on metadata for table '%s'.", m.tableName)
	}

	// Index on embedding (HNSW is often a good default, requires pgvector >= 0.5.0)
	// Cosine distance is 1 - cosine similarity. Use <-> operator for cosine distance.
	// Use <=> for L2 distance, <+> for inner product.
	// Adjust lists/ef_construction based on data size/needs.
	createVectorIndexSQL := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_embedding_hnsw ON %s USING HNSW (embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64);`, m.sanitizeIdentifier(m.tableName), m.tableName)
	// Fallback for older pgvector or different preference: IVFFlat
	// createVectorIndexSQL := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_embedding_ivfflat ON %s USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);`, m.sanitizeIdentifier(m.tableName), m.tableName)

	_, err = m.pool.Exec(ctx, createVectorIndexSQL)
	if err != nil {
		// Log warning, but don't fail initialization
		log.Printf("Warning: Failed to create HNSW index on embedding for table '%s' (consider IVFFlat or check pgvector version/config): %v", m.tableName, err)
	} else {
		log.Printf("Ensured HNSW index on embedding for table '%s'.", m.tableName)
	}

	// Trigger for updated_at
	triggerFuncSQL := `
        CREATE OR REPLACE FUNCTION update_updated_at_column()
        RETURNS TRIGGER AS $$
        BEGIN
           NEW.updated_at = NOW();
           RETURN NEW;
        END;
        $$ language 'plpgsql';
    `
	_, err = m.pool.Exec(ctx, triggerFuncSQL)
	if err != nil {
		log.Printf("Warning: Failed to create updated_at trigger function: %v", err)
	} else {
		triggerSQL := fmt.Sprintf(`
            DROP TRIGGER IF EXISTS update_%s_updated_at ON %s; -- Drop existing trigger first
            CREATE TRIGGER update_%s_updated_at
            BEFORE UPDATE ON %s
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
        `, m.sanitizeIdentifier(m.tableName), m.tableName, m.sanitizeIdentifier(m.tableName), m.tableName)
		_, err = m.pool.Exec(ctx, triggerSQL)
		if err != nil {
			log.Printf("Warning: Failed to create updated_at trigger for table '%s': %v", m.tableName, err)
		} else {
			log.Printf("Ensured updated_at trigger for table '%s'.", m.tableName)
		}
	}

	return nil
}

// Store saves a vector embedding and its associated metadata to PgVector using UPSERT.
func (m *PgVectorMemory) Store(ctx context.Context, id string, embedding []float32, metadata map[string]any) error {
	if id == "" {
		return errors.New("item ID cannot be empty")
	}
	if len(embedding) == 0 {
		return errors.New("embedding cannot be empty")
	}

	// Convert metadata map to JSONB
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata to JSON: %w", err)
	}

	// Convert embedding to pgvector.Vector
	vec := pgvector.NewVector(embedding)

	// Use INSERT ... ON CONFLICT for UPSERT semantics
	upsertSQL := fmt.Sprintf(`
        INSERT INTO %s (id, embedding, metadata)
        VALUES ($1, $2, $3)
        ON CONFLICT (id) DO UPDATE SET
            embedding = EXCLUDED.embedding,
            metadata = EXCLUDED.metadata,
            updated_at = NOW();
    `, m.tableName) // Use %s for table name

	_, err = m.pool.Exec(ctx, upsertSQL, id, vec, metadataJSON)
	if err != nil {
		// Check if the error is due to incorrect vector dimensions
		if strings.Contains(err.Error(), "invalid vector dimensions") {
			return fmt.Errorf("dimension mismatch: expected %d dimensions based on table schema, but got %d for id '%s'", 0, len(embedding), id) // TODO: Fetch expected dimension from config/schema
		}
		return fmt.Errorf("failed to upsert item '%s': %w", id, err)
	}

	return nil
}

// Query performs a vector similarity search in PgVector.
func (m *PgVectorMemory) Query(ctx context.Context, embedding []float32, topK int) ([]agentflow.QueryResult, error) {
	if len(embedding) == 0 {
		return nil, errors.New("query embedding cannot be empty")
	}
	if topK <= 0 {
		return nil, errors.New("topK must be positive")
	}

	// Convert query embedding to pgvector.Vector
	queryVec := pgvector.NewVector(embedding)

	// Use the cosine distance operator '<->' (lower distance is better)
	// Similarity = 1 - distance
	querySQL := fmt.Sprintf(`
        SELECT id, metadata, 1 - (embedding <-> $1) AS similarity_score
        FROM %s
        ORDER BY embedding <-> $1
        LIMIT $2;
    `, m.tableName) // Use %s for table name

	rows, err := m.pool.Query(ctx, querySQL, queryVec, topK)
	if err != nil {
		return nil, fmt.Errorf("PgVector query failed: %w", err)
	}
	defer rows.Close()

	results := []agentflow.QueryResult{}
	for rows.Next() {
		var (
			id           string
			metadataJSON []byte
			score        float32
		)
		err := rows.Scan(&id, &metadataJSON, &score)
		if err != nil {
			log.Printf("Warning: Failed to scan row during PgVector query: %v", err)
			continue // Skip problematic row
		}

		metadata := make(map[string]any)
		if metadataJSON != nil {
			err = json.Unmarshal(metadataJSON, &metadata)
			if err != nil {
				log.Printf("Warning: Failed to unmarshal metadata JSON for id '%s': %v", id, err)
				// Keep going, but metadata will be empty for this result
			}
		}

		results = append(results, agentflow.QueryResult{
			ID:       id,
			Metadata: metadata,
			Score:    score, // Score is calculated as 1 - distance
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating PgVector query results: %w", err)
	}

	return results, nil
}

// sanitizeIdentifier basic sanitization for SQL identifiers (table/index names)
// This is NOT foolproof against all SQL injection if table names come from untrusted input,
// but sufficient if table names are controlled internally or via config.
func (m *PgVectorMemory) sanitizeIdentifier(identifier string) string {
	// Replace potentially harmful characters. Allow alphanumeric and underscore.
	sanitized := strings.Builder{}
	for _, r := range identifier {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			sanitized.WriteRune(r)
		} else {
			sanitized.WriteRune('_') // Replace others with underscore
		}
	}
	// Basic check to prevent empty or all-underscore names
	result := sanitized.String()
	if result == "" || strings.Trim(result, "_") == "" {
		return "invalid_identifier"
	}
	// Ensure it doesn't start with a number (though quoted identifiers might allow it)
	if result[0] >= '0' && result[0] <= '9' {
		return "_" + result
	}
	// In real SQL, identifiers might need quoting (""), but Sprintf doesn't handle that well.
	// Assume basic names are okay here.
	return result
}

// Close closes the underlying database connection pool.
func (m *PgVectorMemory) Close() {
	if m.pool != nil {
		m.pool.Close()
		log.Println("PgVectorMemory connection pool closed.")
	}
}
