-- Initialize database for RAG knowledge base example

-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create documents table for storing document chunks with embeddings
CREATE TABLE IF NOT EXISTS documents (
    id SERIAL PRIMARY KEY,
    content TEXT NOT NULL,
    embedding vector(1536), -- OpenAI ada-002 dimensions, adjust if using different model
    metadata JSONB DEFAULT '{}',
    source VARCHAR(500),
    chunk_index INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index for vector similarity search
CREATE INDEX IF NOT EXISTS documents_embedding_idx 
ON documents USING ivfflat (embedding vector_cosine_ops) 
WITH (lists = 100);

-- Create index for metadata queries
CREATE INDEX IF NOT EXISTS documents_metadata_idx ON documents USING GIN (metadata);

-- Create index for source queries
CREATE INDEX IF NOT EXISTS documents_source_idx ON documents (source);

-- Create full-text search index for hybrid search
ALTER TABLE documents ADD COLUMN IF NOT EXISTS content_tsvector tsvector;
CREATE INDEX IF NOT EXISTS documents_content_search_idx ON documents USING GIN (content_tsvector);

-- Create trigger to update tsvector column
CREATE OR REPLACE FUNCTION update_content_tsvector() RETURNS trigger AS $$
BEGIN
    NEW.content_tsvector := to_tsvector('english', NEW.content);
    NEW.updated_at := CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_documents_tsvector ON documents;
CREATE TRIGGER update_documents_tsvector 
    BEFORE INSERT OR UPDATE ON documents 
    FOR EACH ROW EXECUTE FUNCTION update_content_tsvector();

-- Create query_logs table for analytics
CREATE TABLE IF NOT EXISTS query_logs (
    id SERIAL PRIMARY KEY,
    query TEXT NOT NULL,
    results_count INTEGER DEFAULT 0,
    processing_time_ms INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index for query analytics
CREATE INDEX IF NOT EXISTS query_logs_created_at_idx ON query_logs (created_at);

-- Insert some sample data for testing
INSERT INTO documents (content, source, metadata) VALUES 
(
    'AgenticGoKit is a Go framework for building multi-agent systems. It provides orchestration capabilities, memory management, and tool integration for creating sophisticated AI agent workflows.',
    'sample_docs/agenticgokit_intro.txt',
    '{"type": "documentation", "category": "framework", "tags": ["go", "agents", "framework"]}'
),
(
    'Vector databases are specialized databases designed to store and query high-dimensional vectors efficiently. They enable semantic search, recommendation systems, and similarity matching at scale.',
    'sample_docs/vector_databases.txt',
    '{"type": "documentation", "category": "database", "tags": ["vectors", "search", "database"]}'
),
(
    'RAG (Retrieval-Augmented Generation) combines information retrieval with text generation. It retrieves relevant context from a knowledge base and uses it to generate more accurate and informed responses.',
    'sample_docs/rag_explanation.txt',
    '{"type": "documentation", "category": "ai", "tags": ["rag", "retrieval", "generation"]}'
);

-- Update the tsvector column for existing data
UPDATE documents SET content_tsvector = to_tsvector('english', content) WHERE content_tsvector IS NULL;

-- Create a view for easy querying
CREATE OR REPLACE VIEW document_search AS
SELECT 
    id,
    content,
    source,
    metadata,
    chunk_index,
    created_at,
    -- Function to calculate similarity (placeholder - actual similarity calculated in application)
    0.0 as similarity_score
FROM documents;

-- Create function for hybrid search (combines vector and text search)
CREATE OR REPLACE FUNCTION hybrid_search(
    query_text TEXT,
    query_embedding vector(1536),
    similarity_threshold FLOAT DEFAULT 0.7,
    limit_count INTEGER DEFAULT 5
) RETURNS TABLE (
    id INTEGER,
    content TEXT,
    source VARCHAR(500),
    metadata JSONB,
    similarity_score FLOAT,
    text_rank FLOAT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        d.id,
        d.content,
        d.source,
        d.metadata,
        (1 - (d.embedding <=> query_embedding)) as similarity_score,
        ts_rank(d.content_tsvector, plainto_tsquery('english', query_text)) as text_rank
    FROM documents d
    WHERE 
        (1 - (d.embedding <=> query_embedding)) > similarity_threshold
        OR d.content_tsvector @@ plainto_tsquery('english', query_text)
    ORDER BY 
        (0.7 * (1 - (d.embedding <=> query_embedding)) + 0.3 * ts_rank(d.content_tsvector, plainto_tsquery('english', query_text))) DESC
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

-- Grant permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO agentflow;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO agentflow;

-- Display setup completion message
DO $$
BEGIN
    RAISE NOTICE 'RAG Knowledge Base database initialized successfully!';
    RAISE NOTICE 'Tables created: documents, query_logs';
    RAISE NOTICE 'Indexes created for vector search, metadata, and full-text search';
    RAISE NOTICE 'Sample data inserted for testing';
END $$;