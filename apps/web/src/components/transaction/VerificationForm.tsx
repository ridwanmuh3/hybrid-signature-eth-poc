import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { Search } from "lucide-react";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  FormDescription,
} from "@/components/ui/form";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import { useVerifyTransaction } from "@/features/transaction/useVerifyTransaction";
import { Spinner } from "../ui/spinner";
import { useTxVerifyStore } from "@/stores/verification";
import { verificationFormSchema } from "@/schemas/verificationFormSchema";
import { useTxDataStore } from "@/stores/transaction";
import type { AxiosError } from "axios";
import type { TxData, VerifyResponse } from "@/types/transaction";

const VerificationForm = () => {
  const form = useForm<z.infer<typeof verificationFormSchema>>({
    resolver: zodResolver(verificationFormSchema),
    defaultValues: {
      transactionHash: "",
      publicKey: "",
    },
  });

  const { setVerifyResult } = useTxVerifyStore();
  const { setTxData } = useTxDataStore();
  const verifyTransaction = useVerifyTransaction();

  const onSubmit = async (values: z.infer<typeof verificationFormSchema>) => {
    try {
      const { data } = await verifyTransaction.mutateAsync({
        tx_hash: values.transactionHash,
        public_key: values.publicKey,
      });

      if (data.tx_hash) {
        setTxData(data as unknown as TxData);
      }

      setVerifyResult({
        valid: data.valid,
        message: data.verify_message ?? data.message,
        note: data.note,
      });
    } catch (err: unknown) {
      const e = err as AxiosError;
      if (e.response) {
        const respData = e.response.data as VerifyResponse;

        if (respData?.tx_hash) {
          setTxData(respData as unknown as TxData);
        }

        if (e.response.status === 401) {
          setVerifyResult({
            valid: false,
            message: respData?.verify_message ?? "Verification Failed",
            note: respData?.note ?? "Invalid signature or public key",
          });
          return;
        }

        if (e.response.status === 404) {
          alert("Transaction not found on-chain");
          return;
        }

        alert(`Error: ${JSON.stringify(respData)}`);
      } else {
        alert("Network error");
      }
    }
  };

  return (
    <Card className="max-h-fit gap-3 border-input shadow-sm bg-background animate-in fade-in slide-in-from-top-4 duration-500">
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          Verify Transaction
        </CardTitle>
        <CardDescription>
          One-click verification. The oracle fetches on-chain data and validates
          both ECDSA and MLDSA signatures against your Public Key.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
            <div className="space-y-4">
              <FormField
                control={form.control}
                name="transactionHash"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel className="font-semibold text-sm">
                      Transaction Hash
                    </FormLabel>
                    <FormControl>
                      <div className="relative">
                        <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                        <Input
                          placeholder="0x..."
                          {...field}
                          className="pl-9 font-mono text-xs"
                        />
                      </div>
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="publicKey"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel className="font-semibold text-sm">
                      Signer Public Key
                    </FormLabel>
                    <FormControl>
                      <Textarea
                        className="font-mono text-[10px] max-h-50"
                        placeholder="Paste the Public Key here..."
                        {...field}
                      />
                    </FormControl>
                    <FormDescription>
                      The identity key needed to verify the digital signature.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
            <Button
              className="w-full cursor-pointer"
              type="submit"
              disabled={verifyTransaction.isPending}
            >
              {verifyTransaction.isPending ? <Spinner /> : "Verify Transaction"}
            </Button>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
};

export default VerificationForm;
