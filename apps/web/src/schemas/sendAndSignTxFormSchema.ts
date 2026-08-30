import { allAlgorithms, signatureAlgorithmMode } from "@/constants";
import z from "zod";

export const sendAndSignTxFormSchema = z.object({
  receiverAddress: z
    .string()
    .min(1, "Receiver address is required")
    .regex(/^0x[a-fA-F0-9]{40}$/, "Invalid Ethereum Address"),
  message: z.string().min(1, "Message is required"),
  amount: z
    .string()
    .min(1, "Amount is required")
    .regex(/^\d+(\.\d{1,18})?$/, "Invalid amount format"),
  privateKey: z
    .string()
    .min(1, "Private key is required")
    .regex(/^0x[a-fA-F0-9]+$/, "Private key must be a hex string (0x...)"),
  algorithm: z.enum(allAlgorithms, { error: "Invalid algorithm" }),
  mode: z.enum(signatureAlgorithmMode, {
    error: "Mode is required",
  }),
});
