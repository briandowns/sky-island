package mocks

import "github.com/stretchr/testify/mock"

type NetworkServicer struct {
	mock.Mock
}

// Allocate provides a mock function with given fields:
func (_m *NetworkServicer) Allocate() (string, error) {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Pool provides a mock function with given fields:
func (_m *NetworkServicer) Pool() map[string]byte {
	ret := _m.Called()

	var r0 map[string]byte
	if rf, ok := ret.Get(0).(func() map[string]byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]byte)
		}
	}

	return r0
}

// UpdateIPState provides a mock function with given fields: _a0, _a1
func (_m *NetworkServicer) UpdateIPState(_a0 string, _a1 byte) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, byte) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
