import { Card } from "@/components/ui/card";
import WalletDetails from "./WalletDetails";
import DepositWithdraw from "./DepositWithdraw";

const WalletInfo = () => {
  return (
    <Card className="max-h-fit gap-3 border-input shadow-sm bg-background animate-in fade-in slide-in-from-top-4 duration-500">
      <WalletDetails />
      <div className="px-6 pb-6">
        <DepositWithdraw />
      </div>
    </Card>
  );
};

export default WalletInfo;
