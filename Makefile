PROTO_DIR = proto
PB_DIR = pb
DOCS_DIR = docs

directories:
	mkdir -p $(PB_DIR)/subledger
	mkdir -p $(PB_DIR)/accounts
	mkdir -p $(DOCS_DIR)

proto: directories
	@echo "Generating protobuf files..."
	@protoc --proto_path=$(PROTO_DIR) \
	       --go_out=$(PB_DIR) --go_opt=paths=source_relative \
	       --go-grpc_out=$(PB_DIR) --go-grpc_opt=paths=source_relative \
	       $(PROTO_DIR)/subledger/*.proto
	@protoc --proto_path=$(PROTO_DIR) \
	       --go_out=$(PB_DIR) --go_opt=paths=source_relative \
	       --go-grpc_out=$(PB_DIR) --go-grpc_opt=paths=source_relative \
	       $(PROTO_DIR)/accounts/*.proto
	@echo "Protobuf generation complete"

swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/gateway/main.go -o docs --parseInternal
	@echo "Swagger docs generated at http://localhost:8080/swagger/index.html"

gen: proto swagger

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

docker-clean:
	docker compose down -v

help:
	@echo "Available commands:"
	@echo "  make proto         - Generate protobuf files"
	@echo "  make swagger       - Generate Swagger documentation"
	@echo "  make gen           - Generate both proto and swagger"
	@echo "  make docker-up     - Start all services with docker compose"
	@echo "  make docker-down   - Stop all services"
	@echo "  make docker-logs   - View docker logs"
	@echo "  make docker-clean  - Stop services and remove volumes"