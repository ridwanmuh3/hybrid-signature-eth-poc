export const pqcAlgorithms = [
  "ML-DSA-44",
  "ML-DSA-65",
  "ML-DSA-87",
  "Falcon-512",
  "Falcon-1024",
  "MAYO-1",
  "MAYO-3",
  "MAYO-5",
  "SNOVA_24_5_4",
  "SNOVA_37_8_4",
  "SNOVA_60_10_4",
] as const;

export const signatureAlgorithmMode: Readonly<["single", "hybrid"]> = [
  "single",
  "hybrid",
];

export const allAlgorithms = ["ECDSA", ...pqcAlgorithms] as const;
