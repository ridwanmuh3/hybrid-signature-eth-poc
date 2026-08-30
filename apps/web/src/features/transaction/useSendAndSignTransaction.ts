import { axiosInstance } from "@/lib/utils";
import type { TxResult } from "@/types/transaction";
import { useMutation } from "@tanstack/react-query";

export const useSendAndSignTransaction = () => {
  return useMutation({
    mutationFn: async (body: any) => {
      const response = await axiosInstance.post<TxResult>("/api/sign", body);

      return response;
    },
  });
};
