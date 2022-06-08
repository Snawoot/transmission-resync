package main

import (
	"time"
)

type ChainEntry struct {
	Command []string
	Timeout time.Duration
}

type Chain []*ChainEntry
