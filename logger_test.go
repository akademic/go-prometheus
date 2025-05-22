package prometheus

// MockLogger implements Logger interface for testing
type MockLogger struct{}

func (m *MockLogger) Info(format string, args ...any)  {}
func (m *MockLogger) Error(format string, args ...any) {}
