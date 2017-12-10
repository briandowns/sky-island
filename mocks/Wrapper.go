package mocks

import "github.com/stretchr/testify/mock"

type Wrapper struct {
	mock.Mock
}

// Output provides a mock function with given fields: name, args
func (_m *Wrapper) Output(name string, args ...string) ([]byte, error) {
	ret := _m.Called(name, args)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(string, ...string) []byte); ok {
		r0 = rf(name, args...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, ...string) error); ok {
		r1 = rf(name, args...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CombinedOutput provides a mock function with given fields: name, args
func (_m *Wrapper) CombinedOutput(name string, args ...string) ([]byte, error) {
	ret := _m.Called(name, args)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(string, ...string) []byte); ok {
		r0 = rf(name, args...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, ...string) error); ok {
		r1 = rf(name, args...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
