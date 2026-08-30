import { createPublicClient, http } from "viem";
import { ganachenet, hardhatnet } from "../chainnet/index";

export const getChainClient = async (net: string) => {
  const client = createPublicClient({
    chain: net === "hardhat" ? hardhatnet : ganachenet,
    transport: http(),
  });
  return client;
};
