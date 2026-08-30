import type { TxData, TxResult } from "@/types/transaction";
import { create } from "zustand";

export type TransactionResultState = {
  txResult: TxResult;
  setTxResult: (txResult: TxResult) => void;
};

export type TransactionDataState = {
  txData: TxData;
  setTxData: (txData: TxData) => void;
};

export const useTxResultStore = create<TransactionResultState>()((set) => ({
  txResult: {
    status: "",
    signing_algorithm: "",
    transaction_hash: "0x",
    user_nonce_used: 0,
    signature: "0x",
    signature_length: 0,
    gas_price: 0,
  },
  setTxResult: (txResult) => set(() => ({ txResult: txResult })),
}));

export const useTxDataStore = create<TransactionDataState>()((set) => ({
  txData: {
    tx_hash: "",
    status: 0,
    block_number: 0,
    relayer_addr: "",
    relayer_nonce: 0,
    user_nonce: "0",
    sender: "",
    receiver: "",
    signing_algorithm: "",
    signing_mode: "",
    amount_wei: "0",
    message: "",
    is_hybrid: false,
    signature_full: "",
    signature_ecdsa: "",
    signature_pqs: "",
  },
  setTxData: (txData) => set(() => ({ txData: txData })),
}));
