package utils

import (
	"os/exec"
)

// Wrapper
type Wrapper interface {
	Output(name string, args ...string) ([]byte, error)
	CombinedOutput(name string, args ...string) ([]byte, error)
}

// Wrap
type Wrap struct{}

// Output
func (Wrap) Output(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

// CombinedOutput
func (Wrap) CombinedOutput(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

// NoOpWrapper
type NoOpWrapper struct{}

// Output
func (NoOpWrapper) Output(name string, args ...string) ([]byte, error) {
	return nil, nil
}

// CombinedOutput
func (NoOpWrapper) CombinedOutput(name string, args ...string) ([]byte, error) {
	return nil, nil
}
