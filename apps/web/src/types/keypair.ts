export type Keypair = Partial<{
  status: string;
  algorithm: string;
  mode: string;
  private_key: `0x${string}`;
  public_key: `0x${string}`;
}>;
