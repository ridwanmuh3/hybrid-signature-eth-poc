import { HAS_CONTRACT_ADDRESS, RESOLVED_CONTRACT_ADDRESS } from "@/constants";
import { useBalance, useChainId, useWatchBlockNumber } from "wagmi";

export const useContractEthBalance = () => {
  const chainId = useChainId();

  const query = useBalance({
    address: RESOLVED_CONTRACT_ADDRESS,
    chainId,
    query: { enabled: HAS_CONTRACT_ADDRESS },
  });

  useWatchBlockNumber({
    chainId,
    enabled: HAS_CONTRACT_ADDRESS,
    onBlockNumber: () => {
      void query.refetch();
    },
  });

  return {
    balance: query.data?.value ?? 0n,
    refetch: query.refetch,
  };
};
