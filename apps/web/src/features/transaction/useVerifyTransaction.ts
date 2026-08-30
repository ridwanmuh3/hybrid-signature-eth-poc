import { axiosInstance } from "@/lib/utils";
import type { VerifyResponse } from "@/types/transaction";
import { useMutation } from "@tanstack/react-query";

export const useVerifyTransaction = () => {
  return useMutation({
    mutationFn: async (body: { tx_hash: string; public_key: string }) => {
      const response = await axiosInstance.post<VerifyResponse>(
        "/api/verify",
        body
      );

      return response;
    },
  });
};
