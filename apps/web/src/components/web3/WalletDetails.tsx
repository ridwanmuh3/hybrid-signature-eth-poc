import { useConnection, useBytecode } from "wagmi";
import { CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Wallet } from "lucide-react";
import { formatEther } from "viem";
import { useKeypairStore } from "@/stores/keypair";
import {
  HAS_CONTRACT_ADDRESS,
  RESOLVED_CONTRACT_ADDRESS,
} from "@/constants";
import { useContractLedgerBalance } from "@/features/web3/useContractLedgerBalance";
import { useContractEthBalance } from "@/features/web3/useContractEthBalance";

const WalletDetails = () => {
  const { address, isConnected, chain } = useConnection();
  const { keypair } = useKeypairStore();

  const { balance: ledgerBalance } = useContractLedgerBalance(address);
  const { balance: contractEthBalance } = useContractEthBalance();
  const contractBytecode = useBytecode({
    address: RESOLVED_CONTRACT_ADDRESS,
    query: { enabled: HAS_CONTRACT_ADDRESS },
  });

  const roundedLedgerBalance = Number(formatEther(ledgerBalance)).toFixed(4);
  const roundedContractEthBalance = Number(
    formatEther(contractEthBalance),
  ).toFixed(4);
  const hasContractCode = Boolean(
    contractBytecode.data && contractBytecode.data !== "0x",
  );
  const canUseContract = HAS_CONTRACT_ADDRESS && hasContractCode;
  const contractStatus = !HAS_CONTRACT_ADDRESS
    ? "Not configured"
    : contractBytecode.isLoading
      ? "Checking deployment"
      : hasContractCode
        ? "Ready"
        : "No deployed contract on this network";

  return (
    <>
      <CardHeader>
        <div className="flex gap-3 items-center">
          <div className="p-2.5 border border-input text-foreground bg-muted rounded-md">
            <Wallet className="size-5 text-muted-foreground" />
          </div>
          <CardTitle className="text-base">Wallet Info</CardTitle>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div className="flex gap-1 justify-between flex-col">
            <p className="text-foreground font-semibold text-sm">
              Contract Address
            </p>
            <p className="text-muted-foreground text-sm break-all">
              {HAS_CONTRACT_ADDRESS
                ? RESOLVED_CONTRACT_ADDRESS
                : "Not configured"}
            </p>
          </div>
          <div className="flex gap-1 justify-between flex-col">
            <p className="text-foreground font-semibold text-sm">
              Contract Status
            </p>
            <p className="text-muted-foreground text-sm break-all">
              {contractStatus}
            </p>
          </div>
          <div className="flex gap-1 justify-between flex-col">
            <p className="text-foreground font-semibold text-sm">
              Contract Wallet Balance
            </p>
            <p className="text-muted-foreground text-sm break-all">
              {canUseContract
                ? `${roundedContractEthBalance} ETH`
                : "Not available"}
            </p>
          </div>
        </div>
        {isConnected ? (
          <div className="space-y-4 mt-4 border-t border-border pt-4">
            <div className="flex gap-1 justify-between flex-col">
              <p className="text-foreground font-semibold text-sm">
                Your Address
              </p>
              <p className="text-muted-foreground text-sm break-all">
                {address}
              </p>
            </div>
            <div className="flex gap-1 justify-between flex-col">
              <p className="text-foreground font-semibold text-sm">Network</p>
              <p className="text-muted-foreground text-sm break-all">
                {chain?.name}
              </p>
            </div>
            <div className="flex gap-1 justify-between flex-col">
              <p className="text-foreground font-semibold text-sm">
                Your Ledger Balance
              </p>
              <p className="text-muted-foreground text-sm break-all">
                {roundedLedgerBalance} ETH
              </p>
            </div>
            {keypair.algorithm && (
              <div className="flex gap-1 justify-between flex-col">
                <p className="text-foreground font-semibold text-sm">
                  Signature Algorithm
                </p>
                <p className="text-muted-foreground text-sm break-all">
                  {keypair.algorithm === "ECDSA"
                    ? "Secp256k1"
                    : keypair.algorithm}
                </p>
              </div>
            )}
          </div>
        ) : (
          <p className="text-muted-foreground break-all text-center">
            Please connect your wallet!
          </p>
        )}
      </CardContent>
    </>
  );
};

export default WalletDetails;
