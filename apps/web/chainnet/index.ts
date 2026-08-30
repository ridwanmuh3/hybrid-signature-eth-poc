import { defineChain } from "viem";

export const foundrynet = defineChain({
  id: 31337,
  name: "Foundry Local Net",
  nativeCurrency: { name: "Ether", symbol: "ETH", decimals: 18 },
  rpcUrls: {
    default: { http: ["http://localhost:8540"] },
  },
  blockExplorers: {
    default: { name: "Etherscan", url: "https://etherscan.io" },
  },
});

export const sepolianet = defineChain({
  id: 11155111,
  name: "Sepolia Development Net",
  nativeCurrency: { name: "Ether", symbol: "ETH", decimals: 18 },
  rpcUrls: {
    default: {
      http: [
        import.meta.env.VITE_SEPOLIA_RPC_URL ||
          "https://ethereum-sepolia-rpc.publicnode.com",
      ],
    },
  },
  blockExplorers: {
    default: { name: "Etherscan", url: "https://etherscan.io" },
  },
});
