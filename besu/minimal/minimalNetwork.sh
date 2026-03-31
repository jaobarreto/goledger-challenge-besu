#!/bin/bash

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/../colors.sh"

BESU=./bin/besu

wait_for_rpc() {
    local node_name="$1"
    local rpc_port="$2"
    local attempts=60
    local count=0

    echo -e "${BLUE}Waiting for ${node_name} to be responsive...${NC}"
    until curl -s -X POST --data '{"jsonrpc":"2.0","method":"web3_clientVersion","params":[],"id":1}' -H "Content-Type: application/json" "http://localhost:${rpc_port}/" > /dev/null 2>&1; do
        printf '.'
        sleep 2
        count=$((count + 1))
        if [ "$count" -ge "$attempts" ]; then
            echo -e "\n${RED}${node_name} did not become responsive on port ${rpc_port}.${NC}"
            docker logs --tail 50 "$node_name" || true
            exit 1
        fi
    done
    echo -e "\n${GREEN}${node_name} is responsive!${NC}\n"
}

wait_for_enode() {
    local attempts=30
    local count=0
    local enode_response
    local enode_url

    echo -e "${BLUE}Waiting for besu.node-1 P2P to be ready...${NC}" >&2
    while true; do
        enode_response=$(curl -s -X POST --data '{"jsonrpc":"2.0","method":"net_enode","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545/)
        enode_url=$(echo "$enode_response" | jq -r '.result // empty')

        if [[ "$enode_url" =~ ^enode:// ]]; then
            echo "$enode_url"
            return 0
        fi

        printf '.' >&2
        sleep 2
        count=$((count + 1))
        if [ "$count" -ge "$attempts" ]; then
            echo -e "\n${RED}Could not determine a valid enode URL from besu.node-1.${NC}" >&2
            echo -e "${YELLOW}Last JSON-RPC response:${NC} $enode_response" >&2
            docker logs --tail 50 besu.node-1 || true
            exit 1
        fi
    done
}

mkdir tmp && cd tmp
../$BESU operator generate-blockchain-config --config-file=../minimal/config.json --to=network --private-key-file-name=key

cd ..

counter=1
for folder in tmp/network/keys/*; do
    mkdir -p "nodes/node-$counter/data"
    cp -r "$folder"/* "nodes/node-$counter/data"
    ((counter++))
done

mkdir -p genesis
cp tmp/network/genesis.json genesis/genesis.json

if [ "$OS" = "Darwin" ]; then
    rm -rf tmp
else
    sudo rm -rf tmp
fi
echo

echo -e "${BLUE}Starting docker network 'besu_test_network'...${NC}"
docker network create --driver bridge besu_test_network
if [ $? -eq 0 ]; then
    echo -e "${GREEN}Docker network created successfully.${NC}\n"
else
    echo -e "${YELLOW}Docker network may already exist. Continuing...${NC}\n"
fi

echo -e "${BLUE}Starting besu.node-1 on docker...${NC}"
docker run -d \
    --name besu.node-1 \
    --user root \
    -v "$(pwd)/nodes/node-1/data:/opt/besu/data" \
    -v "$(pwd)/genesis:/opt/besu/genesis" \
    -v "$(pwd)/minimal/config.toml:/opt/besu/config.toml" \
    -p 30303:30303 \
    -p 8545:8545 \
    -p 8546:8546 \
    -p 30303:30303/udp \
    -p 8545:8545/udp \
    --network besu_test_network \
    --restart always \
    hyperledger/besu:25.4.1 \
    --config-file=/opt/besu/config.toml

echo

wait_for_rpc "besu.node-1" 8545
ENODE_URL=$(wait_for_enode)

printf '%s\n' "$ENODE_URL" > minimal/bootnodes.txt

HOST_IP=$(docker container inspect besu.node-1 | jq -r '.[0].NetworkSettings.Networks.besu_test_network.IPAddress')
sed -i.bak -E "s/(127\.0\.0\.1|0\.0\.0\.0)/$HOST_IP/g" minimal/bootnodes.txt

BOOTNODE_URL=$(cat ./minimal/bootnodes.txt)

echo -e "${BLUE}Starting besu.node-2 on docker...${NC}"
docker run -d \
    --name besu.node-2 \
    --user root \
    -v "$(pwd)/nodes/node-2/data:/opt/besu/data" \
    -v "$(pwd)/genesis:/opt/besu/genesis" \
    -v "$(pwd)/minimal/config.toml:/opt/besu/config.toml" \
    -v "$(pwd)/minimal/bootnodes.txt:/opt/besu/bootnodes.txt" \
    -p 30304:30303 \
    -p 8547:8545 \
    -p 8548:8546 \
    -p 30304:30303/udp \
    -p 8547:8545/udp \
    --network besu_test_network \
    --restart always \
    hyperledger/besu:25.4.1 \
    --config-file=/opt/besu/config.toml --bootnodes="$BOOTNODE_URL"

echo

wait_for_rpc "besu.node-2" 8547

echo -e "${BLUE}Starting besu.node-3 on docker...${NC}"
docker run -d \
    --name besu.node-3 \
    --user root \
    -v "$(pwd)/nodes/node-3/data:/opt/besu/data" \
    -v "$(pwd)/genesis:/opt/besu/genesis" \
    -v "$(pwd)/minimal/config.toml:/opt/besu/config.toml" \
    -v "$(pwd)/minimal/bootnodes.txt:/opt/besu/bootnodes.txt" \
    -p 30305:30303 \
    -p 8549:8545 \
    -p 8550:8546 \
    -p 30305:30303/udp \
    -p 8549:8545/udp \
    --network besu_test_network \
    --restart always \
    hyperledger/besu:25.4.1 \
    --config-file=/opt/besu/config.toml --bootnodes="$BOOTNODE_URL"

echo

wait_for_rpc "besu.node-3" 8549

echo -e "${BLUE}Starting besu.node-4 on docker...${NC}"
docker run -d \
    --name besu.node-4 \
    --user root \
    -v "$(pwd)/nodes/node-4/data:/opt/besu/data" \
    -v "$(pwd)/genesis:/opt/besu/genesis" \
    -v "$(pwd)/minimal/config.toml:/opt/besu/config.toml" \
    -v "$(pwd)/minimal/bootnodes.txt:/opt/besu/bootnodes.txt" \
    -p 30306:30303 \
    -p 8551:8545 \
    -p 8552:8546 \
    -p 30306:30303/udp \
    -p 8551:8545/udp \
    --network besu_test_network \
    --restart always \
    hyperledger/besu:25.4.1 \
    --config-file=/opt/besu/config.toml --bootnodes="$BOOTNODE_URL"

echo

wait_for_rpc "besu.node-4" 8551

echo -e "${GREEN}============================="
echo -e "Network started successfully!"
echo -e "=============================${NC}\n"