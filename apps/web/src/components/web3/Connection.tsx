import { useDisconnect } from "wagmi";
import { Button } from "@/components/ui/button";

const Connection = () => {
  const disconnect = useDisconnect();

  const handleDisconnect = () => {
    const confirm = window.confirm("Are you sure you want to disconnect?");
    if (confirm) {
      disconnect.mutate();
    }
  };

  return (
    <>
      <Button
        className="cursor-pointer hover:bg-zinc-600"
        onClick={handleDisconnect}
      >
        Disconnect
      </Button>
    </>
  );
};

export default Connection;
