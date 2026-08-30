import type { Keypair } from "@/types/keypair";
import { create } from "zustand";

export type KeypairState = {
  keypair: Keypair;
  setKeypair: (keys: Keypair) => void;
};

export const useKeypairStore = create<KeypairState>()((set) => ({
  keypair: {
    status: "",
    algorithm: "",
    mode: "",
    private_key: "0x",
    public_key: "0x",
  },
  setKeypair: (keypair) => set(() => ({ keypair: keypair })),
}));
