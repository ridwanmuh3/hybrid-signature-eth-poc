import { FileText, CheckCircle2, XCircle } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useTxDataStore } from "@/stores/transaction";
import { formatAmount } from "@/lib/utils";
import { Alert, AlertDescription, AlertTitle } from "../ui/alert";
import { useTxVerifyStore } from "@/stores/verification";

const VerificationResult = () => {
  const { verifyResult } = useTxVerifyStore();
  const { txData } = useTxDataStore();
  const isDataLoaded = txData && txData.tx_hash && txData.tx_hash.length > 2;

  return (
    <>
      {isDataLoaded && (
        <Card className="mt-6 max-h-fit bg-background border-input shadow-sm animate-in fade-in slide-in-from-bottom-4 duration-500">
          <CardHeader className="border-b border-border">
            <div className="flex gap-3 items-center">
              <div className="p-2.5 border border-input bg-muted text-muted-foreground rounded-md">
                <FileText className="size-5" />
              </div>
              <div>
                <CardTitle className="text-base">On-Chain Data Found</CardTitle>
                <div className="flex items-center gap-1 text-sm text-muted-foreground font-normal mt-0.5">
                  <span className="font-mono">
                    Block #{txData.block_number}
                  </span>
                  <span className="mx-1">•</span>
                  <span>Fetched successfully</span>
                </div>
              </div>
            </div>
          </CardHeader>
          <CardContent className="">
            <div className="flex flex-col gap-5">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="flex flex-col gap-1.5">
                  <p className="text-sm font-semibold  tracking-wider">
                    Amount
                  </p>
                  <div className="py-2 px-3 border border-input bg rounded-md shadow-sm">
                    <p className="font-mono text-sm text-muted-foreground  ">
                      {formatAmount(txData.amount_wei, "wei")} ETH
                    </p>
                  </div>
                </div>
                <div className="flex flex-col gap-1.5">
                  <p className="text-sm font-semibold  tracking-wider">
                    User Nonce
                  </p>
                  <div className="py-2 px-3 border border-input bg rounded-md shadow-sm">
                    <p className="font-mono text-sm text-muted-foreground  ">
                      {txData.user_nonce}
                    </p>
                  </div>
                </div>
                <div className="col-span-1 md:col-span-2 flex flex-col gap-1.5">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <p className="text-sm font-semibold  tracking-wider mb-1.5">
                        Sender
                      </p>
                      <div className="py-2 px-3 border border-input bg rounded-md shadow-sm">
                        <p className="font-mono text-sm break-all text-muted-foreground">
                          {txData.sender}
                        </p>
                      </div>
                    </div>
                    <div>
                      <p className="text-sm font-semibold  tracking-wider mb-1.5">
                        Receiver
                      </p>
                      <div className="py-2 px-3 border border-input bg rounded-md shadow-sm">
                        <p className="font-mono text-sm break-all text-muted-foreground">
                          {txData.receiver}
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
                <div className="flex flex-col gap-1.5">
                  <p className="text-sm font-semibold  tracking-wider">
                    Algorithm
                  </p>
                  <div className="py-2 px-3 border border-input bg rounded-md shadow-sm">
                    <p className="font-mono text-sm text-muted-foreground">
                      {txData.signing_algorithm === "ECDSA"
                        ? "Secp256k1"
                        : txData.signing_algorithm}
                    </p>
                  </div>
                </div>
                <div className="flex flex-col gap-1.5">
                  <p className="text-sm font-semibold tracking-wider">Mode</p>
                  <div className="py-2 px-3 border border-input bg rounded-md shadow-sm">
                    <p className="font-mono text-sm text-muted-foreground">
                      {txData.signing_mode}
                    </p>
                  </div>
                </div>
                <div className="col-span-1 md:col-span-2 flex flex-col gap-1.5">
                  <p className="text-sm font-semibold tracking-wider">
                    Message
                  </p>
                  <div className="py-2 px-3 border border-input bg rounded-md shadow-sm">
                    <p className="font-mono text-sm text-muted-foreground italic">
                      "{txData.message}"
                    </p>
                  </div>
                </div>
              </div>
            </div>
            {verifyResult.note && (
              <div className="mt-6 animate-in fade-in slide-in-from-top-4 duration-500">
                <Alert
                  variant={verifyResult.valid ? "default" : "destructive"}
                  className={`border transition-all duration-300 ${
                    verifyResult.valid
                      ? "bg-green-50 border-green-200 text-green-800"
                      : "bg-red-50 border-red-200 text-red-800"
                  }`}
                >
                  {verifyResult.valid ? (
                    <CheckCircle2 className="h-4 w-4 text-green-600" />
                  ) : (
                    <XCircle className="h-4 w-4 text-red-600" />
                  )}
                  <AlertTitle className="font-semibold">
                    {verifyResult.valid
                      ? "Valid Transaction"
                      : "Verification Failed"}
                  </AlertTitle>
                  <AlertDescription className="text-sm mt-1 opacity-90">
                    {verifyResult.note}
                  </AlertDescription>
                </Alert>
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </>
  );
};

export default VerificationResult;
