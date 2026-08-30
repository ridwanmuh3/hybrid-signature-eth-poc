// onchain_test.go — Real on-chain measurements for the article.
//
// This is the only test that requires a live EVM RPC (Anvil). Set the
// ONCHAIN_RPC environment variable (default port from `make run-node` is
// http://localhost:8540) and the test will:
//
//   1. Deploy a fresh Transactions contract with a dedicated test oracle.
//   2. For each algorithm scenario, deposit funds for a fresh sender, then
//      submit a sendTransaction with a real ECDSA signature (hybrid mode)
//      or empty ECDSA signature + real PQC signature (single PQC mode).
//   3. Capture wall-clock submission→inclusion latency and the actual
//      receipt.GasUsed (not the calldata estimate from functional_test.go).
//
// Run:
//
//   make run-node                                                   # terminal 1
//   cd apps/oracle && ONCHAIN_RPC=http://localhost:8540 \
//     go test ./test/... -run TestOnchain -v -timeout 600s          # terminal 2
//
// Without ONCHAIN_RPC the test is skipped — keeping the unit-test loop fast
// for users without a local node.

package integration_test

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"
	"text/tabwriter"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"

	transactions "github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/artifacts"
	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/signature"
	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/util"
)

// Anvil's first default account — funded with 10 000 ETH at boot. Reused
// here as the test oracle key. Anyone running `anvil` locally has the same
// keypair, so committing it is harmless.
const anvilOraclePrivKeyHex = "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

type onchainResult struct {
	scenario   fnScenario
	gasUsed    uint64
	sigBytes   int
	mineMillis float64
}

func TestOnchain_GasAndLatency(t *testing.T) {
	rpc := os.Getenv("ONCHAIN_RPC")
	if rpc == "" {
		t.Skip("set ONCHAIN_RPC=http://localhost:8540 (or your anvil RPC) to run on-chain measurements")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	client, err := ethclient.Dial(rpc)
	require.NoError(t, err, "dial %s", rpc)
	defer client.Close()

	chainID, err := client.ChainID(ctx)
	require.NoError(t, err, "fetch chain id")
	t.Logf("connected to chain id %s via %s", chainID, rpc)

	oracleSk, err := crypto.HexToECDSA(strings.TrimPrefix(anvilOraclePrivKeyHex, "0x"))
	require.NoError(t, err)
	oracleAddr := crypto.PubkeyToAddress(oracleSk.PublicKey)

	auth, err := bind.NewKeyedTransactorWithChainID(oracleSk, chainID)
	require.NoError(t, err)
	auth.Context = ctx

	t.Logf("deploying Transactions contract (oracle=%s) ...", oracleAddr.Hex())
	contractAddr, deployTx, contract, err := transactions.DeployTransactions(auth, client, oracleAddr)
	require.NoError(t, err, "deploy contract")
	rcpt, err := bind.WaitMined(ctx, client, deployTx)
	require.NoError(t, err, "wait deploy")
	require.Equal(t, uint64(1), rcpt.Status, "deploy reverted")
	t.Logf("contract deployed at %s", contractAddr.Hex())

	results := make([]onchainResult, 0, len(fnScenarios))
	for _, sc := range fnScenarios {
		results = append(results, runOnchainScenario(ctx, t, client, contract, oracleSk, chainID, sc))
	}

	printOnchainReport(t, results)
}

func runOnchainScenario(
	ctx context.Context,
	t *testing.T,
	client *ethclient.Client,
	contract *transactions.Transactions,
	oracleSk *ecdsa.PrivateKey,
	chainID *big.Int,
	sc fnScenario,
) onchainResult {
	t.Helper()

	// 1. Build a fresh sender keypair so each scenario starts from nonce=0.
	senderSk, err := crypto.GenerateKey()
	require.NoError(t, err, "[%s] sender keygen", sc.name)
	senderAddr := crypto.PubkeyToAddress(senderSk.PublicKey)

	// 2. Fund sender from oracle's balance and deposit so the contract has
	//    funds to debit on sendTransaction.
	depositAmount := big.NewInt(0).Mul(big.NewInt(2), big.NewInt(1e18)) // 2 ETH
	fundSenderAndDeposit(ctx, t, client, contract, oracleSk, chainID, senderAddr, senderSk, depositAmount)

	// 3. Build the algorithm-specific keypair so the on-chain ECDSA recovery
	//    (when needsEcdsa) lines up with the funded sender address.
	var ks fnKeySet
	switch {
	case sc.mode == "hybrid":
		ks = remintHybridForSender(t, sc, senderSk)
	case sc.algorithm == "ECDSA":
		ks = fnKeySet{
			priv:   crypto.FromECDSA(senderSk),
			pub:    crypto.FromECDSAPub(&senderSk.PublicKey),
			sender: senderAddr,
		}
	default:
		ks = fnMakeKeySet(t, sc) // pure-PQC: contract does no ECDSA check
		ks.sender = senderAddr
	}

	amount := big.NewInt(1e15) // 0.001 ETH
	receiver := common.HexToAddress("0x000000000000000000000000000000000000dEaD")

	hashData, err := util.HashEthSignedData(
		chainID,
		big.NewInt(1), // userNonce starts at 0; first valid nonce is 1
		senderAddr, receiver,
		sc.algorithm, sc.mode,
		"onchain-bench", amount,
	)
	require.NoError(t, err, "[%s] hash", sc.name)

	signer := signature.NewTransactionSigner(sc.algorithm, sc.mode)
	ecdsaSig, pqSig, finalSig, err := signer.Sign(hashData, ks.priv)
	require.NoError(t, err, "[%s] sign", sc.name)

	// 4. Submit and time inclusion.
	auth, err := bind.NewKeyedTransactorWithChainID(oracleSk, chainID)
	require.NoError(t, err)
	auth.Context = ctx

	pqSigToSend := pqSig
	if sc.mode != "hybrid" && sc.algorithm == "ECDSA" {
		pqSigToSend = []byte{}
	}

	startSubmit := time.Now()
	tx, err := contract.SendTransaction(auth, transactions.TransactionsTxParams{
		Nonce:            big.NewInt(1),
		Sender:           senderAddr,
		Receiver:         receiver,
		SigningAlgorithm: sc.algorithm,
		SigningMode:      sc.mode,
		Message:          "onchain-bench",
		Amount:           amount,
		EcdsaSignature:   ecdsaSig,
		PqSignature:      pqSigToSend,
	})
	require.NoError(t, err, "[%s] submit", sc.name)

	receipt, err := bind.WaitMined(ctx, client, tx)
	mineMillis := float64(time.Since(startSubmit).Microseconds()) / 1000.0
	require.NoError(t, err, "[%s] wait mined", sc.name)
	require.Equal(t, uint64(1), receipt.Status, "[%s] tx reverted: hash=%s", sc.name, tx.Hash().Hex())

	t.Logf("[%s] gas=%d mined_in=%.1fms sig=%dB tx=%s",
		sc.name, receipt.GasUsed, mineMillis, len(finalSig), tx.Hash().Hex())

	return onchainResult{
		scenario:   sc,
		gasUsed:    receipt.GasUsed,
		sigBytes:   len(finalSig),
		mineMillis: mineMillis,
	}
}

// fundSenderAndDeposit transfers ETH from oracle EOA to sender EOA, then has
// the sender call deposit() so contract.balances[sender] is large enough to
// cover the test transfer.
func fundSenderAndDeposit(
	ctx context.Context,
	t *testing.T,
	client *ethclient.Client,
	contract *transactions.Transactions,
	oracleSk *ecdsa.PrivateKey,
	chainID *big.Int,
	senderAddr common.Address,
	senderSk *ecdsa.PrivateKey,
	depositAmount *big.Int,
) {
	t.Helper()

	// Send a bit extra so the sender has gas headroom for the deposit call.
	fundAmount := new(big.Int).Add(depositAmount, big.NewInt(1e17)) // +0.1 ETH headroom
	sendETH(ctx, t, client, oracleSk, chainID, senderAddr, fundAmount)

	senderAuth, err := bind.NewKeyedTransactorWithChainID(senderSk, chainID)
	require.NoError(t, err)
	senderAuth.Context = ctx
	senderAuth.Value = depositAmount

	tx, err := contract.Deposit(senderAuth)
	require.NoError(t, err, "deposit")
	rcpt, err := bind.WaitMined(ctx, client, tx)
	require.NoError(t, err)
	require.Equal(t, uint64(1), rcpt.Status, "deposit reverted")
}

// sendETH transfers `value` from a signer's EOA to `to` via a plain legacy tx.
func sendETH(
	ctx context.Context,
	t *testing.T,
	client *ethclient.Client,
	fromSk *ecdsa.PrivateKey,
	chainID *big.Int,
	to common.Address,
	value *big.Int,
) {
	t.Helper()

	fromAddr := crypto.PubkeyToAddress(fromSk.PublicKey)
	nonce, err := client.PendingNonceAt(ctx, fromAddr)
	require.NoError(t, err, "pending nonce")

	gasPrice, err := client.SuggestGasPrice(ctx)
	require.NoError(t, err, "suggest gas price")

	rawTx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &to,
		Value:    value,
		Gas:      21000,
		GasPrice: gasPrice,
	})
	signedTx, err := types.SignTx(rawTx, types.NewEIP155Signer(chainID), fromSk)
	require.NoError(t, err, "sign legacy tx")

	require.NoError(t, client.SendTransaction(ctx, signedTx), "send eth")
	rcpt, err := bind.WaitMined(ctx, client, signedTx)
	require.NoError(t, err, "wait fund")
	require.Equal(t, uint64(1), rcpt.Status, "fund tx reverted")
}

func printOnchainReport(t *testing.T, results []onchainResult) {
	const gasPrice = 30e9 // 30 gwei
	const ethUSD = 3000.0

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight|tabwriter.Debug)
	fmt.Fprintln(w, "\n=========================================================================")
	fmt.Fprintln(w, "ON-CHAIN GAS + INCLUSION LATENCY  (real receipts, 30 gwei, ETH=$3000)")
	fmt.Fprintln(w, "=========================================================================")
	fmt.Fprintln(w, "Scenario\t Sig (B)\t Real Gas\t ETH Cost\t USD Cost\t Mined In (ms)")
	fmt.Fprintln(w, strings.Repeat("-", 90))
	for _, r := range results {
		ethCost := float64(r.gasUsed) * gasPrice / 1e18
		fmt.Fprintf(w, "%s\t %d\t %d\t %.6f\t $%.4f\t %.1f\n",
			r.scenario.name, r.sigBytes, r.gasUsed, ethCost, ethCost*ethUSD, r.mineMillis)
	}
	w.Flush()
	t.Logf("on-chain report printed (%d scenarios)", len(results))
}
