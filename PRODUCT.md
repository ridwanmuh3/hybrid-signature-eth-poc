# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Stack

- **Client:** React 19 + TypeScript + Vite + Tailwind CSS v4, React Router v7, Wagmi v3 + Viem v2 (multi-chain wallet + contract), TanStack Query, Zustand (state), Zod + React Hook Form (validation), lucide-react icons. Dev: ESLint + Vitest. Build: `pnpm dev` / `pnpm build`.
- **Oracle server:** Go 1.25 + Fiber v2 REST API on `:9000`, liboqs-go (11 PQC algorithms), go-ethereum (JSON-RPC + `abigen`-bound `Transactions` contract), Viper config. Build/test via `just`.
- **Smart contract:** Solidity 0.8.31, Foundry (Forge/Anvil, chain ID 31337), deployed at `0x73511669fd4dE447feD18BB79bAFeAC93aB7F31f` on test; oracle address set immutable at deploy.
- **Frontend↔Oracle:** Axios calls to `http://localhost:9000` (`/api/generate-key`, `/api/sign`, `/api/transactions/:txhash`, `/api/verify`).

## Users

Primary: cryptography/Ethereum researchers and engineers running or evaluating the hybrid signature experiment. They connect a wallet (Sepolia or local Anvil), generate a keypair, fund a contract ledger via `deposit()`, submit a dual-signed transaction routed through the oracle, and verify the result. Secondary: reviewers reading the README and `results/` artifacts who need reproducible commands (`just bench`, `just test-all`, `just benchmark-onchain`) and a 54-test security matrix.

Job: prove feasibility and measure overhead of a hybrid ECDSA + post-quantum signature scheme on EVM, with on-chain ECDSA verification via `ecrecover` and an on-chain `keccak256(pqSig)` integrity anchor for the PQC component.

## Product Purpose

Demonstrate that a hybrid transaction signature scheme can be deployed on Ethereum with acceptable overhead by combining Secp256k1 ECDSA (EVM-native) with a post-quantum algorithm (NIST FIPS 202/204 + Round-2 additional signatures). The oracle produces a TLV-wrapped hybrid signature, verifies both components off-chain, and submits ECDSA + PQC hash to the contract, which enforces oracle-only submission, nonce replay protection, ECDSA verification, and malleability rejection. Success = a reproducible, test-backed experiment with documented gas cost, signing latency, and signature size across 23 single and hybrid schemes.

## Positioning

Independent concatenation (`σ_ECDSA ‖ σ_PQC`) rather than compact/composite binding (no shared cross-scheme derivation). Security reduces to the "at least one component secure" theorem (Bindel et al. 2019). The EVM constraint (Solidity `ecrecover` verifies ECDSA independently) and algorithm agility (swap PQC scheme without touching the ECDSA layer or contract) are the differentiating, non-copyable commitments versus prior hybrid-key work.

## Operating Context

- Local: Anvil node on `http://localhost:8540`, chain ID 31337; Foundry handles compile/inspect/deployment.
- Testnet: Sepolia (`sepolianet`); contract address from `VITE_CONTRACT_ADDRESS` (`apps/web/.env`).
- Workflows: generate keypair → deposit ETH to contract ledger → fill transaction form (receiver, amount, message, private key, algorithm, mode) → oracle signs with ECDSA + PQC → contract verifies ECDSA, stores PQC hash → verify via tx hash + public key.
- Tooling: `make` recipes drive benchmarks, keygen, deployment, and integration/functional tests; artifacts regenerated via `forge inspect` + `abigen`.

## Capabilities and Constraints

- Capabilities: 11 PQC algorithms (ML-DSA-44/65/87, Falcon-512/1024, MAYO-1/3/5, SNOVA_24_5_4/37_8_4/60_10_4) in `single` and `hybrid` modes; TLV wire format (`HYBS` magic, version `0x01`, tagged fields); 23-scheme performance matrix; per-sender monotonic nonce; 54 security tests (39 Go + 15 Foundry) covering stripping, substitution, tampering, mode confusion, replay, front-running, TLV framing, malleability, reentrancy.
- Constraints: PQC verification stays off-chain (no EVM precompile) — only `keccak256(pqSig)` is anchored on-chain; the oracle is a trusted relayer by design; signed payload binds `chainId` but not `verifyingContract`, so production should adopt EIP-712 or explicit contract binding; ECDSA and PQC components are independently validated (no compact cross-binding).

## Brand Commitments

- Project name: "Hybrid PQS PoC" (browser title).
- Wire format: TLV blobs begin with magic bytes `HYBS`, version `0x01`, with tags `TagAlgorithm`, ECDSA material, and PQC material; bounded parsing rejects invalid magic, unsupported version, duplicate tags, truncation, oversized blobs.
- Domain separation: outer hash = `keccak256("\x19ETHEREUM-ORACLE-SIGNED:\n32" || messageHash)`; identical prefix used in Go oracle and Solidity contract.
- Algorithm labels: `"ML-DSA-65"`, `"Hybrid-Secp256k1-ML-DSA-65"`, etc., included in the signed tuple for algorithm binding.

## Evidence on Hand

- `README.md` (root): full architecture, message format, security properties, limitations, related work, test instructions.
- `apps/oracle/README.md` (deleted): performance metrics, message/domain-separation spec, signature-size table, blockchain assumptions, on-chain vs off-chain trade-off.
- `apps/oracle/results/`: `security_test.txt`, `functional_performance.txt`, `go_benchmarks.txt`, `keygen.txt`.
- removed, `security-test-result-new.yxy`: security test outputs.
- Test suites: `oracle-server/test/security_test.go`, `oracle-server/test/functional_test.go`, `oracle-server/test/performance_test.go`, `oracle-server/test/onchain_test.go`; `ethereum/test/Transactions.t.sol`.
- Source of truth for signing: `oracle-server/internal/signature/` (TLV, signer, verifier, signature), `ethereum/src/Transactions.sol`.

## Product Principles

1. Correctness over cleverness — independent verification; never merge signature components into a non-auditable compact form.
2. Algorithm agility — swap PQC scheme without changing the ECDSA layer or the contract.
3. Security by testing — every capability change must add or update a test (Go verifier + Foundry contract).
4. Reproducibility — benchmark and security commands must run from `make` with warm-state counts recorded.

## Accessibility & Inclusion

No product-specific accessibility requirement established yet. Existing implementation relies on Semantic UI form primitives (Radix) and native form controls; adopt WCAG 2.1 AA as a baseline when auditing the client surface.

---

<!-- impeccable:product-init complete; do not write DESIGN.md here. -->
