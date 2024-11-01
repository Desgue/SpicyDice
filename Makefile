IMAGE_NAME ?= spicy_dice_server

# Default target
all: build run

# Build the Docker image
build:
	@echo "Building Docker image..."
	docker build -t $(IMAGE_NAME) .

# Run the Docker image
run:
	@echo "Running your docker image..."
	docker run -p 8080:8080 $(IMAGE_NAME)

# Clean up local Docker images (optional target)
clean:
	@echo "Cleaning up local Docker images..."
	docker rmi $(IMAGE_NAME):latest || true