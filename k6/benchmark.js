// ============================================================================
// K6 Benchmark Test - Hybrid PQC Scheme Evaluation
// ============================================================================
// Schemes compared (15 total):
//   Single:  ECDSA, ML-DSA-44, ML-DSA-65, Falcon-512, Falcon-1024,
//            SLH-DSA-SHA2-128S, SLH-DSA-SHA2-192S, SLH-DSA-SHA2-256S
//   Hybrid:  ECDSA + each of the 7 PQC schemes above
//
// Configuration: 1 VU, 101 iterations (iteration 0 = warmup, excluded)
// ============================================================================

import http from "k6/http";
import { check, group } from "k6";
import { Trend, Counter } from "k6/metrics";
import exec from "k6/execution";
import {
  BASE_URL, HEADERS, RECEIVER, HARDHAT_ACCOUNTS, TOTAL_ITERATIONS,
} from "./config.js";
import { textSummary } from "https://jslib.k6.io/k6-summary/0.1.0/index.js";

// ============================================================================
// Scheme definitions
// ============================================================================
// keygenAlgo : algorithm passed to /api/generate-key
// signAlgo   : algorithm passed to /api/sign and /api/verify
// mode       : "single" or "hybrid"
// accountIdx : HARDHAT_ACCOUNTS index (ECDSA single + hybrid schemes)
// sender     : arbitrary address for single PQC (no on-chain ECDSA check)

const SCHEMES = [
  { key: "ecdsa",              keygenAlgo: "ECDSA",                           signAlgo: "ECDSA",                                       mode: "single", accountIdx: 1 },
  { key: "mldsa44",            keygenAlgo: "ML-DSA-44",                       signAlgo: "ML-DSA-44",                                   mode: "single", sender: "0x4444444444444444444444444444444444444444" },
  { key: "mldsa65",            keygenAlgo: "ML-DSA-65",                       signAlgo: "ML-DSA-65",                                   mode: "single", sender: "0x1111111111111111111111111111111111111111" },
  { key: "falcon512",          keygenAlgo: "Falcon-512",                      signAlgo: "Falcon-512",                                  mode: "single", sender: "0x5555555555555555555555555555555555555555" },
  { key: "falcon1024",         keygenAlgo: "Falcon-1024",                     signAlgo: "Falcon-1024",                                 mode: "single", sender: "0x6666666666666666666666666666666666666666" },
  { key: "slh128s",            keygenAlgo: "SLH_DSA_PURE_SHA2_128S",          signAlgo: "SLH_DSA_PURE_SHA2_128S",                      mode: "single", sender: "0x7777777777777777777777777777777777777777" },
  { key: "slh192s",            keygenAlgo: "SLH_DSA_PURE_SHA2_192S",          signAlgo: "SLH_DSA_PURE_SHA2_192S",                      mode: "single", sender: "0x8888888888888888888888888888888888888888" },
  { key: "slh256s",            keygenAlgo: "SLH_DSA_PURE_SHA2_256S",          signAlgo: "SLH_DSA_PURE_SHA2_256S",                      mode: "single", sender: "0x9999999999999999999999999999999999999999" },
  { key: "hybrid_mldsa44",     keygenAlgo: "ML-DSA-44",                       signAlgo: "Hybrid-Secp256k1-ML-DSA-44",                  mode: "hybrid", accountIdx: 2 },
  { key: "hybrid_mldsa65",     keygenAlgo: "ML-DSA-65",                       signAlgo: "Hybrid-Secp256k1-ML-DSA-65",                  mode: "hybrid", accountIdx: 3 },
  { key: "hybrid_falcon512",   keygenAlgo: "Falcon-512",                      signAlgo: "Hybrid-Secp256k1-Falcon-512",                 mode: "hybrid", accountIdx: 4 },
  { key: "hybrid_falcon1024",  keygenAlgo: "Falcon-1024",                     signAlgo: "Hybrid-Secp256k1-Falcon-1024",                mode: "hybrid", accountIdx: 5 },
  { key: "hybrid_slh128s",     keygenAlgo: "SLH_DSA_PURE_SHA2_128S",          signAlgo: "Hybrid-Secp256k1-SLH_DSA_PURE_SHA2_128S",     mode: "hybrid", accountIdx: 6 },
  { key: "hybrid_slh192s",     keygenAlgo: "SLH_DSA_PURE_SHA2_192S",          signAlgo: "Hybrid-Secp256k1-SLH_DSA_PURE_SHA2_192S",     mode: "hybrid", accountIdx: 7 },
  { key: "hybrid_slh256s",     keygenAlgo: "SLH_DSA_PURE_SHA2_256S",          signAlgo: "Hybrid-Secp256k1-SLH_DSA_PURE_SHA2_256S",     mode: "hybrid", accountIdx: 8 },
];

// ============================================================================
// Custom metrics — one set per scheme
// ============================================================================

const keypairMetrics = {};
const signMetrics    = {};
const verifyMetrics  = {};
const payloadMetrics = {};
const sigLenMetrics  = {};
const gasUsedMetrics = {};
const gasPriceMetrics = {};

for (const { key } of SCHEMES) {
  keypairMetrics[key]  = new Trend(`keypair_gen_${key}_ms`, true);
  signMetrics[key]     = new Trend(`sign_exec_${key}_ms`, true);
  verifyMetrics[key]   = new Trend(`verify_exec_${key}_ms`, true);
  payloadMetrics[key]  = new Trend(`tx_payload_size_${key}_bytes`);
  sigLenMetrics[key]   = new Trend(`signature_length_${key}_bytes`);
  gasUsedMetrics[key]  = new Trend(`gas_used_${key}`);
  gasPriceMetrics[key] = new Trend(`gas_price_${key}`);
}

const successOps = new Counter("successful_operations");
const failedOps  = new Counter("failed_operations");

// ============================================================================
// K6 Options
// ============================================================================

export const options = {
  scenarios: {
    benchmark: {
      executor: "per-vu-iterations",
      vus: 1,
      iterations: TOTAL_ITERATIONS,
      maxDuration: "120m",
    },
  },
};

// ============================================================================
// Setup: pre-generate signing keys for all schemes
// ============================================================================

export function setup() {
  console.log(`=== SETUP: Generating keys for ${SCHEMES.length} schemes ===`);

  const result = {};

  for (const s of SCHEMES) {
    if (s.key === "ecdsa") {
      const acct = HARDHAT_ACCOUNTS[s.accountIdx];
      result[s.key] = {
        privateKey: acct.privateKey,
        publicKey: acct.address,
        sender: acct.address,
      };
      console.log(`  ${s.key}: using hardhat account ${acct.address}`);
      continue;
    }

    if (s.mode === "single") {
      const res = http.post(
        `${BASE_URL}/api/generate-key`,
        JSON.stringify({ algorithm: s.keygenAlgo, mode: "single" }),
        { headers: HEADERS }
      );
      check(res, { [`setup: ${s.key} keygen OK`]: (r) => r.status === 200 });
      if (res.status !== 200) throw new Error(`Keygen failed for ${s.key}: ${res.body}`);
      const d = res.json();
      result[s.key] = { privateKey: d.private_key, publicKey: d.public_key, sender: s.sender };
    } else {
      const acct = HARDHAT_ACCOUNTS[s.accountIdx];
      const res = http.post(
        `${BASE_URL}/api/generate-key`,
        JSON.stringify({ algorithm: s.keygenAlgo, mode: "hybrid", ecdsa_private_key: acct.privateKey }),
        { headers: HEADERS }
      );
      check(res, { [`setup: ${s.key} keygen OK`]: (r) => r.status === 200 });
      if (res.status !== 200) throw new Error(`Keygen failed for ${s.key}: ${res.body}`);
      const d = res.json();
      result[s.key] = { privateKey: d.private_key, publicKey: d.public_key, sender: acct.address };
    }

    console.log(`  ${s.key}: keypair generated`);
  }

  console.log("=== SETUP COMPLETE ===");
  return result;
}

// ============================================================================
// Main test function
// ============================================================================

export default function (setupData) {
  const iter = exec.vu.iterationInScenario;
  const isWarmup = iter === 0;

  if (isWarmup) {
    console.log("--- Warmup iteration (excluded from metrics) ---");
  }

  // ==========================================================================
  // METRIC 1: Keypair Generation Times
  // ==========================================================================
  group("Metric 1 - Keypair Generation", () => {
    for (const s of SCHEMES) {
      let body;
      if (s.key === "ecdsa") {
        body = { algorithm: "ECDSA", mode: "single" };
      } else if (s.mode === "single") {
        body = { algorithm: s.keygenAlgo, mode: "single" };
      } else {
        const acct = HARDHAT_ACCOUNTS[s.accountIdx];
        body = { algorithm: s.keygenAlgo, mode: "hybrid", ecdsa_private_key: acct.privateKey };
      }

      const res = http.post(
        `${BASE_URL}/api/generate-key`,
        JSON.stringify(body),
        { headers: HEADERS, tags: { operation: `keypair_${s.key}` } }
      );
      check(res, { [`${s.key} keygen OK`]: (r) => r.status === 200 });

      if (!isWarmup) {
        if (res.status === 200) {
          keypairMetrics[s.key].add(res.timings.duration);
          successOps.add(1);
        } else {
          failedOps.add(1);
        }
      }
    }
  });

  // ==========================================================================
  // METRICS 2, 4, 5: Signing + Payload Size + Gas Cost
  // ==========================================================================
  const signResults = {};

  group("Metric 2,4,5 - Signing Transactions", () => {
    for (const s of SCHEMES) {
      const sd = setupData[s.key];
      const res = http.post(
        `${BASE_URL}/api/sign`,
        JSON.stringify({
          sender: sd.sender,
          receiver: RECEIVER,
          message: `k6-${s.key}-${iter}`,
          amount: "1000000",
          private_key: sd.privateKey,
          algorithm: s.signAlgo,
          mode: s.mode,
        }),
        { headers: HEADERS, tags: { operation: `sign_${s.key}` } }
      );
      check(res, { [`${s.key} sign OK`]: (r) => r.status === 200 });

      if (res.status === 200) {
        signResults[s.key] = res.json();
        if (!isWarmup) {
          const body = res.json();
          signMetrics[s.key].add(res.timings.duration);
          payloadMetrics[s.key].add(res.body.length);
          sigLenMetrics[s.key].add(body.signature_length);
          gasUsedMetrics[s.key].add(body.gas_used);
          gasPriceMetrics[s.key].add(body.gas_price);
          successOps.add(1);
        }
      } else {
        failedOps.add(1);
        console.error(`${s.key} sign failed [iter ${iter}]: ${res.body}`);
      }
    }
  });

  // ==========================================================================
  // METRIC 2: Verification Execution Times
  // ==========================================================================
  group("Metric 2 - Verification", () => {
    for (const s of SCHEMES) {
      const sr = signResults[s.key];
      if (!sr) continue;

      const sd = setupData[s.key];
      const publicKey = s.key === "ecdsa" ? sd.sender : sd.publicKey;

      const res = http.post(
        `${BASE_URL}/api/verify`,
        JSON.stringify({
          sender: sd.sender,
          receiver: RECEIVER,
          amount: "1000000",
          message: `k6-${s.key}-${iter}`,
          nonce: sr.user_nonce_used,
          signature: sr.signature,
          public_key: publicKey,
          mode: s.mode,
          algorithm: s.signAlgo,
        }),
        { headers: HEADERS, tags: { operation: `verify_${s.key}` } }
      );

      if (!isWarmup) {
        verifyMetrics[s.key].add(res.timings.duration);
        const body = res.json();
        check(res, {
          [`${s.key} verify OK`]: () => s.key === "ecdsa"
            ? (res.status === 200 || res.status === 401)
            : (res.status === 200 && body.valid),
        });
        if (body.valid) {
          successOps.add(1);
        } else {
          failedOps.add(1);
          if (s.key !== "ecdsa") {
            console.warn(`${s.key} verify failed [iter ${iter}]: ${JSON.stringify(body)}`);
          }
        }
      }
    }
  });
}

// ============================================================================
// Custom Summary
// ============================================================================

export function handleSummary(data) {
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

  const pct = (base, comp) =>
    base && comp && base.avg > 0
      ? ((comp.avg - base.avg) / base.avg) * 100
      : null;

  const schemes = {};
  for (const { key } of SCHEMES) {
    schemes[key] = {
      keypair_gen_ms:          extract(`keypair_gen_${key}_ms`),
      sign_exec_ms:            extract(`sign_exec_${key}_ms`),
      verify_exec_ms:          extract(`verify_exec_${key}_ms`),
      payload_bytes:           extract(`tx_payload_size_${key}_bytes`),
      signature_length_bytes:  extract(`signature_length_${key}_bytes`),
      gas_used:                extract(`gas_used_${key}`),
      gas_price:               extract(`gas_price_${key}`),
    };
  }

  const ecdsaS = schemes.ecdsa;
  const overhead = {};
  for (const { key } of SCHEMES) {
    if (key === "ecdsa") continue;
    const s = schemes[key];
    overhead[key] = {
      keypair_gen_overhead_pct:    pct(ecdsaS.keypair_gen_ms, s.keypair_gen_ms),
      signing_overhead_pct:        pct(ecdsaS.sign_exec_ms,   s.sign_exec_ms),
      verification_overhead_pct:   pct(ecdsaS.verify_exec_ms, s.verify_exec_ms),
      gas_overhead_pct:            pct(ecdsaS.gas_used,        s.gas_used),
      payload_overhead_pct:        pct(ecdsaS.payload_bytes,   s.payload_bytes),
    };
  }

  const report = {
    test: "Hybrid PQC Benchmark",
    total_iterations: TOTAL_ITERATIONS,
    measured_iterations: TOTAL_ITERATIONS - 1,
    timestamp: new Date().toISOString(),
    schemes,
    overhead_vs_ecdsa: overhead,
    operations: {
      successful: data.metrics.successful_operations?.values.count ?? 0,
      failed: data.metrics.failed_operations?.values.count ?? 0,
    },
  };

  return {
    stdout: textSummary(data, { indent: " ", enableColors: true }),
    "results/benchmark_report.json": JSON.stringify(report, null, 2),
  };
}
