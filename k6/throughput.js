// ============================================================================
// K6 Throughput (TPS) Test - Hybrid PQC Scheme Evaluation
// ============================================================================
// Metric 3: Throughput (TPS) — max transaction processing capacity
//           under high workload (10 concurrent VUs).
//
// 15 sequential scenarios (15 min offset each):
//   Single:  ECDSA, ML-DSA-44, ML-DSA-65, Falcon-512, Falcon-1024,
//            SLH-DSA-SHA2-128S, SLH-DSA-SHA2-192S, SLH-DSA-SHA2-256S
//   Hybrid:  ECDSA + each of the 7 PQC schemes above
//
// Each VU uses a separate sender address to avoid nonce conflicts.
// Iteration 0 per VU = warmup (excluded from custom metrics).
// ============================================================================

import http from "k6/http";
import { check } from "k6";
import { Trend, Counter, Rate, Gauge } from "k6/metrics";
import exec from "k6/execution";
import {
  BASE_URL, HEADERS, RECEIVER, HARDHAT_ACCOUNTS, TOTAL_ITERATIONS,
  mldsaSenderForVU, mldsa44SenderForVU,
} from "./config.js";
import { textSummary } from "https://jslib.k6.io/k6-summary/0.1.0/index.js";

// ============================================================================
// Scheme list — same key/signAlgo/mode/accountIdx as benchmark.js
// ============================================================================

const SCHEMES = [
  { key: "ecdsa",             signAlgo: "ECDSA",                                      mode: "single", keygenAlgo: "ECDSA" },
  { key: "mldsa44",           signAlgo: "ML-DSA-44",                                  mode: "single", keygenAlgo: "ML-DSA-44" },
  { key: "mldsa65",           signAlgo: "ML-DSA-65",                                  mode: "single", keygenAlgo: "ML-DSA-65" },
  { key: "falcon512",         signAlgo: "Falcon-512",                                 mode: "single", keygenAlgo: "Falcon-512" },
  { key: "falcon1024",        signAlgo: "Falcon-1024",                                mode: "single", keygenAlgo: "Falcon-1024" },
  { key: "slh128s",           signAlgo: "SLH_DSA_PURE_SHA2_128S",                     mode: "single", keygenAlgo: "SLH_DSA_PURE_SHA2_128S" },
  { key: "slh192s",           signAlgo: "SLH_DSA_PURE_SHA2_192S",                     mode: "single", keygenAlgo: "SLH_DSA_PURE_SHA2_192S" },
  { key: "slh256s",           signAlgo: "SLH_DSA_PURE_SHA2_256S",                     mode: "single", keygenAlgo: "SLH_DSA_PURE_SHA2_256S" },
  { key: "hybrid_mldsa44",    signAlgo: "Hybrid-Secp256k1-ML-DSA-44",                 mode: "hybrid", keygenAlgo: "ML-DSA-44" },
  { key: "hybrid_mldsa65",    signAlgo: "Hybrid-Secp256k1-ML-DSA-65",                 mode: "hybrid", keygenAlgo: "ML-DSA-65" },
  { key: "hybrid_falcon512",  signAlgo: "Hybrid-Secp256k1-Falcon-512",                mode: "hybrid", keygenAlgo: "Falcon-512" },
  { key: "hybrid_falcon1024", signAlgo: "Hybrid-Secp256k1-Falcon-1024",               mode: "hybrid", keygenAlgo: "Falcon-1024" },
  { key: "hybrid_slh128s",    signAlgo: "Hybrid-Secp256k1-SLH_DSA_PURE_SHA2_128S",    mode: "hybrid", keygenAlgo: "SLH_DSA_PURE_SHA2_128S" },
  { key: "hybrid_slh192s",    signAlgo: "Hybrid-Secp256k1-SLH_DSA_PURE_SHA2_192S",    mode: "hybrid", keygenAlgo: "SLH_DSA_PURE_SHA2_192S" },
  { key: "hybrid_slh256s",    signAlgo: "Hybrid-Secp256k1-SLH_DSA_PURE_SHA2_256S",    mode: "hybrid", keygenAlgo: "SLH_DSA_PURE_SHA2_256S" },
];

// ============================================================================
// Custom metrics — one set per scheme
// ============================================================================

const signLatency  = {};
const successRate  = {};
const totalTx      = {};
const tpsGauge     = {};

for (const { key } of SCHEMES) {
  signLatency[key] = new Trend(`throughput_sign_latency_${key}_ms`, true);
  successRate[key] = new Rate(`throughput_success_rate_${key}`);
  totalTx[key]     = new Counter(`throughput_total_tx_${key}`);
  tpsGauge[key]    = new Gauge(`tps_${key}`);
}

// ============================================================================
// K6 Options — 15 sequential scenarios (15 min apart)
// ============================================================================

const VUS   = parseInt(__ENV.VUS   || "10");
const ITERS = parseInt(__ENV.ITERATIONS || "101");
const OFFSET_MIN = parseInt(__ENV.OFFSET_MIN || "15");

function startAt(idx) {
  return `${idx * OFFSET_MIN}m`;
}

export const options = {
  scenarios: Object.fromEntries(
    SCHEMES.map((s, i) => [
      `${s.key}_throughput`,
      {
        executor: "per-vu-iterations",
        vus: VUS,
        iterations: ITERS,
        maxDuration: `${OFFSET_MIN + 5}m`,
        exec: `runScenario`,
        startTime: startAt(i),
        env: { SCENARIO_KEY: s.key },
      },
    ])
  ),
};

// ============================================================================
// Sender address helpers
// ============================================================================

function senderForVU(schemeKey, vuIdx) {
  if (schemeKey === "ecdsa" || schemeKey.startsWith("hybrid_")) {
    return HARDHAT_ACCOUNTS[vuIdx % HARDHAT_ACCOUNTS.length].address;
  }
  // Single PQC: unique arbitrary address per VU
  const schemeIdx = SCHEMES.findIndex((s) => s.key === schemeKey);
  const hex = vuIdx.toString(16).padStart(4, "0");
  const idx = schemeIdx.toString(16).padStart(4, "0");
  return `0x00000000000000000000000000000000${idx}${hex}`;
}

function privateKeyForVU(accounts, schemeKey, vuIdx) {
  const account = accounts[vuIdx];
  return account[schemeKey]?.privateKey ?? null;
}

// ============================================================================
// Setup: pre-generate all keys for all VUs
// ============================================================================

export function setup() {
  console.log(`=== SETUP: Generating keys for ${VUS} VUs × ${SCHEMES.length} schemes ===`);

  const accounts = [];

  for (let i = 0; i < VUS; i++) {
    const acct = HARDHAT_ACCOUNTS[i % HARDHAT_ACCOUNTS.length];
    const vuKeys = { vuIndex: i };

    for (const s of SCHEMES) {
      if (s.key === "ecdsa") {
        vuKeys[s.key] = {
          privateKey: acct.privateKey,
          sender: acct.address,
        };
        continue;
      }

      if (s.mode === "single") {
        const res = http.post(
          `${BASE_URL}/api/generate-key`,
          JSON.stringify({ algorithm: s.keygenAlgo, mode: "single" }),
          { headers: HEADERS }
        );
        if (res.status !== 200) throw new Error(`Keygen failed VU${i} ${s.key}: ${res.body}`);
        const d = res.json();
        vuKeys[s.key] = {
          privateKey: d.private_key,
          sender: senderForVU(s.key, i),
        };
      } else {
        const res = http.post(
          `${BASE_URL}/api/generate-key`,
          JSON.stringify({ algorithm: s.keygenAlgo, mode: "hybrid", ecdsa_private_key: acct.privateKey }),
          { headers: HEADERS }
        );
        if (res.status !== 200) throw new Error(`Keygen failed VU${i} ${s.key}: ${res.body}`);
        const d = res.json();
        vuKeys[s.key] = {
          privateKey: d.private_key,
          sender: acct.address,
        };
      }
    }

    accounts.push(vuKeys);
    console.log(`  VU ${i}: all keys generated`);
  }

  console.log("=== SETUP COMPLETE ===");
  return { accounts, startTime: Date.now() };
}

// ============================================================================
// Generic scenario runner (exec target for all scenarios)
// ============================================================================

export function runScenario(data) {
  const scenarioKey = __ENV.SCENARIO_KEY;
  const scheme = SCHEMES.find((s) => s.key === scenarioKey);
  if (!scheme) {
    console.error(`Unknown SCENARIO_KEY: ${scenarioKey}`);
    return;
  }

  const iter  = exec.vu.iterationInScenario;
  const isWarmup = iter === 0;
  const vuIdx = (exec.vu.idInTest - 1) % data.accounts.length;
  const acct  = data.accounts[vuIdx];
  const vuKey = acct[scenarioKey];

  const res = http.post(
    `${BASE_URL}/api/sign`,
    JSON.stringify({
      sender: vuKey.sender,
      receiver: RECEIVER,
      message: `k6-tps-${scenarioKey}-vu${vuIdx}-${iter}`,
      amount: "1000000",
      private_key: vuKey.privateKey,
      algorithm: scheme.signAlgo,
      mode: scheme.mode,
    }),
    { headers: HEADERS, tags: { scenario: `${scenarioKey}_throughput` } }
  );

  const success = res.status === 200;

  if (!isWarmup) {
    signLatency[scenarioKey].add(res.timings.duration);
    successRate[scenarioKey].add(success);
    if (success) totalTx[scenarioKey].add(1);
  }

  if (!success) {
    console.error(`${scenarioKey} TPS [VU${vuIdx} iter${iter}]: ${res.status} - ${res.body}`);
  }
}

// ============================================================================
// Custom Summary
// ============================================================================

export function handleSummary(data) {
  const computeTPS = (latencyKey) => {
    const m = data.metrics[latencyKey];
    if (!m || !m.values || !m.values.avg) return 0;
    return VUS / (m.values.avg / 1000);
  };

  const extract = (name) => {
    const m = data.metrics[name];
    if (!m || !m.values) return null;
    return {
      avg: m.values.avg,
      min: m.values.min,
      max: m.values.max,
      med: m.values.med,
      "p(90)": m.values["p(90)"],
      "p(95)": m.values["p(95)"],
    };
  };

  const schemes = {};
  for (const { key } of SCHEMES) {
    const latencyKey = `throughput_sign_latency_${key}_ms`;
    schemes[key] = {
      sign_latency:        extract(latencyKey),
      success_rate:        data.metrics[`throughput_success_rate_${key}`]?.values.rate ?? 0,
      total_successful_tx: data.metrics[`throughput_total_tx_${key}`]?.values.count ?? 0,
      estimated_tps:       computeTPS(latencyKey),
    };
  }

  // TPS comparison (all vs ECDSA)
  const ecdsaTps = schemes.ecdsa?.estimated_tps ?? 0;
  const tpsComparison = {};
  for (const { key } of SCHEMES) {
    if (key === "ecdsa") continue;
    const t = schemes[key].estimated_tps;
    tpsComparison[key] = {
      estimated_tps: t,
      vs_ecdsa_pct: ecdsaTps > 0 ? ((t - ecdsaTps) / ecdsaTps) * 100 : null,
    };
  }

  const report = {
    test: "Hybrid PQC Throughput (TPS)",
    vus: VUS,
    iterations_per_vu: ITERS,
    measured_iterations_per_vu: ITERS - 1,
    timestamp: new Date().toISOString(),
    schemes,
    tps_comparison: tpsComparison,
  };

  return {
    stdout: textSummary(data, { indent: " ", enableColors: true }),
    "results/throughput_report.json": JSON.stringify(report, null, 2),
  };
}
