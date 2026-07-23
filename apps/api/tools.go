//go:build tools

package tools

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/securego/gosec/v2/cmd/gosec"
	_ "go.uber.org/mock/mockgen"
	_ "golang.org/x/vuln/cmd/govulncheck"
)
