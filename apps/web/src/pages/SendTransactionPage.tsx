import { NavLink } from "react-router";
import { KeyRound } from "lucide-react";
import { TransactionForm, WalletInfo, TransactionResult } from "@/components";
import { Card, CardContent } from "@/components/ui/card";
import { useTxResultStore } from "@/stores/transaction";
import { useKeypairStore } from "@/stores/keypair";

const NoKeypairPrompt = () => (
  <Card className="max-h-fit gap-3 border-amber-200 bg-amber-50 shadow-sm animate-in fade-in slide-in-from-top-4 duration-500">
    <CardContent className="pt-6 flex flex-col items-center gap-4 text-center">
      <div className="p-3 border border-amber-300 bg-amber-100 rounded-md">
        <KeyRound className="size-6 text-amber-700" />
      </div>
      <div>
        <p className="font-semibold text-amber-900 text-sm">No keypair detected</p>
        <p className="text-amber-700 text-sm mt-1">
          Generate a keypair before sending a transaction.
        </p>
      </div>
      <NavLink
        to="/generate-keypair"
        className="px-4 py-2 bg-amber-700 text-white text-sm font-semibold rounded-md hover:bg-amber-800"
      >
        Generate Keypair
      </NavLink>
    </CardContent>
  </Card>
);

const SendTransactionPage = () => {
  const { txResult } = useTxResultStore();
  const { keypair } = useKeypairStore();
  const hasKeypair = keypair.public_key !== "0x";

  return (
    <div className="max-w-5xl mx-auto p-8 space-y-8">
      <h1 className="text-2xl font-semibold text-foreground">Send Transaction</h1>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
        <WalletInfo />
        {hasKeypair ? <TransactionForm /> : <NoKeypairPrompt />}
      </div>
      {hasKeypair && txResult.signature !== "0x" && txResult.transaction_hash !== "0x" ? (
        <TransactionResult txResult={txResult} />
      ) : null}
    </div>
  );
};

export default SendTransactionPage;
