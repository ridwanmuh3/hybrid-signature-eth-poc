package integration_test

import (
	"testing"

	"github.com/spf13/viper"
)

func newViperConfig(t *testing.T) *viper.Viper {
	t.Helper()
	v := viper.New()
	v.Set("CHAIN_ID", int64(31337))
	return v
}
