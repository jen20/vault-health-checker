package main

type vaultStatus uint8

const (
	vaultStatusActive vaultStatus = iota
	vaultStatusStandby
	vaultStatusUnhealthy
)
