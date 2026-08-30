import { Copy } from "lucide-react";
import { Button } from "../ui/button";

type PropsType = {
  className: string | undefined;
  content: string;
};

const CopyClipboardButton = ({ className, content }: PropsType) => {
  return (
    <>
      <Button
        className={`cursor-pointer ${className}`}
        variant="ghost"
        onClick={() => navigator.clipboard.writeText(content)}
      >
        <Copy className="size-4" />
      </Button>
    </>
  );
};

export default CopyClipboardButton;
