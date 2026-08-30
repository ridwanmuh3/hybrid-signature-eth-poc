import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "@/assets/globals.css";
import Providers from "./providers.tsx";
import { Route, Routes } from "react-router";
import Layout from "./layout.tsx";
import {
  GenerateKeypairPage,
  SendTransactionPage,
  VerifyTransactionPage,
} from "@/pages";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <Providers>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<SendTransactionPage />} />
          <Route path="/verify" element={<VerifyTransactionPage />} />
          <Route path="/generate-keypair" element={<GenerateKeypairPage />} />
        </Route>
      </Routes>
    </Providers>
  </StrictMode>
);
