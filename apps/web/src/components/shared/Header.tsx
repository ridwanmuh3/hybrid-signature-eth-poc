import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import WalletOptions from "../web3/WalletOptions";
import { useConnection } from "wagmi";
import Connection from "../web3/Connection";
import { NavLink, type NavLinkRenderProps } from "react-router";

const Header = () => {
  const { isConnected } = useConnection();

  const navLinkClassHandler = ({ isActive }: NavLinkRenderProps) => {
    const baseStyle = "hover:text-muted-foreground hover:underline font-bold ";
    return isActive
      ? `${baseStyle} text-zinc-800`
      : `${baseStyle} text-muted-foreground`;
  };

  return (
    <>
      <header className="border-b bg-background">
        <nav className="w-full max-w-5xl mx-auto px-8 py-5 flex justify-between items-center">
          <div className="flex gap-4">
            <NavLink className={navLinkClassHandler} to="/generate-keypair">
              Generate Keypair
            </NavLink>
            <NavLink className={navLinkClassHandler} to="/">
              Send Transaction
            </NavLink>
            <NavLink className={navLinkClassHandler} to="/verify">
              Verify Transaction
            </NavLink>
          </div>
          <div>
            <DropdownMenu>
              {isConnected ? (
                <Connection />
              ) : (
                <DropdownMenuTrigger className="px-4 py-2 bg-zinc-900 rounded-md text-zinc-50 text-sm font-semibold cursor-pointer">
                  Connect
                </DropdownMenuTrigger>
              )}

              <DropdownMenuContent>
                <DropdownMenuLabel>Connect Wallet</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <WalletOptions />
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </nav>
      </header>
    </>
  );
};

export default Header;
