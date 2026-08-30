import { useEffect, useState } from "react";
import {
  useConnection,
  useWriteContract,
  useWaitForTransactionReceipt,
} from "wagmi";
import { parseEther } from "viem";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import {
  CONTRACT_ABI,
  HAS_CONTRACT_ADDRESS,
  RESOLVED_CONTRACT_ADDRESS,
} from "@/constants";

const DepositWithdraw = () => {
  const { isConnected } = useConnection();

  const [depositAmount, setDepositAmount] = useState("");
  const [withdrawAmount, setWithdrawAmount] = useState("");
  const [depositError, setDepositError] = useState("");
  const [withdrawError, setWithdrawError] = useState("");

  const canUseContract = HAS_CONTRACT_ADDRESS;

  const {
    writeContractAsync: depositAsync,
    isPending: isDepositing,
    data: depositTxHash,
  } = useWriteContract();
  const {
    writeContractAsync: withdrawAsync,
    isPending: isWithdrawing,
    data: withdrawTxHash,
  } = useWriteContract();

  const depositReceipt = useWaitForTransactionReceipt({ hash: depositTxHash });
  const withdrawReceipt = useWaitForTransactionReceipt({
    hash: withdrawTxHash,
  });

  // Refetch parent balances via page-level refresh (simple approach)
  useEffect(() => {
    if (depositReceipt.data?.transactionHash) {
      window.dispatchEvent(new CustomEvent("balance-updated"));
    }
  }, [depositReceipt.data?.transactionHash]);

  useEffect(() => {
    if (withdrawReceipt.data?.transactionHash) {
      window.dispatchEvent(new CustomEvent("balance-updated"));
    }
  }, [withdrawReceipt.data?.transactionHash]);

  const handleDeposit = async () => {
    setDepositError("");
    if (
      !depositAmount ||
      isNaN(Number(depositAmount)) ||
      Number(depositAmount) <= 0
    ) {
      setDepositError("Enter a valid ETH amount");
      return;
    }
    try {
      await depositAsync({
        address: RESOLVED_CONTRACT_ADDRESS,
        abi: CONTRACT_ABI,
        functionName: "deposit",
        value: parseEther(depositAmount),
      });
      setDepositAmount("");
    } catch (err: unknown) {
      setDepositError("Deposit failed: " + (err as Error).message);
    }
  };

  const handleWithdraw = async () => {
    setWithdrawError("");
    if (
      !withdrawAmount ||
      isNaN(Number(withdrawAmount)) ||
      Number(withdrawAmount) <= 0
    ) {
      setWithdrawError("Enter a valid ETH amount");
      return;
    }
    try {
      await withdrawAsync({
        address: RESOLVED_CONTRACT_ADDRESS,
        abi: CONTRACT_ABI,
        functionName: "withdraw",
        args: [parseEther(withdrawAmount)],
      });
      setWithdrawAmount("");
    } catch (err: unknown) {
      setWithdrawError("Withdraw failed: " + (err as Error).message);
    }
  };

  if (!isConnected) return null;

  return (
    <div className="border-t border-border pt-3 space-y-3">
      <div className="flex gap-2 flex-col sm:flex-row sm:items-end">
        <div className="flex-1 flex flex-col gap-1">
          <Input
            type="number"
            step="0.0001"
            min="0"
            placeholder="ETH amount"
            value={depositAmount}
            onChange={(e) => setDepositAmount(e.target.value)}
            className="text-xs"
            aria-invalid={!!depositError}
            aria-describedby={depositError ? "deposit-error" : undefined}
          />
          {depositError && (
            <p
              id="deposit-error"
              className="text-destructive text-xs"
              role="alert"
            >
              {depositError}
            </p>
          )}
        </div>
        <Button
          size="sm"
          onClick={handleDeposit}
          disabled={isDepositing || !canUseContract}
          className="cursor-pointer shrink-0"
        >
          {isDepositing ? <Spinner /> : "Deposit"}
        </Button>
      </div>
      <div className="flex gap-2 flex-col sm:flex-row sm:items-end">
        <div className="flex-1 flex flex-col gap-1">
          <Input
            type="number"
            step="0.0001"
            min="0"
            placeholder="ETH amount"
            value={withdrawAmount}
            onChange={(e) => setWithdrawAmount(e.target.value)}
            className="text-xs"
            aria-invalid={!!withdrawError}
            aria-describedby={withdrawError ? "withdraw-error" : undefined}
          />
          {withdrawError && (
            <p
              id="withdraw-error"
              className="text-destructive text-xs"
              role="alert"
            >
              {withdrawError}
            </p>
          )}
        </div>
        <Button
          size="sm"
          variant="outline"
          onClick={handleWithdraw}
          disabled={isWithdrawing || !canUseContract}
          className="cursor-pointer shrink-0"
        >
          {isWithdrawing ? <Spinner /> : "Withdraw"}
        </Button>
      </div>
    </div>
  );
};

export default DepositWithdraw;
