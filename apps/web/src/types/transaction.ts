export type TxResult = Partial<{
  status: string;
  signing_algorithm: string;
  transaction_hash: `0x${string}`;
  user_nonce_used: number;
  sender: string;
  receiver: string;
  amount_wei: string;
  signature: `0x${string}`;
  signature_length: number;
  gas_price: number;
  gas_used: number;
}>;

export type TxData = Partial<{
  tx_hash: string;
  status: number;
  block_number: number;
  relayer_addr: string;
  relayer_nonce: number;
  user_nonce: string;
  sender: string;
  receiver: string;
  signing_algorithm: string;
  signing_mode: string;
  amount_wei: string;
  message: string;
  is_hybrid: boolean;
  signature_full: string;
  signature_ecdsa: string;
  signature_pqs: string;
  gas_price: number;
  gas_used: number;
}>;

export type VerifyResult = Partial<{
  valid: boolean;
  message: string;
  verify_message: string;
  note: string;
}>;

export type VerifyResponse = VerifyResult & TxData;
