package main

import (
	"fmt"
	"net"

	log "github.com/hashicorp/go-hclog"
)

type tcpListener struct {
	addr         string
	standbyOK    bool
	statusChange <-chan vaultStatus
	logger       log.Logger

	listener net.Listener
}

func newTCPListener(addr string, standbyOK bool, logger log.Logger, statusChange <-chan vaultStatus) *tcpListener {
	return &tcpListener{
		addr:         addr,
		standbyOK:    standbyOK,
		logger:       logger,
		statusChange: statusChange,
	}
}

func (tl *tcpListener) run() error {
	for {
		status := <-tl.statusChange

		shouldRun := false

		switch status {
		case vaultStatusActive:
			tl.logger.Info("Vault Status: Healthy (Active)")
			shouldRun = true
		case vaultStatusStandby:
			if tl.standbyOK {
				tl.logger.Info("Vault Status: Healthy (Standby)")
				shouldRun = true
			} else {
				tl.logger.Info("Vault Status: Unhealthy (Standby)")
				shouldRun = false
			}
		case vaultStatusUnhealthy:
			tl.logger.Info("Vault Status: Unhealthy")
			shouldRun = false
		}

		if shouldRun {
			if tl.listener != nil {
				continue
			}

			listener, err := net.Listen("tcp", tl.addr)
			if err != nil {
				tl.logger.Error(fmt.Sprintf("TCP Listener Error: %s", err))
			}
			tl.listener = listener
			tl.logger.Info(fmt.Sprintf("Listening on %s...", tl.addr))
		} else {
			if tl.listener != nil {
				tl.listener.Close()
				tl.logger.Info(fmt.Sprintf("Listener %s closed", tl.addr))
			}
		}
	}
}
