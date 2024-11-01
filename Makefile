# Image and container names
SERVER_IMAGE ?= spicy_dice_server
DB_IMAGE ?= spicy_dice_db

# SERVER ONLY OPERATIONS
.PHONY: all-server
all-server: build-server run-server

.PHONY: build-server
build-server:
	@echo "Building server Docker image..."
	docker build  -t $(SERVER_IMAGE) .

.PHONY: run-server
run-server:
	@echo "Running server container..."
	docker run --name $(SERVER_IMAGE) -p 8080:8080 $(SERVER_IMAGE)

.PHONY: clean-server
clean-server:
	@echo "Stopping and cleaning up server container and image..."
	docker stop $(SERVER_IMAGE) || true
	docker rm $(SERVER_IMAGE) || true
	docker rmi $(SERVER_IMAGE):latest || true

# DATABASE ONLY OPERATIONS
.PHONY: all-db
all-db: build-db run-db

.PHONY: build-db
build-db:
	@echo "Building database Docker image..."
	docker build -t $(DB_IMAGE) ./postgres	

.PHONY: run-db
run-db:
	@echo "Running database container..."
	docker run --name $(DB_IMAGE) -p 5432:5432 $(DB_IMAGE)

.PHONY: clean-db
clean-db:
	@echo "Stopping and cleaning up database container and image..."
	docker stop $(DB_IMAGE) || true
	docker rm $(DB_IMAGE) || true
	docker rmi $(DB_IMAGE):latest || true

# DOCKER COMPOSE OPERATIONS
.PHONY: up
up:
	@echo "Starting all services with Docker Compose..."
	docker-compose up --build

.PHONY: up-d
up-d:
	@echo "Starting all services with Docker Compose in detached mode..."
	docker-compose up --build -d

.PHONY: down
down:
	@echo "Stopping all services..."
	docker-compose down

.PHONY: clean
clean: down
	@echo "Cleaning up Docker images..."
	docker rmi $(SERVER_IMAGE):latest || true
	docker rmi $(DB_IMAGE):latest || true
	docker-compose rm -f


# HELPERS
.PHONY: ps
ps:
	@echo "Showing running containers..."
	docker-compose ps

.PHONY: restart
restart: down up

.PHONY: logs
logs:
	@echo "Showing logs for all services..."
	docker-compose logs -f

	.PHONY: logs-server
logs-server:
	@echo "Showing logs for server..."
	docker-compose logs -f app

.PHONY: logs-db
logs-db:
	@echo "Showing logs for database..."
	docker-compose logs -f db