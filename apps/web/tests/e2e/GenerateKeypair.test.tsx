import { expect, test, vi, beforeEach } from "vitest";
import { cleanup, render } from "vitest-browser-react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { userEvent } from "@vitest/browser/context";
import GenerateKeypairPage from "@/pages/GenerateKeypairPage";

// --- 1. Mocks Setup ---
const mockMutateAsync = vi.fn();

// Mocking Feature Hook (Tanstack Query)
vi.mock("@/features/keypair/useGenerateKeypair", () => ({
  useGenerateKeypair: () => ({
    mutateAsync: mockMutateAsync,
    isPending: false,
  }),
}));

// Mocking Download Blob/URL (Agar tidak error di environment test)
window.URL.createObjectURL = vi.fn(() => "blob:mock-url");
window.URL.revokeObjectURL = vi.fn();

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false } },
});

beforeEach(() => {
  cleanup();
  vi.clearAllMocks();
});

test("Alur Lengkap: Mengisi form, mendapatkan hasil, dan melihat Private Key", async () => {
  const { getByText, getByPlaceholder, getByRole } = await render(
    <QueryClientProvider client={queryClient}>
      <GenerateKeypairPage />
    </QueryClientProvider>,
  );

  const algoTrigger = getByRole("combobox", { name: /signature algorithm/i });
  await userEvent.click(algoTrigger);
  await userEvent.click(getByText("ML-DSA-65 (Post-Quantum)"));

  const modeTrigger = getByRole("combobox", { name: /operation mode/i });
  await userEvent.click(modeTrigger);
  await userEvent.click(getByText(/Hybrid Scheme/i));

  const privKeyInput = getByPlaceholder(/Paste your Private Key/i);
  const dummyEcdsa =
    "0xabc1234567890defabc1234567890defabc1234567890defabc1234567890def";
  await userEvent.fill(privKeyInput, dummyEcdsa);

  // --- STEP 2: Eksekusi Mutation ---

  // Mock data yang dikembalikan backend
  const mockPqResult = {
    data: {
      private_key: "0xPQ_SECRET_KEY_DUMMY_LONG_TEXT_FOR_TESTING",
      public_key: "0xPQ_PUBLIC_KEY_DUMMY_LONG_TEXT_FOR_TESTING",
      address: "0x123456789",
    },
  };
  mockMutateAsync.mockResolvedValue(mockPqResult);

  // Klik Generate
  const submitBtn = getByRole("button", { name: /Generate Keypair/i });
  await userEvent.click(submitBtn);

  // --- STEP 3: Verifikasi Hasil (GenerateKeypairResult) Muncul ---

  // Pastikan judul result muncul
  await expect.element(getByText("Keypair Result")).toBeInTheDocument();

  // Verifikasi Public Key muncul (tidak di-blur)
  await expect.element(getByText(/0xPQ_PUBLIC_KEY/i)).toBeInTheDocument();

  // --- STEP 4: Interaksi di Card Result ---

  // Verifikasi Private Key awalnya tersembunyi (ter-blur)
  const privateKeyField = getByText(/0x••••••••/);
  expect(privateKeyField).toHaveClass("blur-[3px]");

  // Klik icon Mata (Eye) untuk reveal Private Key
  // Kita cari button berdasarkan icon atau tipe button di dalam section Private Key
  const toggleVisibilityBtn = getByRole("button", { name: "" }); // icon button biasanya tidak punya text
  await userEvent.click(toggleVisibilityBtn);

  // Verifikasi Private Key sekarang terlihat dan blur hilang
  await expect.element(getByText(/0xPQ_SECRET_KEY/i)).toBeInTheDocument();
  expect(getByText(/0xPQ_SECRET_KEY/i)).not.toHaveClass("blur-[3px]");

  // --- STEP 5: Verifikasi Fitur Download ---

  const downloadBtn = getByRole("button", { name: /Save .json/i });
  expect(downloadBtn).toBeEnabled();
  await userEvent.click(downloadBtn);

  // Verifikasi apakah fungsi download terpanggil (lewat mock URL)
  expect(window.URL.createObjectURL).toHaveBeenCalled();
});
