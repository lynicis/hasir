package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadSops(t *testing.T) {
	// Set the age private key for sops decryption
	os.Setenv("SOPS_AGE_KEY", "AGE-SECRET-KEY-1EKP28WQ03K737TYUULLCT8QFXPSSZ6UC0U2S2ZFH50SJCE3DU8WS38T3WH")
	defer os.Unsetenv("SOPS_AGE_KEY")

	// Read the sops encrypted config file
	configPath := filepath.Join("testdata", "config.sops.json")
	reader := &JsonConfig{ConfigPath: configPath}
	cfg := reader.Read()

	assert.Equal(t, "https://sops-decrypted.example.com", cfg.DashboardUrl)
}
