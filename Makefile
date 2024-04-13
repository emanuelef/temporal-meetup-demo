.PHONY: dev
dev: check-env
	@echo "Start Temporal"
	temporal server start-dev >/dev/null 2>&1 &
	docker compose up --build --force-recreate --remove-orphans --detach
	@echo "Temporal Meetup Demo Started"
	@echo "Temporal UI http://localhost:8233/namespaces/default/workflows"
	@echo "curl localhost:8080/start to start a Temporal Worflow"

.PHONY: rebuild
rebuild: check-env
	@echo "Rebuild local Docker images"
	docker-compose build --no-cache

.PHONY: start
start: check-env
	@echo "Start Temporal"
	temporal server start-dev >/dev/null 2>&1 &
	docker compose -f docker-compose-ghcr.yml up --force-recreate --remove-orphans --detach
	@echo "Temporal Meetup Demo Started"
	@echo "Temporal UI http://localhost:8233/namespaces/default/workflows"
	@echo "curl localhost:8080/start to start a Temporal Worflow"

.PHONY: check-env
check-env:
	@echo "Checking for .env file"
	@test -f .env && echo ".env file found for OTeL data configuration" || \
	(echo ".env file is needed to specify where to send OTeL data, run make create-env"; exit 1)

.PHONY: create-env
create-env:
	@echo "Creating or overriding .env file"
	@read -s -p "Enter your Configuration API key: " apiKey; \
	sed "s/your_key_here/$$apiKey/" .env.example >.env
	@echo "\n.env file created or overridden with OTeL data configuration"

.PHONY: stop
stop:
	@echo "Stopping Temporal"
	kill `pgrep -f "temporal server start-dev"` # Quite brutal way to stop it
	docker compose down --remove-orphans --volumes
	@echo ""
	@echo "Temporal Meetup Demo Stopped"

.PHONY: trigger
trigger:
	curl localhost:8080/start