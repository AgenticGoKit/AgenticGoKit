#!/bin/bash

# RAG Knowledge Base Setup Script
# This script sets up the PostgreSQL database with pgvector for the RAG example

set -e

echo "🚀 Setting up RAG Knowledge Base..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running. Please start Docker and try again.${NC}"
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose > /dev/null 2>&1 && ! docker compose version > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker Compose is not available. Please install Docker Compose.${NC}"
    exit 1
fi

# Use docker compose or docker-compose based on availability
DOCKER_COMPOSE_CMD="docker compose"
if ! docker compose version > /dev/null 2>&1; then
    DOCKER_COMPOSE_CMD="docker-compose"
fi

echo -e "${BLUE}📦 Starting PostgreSQL with pgvector...${NC}"

# Start PostgreSQL container
$DOCKER_COMPOSE_CMD up -d postgres

echo -e "${YELLOW}⏳ Waiting for PostgreSQL to be ready...${NC}"

# Wait for PostgreSQL to be ready
max_attempts=30
attempt=1

while [ $attempt -le $max_attempts ]; do
    if docker exec rag-postgres pg_isready -U agentflow -d agentflow > /dev/null 2>&1; then
        echo -e "${GREEN}✅ PostgreSQL is ready!${NC}"
        break
    fi
    
    if [ $attempt -eq $max_attempts ]; then
        echo -e "${RED}❌ PostgreSQL failed to start after $max_attempts attempts${NC}"
        echo -e "${YELLOW}💡 Try running: $DOCKER_COMPOSE_CMD logs postgres${NC}"
        exit 1
    fi
    
    echo -e "${YELLOW}   Attempt $attempt/$max_attempts - waiting...${NC}"
    sleep 2
    ((attempt++))
done

echo -e "${BLUE}🔧 Verifying database setup...${NC}"

# Verify pgvector extension
if docker exec rag-postgres psql -U agentflow -d agentflow -c "SELECT extname FROM pg_extension WHERE extname = 'vector';" | grep -q vector; then
    echo -e "${GREEN}✅ pgvector extension is installed${NC}"
else
    echo -e "${RED}❌ pgvector extension not found${NC}"
    exit 1
fi

# Verify tables exist
if docker exec rag-postgres psql -U agentflow -d agentflow -c "\dt" | grep -q documents; then
    echo -e "${GREEN}✅ Documents table created${NC}"
else
    echo -e "${RED}❌ Documents table not found${NC}"
    exit 1
fi

# Check sample data
sample_count=$(docker exec rag-postgres psql -U agentflow -d agentflow -t -c "SELECT COUNT(*) FROM documents;" | tr -d ' ')
if [ "$sample_count" -gt 0 ]; then
    echo -e "${GREEN}✅ Sample data loaded ($sample_count documents)${NC}"
else
    echo -e "${YELLOW}⚠️  No sample data found${NC}"
fi

echo -e "${BLUE}🔍 Testing vector search functionality...${NC}"

# Test vector search
if docker exec rag-postgres psql -U agentflow -d agentflow -c "SELECT content FROM documents WHERE embedding IS NOT NULL LIMIT 1;" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Vector search functionality verified${NC}"
else
    echo -e "${YELLOW}⚠️  Vector search test skipped (no embeddings yet)${NC}"
fi

echo -e "${BLUE}📊 Database connection info:${NC}"
echo -e "   Host: localhost"
echo -e "   Port: 5432"
echo -e "   Database: agentflow"
echo -e "   Username: agentflow"
echo -e "   Password: password"

echo -e "${BLUE}🔗 Connection string:${NC}"
echo -e "   postgres://agentflow:password@localhost:5432/agentflow?sslmode=disable"

echo -e "${GREEN}🎉 Setup complete!${NC}"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo -e "1. Set your API keys:"
echo -e "   ${YELLOW}export OPENAI_API_KEY='your-openai-api-key'${NC}"
echo -e "   ${YELLOW}# OR for local embeddings:${NC}"
echo -e "   ${YELLOW}export EMBEDDING_PROVIDER='ollama'${NC}"
echo -e "   ${YELLOW}ollama pull nomic-embed-text:latest${NC}"
echo ""
echo -e "2. Install Go dependencies:"
echo -e "   ${YELLOW}go mod tidy${NC}"
echo ""
echo -e "3. Try the example:"
echo -e "   ${YELLOW}# Ingest a document${NC}"
echo -e "   ${YELLOW}echo 'Sample document content' > test.txt${NC}"
echo -e "   ${YELLOW}go run . --mode ingest --path test.txt${NC}"
echo ""
echo -e "   ${YELLOW}# Query the knowledge base${NC}"
echo -e "   ${YELLOW}go run . --mode query --question 'What is this document about?'${NC}"
echo ""
echo -e "4. To stop the database:"
echo -e "   ${YELLOW}$DOCKER_COMPOSE_CMD down${NC}"

# Create .env.example if it doesn't exist
if [ ! -f .env.example ]; then
    echo -e "${BLUE}📝 Creating .env.example file...${NC}"
    cat > .env.example << 'EOF'
# LLM Provider Configuration
OPENAI_API_KEY=your-openai-api-key-here
# AZURE_OPENAI_API_KEY=your-azure-api-key
# AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/
# AZURE_OPENAI_DEPLOYMENT=your-deployment-name

# Embedding Configuration
EMBEDDING_PROVIDER=openai
EMBEDDING_MODEL=text-embedding-ada-002
# For local embeddings:
# EMBEDDING_PROVIDER=ollama
# EMBEDDING_MODEL=nomic-embed-text:latest

# LLM Provider
LLM_PROVIDER=openai

# Database Configuration
DATABASE_URL=postgres://agentflow:password@localhost:5432/agentflow?sslmode=disable
EOF
    echo -e "${GREEN}✅ Created .env.example${NC}"
fi

echo -e "${BLUE}💡 Tip: Copy .env.example to .env and update with your API keys${NC}"