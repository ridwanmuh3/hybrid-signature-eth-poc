# Hybrid PQS PoC — Root Justfile
# Run `just` to see all available recipes.

set dotenv-load

# ── Local development ────────────────────────────────────────────────

# Start Anvil local Ethereum node (port 8540)
run-node:
    anvil --port 8540 --block-gas-limit 50000000

# Start the client dev server (Vite)
run-client:
    cd apps/web && pnpm dev

# ── Contract workflow ────────────────────────────────────────────────

# Compile Transactions.sol and generate Go ABI bindings
compile-contract:
    cd apps/oracle && just compile-contract

# Deploy contract via Foundry script
# Usage: just forge-deploy KEY=0x... ADDR=0x...
forge-deploy KEY ADDR:
    cd apps/ethereum && ORACLE_ADDRESS={{ADDR}} forge script script/Deploy.s.sol:Deploy \
        --rpc-url http://localhost:8540 --broadcast --private-key {{KEY}}

# ── Oracle server ────────────────────────────────────────────────────

# Start oracle server (dev mode, hot reload)
run-oracle:
    cd apps/oracle && just run

# Start oracle server with a specific contract address
# Usage: just run-oracle-with ADDR=0x...
run-oracle-with ADDR:
    @if [ -z "{{ADDR}}" ]; then echo "❌ ADDR is required. Usage: just run-oracle-with ADDR=0x..."; exit 1; fi
    sed -i 's/^CONTRACT_ADDRESS=.*/CONTRACT_ADDRESS={{ADDR}}/' ./apps/oracle/.env
    cd apps/oracle && just run

# ── Docker Compose ──────────────────────────────────────────────────

# Start full stack via Docker Compose (Anvil + Oracle + Client)
dev:
    docker compose up --build

# Start full stack in detached mode
dev-detach:
    docker compose up --build -d

# Stop all containers
dev-down:
    docker compose down

# View logs from all services
dev-logs:
    docker compose logs -f

# ── Testing ──────────────────────────────────────────────────────────

# Run all security tests (Go + Solidity)
test-security: _test-security-go _test-security-foundry

_test-security-go:
    cd apps/oracle && go test ./test/... -run TestSecurity_ -v -count=1

_test-security-foundry:
    cd apps/ethereum && forge test --match-test "test_Security" -v

# Run full test suite
test-all: _test-security-go _test-security-foundry
    cd apps/oracle && go test ./test/... -v -count=1 -timeout 300s

# ── Benchmarks ───────────────────────────────────────────────────────

BENCH_TIMEOUT := "300s"
BENCH_TIME := "10s"
ONCHAIN_RPC := ""

# Run performance benchmarks (crypto layer only)
benchmark:
    mkdir -p apps/oracle/results
    bash -o pipefail -c "cd apps/oracle && go test ./test/... -count=1 -run TestFunctional_Performance -v -timeout {{BENCH_TIMEOUT}} 2>&1 | tee apps/oracle/results/functional_performance.txt"
    bash -o pipefail -c "cd apps/oracle && go test ./test/... -count=1 -run TestFunctional_KeyGen -v -timeout {{BENCH_TIMEOUT}} 2>&1 | tee apps/oracle/results/keygen.txt"
    bash -o pipefail -c "cd apps/oracle && go test ./test/... -count=1 -bench=. -benchtime={{BENCH_TIME}} -benchmem -run '^$$' 2>&1 | tee apps/oracle/results/go_benchmarks.txt"

# Run on-chain gas + latency benchmarks (requires running Anvil)
benchmark-onchain:
    @if [ -z "{{ONCHAIN_RPC}}" ]; then echo "❌ ONCHAIN_RPC required. Usage: just benchmark-onchain ONCHAIN_RPC=http://localhost:8540"; exit 1; fi
    mkdir -p apps/oracle/results
    bash -o pipefail -c "cd apps/oracle && ONCHAIN_RPC={{ONCHAIN_RPC}} go test ./test/... -count=1 -run TestOnchain_GasAndLatency -v -timeout 600s 2>&1 | tee apps/oracle/results/onchain_gas_latency.txt"

# Run benchmarks with on-chain measurements
benchmark-full:
    mkdir -p apps/oracle/results
    bash -o pipefail -c "cd apps/oracle && go test ./test/... -count=1 -run TestFunctional_Performance -v -timeout {{BENCH_TIMEOUT}} 2>&1 | tee apps/oracle/results/functional_performance.txt"
    bash -o pipefail -c "cd apps/oracle && go test ./test/... -count=1 -run TestFunctional_KeyGen -v -timeout {{BENCH_TIMEOUT}} 2>&1 | tee apps/oracle/results/keygen.txt"
    bash -o pipefail -c "cd apps/oracle && go test ./test/... -count=1 -bench=. -benchtime={{BENCH_TIME}} -benchmem -run '^$$' 2>&1 | tee apps/oracle/results/go_benchmarks.txt"
    @if [ -n "{{ONCHAIN_RPC}}" ]; then \
        bash -o pipefail -c "cd apps/oracle && ONCHAIN_RPC={{ONCHAIN_RPC}} go test ./test/... -count=1 -run TestOnchain_GasAndLatency -v -timeout 600s 2>&1 | tee apps/oracle/results/onchain_gas_latency.txt"; \
    else \
        echo "ℹ️  Skipping on-chain benchmark. Set ONCHAIN_RPC=http://localhost:8540 to include it."; \
    fi

# ── Utilities ────────────────────────────────────────────────────────

# Generate a test wallet keypair
genkey:
    cd apps/oracle && just genkey-contract

# Show help
help:
    @just --list
