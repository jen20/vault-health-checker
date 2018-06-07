// +build mage

package main

import (
	"context"
	"path"

	"github.com/magefile/mage/sh"
)

const (
	rootPkg = "github.com/jen20/vault-health-checker"

	dirGoreleaser = "goreleaser"
)

func ReleaseVaultHealthChecker(ctx context.Context) error {
	config := path.Join(dirGoreleaser, "vault-health-checker.yml")

	return sh.RunV("goreleaser", "--rm-dist", "--config", config)
}
