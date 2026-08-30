import { allAlgorithms, signatureAlgorithmMode } from "@/constants";
import z from "zod";

export const generateKeypairFormSchema = z.object({
  algorithm: z.enum(allAlgorithms, { error: "Invalid algorithm" }),
  mode: z.enum(signatureAlgorithmMode, {
    error: "Mode is required",
  }),
  ecdsaPrivateKey: z
    .string()
    .optional()
    .refine((val) => {
      if (!val) return true;
      return /^0x[a-fA-F0-9]{64}$|^[a-fA-F0-9]{64}$/.test(val);
    }, "Invalid Hex Private Key format"),
});
