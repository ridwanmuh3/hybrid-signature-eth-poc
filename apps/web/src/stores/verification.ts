import type { VerifyResult } from "@/types/transaction";
import { create } from "zustand";

export type VerificationResultState = {
  verifyResult: VerifyResult;
  setVerifyResult: (verifyResult: VerifyResult) => void;
};

export const useTxVerifyStore = create<VerificationResultState>()((set) => ({
  verifyResult: {
    valid: false,
    message: "",
    note: "",
  },
  setVerifyResult: (verifyResult) =>
    set(() => ({ verifyResult: verifyResult })),
}));
