package service

import "time"

type CacheInterface interface {
	Set(string, interface{}, time.Duration)
	Get(string) (interface{}, bool)
	Pop(string) (interface{}, bool)
	Remove(string)
	GetTTL(string) (time.Duration, bool)
}
