import z from "zod";

export const verificationFormSchema = z.object({
  transactionHash: z
    .string()
    .regex(/^0x[a-fA-F0-9]{64}$/, "Invalid transaction hash format"),
  publicKey: z
    .string()
    .min(1, "Public key is required")
    .regex(/^0x[a-fA-F0-9]+$/, "Public key must be a hex string (0x...)"),
});
