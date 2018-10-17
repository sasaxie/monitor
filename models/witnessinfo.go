package models

import "sync"

type WitnessInfo struct {
	Info map[string]int64
	Lock *sync.Mutex
}
