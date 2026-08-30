import {
  GenerateKeypairForm,
  GenerateKeypairResult,
  WalletInfo,
} from "@/components";
import { useKeypairStore } from "@/stores/keypair";

const GenerateKeypairPage = () => {
  const { keypair, setKeypair } = useKeypairStore();
  return (
    <>
      <div className="max-w-5xl mx-auto p-8 space-y-8">
        <h1 className="text-2xl font-semibold text-foreground">Generate Keypair</h1>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
          <WalletInfo />
          <GenerateKeypairForm keypair={keypair} setKeypair={setKeypair} />
        </div>
        {keypair.private_key !== "0x" && keypair.public_key !== "0x" ? (
          <GenerateKeypairResult keypair={keypair} />
        ) : null}
      </div>
    </>
  );
};

export default GenerateKeypairPage;
