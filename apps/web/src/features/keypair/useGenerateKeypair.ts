import { axiosInstance } from "@/lib/utils";
import type { Keypair } from "@/types/keypair";
import { useMutation } from "@tanstack/react-query";

export const useGenerateKeypair = () => {
  return useMutation({
    mutationFn: async (body: any) => {
      const response = await axiosInstance.post<Keypair>(
        "/api/generate-key",
        body
      );

      return response;
    },
  });
};
