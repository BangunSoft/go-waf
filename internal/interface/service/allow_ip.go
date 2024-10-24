package service

type AllowIPInterface interface {
	Check(string) bool
}
