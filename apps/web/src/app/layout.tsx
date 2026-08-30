import { Header } from "@/components";
import { Outlet } from "react-router";

const Layout = () => {
  return (
    <>
      <div className="flex flex-col min-h-screen bg-muted">
        <Header />
        <main className="w-full flex-1">
          <Outlet />
        </main>
      </div>
    </>
  );
};

export default Layout;
