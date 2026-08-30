// performance_test.go — Keygen, signing, verification, and benchmark runs.
//
// Run:
//
//	go test ./test/... -run TestFunctional_(KeyGen|Performance) -v
//	go test ./test/... -bench=. -benchtime=10s

package integration_test

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"text/tabwriter"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/signature"
)

const (
	fnWarmup     = 3
	fnIterations = 50
	gasNonZero   = uint64(16)
	gasZero      = uint64(4)
	gasFixed     = uint64(24_800)
)

type fnKeyGenResult struct {
	sc        fnScenario
	avg       time.Duration
	privBytes int
	pubBytes  int
}

type fnSignVerifyResult struct {
	sc         fnScenario
	signAvg    time.Duration
	verifyAvg  time.Duration
	sigBytes   int
	ecdsaBytes int
	pqBytes    int
	gasEst     uint64
	signCPU    time.Duration
	verifyCPU  time.Duration
	signMemB   uint64
	verifyMemB uint64
}

type fnFullResult struct {
	kg fnKeyGenResult
	sv fnSignVerifyResult
}

func calldataGas(ecdsaSig, pqSig []byte) uint64 {
	gas := gasFixed
	for _, b := range ecdsaSig {
		if b == 0 {
			gas += gasZero
		} else {
			gas += gasNonZero
		}
	}
	for _, b := range pqSig {
		if b == 0 {
			gas += gasZero
		} else {
			gas += gasNonZero
		}
	}
	return gas
}

func fnBenchKeyGen(t *testing.T, sc fnScenario) (fnKeyGenResult, fnKeySet) {
	var total time.Duration
	var privBytes, pubBytes int
	var lastKS fnKeySet

	for i := 0; i < fnWarmup+fnIterations; i++ {
		start := time.Now()
		ks := fnMakeKeySet(t, sc)
		elapsed := time.Since(start)

		if i >= fnWarmup {
			total += elapsed
			lastKS = ks
		}
		if i == fnWarmup {
			privBytes = len(ks.priv)
			pubBytes = len(ks.pub)
		}
	}

	return fnKeyGenResult{
		sc:        sc,
		avg:       total / time.Duration(fnIterations),
		privBytes: privBytes,
		pubBytes:  pubBytes,
	}, lastKS
}

func cpuNanos(ru syscall.Rusage) int64 {
	return (ru.Utime.Sec+ru.Stime.Sec)*1_000_000_000 + int64(ru.Utime.Usec+ru.Stime.Usec)*1000
}

func fnBenchSignVerify(t *testing.T, sc fnScenario, ks fnKeySet) fnSignVerifyResult {
	signer := signature.NewTransactionSigner(sc.algorithm, sc.mode)
	verifier := signature.NewSignatureVerifier(sc.algorithm, sc.mode)

	var totalSign, totalVerify time.Duration
	var sigBytes, ecdsaBytes, pqBytes int
	var gasEst uint64
	var capturedFinalSig []byte

	for i := 0; i < fnWarmup+fnIterations; i++ {
		start := time.Now()
		ecdsaSig, pqSig, finalSig, err := signer.Sign(fnTestHash, ks.priv)
		signElapsed := time.Since(start)
		require.NoError(t, err)

		start = time.Now()
		valid, note := verifier.Verify(fnTestHash, finalSig, ks.pub, ks.sender)
		verifyElapsed := time.Since(start)
		require.True(t, valid, "verify failed for %s (iteration %d): %s", sc.name, i, note)

		if i >= fnWarmup {
			totalSign += signElapsed
			totalVerify += verifyElapsed
		}
		if i == fnWarmup {
			sigBytes = len(finalSig)
			ecdsaBytes = len(ecdsaSig)
			pqBytes = len(pqSig)
			gasEst = calldataGas(ecdsaSig, pqSig)
			capturedFinalSig = finalSig
		}
	}

	var ms0, ms1 runtime.MemStats
	var ru0, ru1 syscall.Rusage

	runtime.GC()
	runtime.ReadMemStats(&ms0)
	syscall.Getrusage(syscall.RUSAGE_SELF, &ru0)
	for i := 0; i < fnIterations; i++ {
		if _, _, _, err := signer.Sign(fnTestHash, ks.priv); err != nil {
			t.Fatalf("sign CPU/mem pass: %v", err)
		}
	}
	syscall.Getrusage(syscall.RUSAGE_SELF, &ru1)
	runtime.ReadMemStats(&ms1)
	signMemB := (ms1.TotalAlloc - ms0.TotalAlloc) / uint64(fnIterations)
	signCPU := time.Duration((cpuNanos(ru1) - cpuNanos(ru0)) / int64(fnIterations))

	runtime.GC()
	runtime.ReadMemStats(&ms0)
	syscall.Getrusage(syscall.RUSAGE_SELF, &ru0)
	for i := 0; i < fnIterations; i++ {
		verifier.Verify(fnTestHash, capturedFinalSig, ks.pub, ks.sender)
	}
	syscall.Getrusage(syscall.RUSAGE_SELF, &ru1)
	runtime.ReadMemStats(&ms1)
	verifyMemB := (ms1.TotalAlloc - ms0.TotalAlloc) / uint64(fnIterations)
	verifyCPU := time.Duration((cpuNanos(ru1) - cpuNanos(ru0)) / int64(fnIterations))

	return fnSignVerifyResult{
		sc:         sc,
		signAvg:    totalSign / time.Duration(fnIterations),
		verifyAvg:  totalVerify / time.Duration(fnIterations),
		sigBytes:   sigBytes,
		ecdsaBytes: ecdsaBytes,
		pqBytes:    pqBytes,
		gasEst:     gasEst,
		signCPU:    signCPU,
		verifyCPU:  verifyCPU,
		signMemB:   signMemB,
		verifyMemB: verifyMemB,
	}
}

func TestFunctional_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping: run without -short flag")
	}

	results := make([]fnFullResult, 0, len(fnScenarios))
	for _, sc := range fnScenarios {
		t.Logf("[%s] key generation ...", sc.name)
		kg, ks := fnBenchKeyGen(t, sc)

		t.Logf("[%s] sign/verify latency ...", sc.name)
		sv := fnBenchSignVerify(t, sc, ks)

		results = append(results, fnFullResult{kg: kg, sv: sv})
	}

	printFunctionalReport(results)
}

func TestFunctional_KeyGen(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping: run without -short flag")
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight|tabwriter.Debug)
	fmt.Fprintln(writer, "\n[KEY GENERATION BENCHMARK]")
	fmt.Fprintf(writer, "Warmup: %d | Iterations: %d\n", fnWarmup, fnIterations)
	fmt.Fprintln(writer, "Algorithm\t PrivKey (B)\t PubKey (B)\t Total Key (B)\t KeyGen Avg (ms)")
	fmt.Fprintln(writer, strings.Repeat("-", 80))

	for _, sc := range fnScenarios {
		t.Logf("[%s] key generation ...", sc.name)
		kg, _ := fnBenchKeyGen(t, sc)
		fmt.Fprintf(writer, "%s\t %d\t %d\t %d\t %.3f\n",
			kg.sc.name,
			kg.privBytes, kg.pubBytes,
			kg.privBytes+kg.pubBytes,
			msf(kg.avg))
	}
	writer.Flush()
}

func printFunctionalReport(results []fnFullResult) {
	const gasPrice = 30e9
	const ethUSD = 3000.0

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight|tabwriter.Debug)
	separator := strings.Repeat("=", 110)

	fmt.Fprintln(writer, "\n"+separator)
	fmt.Fprintln(writer, "ORACLE FUNCTIONAL PERFORMANCE REPORT")
	fmt.Fprintf(writer, "Warmup: %d | Iterations: %d\n", fnWarmup, fnIterations)
	fmt.Fprintln(writer, separator)

	fmt.Fprintln(writer, "\n[1] KEY GENERATION")
	fmt.Fprintln(writer, "Algorithm\t PrivKey (B)\t PubKey (B)\t Total Key (B)\t KeyGen Avg (ms)")
	fmt.Fprintln(writer, strings.Repeat("-", 80))
	for _, result := range results {
		kg := result.kg
		fmt.Fprintf(writer, "%s\t %d\t %d\t %d\t %.3f\n",
			kg.sc.name,
			kg.privBytes, kg.pubBytes,
			kg.privBytes+kg.pubBytes,
			msf(kg.avg))
	}

	fmt.Fprintln(writer, "\n[2] SIGNING & VERIFICATION LATENCY / PAYLOAD SIZES")
	fmt.Fprintln(writer, "Algorithm\t ECDSA Sig (B)\t PQC Sig (B)\t Total Sig (B)\t Sign (ms)\t Verify (ms)\t Round-trip (ms)")
	fmt.Fprintln(writer, strings.Repeat("-", 100))
	for _, result := range results {
		sv := result.sv
		fmt.Fprintf(writer, "%s\t %d\t %d\t %d\t %.3f\t %.3f\t %.3f\n",
			sv.sc.name,
			sv.ecdsaBytes, sv.pqBytes, sv.sigBytes,
			msf(sv.signAvg), msf(sv.verifyAvg), msf(sv.signAvg+sv.verifyAvg))
	}

	fmt.Fprintln(writer, "\n[3] GAS ESTIMATION  (calldata only, 30 gwei, ETH=$3000)")
	fmt.Fprintln(writer, "Algorithm\t Sig Payload (B)\t Est. Gas\t ETH Cost\t USD Cost")
	fmt.Fprintln(writer, strings.Repeat("-", 80))
	for _, result := range results {
		sv := result.sv
		ethCost := float64(sv.gasEst) * gasPrice / 1e18
		fmt.Fprintf(writer, "%s\t %d\t %d\t %.6f ETH\t $%.4f\n",
			sv.sc.name, sv.sigBytes, sv.gasEst, ethCost, ethCost*ethUSD)
	}

	fmt.Fprintln(writer, "\n[4] CPU & MEMORY USAGE  (per operation, user+sys CPU time)")
	fmt.Fprintln(writer, "Algorithm\t Sign CPU (μs)\t Verify CPU (μs)\t Sign Alloc (KB)\t Verify Alloc (KB)")
	fmt.Fprintln(writer, strings.Repeat("-", 90))
	for _, result := range results {
		sv := result.sv
		fmt.Fprintf(writer, "%s\t %.1f\t %.1f\t %.2f\t %.2f\n",
			sv.sc.name,
			float64(sv.signCPU.Microseconds()),
			float64(sv.verifyCPU.Microseconds()),
			float64(sv.signMemB)/1024.0,
			float64(sv.verifyMemB)/1024.0)
	}

	writer.Flush()
}

func msf(duration time.Duration) float64 {
	return float64(duration.Microseconds()) / 1000.0
}

func BenchmarkKeyGen(b *testing.B) {
	for _, sc := range fnScenarios {
		sc := sc
		b.Run(sc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = fnMakeKeySet(b, sc)
			}
		})
	}
}

func BenchmarkSign(b *testing.B) {
	for _, sc := range fnScenarios {
		sc := sc
		b.Run(sc.name, func(b *testing.B) {
			ks := fnMakeKeySet(b, sc)
			signer := signature.NewTransactionSigner(sc.algorithm, sc.mode)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, _, _, err := signer.Sign(fnTestHash, ks.priv); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkVerify(b *testing.B) {
	for _, sc := range fnScenarios {
		sc := sc
		b.Run(sc.name, func(b *testing.B) {
			ks := fnMakeKeySet(b, sc)
			signer := signature.NewTransactionSigner(sc.algorithm, sc.mode)
			_, _, finalSig, err := signer.Sign(fnTestHash, ks.priv)
			if err != nil {
				b.Fatal(err)
			}

			verifier := signature.NewSignatureVerifier(sc.algorithm, sc.mode)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if valid, _ := verifier.Verify(fnTestHash, finalSig, ks.pub, ks.sender); !valid {
					b.Fatal("verify failed")
				}
			}
		})
	}
}
