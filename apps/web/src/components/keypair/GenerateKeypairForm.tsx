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
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { pqcAlgorithms } from "@/constants";
import { useGenerateKeypair } from "@/features/keypair/useGenerateKeypair";
import { type KeypairState } from "@/stores/keypair";
import { Spinner } from "@/components/ui/spinner";
import { generateKeypairFormSchema } from "@/schemas/keypairFormSchema";

const GenerateKeypairForm = (keypair: KeypairState) => {
  const form = useForm<z.infer<typeof generateKeypairFormSchema>>({
    resolver: zodResolver(generateKeypairFormSchema),
    defaultValues: {
      algorithm: "ML-DSA-65",
      mode: "hybrid",
      ecdsaPrivateKey: "",
    },
  });

  const generateKeypair = useGenerateKeypair();
  const selectedAlgorithm = form.watch("algorithm");
  const selectedMode = form.watch("mode");

  const [formError, setFormError] = useState("");
  const [successMessage, setSuccessMessage] = useState("");

  const onSubmit = async (
    values: z.infer<typeof generateKeypairFormSchema>,
  ) => {
    setFormError("");
    setSuccessMessage("");

    try {
      const payload = {
        algorithm: values.algorithm,
        mode: values.mode,
        ecdsa_private_key:
          values.mode === "hybrid" && values.ecdsaPrivateKey
            ? values.ecdsaPrivateKey
            : undefined,
      };

      const { data } = await generateKeypair.mutateAsync(payload);

      keypair.setKeypair({ ...data });
      setSuccessMessage("Keypair generated successfully");
    } catch (err) {
      console.error("Failed to generate keypair:", err);
      setFormError("Keypair generation failed. Check your inputs and try again.");
    }
  };

  return (
    <>
      <Card className="max-h-fit bg-background border-input shadow-sm animate-in fade-in slide-in-from-top-4 duration-500">
        <CardHeader>
          <CardTitle className="text-base">Generate Keypair</CardTitle>
          <CardDescription className="text-sm text-muted-foreground">
            Create a new keypair identity based on chosen scheme.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {(formError || successMessage) && (
            <div className="mb-4">
              {formError && (
                <p className="text-destructive text-sm" role="alert">{formError}</p>
              )}
              {successMessage && (
                <p className="text-green-700 text-sm" role="status">{successMessage}</p>
              )}
            </div>
          )}
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
              <FormField
                control={form.control}
                name="algorithm"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel className="font-semibold text-sm">
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
                      <FormControl>
                        <SelectTrigger className="w-full">
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
                    <FormLabel className="font-semibold text-sm">
                      Operation Mode
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
                        <SelectItem value="single">
                          Single Scheme
                          {selectedAlgorithm === "ECDSA"
                            ? " (Standard)"
                            : " (Pure Post-Quantum)"}
                        </SelectItem>
                        <SelectItem
                          // Hybrid tidak boleh dipilih jika algoritmanya ECDSA
                          disabled={selectedAlgorithm === "ECDSA"}
                          value="hybrid"
                        >
                          Hybrid Scheme (Secp256k1 + PQC)
                        </SelectItem>
                      </SelectContent>
                    </Select>
                  </FormItem>
                )}
              />
              {selectedMode === "hybrid" && (
                <FormField
                  control={form.control}
                  name="ecdsaPrivateKey"
                  render={({ field }) => (
                    <FormItem className="p-4 border border-dashed border-input rounded-md bg-muted">
                      <FormLabel className="font-semibold">
                        Link Existing Wallet (Mandatory)
                      </FormLabel>
                      <FormControl>
                        <input
                          className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                          placeholder="0x... (Paste your Private Key)"
                          {...field}
                          type="password"
                        />
                      </FormControl>
                      <FormMessage />
                      <p className="text-xs text-muted-foreground">
                        Provide your wallet private key.
                      </p>
                    </FormItem>
                  )}
                />
              )}
              <div className="flex justify-end">
                <Button
                  className="w-full cursor-pointer"
                  type="submit"
                  disabled={
                    generateKeypair.isPending ||
                    (form.watch("mode") === "single" &&
                      form.watch("algorithm") === "ECDSA") ||
                    (form.watch("mode") === "hybrid" &&
                      form.watch("ecdsaPrivateKey") === "")
                  }
                >
                  {generateKeypair.isPending ? (
                    <Spinner />
                  ) : form.watch("mode") === "single" &&
                    form.watch("algorithm") === "ECDSA" ? (
                    "Use your existing wallet private key"
                  ) : (
                    "Generate Keypair"
                  )}
                </Button>
              </div>
            </form>
          </Form>
        </CardContent>
      </Card>
    </>
  );
};

export default GenerateKeypairForm;
