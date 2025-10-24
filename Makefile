.PHONY: build up down logs test fmt vet

build:
	@docker-compose build
	@echo "Build complete."

up:
	@docker-compose up -d --remove-orphans
	@echo "Services are up and running."

down:
	@docker-compose down
	@echo "Services have been stopped."

logs:
	@docker-compose logs -f
	@echo "Displaying logs."

