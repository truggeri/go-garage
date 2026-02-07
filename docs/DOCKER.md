# Docker Setup Guide

This document provides instructions for building and running Go-Garage using Docker and Docker Compose.

## Prerequisites

- Docker Engine 20.10 or later
- Docker Compose 2.0 or later

## Quick Start

### Using Docker Compose (Recommended)

1. Build and start the application:
   ```bash
   docker-compose up -d
   ```

2. View logs:
   ```bash
   docker-compose logs -f app
   ```

3. Stop the application:
   ```bash
   docker-compose down
   ```

### Using Docker CLI

1. Build the image:
   ```bash
   docker build -t go-garage .
   ```

2. Run the container:
   ```bash
   docker run -d \
     -p 8080:8080 \
     -v garage-data:/app/data \
     -e ENVIRONMENT=development \
     -e LOG_LEVEL=debug \
     --name go-garage-app \
     go-garage
   ```

3. View logs:
   ```bash
   docker logs -f go-garage-app
   ```

4. Stop and remove the container:
   ```bash
   docker stop go-garage-app && docker rm go-garage-app
   ```

## Configuration

### Environment Variables

The Docker container accepts the following environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `ENVIRONMENT` | `development` | Application environment (development/production) |
| `APP_PORT` | `8080` | Port the server listens on |
| `SERVER_HOST` | `0.0.0.0` | Host the server binds to |
| `LOG_LEVEL` | `info` | Logging level (debug/info/warn/error) |
| `LOG_FORMAT` | `json` | Log output format (json/text) |
| `DB_PATH` | `/app/data/go-garage.db` | Path to SQLite database file |
| `JWT_SECRET` | _(none)_ | JWT secret key (required in production) |

### Volume Configuration

The Docker setup uses named volumes for data persistence:

- **garage-data**: Stores the SQLite database file
  - Mount point: `/app/data`
  - Purpose: Persist vehicle and maintenance records

### Network Configuration

The application runs in a bridge network (`garage-network`) for container isolation and communication.

## Development Workflow

### Local Development with Docker Compose

1. Copy the override example:
   ```bash
   cp docker-compose.override.yml.example docker-compose.override.yml
   ```

2. Edit `docker-compose.override.yml` to customize your local environment

3. Start the development environment:
   ```bash
   docker-compose up
   ```

### Rebuilding After Code Changes

```bash
docker-compose up --build
```

Or rebuild without cache:
```bash
docker-compose build --no-cache
docker-compose up
```

### Accessing the Application

Once running, the application is available at:
- Health check: http://localhost:8080/health
- Main application: http://localhost:8080

### Viewing Logs

```bash
# All logs
docker-compose logs

# Follow logs in real-time
docker-compose logs -f

# Logs for specific service
docker-compose logs -f app
```

## Health Checks

The Docker Compose setup includes a health check that:
- Runs every 30 seconds
- Times out after 10 seconds
- Retries up to 3 times
- Waits 10 seconds after container start

View health status:
```bash
docker-compose ps
```

Or using Docker CLI:
```bash
docker inspect --format='{{.State.Health.Status}}' go-garage-app
```

## Production Deployment

### Building for Production

1. Set production environment variables:
   ```bash
   export ENVIRONMENT=production
   export JWT_SECRET=your-secret-key-here
   ```

2. Build and run:
   ```bash
   docker-compose -f docker-compose.yml up -d
   ```

### Security Considerations

- The container runs as a non-root user (`appuser`, UID 1000)
- Only port 8080 is exposed
- Secrets should be provided via environment variables or Docker secrets
- Database is stored in a named volume, separate from the container

### Resource Limits

For production, consider adding resource limits to `docker-compose.yml`:

```yaml
services:
  app:
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 256M
```

## Troubleshooting

### Container Won't Start

1. Check logs:
   ```bash
   docker-compose logs app
   ```

2. Verify port 8080 is available:
   ```bash
   lsof -i :8080
   ```

3. Check configuration:
   ```bash
   docker-compose config
   ```

### Database Issues

1. Check volume permissions:
   ```bash
   docker exec go-garage-app ls -la /app/data
   ```

2. Reset database (WARNING: deletes all data):
   ```bash
   docker-compose down -v
   docker-compose up -d
   ```

### Build Failures

1. Clear build cache:
   ```bash
   docker-compose build --no-cache
   ```

2. Check Docker version:
   ```bash
   docker --version
   docker-compose --version
   ```

3. Verify Go version in Dockerfile matches project requirements

## Multi-Stage Build Details

The Dockerfile uses a multi-stage build:

1. **Builder stage** (golang:1.24-alpine):
   - Installs build dependencies (gcc, musl-dev, sqlite-dev)
   - Downloads Go modules
   - Compiles the application with CGO enabled for SQLite support
   - Creates a statically-linked binary

2. **Runtime stage** (alpine:latest):
   - Minimal Alpine Linux base
   - Only includes CA certificates and SQLite runtime libraries
   - Creates non-root user for security
   - Copies compiled binary from builder stage
   - Final image size: ~30-40MB

## Clean Up

Remove containers, networks, and volumes:
```bash
docker-compose down -v
```

Remove images:
```bash
docker rmi go-garage
```

Complete cleanup (removes all Docker resources):
```bash
docker system prune -a --volumes
```

## Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Go-Garage README](../README.md)
