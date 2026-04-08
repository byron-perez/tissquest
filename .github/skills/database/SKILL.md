---
name: database
description: "Skill for database operations in tissquest - migrations, setup, backend switching"
---

# Database Operations Skill

## Use When
- Setting up database connections
- Running migrations
- Switching between SQLite and PostgreSQL
- Troubleshooting database issues

## Workflow
1. Check current DB_TYPE environment variable
2. Copy .env.example to .env if needed
3. Configure connection parameters
4. Run migration.RunMigration() or build the app to trigger auto-migration
5. For switching backends, update DB_TYPE and restart the application

## Assets
- .env.example: Template for environment variables
- internal/persistence/migration/migration.go: Migration logic
- Dockerfile: Containerized database setup