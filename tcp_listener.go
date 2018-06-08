package main

import (
	"fmt"
	"net"
	"time"

	log "github.com/hashicorp/go-hclog"
)

type tcpListener struct {
	addr         string
	standbyOK    bool
	statusChange <-chan vaultStatus
	logger       log.Logger

	listener net.Listener
	shutdown chan struct{}
}

func newTCPListener(addr string, standbyOK bool, logger log.Logger, statusChange <-chan vaultStatus) *tcpListener {
	return &tcpListener{
		addr:         addr,
		standbyOK:    standbyOK,
		logger:       logger,
		statusChange: statusChange,

		shutdown: make(chan struct{}, 1),
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

			go tl.runListener()
		} else {
			if tl.listener != nil {
				tl.shutdown <- struct{}{}
				tl.listener.Close()
				tl.logger.Info(fmt.Sprintf("Listener %s closed", tl.addr))
			}
		}
	}
}

func (tl *tcpListener) runListener() {
	listener, err := net.Listen("tcp", tl.addr)
	if err != nil {
		tl.logger.Error(fmt.Sprintf("TCP Listener Error: %s", err))
		return
	}
	tl.logger.Info(fmt.Sprintf("Listening on %s...", tl.addr))
	tl.listener = listener

	for {
		conn, err := listener.Accept()
		if err != nil {
			tl.logger.Error(fmt.Sprintf("Error accepting connection: %s", err))

			select {
			case <-tl.shutdown:
				break
			default:
				continue
			}
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	time.Sleep(500 * time.Millisecond)
}
