#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

if [ -z "$1" ]; then
    echo -e "${RED}Error: Registry prefix is required.${NC}"
    echo "Usage: $0 <registry-prefix>"
    echo "Example: $0 myuser (for Docker Hub) or gcr.io/my-project"
    exit 1
fi

REGISTRY_PREFIX=$1
TAG=${2:-latest}

echo -e "${GREEN}Building images for registry prefix: ${REGISTRY_PREFIX}...${NC}"

# Ensure we are at the project root
if [ ! -f "docker-compose.yaml" ]; then
    echo -e "${RED}Error: Please run this script from the project root.${NC}"
    exit 1
fi

# Function to build and push
build_and_push() {
    SERVICE_NAME=$1
    SERVICE_PATH=$2
    IMAGE_NAME="${REGISTRY_PREFIX}/${SERVICE_NAME}:${TAG}"

    echo -e "${GREEN}Building and pushing ${IMAGE_NAME}...${NC}"
    
    # Use buildx to ensure linux/amd64 compatibility (important if building on Apple Silicon)
    docker buildx build --platform linux/amd64 \
        --build-arg SERVICE_PATH=${SERVICE_PATH} \
        -f build/Dockerfile \
        -t ${IMAGE_NAME} \
        --push .
}

# Build services
# Note: SERVICE_PATH must match the directory in cmd/
build_and_push "gateway" "gateway"
build_and_push "accounts" "accounts"
build_and_push "subledger" "subledger"

echo -e "${GREEN}All images built and pushed successfully!${NC}"
echo -e "You can now deploy on your VM using:"
echo -e "export REGISTRY_PREFIX=${REGISTRY_PREFIX}"
echo -e "docker compose -f docker-compose.yaml -f docker-compose.prod.yaml up -d"
