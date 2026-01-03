#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting setup for demo_wallet deployment...${NC}"

echo -e "${GREEN}Updating system packages...${NC}"
sudo apt-get update
sudo apt-get upgrade -y

echo -e "${GREEN}Removing conflicting Docker packages...${NC}"
for pkg in docker.io docker-doc docker-compose docker-compose-v2 podman-docker containerd runc; do 
    sudo apt-get remove -y $pkg || true
done

echo -e "${GREEN}Installing Official Docker Engine...${NC}"
sudo apt-get install -y ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor --yes -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
echo -e "${GREEN}Verifying Docker Compose version...${NC}"
docker compose version

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

echo -e "${GREEN}Setup complete!${NC}"
echo -e "You can now run: ${GREEN}docker compose --env-file .env -f docker-compose.yaml -f docker-compose.prod.yaml up -d${NC}"
echo -e "Make sure to create .env and export REGISTRY_PREFIX first!"