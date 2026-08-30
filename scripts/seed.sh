#!/usr/bin/env bash
set -euo pipefail

# ── Configuration ────────────────────────────────────────────────────
ANVIL_RPC="${ANVIL_RPC:-http://anvil:8540}"
ORACLE_KEY="${ORACLE_KEY:-0x2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6}"
SENDER_KEY="${SENDER_KEY:-0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80}"

ORACLE_ADDR="0xa0Ee7A142d267C1f36714E4a8F75612F20a79720"
SENDER_ADDR="0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"

echo "⏳ Waiting for Anvil at ${ANVIL_RPC}..."
for i in $(seq 1 30); do
    if curl -sf -X POST "${ANVIL_RPC}" \
        -H "Content-Type: application/json" \
        -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
        > /dev/null 2>&1; then
        echo "✅ Anvil is ready"
        break
    fi
    if [ "$i" -eq 30 ]; then
        echo "❌ Anvil did not start in time"
        exit 1
    fi
    sleep 1
done

# ── Deploy contract ──────────────────────────────────────────────────
echo "📦 Deploying Transactions.sol..."
cd /workspace/apps/ethereum

CONTRACT_ADDR=$(forge script script/Deploy.s.sol:Deploy \
    --rpc-url "${ANVIL_RPC}" \
    --broadcast \
    --private-key "${ORACLE_KEY}" \
    2>&1 | grep -oP 'Transactions deployed at: \K0x[a-fA-F0-9]{40}' | head -1)

if [ -z "${CONTRACT_ADDR}" ]; then
    echo "❌ Contract deployment failed"
    exit 1
fi
echo "✅ Contract deployed at: ${CONTRACT_ADDR}"

# ── Write oracle .env ───────────────────────────────────────────────
echo "📝 Writing oracle .env..."
cat > /workspace/apps/oracle/.env <<EOF
ETHEREUM_URL=${ANVIL_RPC}
CHAIN_ID=31337
ORACLE_PRIVATE_KEY=${ORACLE_KEY}
ORACLE_ADDRESS=${ORACLE_ADDR}
CONTRACT_ADDRESS=${CONTRACT_ADDR}
TEST_SENDER_ADDRESS=${SENDER_ADDR}
TEST_RECEIVER_ADDRESS=0x70997970C51812dc3A010C7d01b50e0d17dc79C8
TEST_ENABLED=true
EOF

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "  🚀 Seed complete!"
echo ""
echo "  Contract:  ${CONTRACT_ADDR}"
echo "  Oracle:    ${ORACLE_ADDR}"
echo "  Sender:    ${SENDER_ADDR}"
echo "  Anvil RPC: ${ANVIL_RPC}"
echo ""
echo "  Client .env:"
echo "    VITE_CONTRACT_ADDRESS=${CONTRACT_ADDR}"
echo "    VITE_ORACLE_URL=http://localhost:9000"
echo "    VITE_ETHEREUM_RPC_URL=http://localhost:8540"
echo "═══════════════════════════════════════════════════════════════"
