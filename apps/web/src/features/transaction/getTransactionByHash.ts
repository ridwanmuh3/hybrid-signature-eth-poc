import { axiosInstance, queryClient } from "@/lib/utils";
import type { TxData } from "@/types/transaction";

const getTransactionByHash = async (
  transactionHash: `0x${string}` | string,
) => {
  const response = await axiosInstance.get<TxData>(
    `/api/transactions/${transactionHash}`,
  );

  return response;
};

export const fetchTxDataHandler = async (
  transactionHash: `0x${string}` | string,
) => {
  try {
    const { data } = await queryClient.fetchQuery({
      queryKey: ["transaction", transactionHash],
      queryFn: () => getTransactionByHash(transactionHash),
    });

    if (!data) {
      alert("Transaction not found");
      return null;
    }

    return data;
  } catch (err) {
    console.error(err);
    alert("Transaction not found");
    return null;
  }
};
