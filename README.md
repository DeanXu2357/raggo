# Raggo

Raggo is a Go-based RAG (Retrieval-Augmented Generation) service that provides document processing, translation, and knowledge base management capabilities.

## Features

- Document processing and chunking
- Multi-language translation support
- Knowledge base creation and management
- Integration with various LLM providers (including Ollama)
- Vector search capabilities via Weaviate
- Scalable storage with MinIO and PostgreSQL

## Installation

### Prerequisites

- Go 1.20 or later
- Docker and Docker Compose
- PostgreSQL
- MinIO
- Weaviate

### Getting Started

1. Clone the repository:
```bash
git clone https://github.com/yourusername/raggo.git
cd raggo
```

2. Copy the environment file and configure your settings:
```bash
cp .env.example .env
```

3. Start the required services using Docker Compose:
```bash
docker-compose up -d
```

4. Run database migrations:
```bash
./scripts/migration.sh up
```

5. Build and run the service:
```bash
go build
./raggo serve
```

The service should now be running and ready to accept requests.
