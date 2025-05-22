package prometheus

type Logger interface {
	Info(pattern string, args ...any)
	Error(pattern string, args ...any)
}
