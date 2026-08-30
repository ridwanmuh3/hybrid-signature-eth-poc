import { VerificationForm, VerificationResult, WalletInfo } from "@/components";
import { useTxDataStore } from "@/stores/transaction";
import { useTxVerifyStore } from "@/stores/verification";

const VerifyTransactionPage = () => {
  const { verifyResult } = useTxVerifyStore();
  const { txData } = useTxDataStore();
  return (
    <>
      <div className="max-w-5xl mx-auto p-8 space-y-8">
        <h1 className="text-2xl font-semibold text-foreground">Verify Transaction</h1>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
          <WalletInfo />
          <VerificationForm />
        </div>
        {verifyResult.note || txData.tx_hash ? <VerificationResult /> : null}
      </div>
    </>
  );
};

export default VerifyTransactionPage;
