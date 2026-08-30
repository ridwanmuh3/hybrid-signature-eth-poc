import { useConnect, useConnectors } from "wagmi";
import { DropdownMenuItem } from "@/components/ui/dropdown-menu";

const WalletOptions = () => {
  const connect = useConnect();
  const connectors = useConnectors();

  return (
    <>
      {connectors.map((connector) => (
        <DropdownMenuItem
          key={connector.id}
          onClick={() => connect.mutate({ connector })}
        >
          {connector.name}
        </DropdownMenuItem>
      ))}
    </>
  );
};

export default WalletOptions;
