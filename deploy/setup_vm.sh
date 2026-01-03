#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting setup for demo_wallet deployment...${NC}"

# 1. Update system
echo -e "${GREEN}Updating system packages...${NC}"
sudo apt-get update
sudo apt-get upgrade -y

# 2. Install Docker (Official Docker setup for Debian/Ubuntu)
# https://docs.docker.com/engine/install/debian/
if ! command -v docker &> /dev/null; then
    echo -e "${GREEN}Installing Docker...${NC}"
    
    # Add Docker's official GPG key:
    sudo apt-get install -y ca-certificates curl gnupg
    sudo install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    sudo chmod a+r /etc/apt/keyrings/docker.gpg

    # Add the repository to Apt sources:
    echo \
      "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
      $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
      sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

    sudo apt-get update
    sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
    
    # Verify installation
    sudo docker run --rm hello-world
    echo -e "${GREEN}Docker installed successfully.${NC}"
else
    echo -e "${GREEN}Docker is already installed.${NC}"
fi

# 3. Configure Swap (Required for e2-micro's 1GB RAM)
if [ $(sudo swapon --show | wc -l) -eq 0 ]; then
    echo -e "${GREEN}Configuring 4GB Swap file (Required for e2-micro)...${NC}"
    sudo fallocate -l 4G /swapfile
    sudo chmod 600 /swapfile
    sudo mkswap /swapfile
    sudo swapon /swapfile
    echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
    echo -e "${GREEN}Swap configured.${NC}"
else
     echo -e "${GREEN}Swap is already configured.${NC}"
fi

# 4. Final check
echo -e "${GREEN}Setup complete!${NC}"
echo -e "You can now run: ${GREEN}docker compose -f docker-compose.yaml -f docker-compose.prod.yaml up -d${NC}"
echo -e "Make sure to export REGISTRY_PREFIX first!"
