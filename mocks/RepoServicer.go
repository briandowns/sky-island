package mocks

import "github.com/stretchr/testify/mock"

type RepoServicer struct {
	mock.Mock
}

// CloneRepo provides a mock function with given fields: jpath, fname
func (_m *RepoServicer) CloneRepo(jpath string, fname string) error {
	ret := _m.Called(jpath, fname)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(jpath, fname)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveRepo provides a mock function with given fields: repo
func (_m *RepoServicer) RemoveRepo(repo string) error {
	ret := _m.Called(repo)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(repo)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
