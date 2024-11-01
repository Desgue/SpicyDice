# Deployment Guide

## Prerequisites
- Docker
- Docker Compose
- Make (optional)

## Quick Start with Make

### Full Application Deployment
```bash
# Build and start entire application
make up
```

### Alternative Deployment Options
```bash
# Start in background (detached mode)
make up-d

# Stop application
make down

# Restart application
make restart

# Restart application cleaning the volume
make restart-v
```

## Manual Docker Compose Deployment

### Standard Deployment
```bash
# Build and start services
docker-compose up --build

# Start in background
docker-compose up --build -d
```

## Service Management Commands

### With Make
```bash
# View running containers
make ps

# View all logs
make logs

# View server logs
make logs-server

# View database logs
make logs-db
```

### With Docker Compose
```bash
# View running containers
docker-compose ps

# View logs
docker-compose logs

# Stop and remove all services
docker-compose down
```

## Access Points
- Web Application: `http://localhost:8080`
- Database: `localhost:5432`
  - Username: postgres
  - Password: p4ssw0rd

## Cleanup
```bash
# Remove containers, networks, and images
make clean
```

**Note**: Credentials are for development. Modify for production use.
