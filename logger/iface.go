package logger

type L interface {
	Info(s string)
	Error(s string)
	Debug(s string)
}
