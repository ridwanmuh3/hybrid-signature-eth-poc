import { QueryClientProvider } from "@tanstack/react-query";
import { WagmiProvider } from "wagmi";
import { config } from "../../wagmi.config";
import type React from "react";
import { BrowserRouter } from "react-router";
import { queryClient } from "@/lib/utils";

const Providers = ({ children }: React.PropsWithChildren) => {
  return (
    <WagmiProvider config={config}>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>{children}</BrowserRouter>
      </QueryClientProvider>
    </WagmiProvider>
  );
};

export default Providers;
