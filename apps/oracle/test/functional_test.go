// functional_test.go — Sign/verify correctness across all supported schemes.
//
// Run:
//
//	go test ./test/... -run TestFunctional -v

package integration_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ridwanmuh3/hybrid-pqc-poc/apps/oracle/internal/signature"
)

func TestFunctional_SignVerify_Correctness(t *testing.T) {
	for _, sc := range fnScenarios {
		sc := sc
		t.Run(sc.name, func(t *testing.T) {
			t.Parallel()

			ks := fnMakeKeySet(t, sc)
			signer := signature.NewTransactionSigner(sc.algorithm, sc.mode)

			ecdsaSig, pqSig, finalSig, err := signer.Sign(fnTestHash, ks.priv)
			require.NoError(t, err, "Sign must not error")
			require.NotEmpty(t, finalSig, "finalSig must be non-empty")

			if sc.mode == "hybrid" {
				assert.Equal(t, 65, len(ecdsaSig), "ECDSA sig must be 65 bytes")
				assert.NotEmpty(t, pqSig, "PQC sig must be present")
				assert.Greater(t, len(finalSig), 69, "hybrid finalSig must exceed 69 bytes")
			} else if sc.algorithm == "ECDSA" {
				assert.Equal(t, 65, len(finalSig), "ECDSA single sig must be 65 bytes")
			}

			verifier := signature.NewSignatureVerifier(sc.algorithm, sc.mode)
			valid, note := verifier.Verify(fnTestHash, finalSig, ks.pub, ks.sender)
			assert.True(t, valid, "round-trip must succeed: %s", note)

			badHash := common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111")
			valid2, _ := verifier.Verify(badHash, finalSig, ks.pub, ks.sender)
			assert.False(t, valid2, "sig for correct hash must not verify a different hash")
		})
	}
}
