import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import axios from "axios";
import { formatEther } from "viem";
import { QueryClient } from "@tanstack/react-query";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export const axiosInstance = axios.create({
  baseURL: import.meta.env.VITE_ORACLE_URL || "http://localhost:9000",
});

export const queryClient = new QueryClient();

export const trimLongText = (text: string, length: number) =>
  text.slice(0, length) + "...";

export const formatReadableNumber = (num: number) =>
  num.toLocaleString("id").slice(0, -4);

export const formatAmount = (
  wei: string | undefined | number,
  unit: "wei" | "gwei" | undefined
) => {
  try {
    return wei ? formatEther(BigInt(wei), unit) : "0";
  } catch (e) {
    return "0";
  }
};
