import { zeroAddress } from "viem";

export const CONTRACT_ADDRESS = (import.meta.env.VITE_CONTRACT_ADDRESS ??
  "") as `0x${string}` | "";

export const HAS_CONTRACT_ADDRESS = CONTRACT_ADDRESS.length > 0;

export const RESOLVED_CONTRACT_ADDRESS = (HAS_CONTRACT_ADDRESS
  ? CONTRACT_ADDRESS
  : zeroAddress) as `0x${string}`;

export const CONTRACT_ABI = [
  {
    type: "function",
    name: "balances",
    inputs: [{ name: "", type: "address" }],
    outputs: [{ name: "", type: "uint256" }],
    stateMutability: "view",
  },
  {
    type: "function",
    name: "deposit",
    inputs: [],
    outputs: [],
    stateMutability: "payable",
  },
  {
    type: "function",
    name: "withdraw",
    inputs: [{ name: "amount", type: "uint256" }],
    outputs: [],
    stateMutability: "nonpayable",
  },
  {
    type: "event",
    name: "Deposited",
    inputs: [
      { name: "user", type: "address", indexed: true },
      { name: "amount", type: "uint256", indexed: false },
    ],
  },
  {
    type: "event",
    name: "Withdrawn",
    inputs: [
      { name: "user", type: "address", indexed: true },
      { name: "amount", type: "uint256", indexed: false },
    ],
  },
] as const;

export const CONTRACT_CONFIG = {
  address: RESOLVED_CONTRACT_ADDRESS,
  abi: CONTRACT_ABI,
} as const;
