import { createConfig, http } from "wagmi";
import { foundrynet, sepolianet } from "./chainnet/index";

export const config = createConfig({
  chains: [foundrynet, sepolianet],
  transports: {
    [foundrynet.id]: http(),
    [sepolianet.id]: http(),
  },
});
