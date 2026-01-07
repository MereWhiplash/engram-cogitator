.PHONY: build test lint docker kind-up kind-down helm-install helm-uninstall helm-template port-forward

# Go build
build:
	CGO_ENABLED=1 go build -o bin/server ./cmd/server
	CGO_ENABLED=1 go build -o bin/ec-api ./cmd/api
	CGO_ENABLED=1 go build -o bin/ec-shim ./cmd/shim

test:
	CGO_ENABLED=1 go test ./...

lint:
	helm lint charts/engram-cogitator

# Docker
docker:
	docker build -t engram-cogitator:local .

# Kind cluster
kind-up:
	kind create cluster --name engram
	@echo "Cluster ready. Run 'make helm-install' to deploy."

kind-down:
	kind delete cluster --name engram

# Helm
helm-install:
	helm install engram charts/engram-cogitator \
		--set storage.postgres.password=devpassword \
		--set image.repository=engram-cogitator \
		--set image.tag=local \
		--set image.pullPolicy=Never
	@echo "Watching pods... (Ctrl+C when all Running)"
	kubectl get pods -w

helm-upgrade:
	helm upgrade engram charts/engram-cogitator \
		--set storage.postgres.password=devpassword \
		--set image.repository=engram-cogitator \
		--set image.tag=local \
		--set image.pullPolicy=Never

helm-uninstall:
	helm uninstall engram

helm-template:
	helm template engram charts/engram-cogitator \
		--set storage.postgres.password=devpassword

helm-template-mongodb:
	helm template engram charts/engram-cogitator \
		--set storage.driver=mongodb \
		--set storage.mongodb.existingSecret=mongo-secret

# Development
port-forward:
	@echo "API available at http://localhost:8080"
	kubectl port-forward svc/engram-engram-cogitator-api 8080:8080

logs-api:
	kubectl logs -f -l app.kubernetes.io/component=api

logs-ollama:
	kubectl logs -f -l app.kubernetes.io/component=ollama

logs-postgres:
	kubectl logs -f -l app.kubernetes.io/component=postgres

status:
	@echo "=== Pods ==="
	@kubectl get pods
	@echo "\n=== Services ==="
	@kubectl get svc
	@echo "\n=== PVCs ==="
	@kubectl get pvc

# Kind helpers
kind-load:
	kind load docker-image engram-cogitator:local --name engram

# Full local dev cycle
dev: kind-up docker kind-load helm-install
	@echo "Run 'make port-forward' to access the API"

dev-down: helm-uninstall kind-down
