import { describe, it, expect, beforeEach } from "vitest"
import { keypairFormSchema } from "@/schemas/keypairFormSchema"
import { sendAndSignTxFormSchema } from "@/schemas/sendAndSignTxFormSchema"
import { verificationFormSchema } from "@/schemas/verificationFormSchema"

describe("keypairFormSchema", () => {
  beforeEach(() => {
    // Reset any schema state if needed
  })

  it("should validate valid keypair data", () => {
    const validData = {
      publicKey: "0x" + "a".repeat(64),
      mlDsaPublicKey: "0x" + "b".repeat(48),
    }
    const result = keypairFormSchema.safeParse(validData)
    expect(result.success).toBe(true)
  })

  it("should reject missing publicKey", () => {
    const invalidData = {
      publicKey: "",
      mlDsaPublicKey: "0x" + "b".repeat(48),
    }
    const result = keypairFormSchema.safeParse(invalidData)
    expect(result.success).toBe(false)
    expect(result.error?.issues.length).toBeGreaterThan(0)
  })

  it("should reject missing mlDsaPublicKey", () => {
    const invalidData = {
      publicKey: "0x" + "a".repeat(64),
      mlDsaPublicKey: "",
    }
    const result = keypairFormSchema.safeParse(invalidData)
    expect(result.success).toBe(false)
    expect(result.error?.issues.length).toBeGreaterThan(0)
  })
})

describe("sendAndSignTxFormSchema", () => {
  beforeEach(() => {
    // Reset any schema state if needed
  })

  it("should validate valid transaction data", () => {
    const validData = {
      sender: "0x" + "a".repeat(40),
      receiver: "0x" + "b".repeat(40),
      message: "test message",
      amount: "100",
      algorithm: "ECDSA",
      mode: "single" as const,
    }
    const result = sendAndSignTxFormSchema.safeParse(validData)
    expect(result.success).toBe(true)
  })

  it("should reject empty sender", () => {
    const invalidData = {
      sender: "",
      receiver: "0x" + "b".repeat(40),
      message: "test",
      amount: "100",
      algorithm: "ECDSA",
      mode: "single" as const,
    }
    const result = sendAndSignTxFormSchema.safeParse(invalidData)
    expect(result.success).toBe(false)
  })

  it("should reject empty receiver", () => {
    const invalidData = {
      sender: "0x" + "a".repeat(40),
      receiver: "",
      message: "test",
      amount: "100",
      algorithm: "ECDSA",
      mode: "single" as const,
    }
    const result = sendAndSignTxFormSchema.safeParse(invalidData)
    expect(result.success).toBe(false)
  })

  it("should reject zero amount", () => {
    const invalidData = {
      sender: "0x" + "a".repeat(40),
      receiver: "0x" + "b".repeat(40),
      message: "test",
      amount: "0",
      algorithm: "ECDSA",
      mode: "single" as const,
    }
    const result = sendAndSignTxFormSchema.safeParse(invalidData)
    expect(result.success).toBe(false)
  })
})

describe("verificationFormSchema", () => {
  beforeEach(() => {
    // Reset any schema state if needed
  })

  it("should validate valid verification data", () => {
    const validData = {
      signature: "0x" + "a".repeat(130),
      message: "test message",
    }
    const result = verificationFormSchema.safeParse(validData)
    expect(result.success).toBe(true)
  })

  it("should reject empty signature", () => {
    const invalidData = {
      signature: "",
      message: "test message",
    }
    const result = verificationFormSchema.safeParse(invalidData)
    expect(result.success).toBe(false)
  })

  it("should reject empty message", () => {
    const invalidData = {
      signature: "0x" + "a".repeat(130),
      message: "",
    }
    const result = verificationFormSchema.safeParse(invalidData)
    expect(result.success).toBe(false)
  })
})