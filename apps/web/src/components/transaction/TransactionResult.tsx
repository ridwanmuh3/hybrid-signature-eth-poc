import { type TxResult } from "@/types/transaction";
import { Card, CardContent, CardHeader, CardTitle } from "../ui/card";
import { ArrowRight, CoinsIcon, FileJson, Server } from "lucide-react";
import CopyClipboardButton from "../shared/CopyClipboardButton";
import { formatAmount, trimLongText } from "@/lib/utils";
import { Button } from "../ui/button";
import { useConnection } from "wagmi";
import { formatEther } from "viem";
import { useContractLedgerBalance } from "@/features/web3/useContractLedgerBalance";

type PropsType = {
  txResult: TxResult;
};

const TransactionResult = ({ txResult }: PropsType) => {
  const { chain } = useConnection();
  const receiverAddress = txResult.receiver?.startsWith("0x")
    ? (txResult.receiver as `0x${string}`)
    : undefined;
  const { balance: receiverLedgerBalance } = useContractLedgerBalance(receiverAddress);

  const handleDownloadReceipt = () => {
    const data = JSON.stringify({ ...txResult }, null, 4);

    const blob = new Blob([data], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `local-tx-${txResult.transaction_hash?.slice(0, 8)}.json`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  return (
    <Card className="max-h-fit gap-3 bg-background border-input shadow-sm animate-in fade-in slide-in-from-bottom-4 duration-500">
      <CardHeader className="pb-2">
        <div className="flex gap-3 items-center justify-between">
          <div className="flex gap-3 items-center">
            <div className="p-2.5 border border-input bg-muted text-muted-foreground rounded-md">
              <CoinsIcon className="size-5" />
            </div>
            <div>
              <CardTitle>Transaction Sent</CardTitle>
              <div className="flex items-center gap-1 text-xs text-muted-foreground font-normal mt-1">
                <Server className="size-3" />
                <span>{chain?.name}</span>
              </div>
            </div>
          </div>
          <Button
            variant="outline"
            size="sm"
            className="gap-2 text-sm h-8 cursor-pointer"
            onClick={handleDownloadReceipt}
          >
            <FileJson className="size-4" />
            Save Receipt
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col gap-6">
          {txResult.sender && txResult.receiver && (
            <div className="flex items-center gap-3 p-3 bg-muted border border-input rounded-md">
              <div className="flex-1 min-w-0">
                <p className="text-xs text-muted-foreground font-medium mb-0.5">From</p>
                <p className="font-mono text-xs text-muted-foreground truncate">
                  {txResult.sender}
                </p>
              </div>
              <div className="flex flex-col items-center gap-0.5 shrink-0">
                <ArrowRight className="size-4 text-muted-foreground" />
                <p className="text-xs font-semibold text-muted-foreground">
                  {formatAmount(txResult.amount_wei, "wei")} ETH
                </p>
              </div>
              <div className="flex-1 min-w-0 text-right">
                <p className="text-xs text-muted-foreground font-medium mb-0.5">To</p>
                <p className="font-mono text-xs text-muted-foreground truncate">
                  {txResult.receiver}
                </p>
                {receiverAddress && (
                  <p className="text-xs text-green-600 font-semibold mt-0.5">
                    Receiver Ledger Balance:{" "}
                    {Number(formatEther(receiverLedgerBalance)).toFixed(4)}{" "}
                    ETH
                  </p>
                )}
              </div>
            </div>
          )}

          <div className="flex gap-6 justify-between md:flex-row flex-col">
            <div className="flex-1 flex flex-col gap-1">
              <p className="text-foreground font-semibold text-sm">Nonce</p>
              <div className="relative group">
                <p className="text-muted-foreground text-sm break-all py-2.5 px-3 pr-14 border border-input rounded-md shadow-sm bg-background max-h-40 overflow-y-auto font-mono">
                  {txResult.user_nonce_used}
                </p>
              </div>
            </div>
            <div className="flex-1 flex flex-col gap-1">
              <p className="text-foreground font-semibold text-sm">Gas Price</p>
              <div className="relative group">
                <p className="text-muted-foreground text-sm break-all py-2.5 px-3 pr-14 border border-input rounded-md shadow-sm bg-background max-h-40 overflow-y-auto font-mono">
                  {txResult.gas_price
                    ? formatAmount(txResult?.gas_price, "wei")
                    : 0}{" "}
                  ETH
                </p>
              </div>
            </div>
            <div className="flex-1 flex flex-col gap-1">
              <p className="text-foreground font-semibold text-sm">Gas Used</p>
              <div className="relative group">
                <p className="text-muted-foreground text-sm break-all py-2.5 px-3 pr-14 border border-input rounded-md shadow-sm bg-background max-h-40 overflow-y-auto font-mono">
                  {txResult.gas_used ? txResult.gas_used : 0} Unit
                </p>
              </div>
            </div>
          </div>
          <div className="flex gap-1 justify-between flex-col">
            <p className="text-foreground font-semibold text-sm">
              Transaction Hash
            </p>
            <div className="relative group">
              <p className="text-muted-foreground font-mono text-sm break-all py-2.5 px-3 pr-14 border border-input rounded-md shadow-sm bg-background">
                {txResult.transaction_hash}
              </p>
              <CopyClipboardButton
                content={txResult.transaction_hash || ""}
                className="absolute top-1 right-1"
              />
            </div>
          </div>
          <div className="flex gap-1 justify-between flex-col">
            <p className="text-foreground font-semibold text-sm">
              Transaction Signature
            </p>
            <div className="relative group">
              <p className="text-muted-foreground text-sm break-all py-2.5 px-3 pr-14 border border-input rounded-md shadow-sm bg-background max-h-40 overflow-y-auto font-mono">
                {trimLongText(txResult.signature || "", 200)}
              </p>
              <CopyClipboardButton
                content={txResult.signature || ""}
                className="absolute top-1 right-1"
              />
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export default TransactionResult;
