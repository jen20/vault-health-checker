package main

import (
	"fmt"
	"os"
	"time"

	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-sockaddr/template"
	"github.com/sean-/sysexits"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	logLevel := os.Getenv("VAULT_HEALTH_CHECK_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "INFO"
	}
	logger := log.New(&log.LoggerOptions{
		Level:  log.LevelFromString(logLevel),
		Output: os.Stderr,
	})

	serverAddrRaw := os.Getenv("VAULT_HEALTH_CHECK_SERVER_ADDR")
	tcpAddrRaw := os.Getenv("VAULT_HEALTH_CHECK_TCP_ADDR")
	intervalRaw := os.Getenv("VAULT_HEALTH_CHECK_INTERVAL")
	standbyUnhealthy := os.Getenv("VAULT_HEALTH_CHECK_STANDBY_UNHEALTHY")

	interval := 1 * time.Second
	if intervalRaw != "" {
		dur, err := time.ParseDuration(intervalRaw)
		if err != nil {
			logger.Error("Error parsing interval %q: %s", intervalRaw, err)
			os.Exit(sysexits.Usage)
		}
		interval = dur
	}

	if serverAddrRaw == "" {
		serverAddrRaw = "https://{{ GetPrivateIP }}:8200"
	}
	serverAddr, err := template.Parse(serverAddrRaw)
	if err != nil {
		logger.Error("Error parsing Vault server address template %q: %s", serverAddrRaw, err)
		os.Exit(sysexits.Usage)
	}

	if tcpAddrRaw == "" {
		tcpAddrRaw = "{{ GetPrivateIP }}:8210"
	}
	tcpAddr, err := template.Parse(tcpAddrRaw)
	if err != nil {
		logger.Error("Error parsing TCP address template %q: %s", tcpAddrRaw, err)
		os.Exit(sysexits.Usage)
	}
	verifyTLSEnv := os.Getenv("VAULT_HEALTH_CHECK_SKIP_VERIFY")
	verifyTLS := true
	if verifyTLSEnv == "false" {
		verifyTLS = false
	}

	standbyOK := true
	if standbyUnhealthy != "" {
		standbyOK = false
	}

	// Buffered channel allows health checks to proceed even if the processing
	// takes a while (which is not expected).
	statusChannel := make(chan vaultStatus, 10)

	fmt.Fprintf(os.Stderr, "==> Vault NLB Health Checker Configuration:\n\n")
	fmt.Fprintf(os.Stderr, "                           Version: %s (%s %s)\n", version, commit, date)
	fmt.Fprintf(os.Stderr, "              Vault Server Address: %s\n", serverAddr)
	fmt.Fprintf(os.Stderr, "          TCP Health Check Address: %s\n\n", tcpAddr)
	fmt.Fprintf(os.Stderr, "             Health Check Interval: %s\n", interval.String())
	fmt.Fprintf(os.Stderr, "   Treat Standby Nodes as Healthy?: %t\n", standbyOK)
	fmt.Fprintf(os.Stderr, "\n==> NLB Health Checker Started! Log data will stream in below:\n\n")

	healthChecker, err := newVaultHealthChecker(serverAddr, interval, logger, statusChannel, verifyTLS)
	if err != nil {
		logger.Error("Error constructing checker: %s", err)
		os.Exit(1)
	}
	go healthChecker.run()

	tcpListener := newTCPListener(tcpAddr, standbyOK, logger, statusChannel)
	tcpListener.run()
}
