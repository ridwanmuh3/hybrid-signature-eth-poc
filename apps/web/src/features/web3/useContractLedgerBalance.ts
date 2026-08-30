import { CONTRACT_CONFIG, HAS_CONTRACT_ADDRESS } from "@/constants";
import { type Address, zeroAddress } from "viem";
import { useChainId, useReadContract, useWatchBlockNumber } from "wagmi";

export const useContractLedgerBalance = (account?: Address) => {
  const chainId = useChainId();
  const enabled = HAS_CONTRACT_ADDRESS && Boolean(account);

  const query = useReadContract({
    ...CONTRACT_CONFIG,
    functionName: "balances",
    args: [account ?? zeroAddress],
    chainId,
    query: { enabled },
  });

  useWatchBlockNumber({
    chainId,
    enabled,
    onBlockNumber: () => {
      void query.refetch();
    },
  });

  return {
    balance: query.data ?? 0n,
    refetch: query.refetch,
  };
};
