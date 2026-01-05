APP_DIR := go-backup-app
FRONTEND_DIR := $(APP_DIR)/frontend

IMAGE ?= ccbackup
TAG ?= local
PLATFORMS ?= linux/amd64,linux/arm64
OCI_OUT ?= dist/$(IMAGE)_$(TAG).oci.tar

.PHONY: help deps frontend-install frontend-build go-test tidy dev build test docker-build docker-buildx-oci docker-buildx-push clean

help:
	@echo "Targets:"
	@echo "  deps             - Download Go modules"
	@echo "  frontend-install - Install frontend deps (npm ci)"
	@echo "  frontend-build   - Build frontend (vite)"
	@echo "  go-test          - Run Go tests"
	@echo "  tidy             - go mod tidy"
	@echo "  dev              - Run wails dev"
	@echo "  build            - Run wails build"
	@echo "  test             - Run go-test + frontend-build"
	@echo "  docker-build     - Build Docker image (single-platform)"
	@echo "  docker-buildx-oci  - Build multi-platform OCI tar (buildx)"
	@echo "  docker-buildx-push - Build+push multi-platform image (buildx)"
	@echo "  clean            - Remove build artifacts"

deps:
	cd $(APP_DIR) && go mod download

frontend-install:
	cd $(FRONTEND_DIR) && npm ci

frontend-build:
	cd $(FRONTEND_DIR) && npm run build

go-test:
	cd $(APP_DIR) && go test ./...

tidy:
	cd $(APP_DIR) && go mod tidy

dev:
	cd $(APP_DIR) && wails dev

build:
	cd $(APP_DIR) && wails build

test: go-test frontend-build

docker-build:
	docker build -t $(IMAGE):$(TAG) $(APP_DIR)

docker-buildx-oci:
	mkdir -p dist
	docker buildx build --platform $(PLATFORMS) -t $(IMAGE):$(TAG) --output type=oci,dest=$(OCI_OUT) $(APP_DIR)

docker-buildx-push:
	docker buildx build --platform $(PLATFORMS) -t $(IMAGE):$(TAG) --push $(APP_DIR)

clean:
	rm -rf $(APP_DIR)/build/bin $(FRONTEND_DIR)/dist
