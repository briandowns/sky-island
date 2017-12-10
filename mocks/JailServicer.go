package mocks

import "github.com/briandowns/sky-island/jail"
import "github.com/stretchr/testify/mock"

type JailServicer struct {
	mock.Mock
}

// InitializeSystem provides a mock function with given fields:
func (_m *JailServicer) InitializeSystem() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateJail provides a mock function with given fields: _a0, _a1
func (_m *JailServicer) CreateJail(_a0 string, _a1 bool) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, bool) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveJail provides a mock function with given fields: _a0
func (_m *JailServicer) RemoveJail(_a0 string) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// KillJail provides a mock function with given fields: _a0
func (_m *JailServicer) KillJail(_a0 int) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// JailDetails provides a mock function with given fields: _a0
func (_m *JailServicer) JailDetails(_a0 int) (*jail.JLS, error) {
	ret := _m.Called(_a0)

	var r0 *jail.JLS
	if rf, ok := ret.Get(0).(func(int) *jail.JLS); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*jail.JLS)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
