import { useForm } from "react-hook-form";
import { useState } from "react";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { Input } from "@/components/ui/input";
import { Button } from "../ui/button";
import { useConnection } from "wagmi";
import { useSendAndSignTransaction } from "@/features/transaction/useSendAndSignTransaction";
import { useTxResultStore } from "@/stores/transaction";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import { pqcAlgorithms } from "@/constants";
import { parseEther } from "viem";
import { Spinner } from "../ui/spinner";
import { sendAndSignTxFormSchema } from "@/schemas/sendAndSignTxFormSchema";

const TransactionForm = () => {
  const form = useForm<z.infer<typeof sendAndSignTxFormSchema>>({
    resolver: zodResolver(sendAndSignTxFormSchema),
    defaultValues: {
      receiverAddress: "",
      message: "",
      amount: "",
      privateKey: "",
      algorithm: "ML-DSA-65",
      mode: "hybrid",
    },
  });

  const sendAndSignTx = useSendAndSignTransaction();
  const { setTxResult } = useTxResultStore();
  const { address } = useConnection();

  const [formError, setFormError] = useState("");
  const [successMessage, setSuccessMessage] = useState("");

  const selectedAlgorithm = form.watch("algorithm");
  const selectedMode = form.watch("mode");

  const onSubmit = async (values: z.infer<typeof sendAndSignTxFormSchema>) => {
    setFormError("");
    setSuccessMessage("");

    if (!address) {
      setFormError("Please connect your wallet first");
      return;
    }

    try {
      const payload = {
        sender: address,
        receiver: values.receiverAddress,
        message: values.message,
        amount: parseEther(values.amount).toString(),
        private_key: values.privateKey,
        algorithm:
          values.mode === "hybrid"
            ? "Hybrid-Secp256k1-" + values.algorithm
            : values.algorithm,
        mode: values.mode,
      };

      const { data } = await sendAndSignTx.mutateAsync(payload);

      setTxResult(data);
      setSuccessMessage("Transaction submitted successfully");
    } catch (err: unknown) {
      const e = err as Error;
      console.error(e);
      setFormError(e.message);
    }
  };

  return (
    <>
      <Card className="max-h-fit gap-3 border-input  shadow-sm bg-background animate-in fade-in slide-in-from-top-4 duration-500">
        <CardHeader>
          <CardTitle className="text-base">Send Transaction</CardTitle>
          <CardDescription>
            Sign and send payload using Hybrid or Post-Quantum schemes.
          </CardDescription>
        </CardHeader>
        {(formError || successMessage) && (
          <div className="px-6 pb-4">
            {formError && (
              <p className="text-destructive text-sm" role="alert">{formError}</p>
            )}
            {successMessage && (
              <p className="text-green-700 text-sm" role="status">{successMessage}</p>
            )}
          </div>
        )}
        <CardContent>
          <div>
            <Form {...form}>
              <form
                onSubmit={form.handleSubmit(onSubmit)}
                className="space-y-6"
              >
                <FormField
                  control={form.control}
                  name="algorithm"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="text-sm text-foreground font-semibold">
                        Signature Algorithm
                      </FormLabel>
                      <Select
                        onValueChange={(value) => {
                          field.onChange(value);
                          if (value === "ECDSA") {
                            form.setValue("mode", "single");
                          }
                        }}
                        defaultValue={field.value}
                      >
                        <FormControl className="w-full">
                          <SelectTrigger>
                            <SelectValue placeholder="Choose Algorithm" />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          <SelectItem value="ECDSA" className="font-semibold">
                            Secp256k1 (Standard Ethereum)
                          </SelectItem>
                          {pqcAlgorithms.map((alg) => (
                            <SelectItem key={alg} value={alg}>
                              {alg} (Post-Quantum)
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="mode"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="text-sm text-foreground font-semibold">
                        Mode
                      </FormLabel>
                      <Select
                        onValueChange={field.onChange}
                        value={field.value}
                        defaultValue={field.value}
                      >
                        <FormControl>
                          <SelectTrigger className="w-full">
                            <SelectValue placeholder="Choose Mode" />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          <SelectItem value="single">Single Scheme</SelectItem>
                          <SelectItem
                            value="hybrid"
                            disabled={selectedAlgorithm === "ECDSA"}
                          >
                            Hybrid Scheme (Secp256k1 + PQC)
                          </SelectItem>
                        </SelectContent>
                      </Select>

                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="receiverAddress"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="text-sm text-foreground font-semibold">
                        Receiver Address
                      </FormLabel>
                      <FormControl>
                        <Input placeholder="0x..." {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <div className="grid grid-cols-2 gap-4">
                  <FormField
                    control={form.control}
                    name="amount"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel className="text-sm text-foreground font-semibold">
                          Amount (ETH)
                        </FormLabel>
                        <FormControl>
                          <Input
                            type="number"
                            step="0.000000000000000001"
                            placeholder="0.01"
                            {...field}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name="message"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel className="text-sm text-foreground font-semibold">
                          Message (Data)
                        </FormLabel>
                        <FormControl>
                          <Input placeholder="Hello Blockchain" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
                <FormField
                  control={form.control}
                  name="privateKey"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="text-sm text-foreground font-semibold">
                        Private Key
                      </FormLabel>
                      <FormControl>
                        <Input
                          type="password"
                          placeholder={
                            selectedMode === "hybrid"
                              ? "Paste your LONG hybrid key here..."
                              : "0x..."
                          }
                          {...field}
                        />
                      </FormControl>
                      <p className="text-[10px] text-muted-foreground">
                        {selectedMode === "hybrid"
                          ? `Expects a concatenated Hex key (Secp256k1 + ${selectedAlgorithm})`
                          : `Expects a standard 32-byte Hex key or ${selectedAlgorithm} key`}
                      </p>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <div className="flex justify-end">
                  <Button
                    className="w-full cursor-pointer"
                    type="submit"
                    disabled={sendAndSignTx.isPending}
                  >
                    {sendAndSignTx.isPending ? (
                      <Spinner />
                    ) : (
                      "Sign & Send Transaction"
                    )}
                  </Button>
                </div>
              </form>
            </Form>
          </div>
        </CardContent>
      </Card>
    </>
  );
};

export default TransactionForm;
