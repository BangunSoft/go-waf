package service

import "time"

type CacheInterface interface {
	Set(string, []byte, time.Duration)
	Get(string) ([]byte, bool)
	Pop(string) ([]byte, bool)
	Remove(string)
	RemoveByPrefix(string)
	GetTTL(string) (time.Duration, bool)
}
