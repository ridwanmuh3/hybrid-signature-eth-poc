#!/bin/bash
# ============================================================================
# K6 Test Runner for Hybrid PQC Benchmark
# ============================================================================
# Prerequisites:
#   1. k6 installed:        https://grafana.com/docs/k6/latest/set-up/install-k6/
#   2. Hardhat node running: make run-node    (port 8540)
#   3. Contract deployed:    make run-deploy
#   4. Oracle server running: make run-oracle ADDR=<contract_address>
# ============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
RESULTS_DIR="${SCRIPT_DIR}/results"
BASE_URL="${BASE_URL:-http://localhost:9000}"

# Create results directory
mkdir -p "${RESULTS_DIR}"

echo "============================================"
echo "  Hybrid PQC K6 Benchmark Suite"
echo "============================================"
echo "  Oracle URL : ${BASE_URL}"
echo "  Results    : ${RESULTS_DIR}"
echo "============================================"
echo ""

# Health check
echo "[*] Checking oracle server connectivity..."
if ! curl -sf "${BASE_URL}/api/generate-key" -X POST \
    -H "Content-Type: application/json" \
    -d '{"algorithm":"ECDSA","mode":"single"}' > /dev/null 2>&1; then
    echo "[!] ERROR: Cannot reach oracle server at ${BASE_URL}"
    echo "    Make sure Hardhat node, contract, and oracle are running."
    exit 1
fi
echo "[+] Oracle server is reachable."
echo ""

# Parse arguments
TEST="${1:-all}"
VUS="${2:-10}"

run_benchmark() {
    echo "============================================"
    echo "  Running: Benchmark Test (Metrics 1,2,4,5)"
    echo "  Config : 1 VU, 101 iterations"
    echo "============================================"
    k6 run "${SCRIPT_DIR}/benchmark.js" \
        -e BASE_URL="${BASE_URL}" \
        --out json="${RESULTS_DIR}/benchmark_raw.json" \
        2>&1 | tee "${RESULTS_DIR}/benchmark_stdout.log"
    echo ""
    echo "[+] Benchmark complete. Report: ${RESULTS_DIR}/benchmark_report.json"
    echo ""
}

run_throughput() {
    echo "============================================"
    echo "  Running: Throughput Test (Metric 3 - TPS)"
    echo "  Config : ${VUS} VUs, 101 iterations each"
    echo "============================================"
    k6 run "${SCRIPT_DIR}/throughput.js" \
        -e BASE_URL="${BASE_URL}" \
        -e VUS="${VUS}" \
        --out json="${RESULTS_DIR}/throughput_raw.json" \
        2>&1 | tee "${RESULTS_DIR}/throughput_stdout.log"
    echo ""
    echo "[+] Throughput test complete. Report: ${RESULTS_DIR}/throughput_report.json"
    echo ""
}

case "${TEST}" in
    benchmark)
        run_benchmark
        ;;
    throughput)
        run_throughput
        ;;
    all)
        run_benchmark
        run_throughput
        ;;
    *)
        echo "Usage: $0 [benchmark|throughput|all] [vus]"
        echo ""
        echo "  benchmark   - Run sequential benchmark (keypair gen, sign/verify, payload, gas)"
        echo "  throughput  - Run concurrent TPS test"
        echo "  all         - Run both tests (default)"
        echo "  vus         - Number of VUs for throughput test (default: 10)"
        echo ""
        echo "Examples:"
        echo "  $0                          # Run all tests"
        echo "  $0 benchmark                # Run benchmark only"
        echo "  $0 throughput 5             # Run throughput with 5 VUs"
        echo "  BASE_URL=http://host:9000 $0 all"
        exit 1
        ;;
esac

echo "============================================"
echo "  All tests completed!"
echo "  Results saved in: ${RESULTS_DIR}/"
echo "============================================"
