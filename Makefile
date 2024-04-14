# Start building local Docker images
.PHONY: dev
dev: check-env
	@echo "Start Temporal"
	temporal server start-dev >/dev/null 2>&1 &
	docker compose up --build --force-recreate --remove-orphans --detach
	@echo "Temporal Meetup Demo Started"
	@echo "Temporal UI http://localhost:8233/namespaces/default/workflows"
	@echo "curl localhost:8080/start to start a Temporal Worflow"

.PHONY: dev-no-worker
dev-no-worker: check-env
	@echo "Start Temporal"
	temporal server start-dev >/dev/null 2>&1 &
	docker compose -f docker-compose-no-worker.yml up --build --force-recreate --remove-orphans --detach
	@echo "Temporal Meetup Demo Started"
	@echo "Temporal UI http://localhost:8233/namespaces/default/workflows"
	@echo "curl localhost:8080/start to start a Temporal Worflow"

.PHONY: rebuild
rebuild: check-env
	@echo "Rebuild local Docker images"
	docker-compose build --no-cache

# Start using Docker images
.PHONY: start
start: check-env
	@echo "Start Temporal"
	temporal server start-dev >/dev/null 2>&1 &
	docker compose -f docker-compose.yml up --pull always --force-recreate --remove-orphans --detach
	@echo "Temporal Meetup Demo Started"
	@echo "Temporal UI http://localhost:8233/namespaces/default/workflows"
	@echo "curl localhost:8080/start to start a Temporal Worflow"

.PHONY: check-env
check-env:
	@echo "Checking for .env file"
	@test -f .env && echo ".env file found for OTel data configuration" || \
	(echo ".env file is needed to specify where to send OTel data, run make create-env"; exit 1)

.PHONY: create-env
create-env:
	@echo "Creating or overriding .env file"
	@read -s -p "Enter your Configuration API key: " apiKey; \
	sed "s/your_key_here/$$apiKey/" .env.example.honeycomb >.env
	@echo "\n.env file created or overridden with OTel data configuration"

.PHONY: stop
stop:
	@echo "Stopping Temporal"
	kill `pgrep -f "temporal server start-dev"`
	docker compose down --remove-orphans --volumes
	@echo ""
	@echo "Temporal Meetup Demo Stopped"

.PHONY: trigger
trigger:
	curl localhost:8080/start

.PHONY: service
service:
	curl -X POST http://localhost:8080/provision \
  -H "Content-Type: application/json" \
  -d '{ "name": "guestNetwork", "deviceMac": "00:A0:C1:D2:E3:F4" }'
