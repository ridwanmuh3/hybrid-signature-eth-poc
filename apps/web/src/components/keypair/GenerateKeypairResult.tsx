import { useState } from "react";
import { Info, Key, Eye, EyeOff, Download } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "../ui/card";
import CopyClipboardButton from "../shared/CopyClipboardButton";
import { trimLongText } from "@/lib/utils";
import { Alert, AlertDescription, AlertTitle } from "../ui/alert";
import { Button } from "../ui/button";
import type { Keypair } from "@/types/keypair";

type PropsType = Partial<{
  keypair: Keypair;
}>;

const GenerateKeypairResult = ({ keypair }: PropsType) => {
  const [isPrivateKeyVisible, setIsPrivateKeyVisible] = useState(false);

  const handleDownload = () => {
    const data = JSON.stringify({ ...keypair }, null, 2);
    const blob = new Blob([data], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = "keypair.json";
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  return (
    <Card className="max-h-fit gap-3 bg-background border-input shadow-sm">
      <CardHeader className="pb-2">
        <div className="flex gap-3 items-center justify-between">
          <div className="flex gap-3 items-center">
            <div className="p-2.5 border border-input text-foreground bg-muted rounded-md">
              <Key className="size-5" />
            </div>
            <div>
              <CardTitle>Keypair Result</CardTitle>
            </div>
          </div>
          <Button
            variant="outline"
            size="sm"
            className="gap-2 text-sm h-8 cursor-pointer"
            onClick={handleDownload}
            disabled={!keypair?.private_key || !keypair?.public_key}
          >
            <Download className="size-4" />
            Save .json
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col gap-6">
          <div className="flex gap-1 justify-between flex-col">
            <div className="flex justify-between items-center">
              <p className="text-foreground font-semibold text-sm">Private Key</p>
              <button
                onClick={() => setIsPrivateKeyVisible(!isPrivateKeyVisible)}
                className="text-muted-foreground hover:text-muted-foreground transition-colors cursor-pointer"
                type="button"
              >
                {isPrivateKeyVisible ? (
                  <EyeOff className="size-4" />
                ) : (
                  <Eye className="size-4" />
                )}
              </button>
            </div>
            <div className="relative">
              <p
                onClick={() =>
                  !isPrivateKeyVisible && setIsPrivateKeyVisible(true)
                }
                className={`
                  text-muted-foreground text-sm break-all py-2.5 px-3 pr-12 
                  border-input border rounded-md shadow-sm
                  ${
                    !isPrivateKeyVisible
                      ? "blur-[3px] select-none cursor-pointer"
                      : ""
                  }
                `}
              >
                {isPrivateKeyVisible && keypair?.private_key
                  ? trimLongText(keypair?.private_key || "", 100)
                  : "0x••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••"}
              </p>
              <CopyClipboardButton
                key="private-key-clipboard"
                className="absolute top-1 right-1"
                content={keypair?.private_key || ""}
              />
            </div>
            <p className="text-xs text-red-500/80 font-medium mt-1">
              *Never share your private key with anyone.
            </p>
          </div>
          <div className="flex gap-1 justify-between flex-col">
            <p className="text-foreground font-semibold text-sm">Public Key</p>
            <div className="relative">
              <p className="text-muted-foreground text-sm break-all py-2.5 px-3 pr-12 border-input border rounded-md shadow-sm">
                {keypair?.public_key
                  ? trimLongText(keypair?.public_key!, 100)
                  : ""}
              </p>
              <CopyClipboardButton
                key="public-key-clipboard"
                className="absolute top-1 right-1"
                content={keypair?.public_key || ""}
              />
            </div>
          </div>
          <Alert
            className="bg-yellow-50 text-yellow-800 border-yellow-200 flex items-center"
            variant="default"
          >
            <Info className="h-4 w-4 text-yellow-600" />
            <div>
              <AlertTitle className="text-semibold">
                Please save your keypair securely
              </AlertTitle>
              <AlertDescription className="text-yellow-800">
                Store your keypair on secure file or storage!
              </AlertDescription>
            </div>
          </Alert>
        </div>
      </CardContent>
    </Card>
  );
};

export default GenerateKeypairResult;
