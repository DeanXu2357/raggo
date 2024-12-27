#!/bin/sh

# Load environment variables from .env file
set -a
source .env
set +a

# Database URL for migrations
DB_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable"

# Check if migrate is installed
check_migrate() {
    if ! command -v migrate &> /dev/null; then
        echo "golang-migrate is not installed. Installing..."
        go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
        echo "golang-migrate installed successfully"
    else
        echo "golang-migrate is already installed"
    fi
}

# Run migrations up
migrate_up() {
    check_migrate
    echo "Running migrations up..."
    migrate -path database/postgresql/migrations -database "${DB_URL}" up
}

# Run migrations down
migrate_down() {
    check_migrate
    echo "Running migrations down..."
    migrate -path database/postgresql/migrations -database "${DB_URL}" down
}

# Create new migration
migrate_create() {
    check_migrate
    if [ -z "$1" ]; then
        echo "Please provide a migration name"
        echo "Usage: $0 create <migration_name>"
        exit 1
    fi
    echo "Creating new migration: $1"
    migrate create -ext sql -dir database/postgresql/migrations -seq "$1"
}

# Force set migration version
migrate_force() {
    check_migrate
    if [ -z "$1" ]; then
        echo "Please provide a version number"
        echo "Usage: $0 force <version>"
        exit 1
    fi
    echo "Force setting migration version to: $1"
    migrate -path database/postgresql/migrations -database "${DB_URL}" force "$1"
}

# Show current version
migrate_version() {
    check_migrate
    echo "Current migration version:"
    migrate -path database/postgresql/migrations -database "${DB_URL}" version
}

# Show help message
show_help() {
    echo "Usage: $0 <command> [arguments]"
    echo ""
    echo "Commands:"
    echo "  up                    Run all up migrations"
    echo "  down                  Run all down migrations"
    echo "  create <name>         Create a new migration file"
    echo "  force <version>       Force set a specific migration version"
    echo "  version              Show current migration version"
    echo "  help                 Show this help message"
}

# Main script logic
case "$1" in
    "up")
        migrate_up
        ;;
    "down")
        migrate_down
        ;;
    "create")
        migrate_create "$2"
        ;;
    "force")
        migrate_force "$2"
        ;;
    "version")
        migrate_version
        ;;
    "help"|"")
        show_help
        ;;
    *)
        echo "Unknown command: $1"
        show_help
        exit 1
        ;;
esac
