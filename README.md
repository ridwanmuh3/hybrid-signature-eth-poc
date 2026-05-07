# Hybrid Post-Quantum Signature on Ethereum

Proof-of-concept implementation of a hybrid **ECDSA + PQC** transaction signature scheme on an EVM blockchain, using an oracle/relayer architecture. Designed to support a SINTA-3 journal article revision.

---

## Table of Contents

1. [What This Experiment Proves](#1-what-this-experiment-proves)
2. [System Architecture](#2-system-architecture)
3. [Prerequisites](#3-prerequisites)
4. [Quick Start — Four Steps](#4-quick-start--four-steps)
5. [Running Performance Benchmarks](#5-running-performance-benchmarks)
6. [Running Security Tests](#6-running-security-tests)
7. [API Reference](#7-api-reference)
8. [Wire Format — TLV Encoding](#8-wire-format--tlv-encoding)
9. [Scheme Design Decisions](#9-scheme-design-decisions)
10. [Known Limitations](#10-known-limitations)
11. [Reproducing Published Results](#11-reproducing-published-results)
12. [File Map](#12-file-map)

---

## 1. What This Experiment Proves

**Research question:** Can a hybrid ECDSA + post-quantum cryptography (PQC) signature scheme be deployed on Ethereum with acceptable overhead?

**Answer demonstrated here:**

| Claim | Evidence |
|---|---|
| Hybrid signing is feasible on Ethereum | Oracle signs with ECDSA + PQC; contract verifies ECDSA on-chain via `ecrecover` |
| PQC overhead is bounded | Go benchmark: +6.5% sign latency, +63% verify latency vs. pure ECDSA |
| On-chain gas impact is modest | +15.8% gas vs. ECDSA (`keccak256(pqSig)` storage caps on-chain cost) |
| Scheme is quantum-resistant in the relevant threat model | Security if ECDSA *or* PQC is secure (Bindel et al. 2019 "at-least-one" theorem) |
| Attack surface is tested | 47 security tests across reentrancy, replay, front-running, malicious relayer, signature forgery |

**Supported PQC algorithms (NIST FIPS 2024):**
- ML-DSA-44, ML-DSA-65, ML-DSA-87 (FIPS 204 — Dilithium)
- Falcon-512, Falcon-1024
- SLH-DSA variants (FIPS 205 — SPHINCS+)

---

## 2. System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  CLIENT (React/TypeScript)                                  │
│  Submits tx request with private key → Oracle REST API      │
└───────────────────────┬─────────────────────────────────────┘
                        │ HTTP
                        ▼
┌─────────────────────────────────────────────────────────────┐
│  ORACLE SERVER (Go + Fiber + liboqs)                        │
│                                                             │
│  1. Fetch current nonce from contract                       │
│  2. Build message hash (keccak256, domain-prefixed)         │
│  3. Sign with ECDSA (secp256k1) → ecdsaSig (65 B)          │
│  4. Sign with PQC (liboqs)       → pqcSig  (variable)      │
│  5. Encode finalSig as TLV blob                             │
│  6. Submit raw ecdsaSig + pqcSig to contract via oracle key │
│  7. Wait for mining, return tx hash + signature             │
└───────────────────────┬─────────────────────────────────────┘
                        │ JSON-RPC / go-ethereum
                        ▼
┌─────────────────────────────────────────────────────────────┐
│  SMART CONTRACT — Transactions.sol (Solidity 0.8.31)        │
│                                                             │
│  • require(msg.sender == ORACLE)                            │
│  • require(nonce > userNonce[sender])        ← replay guard │
│  • if (ecdsaSig.length == 65) verifySignature()  ← ECDSA   │
│  • balances[sender] -= amount                               │
│  • balances[receiver] += amount                             │
│  • store keccak256(pqSig) on-chain          ← PQC anchor   │
└─────────────────────────────────────────────────────────────┘
                        │ Ethereum local node (Anvil)
```

**Key design choice:** PQC verification is off-chain (oracle). On-chain ECDSA only. `keccak256(pqSig)` is stored as a tamper-evident anchor that a future quantum-capable verifier can use to confirm the PQC signature.

---

## 3. Prerequisites

| Tool | Version | Purpose |
|---|---|---|
| Go | ≥ 1.25 | Oracle server |
| liboqs-go | system lib | PQC operations |
| Foundry (`forge`, `anvil`) | latest | Contract compile, deploy, test |
| Bun | ≥ 1.0 | Client dev server |
| abigen | go-ethereum tooling | ABI → Go bindings |

Install liboqs (Ubuntu/Debian):
```bash
sudo apt install liboqs-dev
```

Install Foundry:
```bash
curl -L https://foundry.paradigm.xyz | bash && foundryup
```

---

## 4. Quick Start — Four Steps

### Step 1 — Start local Ethereum node

```bash
# Uses Anvil (comes with Foundry). Port 8540, high gas limit for PQC calldata.
make run-node
# or: task run-node
```

Leave this terminal running. Default accounts and keys are printed on startup.
**Oracle key** = Anvil account #9: `0x2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6`

### Step 2 — Compile and deploy the contract

```bash
# Compile Transactions.sol → generate Go ABI bindings
make compile-contract

# Deploy. KEY = oracle's private key. ADDR = oracle's address.
make forge-deploy \
  KEY=0x2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6 \
  ADDR=0xa0Ee7A142d267C1f36714E4a8F75612F20a79720
```

Copy the deployed contract address from the output (e.g., `0x700b6A60...`).

### Step 3 — Configure and start the oracle server

```bash
# Paste the contract address from step 2
make run-oracle ADDR=0x700b6A60ce7EaaEA56F065753d8dcB9653dbAD35
```

Oracle listens on `http://localhost:3000`.

### Step 4 — (Optional) Start the client UI

```bash
make run-client
```

Client runs at `http://localhost:5173`.

---

## 5. Running Performance Benchmarks

**Use the Go native test suite.** It directly measures the cryptographic layer without HTTP overhead, giving clean sub-millisecond precision per algorithm.

### 5.1 Sign + Verify latency and TPS (all algorithms)

```bash
cd oracle-server
go test ./test/... -run TestFunctional_Performance -v -timeout 300s
```

Output: tabwriter report with columns — key sizes, sign avg (ms), verify avg (ms), round-trip (ms), TPS, sig payload (bytes), estimated gas.

### 5.2 Key generation time only

```bash
cd oracle-server
go test ./test/... -run TestFunctional_KeyGen -v -timeout 300s
```

### 5.3 Go standard benchmarks (sub-microsecond, `-bench` mode)

```bash
cd oracle-server
go test ./test/... -bench=. -benchtime=10s -benchmem
```

Runs `BenchmarkKeyGen`, `BenchmarkSign`, `BenchmarkVerify` per algorithm.

### 5.4 Sign+verify correctness across all algorithms

```bash
cd oracle-server
go test ./test/... -run TestFunctional_SignVerify_Correctness -v
```

### Benchmark configuration (edit in `test/functional_test.go`)

| Constant | Default | Meaning |
|---|---|---|
| `fnWarmup` | 3 | Iterations discarded before measurement |
| `fnIterations` | 50 | Measured iterations |
| `fnTpsDur` | 5s | TPS measurement window |
| `fnWorkers` | 4 | Concurrent goroutines for TPS |

### Reading the report

The `TestFunctional_Performance` output prints four tables:

1. **Key generation** — privkey size (B), pubkey size (B), avg keygen (ms)
2. **Sign/verify latency** — ECDSA sig (B), PQC sig (B), total (B), sign (ms), verify (ms)
3. **Throughput** — TPS and latency/tx (ms)
4. **Gas estimation** — calldata gas (EIP-2028 model), ETH cost, USD cost at 30 gwei / $3000

---

## 6. Running Security Tests

### 6.1 Oracle-layer security tests (Go)

```bash
cd oracle-server
go test ./test/... -run TestSecurity -v -count=1
```

47 tests, grouped by threat category:

| Group | Coverage |
|---|---|
| A — Signature stripping | Drop ECDSA or PQC component from hybrid sig |
| B — Key/algorithm substitution | Wrong public key, algorithm downgrade |
| C — Message integrity | Tampered sig bit, wrong hash |
| D — Mode confusion | Single sig as hybrid, hybrid sig as single |
| E — Cross-key mixing | **KNOWN LIMITATION** documented: no PQC key commitment |
| F — Replay / nonce | Different nonce → different hash → sig invalid |
| G — Positive sanity | Valid hybrid/single must verify |
| H — Cross-sender replay | Sig for sender A reused as sender B |
| I — Front-running | Tampered receiver/amount hash, concurrent isolation |
| J — Malicious relayer | Oracle signing wrong hash, own-key forgery, nonce replay |
| K — Signature forgery | All-zero, garbage, truncated, wrong-mode |
| L — TLV format validation | Magic, version, algorithm mismatch, round-trip |

### 6.2 Smart-contract security tests (Foundry)

```bash
cd ethereum
forge test --match-test "test_Security" -v
```

16 tests covering:

| Test | Attack |
|---|---|
| `test_Security_Reentrancy_CEIBlocksAttacker` | Reentrancy attack via malicious `receive()` |
| `test_Security_FrontRun_NonOracleCannotSubmitOracleTx` | Front-running by non-oracle |
| `test_Security_FrontRun_NoncePreventsDoubleExecution` | Nonce blocks oracle replay |
| `test_Security_MaliciousRelayer_ShortSigBypassesECDSA` | **Documents vulnerability**: non-65-byte sig skips ECDSA check |
| `test_Security_MaliciousRelayer_CannotInflateAmountBeyondBalance` | Malicious amount rejected |
| `test_Security_MaliciousRelayer_ForgedSignatureRejected` | Oracle forging user's sig |
| `test_Security_Forgery_EmptySigBypassesECDSA` | **Documents vulnerability**: zero-length sig skips check |
| `test_Security_Forgery_CorrectLengthWrongSigRejected` | 65 zeros rejected by `ecrecover` |
| `test_Security_Forgery_SigForDifferentAmountRejected` | Sig bound to exact amount |

### 6.3 Full test suite

```bash
cd ethereum && forge test -v
cd oracle-server && go test ./test/... -v -count=1 -timeout 300s
```

---

## 7. API Reference

Oracle listens on `http://localhost:3000`.

### POST `/api/generate-key`

Generate a keypair. For hybrid mode, provide an existing ECDSA private key so the Ethereum address is preserved.

**Request:**
```json
{
  "algorithm": "ML-DSA-65",
  "mode": "hybrid",
  "ecdsa_private_key": "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
}
```

`mode`: `"single"` | `"hybrid"`. `algorithm`: any supported PQC algorithm, or `"ECDSA"` for single mode.

**Response:**
```json
{
  "status": "success",
  "mode": "hybrid",
  "algorithm": "Hybrid-Secp256k1-ML-DSA-65",
  "private_key": "0x48594253...",
  "public_key": "0x48594253..."
}
```

Keys are TLV-encoded (see §8). Store the private key securely — it is sent in signing requests.

---

### POST `/api/sign`

Sign and submit a transaction to the blockchain.

**Request:**
```json
{
  "sender": "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
  "receiver": "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
  "message": "payment for service",
  "amount": "1000000000000000000",
  "private_key": "0x48594253...",
  "algorithm": "Hybrid-Secp256k1-ML-DSA-65",
  "mode": "hybrid"
}
```

`amount` is in wei (string to avoid integer overflow).

**Response:**
```json
{
  "status": "success",
  "signing_algorithm": "Hybrid-Secp256k1-ML-DSA-65",
  "transaction_hash": "0xabc...",
  "user_nonce_used": 1,
  "sender": "0xf39F...",
  "receiver": "0x7099...",
  "amount_wei": "1000000000000000000",
  "signature": "0x48594253...",
  "signature_length": 3379,
  "gas_price": 1000000000,
  "gas_used": 388057
}
```

---

### POST `/api/verify`

Verify a submitted transaction by on-chain tx hash. Fetches raw calldata from the chain, reconstructs the hash, and runs the full verifier.

**Request:**
```json
{
  "tx_hash": "0xabc...",
  "public_key": "0x48594253..."
}
```

**Response (valid):**
```json
{
  "valid": true,
  "verify_message": "Verification Success",
  "note": "Valid Hybrid Signature (Secp256k1 + ML-DSA-65)",
  "tx_hash": "0xabc...",
  "sender": "0xf39F...",
  "receiver": "0x7099...",
  "signing_algorithm": "Hybrid-Secp256k1-ML-DSA-65",
  "signing_mode": "hybrid"
}
```

---

### GET `/api/transactions/:txhash`

Fetch decoded transaction data (no verification). Useful for inspecting stored nonce, sender, algorithm, and the on-chain `pqSignatureHash`.

---

## 8. Wire Format — TLV Encoding

Hybrid keys and signatures use a self-describing TLV (Type-Length-Value) binary format. This prevents silent algorithm confusion and provides fast corruption detection.

```
[MAGIC 4B] [VERSION 1B] [FIELD...]
FIELD = [TAG 1B] [LENGTH 4B big-endian] [VALUE LENGTH×B]
```

| Constant | Value | Meaning |
|---|---|---|
| Magic | `0x48594253` ("HYBS") | Non-zero sentinel; rejects garbage blobs immediately |
| Version | `0x01` | Format version; incremented on breaking changes |
| `TagAlgorithm` | `0x01` | Algorithm name string (e.g. `"Hybrid-Secp256k1-ML-DSA-65"`) |
| `TagECDSAPriv` | `0x02` | ECDSA private key (32 B secp256k1 scalar) |
| `TagECDSAPub` | `0x03` | ECDSA public key (65 B uncompressed) |
| `TagPQCPriv` | `0x04` | PQC private key (liboqs raw) |
| `TagPQCPub` | `0x05` | PQC public key (liboqs raw) |
| `TagECDSASig` | `0x06` | ECDSA signature (65 B) |
| `TagPQCSig` | `0x07` | PQC signature (variable) |

**Algorithm binding:** Every blob embeds the algorithm name (`TagAlgorithm`). The verifier checks:
1. `sig.algorithm == pub.algorithm` — sig and pub key must agree
2. `sig.algorithm == verifier.algorithm` — blob must match the configured verifier

**Important:** The contract receives raw `ecdsaSig` (65 B) and `pqSig` bytes — not TLV-wrapped — because the Solidity `ecrecover` precompile expects raw bytes. TLV applies only to off-chain key storage and the `finalSig` returned by the API.

Implementation: `oracle-server/internal/signature/tlv.go`

---

## 9. Scheme Design Decisions

### Why oracle/relayer architecture?

On-chain ML-DSA verification would cost millions of gas (lattice polynomial operations are not EVM precompiles). The oracle:
- Verifies both ECDSA and PQC off-chain before submitting
- Submits as a trusted Ethereum account (`require(msg.sender == ORACLE)`)
- Stores `keccak256(pqSig)` on-chain as a tamper-evident anchor

### Why independent concatenation, not composite signatures?

The IETF composite signature standard (draft-ounsworth-pq-composite-sigs) tightly binds both components through shared key derivation. We use `σ_ECDSA ‖ σ_PQC` instead because:

1. The smart contract must independently verify the ECDSA component via `ecrecover` — any compact binding that merges both signatures breaks this.
2. Security follows from the "at-least-one" theorem (Bindel et al., 2019): the hybrid is secure if **either** ECDSA or PQC is secure. No cross-scheme assumptions needed.
3. Algorithm agility: the PQC component can be swapped without changing the Solidity contract.

At ML-DSA-65 scale (3 374 bytes total), the saving from compact binding (≈65 bytes) is less than 2%.

### Message hashing

```
data         = abi.encode(nonce, sender, receiver, algorithm, mode, message, amount)
messageHash  = keccak256(data)
signedHash   = keccak256("\x19ETHEREUM-ORACLE-SIGNED:\n32" || messageHash)
```

The `\x19` prefix follows EIP-191 — it makes the hash byte-incompatible with a valid RLP-encoded Ethereum transaction, preventing signature reuse attacks. The `algorithm` and `mode` fields inside the ABI encoding act as algorithm binding — a sig under `"ML-DSA-65"` cannot be replayed as `"Hybrid-Secp256k1-ML-DSA-65"`.

---

## 10. Known Limitations

These are documented in the security tests and should be addressed in the article.

### L1 — ECDSA check is opt-in in the contract

```solidity
if (_params.ecdsaSignature.length == 65) {
    require(verifySignature(_params), "On-chain ECDSA verification failed");
}
```

A malicious oracle can bypass ECDSA verification by sending any signature shorter or longer than 65 bytes. Tested in `test_Security_MaliciousRelayer_ShortSigBypassesECDSA` and `test_Security_Forgery_EmptySigBypassesECDSA`.

**Mitigation:** Change the condition to `require(_params.ecdsaSignature.length == 65)` unconditionally, or always require ECDSA verification for hybrid mode.

### L2 — No cross-component key commitment

The ECDSA and PQC components are verified independently. An adversary who obtains `kA`'s ECDSA signature for a specific hash can combine it with their own valid PQC signature and pass verification. Tested in `TestSecurity_E1_CrossKeyHybridMixing`.

**Mitigation:** Include `hash(pqcPublicKey)` in the ABI-encoded payload so the ECDSA signature binds to a specific PQC public key.

### L3 — No chain ID in signed payload

The current scheme does not include `chainId` or `verifyingContract` in the signed message. Cross-chain replay is theoretically possible if the same contract is deployed on multiple networks.

**Mitigation:** Adopt EIP-712 domain separation for production.

### L4 — Oracle is a single point of trust

All security depends on the oracle's Ethereum key not being compromised. A compromised oracle can bypass all on-chain checks (except balance validation).

**Mitigation:** Multi-oracle threshold signing, or move PQC verification on-chain once EVM precompiles exist.

---

## 11. Reproducing Published Results

### Performance table (for article Table II / Section C)

```bash
cd oracle-server
go test ./test/... -run TestFunctional_Performance -v -timeout 300s 2>&1 | tee results.txt
```

The report prints directly to stdout. Columns map to article metrics:

| Report column | Article metric |
|---|---|
| `KeyGen Avg (ms)` | Keypair generation time |
| `Sign (ms)` | Signing execution time |
| `Verify (ms)` | Verification execution time |
| `TPS` | Throughput |
| `Total Sig (B)` | Transaction payload size |
| `Est. Gas` | Gas cost efficiency |

### Security test coverage (for article Section D / threat model)

```bash
# Oracle layer
cd oracle-server && go test ./test/... -run TestSecurity -v -count=1 2>&1 | grep -E "^(--- |PASS|FAIL)"

# Contract layer
cd ethereum && forge test --match-test "test_Security|test_Withdraw" -v 2>&1 | grep -E "(PASS|FAIL)"
```

Both should show 0 failures across 47 + 16 tests.

### TLV format correctness (for article Section E on encoding)

```bash
cd oracle-server
go test ./test/... -run TestSecurity_L -v -count=1
```

6 tests verify magic, version, algorithm binding, truncation detection, and round-trip fidelity.

---

## 12. File Map

```
pi/
├── Makefile / Taskfile.yml       ← top-level orchestration commands
│
├── ethereum/                     ← Foundry project
│   ├── src/Transactions.sol      ← smart contract (deposit, withdraw, sendTransaction)
│   ├── script/Deploy.s.sol       ← deployment script
│   └── test/Transactions.t.sol   ← Foundry tests (functional + security)
│
├── oracle-server/                ← Go oracle
│   ├── cmd/app/main.go           ← entry point (Fiber HTTP server)
│   ├── internal/
│   │   ├── signature/
│   │   │   ├── tlv.go            ← TLV codec (magic, version, type tags)
│   │   │   ├── keypair_generator.go  ← ECDSA + PQC key generation
│   │   │   ├── signer.go         ← hybrid signing, private key parsing
│   │   │   └── verifier.go       ← hybrid/single verification
│   │   ├── handler/
│   │   │   └── signature_handler.go  ← HTTP handlers
│   │   ├── chain/chain.go        ← ABI decoder for on-chain tx data
│   │   ├── util/
│   │   │   ├── chain.go          ← BuildSignatureBytes, SendToBlockchain
│   │   │   └── signature.go      ← HashEthSignedData
│   │   ├── model/signature_model.go  ← request/response structs
│   │   └── route/route.go        ← API routes
│   ├── artifacts/Transactions.go ← generated ABI bindings (abigen)
│   └── test/
│       ├── security_test.go      ← 47 security tests (groups A–L)
│       ├── functional_test.go    ← performance benchmarks (all algorithms)
│       └── signing_test.go       ← sign/verify unit tests
│
└── client/                       ← React + TypeScript UI (optional)
    └── src/                      ← wallet connect, tx submission form
```
