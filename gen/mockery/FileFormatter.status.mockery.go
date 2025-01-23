// Code generated by mockery v2.51.0. DO NOT EDIT.

package mockery

import (
	mock "github.com/stretchr/testify/mock"
	status "github.com/walteh/copyrc/pkg/status"
)

// MockFileFormatter_status is an autogenerated mock type for the FileFormatter type
type MockFileFormatter_status struct {
	mock.Mock
}

type MockFileFormatter_status_Expecter struct {
	mock *mock.Mock
}

func (_m *MockFileFormatter_status) EXPECT() *MockFileFormatter_status_Expecter {
	return &MockFileFormatter_status_Expecter{mock: &_m.Mock}
}

// FormatError provides a mock function with given fields: err
func (_m *MockFileFormatter_status) FormatError(err error) string {
	ret := _m.Called(err)

	if len(ret) == 0 {
		panic("no return value specified for FormatError")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(error) string); ok {
		r0 = rf(err)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockFileFormatter_status_FormatError_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FormatError'
type MockFileFormatter_status_FormatError_Call struct {
	*mock.Call
}

// FormatError is a helper method to define mock.On call
//   - err error
func (_e *MockFileFormatter_status_Expecter) FormatError(err interface{}) *MockFileFormatter_status_FormatError_Call {
	return &MockFileFormatter_status_FormatError_Call{Call: _e.mock.On("FormatError", err)}
}

func (_c *MockFileFormatter_status_FormatError_Call) Run(run func(err error)) *MockFileFormatter_status_FormatError_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(error))
	})
	return _c
}

func (_c *MockFileFormatter_status_FormatError_Call) Return(_a0 string) *MockFileFormatter_status_FormatError_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockFileFormatter_status_FormatError_Call) RunAndReturn(run func(error) string) *MockFileFormatter_status_FormatError_Call {
	_c.Call.Return(run)
	return _c
}

// FormatFileOperation provides a mock function with given fields: path, fileType, _a2, isNew, isModified, isRemoved
func (_m *MockFileFormatter_status) FormatFileOperation(path string, fileType string, _a2 string, isNew bool, isModified bool, isRemoved bool) string {
	ret := _m.Called(path, fileType, _a2, isNew, isModified, isRemoved)

	if len(ret) == 0 {
		panic("no return value specified for FormatFileOperation")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string, string, bool, bool, bool) string); ok {
		r0 = rf(path, fileType, _a2, isNew, isModified, isRemoved)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockFileFormatter_status_FormatFileOperation_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FormatFileOperation'
type MockFileFormatter_status_FormatFileOperation_Call struct {
	*mock.Call
}

// FormatFileOperation is a helper method to define mock.On call
//   - path string
//   - fileType string
//   - _a2 string
//   - isNew bool
//   - isModified bool
//   - isRemoved bool
func (_e *MockFileFormatter_status_Expecter) FormatFileOperation(path interface{}, fileType interface{}, _a2 interface{}, isNew interface{}, isModified interface{}, isRemoved interface{}) *MockFileFormatter_status_FormatFileOperation_Call {
	return &MockFileFormatter_status_FormatFileOperation_Call{Call: _e.mock.On("FormatFileOperation", path, fileType, _a2, isNew, isModified, isRemoved)}
}

func (_c *MockFileFormatter_status_FormatFileOperation_Call) Run(run func(path string, fileType string, _a2 string, isNew bool, isModified bool, isRemoved bool)) *MockFileFormatter_status_FormatFileOperation_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string), args[2].(string), args[3].(bool), args[4].(bool), args[5].(bool))
	})
	return _c
}

func (_c *MockFileFormatter_status_FormatFileOperation_Call) Return(_a0 string) *MockFileFormatter_status_FormatFileOperation_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockFileFormatter_status_FormatFileOperation_Call) RunAndReturn(run func(string, string, string, bool, bool, bool) string) *MockFileFormatter_status_FormatFileOperation_Call {
	_c.Call.Return(run)
	return _c
}

// FormatFileStatus provides a mock function with given fields: filename, _a1, metadata
func (_m *MockFileFormatter_status) FormatFileStatus(filename string, _a1 status.FileStatus, metadata map[string]string) string {
	ret := _m.Called(filename, _a1, metadata)

	if len(ret) == 0 {
		panic("no return value specified for FormatFileStatus")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(string, status.FileStatus, map[string]string) string); ok {
		r0 = rf(filename, _a1, metadata)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockFileFormatter_status_FormatFileStatus_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FormatFileStatus'
type MockFileFormatter_status_FormatFileStatus_Call struct {
	*mock.Call
}

// FormatFileStatus is a helper method to define mock.On call
//   - filename string
//   - _a1 status.FileStatus
//   - metadata map[string]string
func (_e *MockFileFormatter_status_Expecter) FormatFileStatus(filename interface{}, _a1 interface{}, metadata interface{}) *MockFileFormatter_status_FormatFileStatus_Call {
	return &MockFileFormatter_status_FormatFileStatus_Call{Call: _e.mock.On("FormatFileStatus", filename, _a1, metadata)}
}

func (_c *MockFileFormatter_status_FormatFileStatus_Call) Run(run func(filename string, _a1 status.FileStatus, metadata map[string]string)) *MockFileFormatter_status_FormatFileStatus_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(status.FileStatus), args[2].(map[string]string))
	})
	return _c
}

func (_c *MockFileFormatter_status_FormatFileStatus_Call) Return(_a0 string) *MockFileFormatter_status_FormatFileStatus_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockFileFormatter_status_FormatFileStatus_Call) RunAndReturn(run func(string, status.FileStatus, map[string]string) string) *MockFileFormatter_status_FormatFileStatus_Call {
	_c.Call.Return(run)
	return _c
}

// FormatHeader provides a mock function with no fields
func (_m *MockFileFormatter_status) FormatHeader() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for FormatHeader")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockFileFormatter_status_FormatHeader_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FormatHeader'
type MockFileFormatter_status_FormatHeader_Call struct {
	*mock.Call
}

// FormatHeader is a helper method to define mock.On call
func (_e *MockFileFormatter_status_Expecter) FormatHeader() *MockFileFormatter_status_FormatHeader_Call {
	return &MockFileFormatter_status_FormatHeader_Call{Call: _e.mock.On("FormatHeader")}
}

func (_c *MockFileFormatter_status_FormatHeader_Call) Run(run func()) *MockFileFormatter_status_FormatHeader_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockFileFormatter_status_FormatHeader_Call) Return(_a0 string) *MockFileFormatter_status_FormatHeader_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockFileFormatter_status_FormatHeader_Call) RunAndReturn(run func() string) *MockFileFormatter_status_FormatHeader_Call {
	_c.Call.Return(run)
	return _c
}

// FormatProgress provides a mock function with given fields: current, total
func (_m *MockFileFormatter_status) FormatProgress(current int, total int) string {
	ret := _m.Called(current, total)

	if len(ret) == 0 {
		panic("no return value specified for FormatProgress")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(int, int) string); ok {
		r0 = rf(current, total)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockFileFormatter_status_FormatProgress_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FormatProgress'
type MockFileFormatter_status_FormatProgress_Call struct {
	*mock.Call
}

// FormatProgress is a helper method to define mock.On call
//   - current int
//   - total int
func (_e *MockFileFormatter_status_Expecter) FormatProgress(current interface{}, total interface{}) *MockFileFormatter_status_FormatProgress_Call {
	return &MockFileFormatter_status_FormatProgress_Call{Call: _e.mock.On("FormatProgress", current, total)}
}

func (_c *MockFileFormatter_status_FormatProgress_Call) Run(run func(current int, total int)) *MockFileFormatter_status_FormatProgress_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int), args[1].(int))
	})
	return _c
}

func (_c *MockFileFormatter_status_FormatProgress_Call) Return(_a0 string) *MockFileFormatter_status_FormatProgress_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockFileFormatter_status_FormatProgress_Call) RunAndReturn(run func(int, int) string) *MockFileFormatter_status_FormatProgress_Call {
	_c.Call.Return(run)
	return _c
}

// FormatRepoInfo provides a mock function with given fields: repo, ref
func (_m *MockFileFormatter_status) FormatRepoInfo(repo string, ref string) string {
	ret := _m.Called(repo, ref)

	if len(ret) == 0 {
		panic("no return value specified for FormatRepoInfo")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string) string); ok {
		r0 = rf(repo, ref)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockFileFormatter_status_FormatRepoInfo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FormatRepoInfo'
type MockFileFormatter_status_FormatRepoInfo_Call struct {
	*mock.Call
}

// FormatRepoInfo is a helper method to define mock.On call
//   - repo string
//   - ref string
func (_e *MockFileFormatter_status_Expecter) FormatRepoInfo(repo interface{}, ref interface{}) *MockFileFormatter_status_FormatRepoInfo_Call {
	return &MockFileFormatter_status_FormatRepoInfo_Call{Call: _e.mock.On("FormatRepoInfo", repo, ref)}
}

func (_c *MockFileFormatter_status_FormatRepoInfo_Call) Run(run func(repo string, ref string)) *MockFileFormatter_status_FormatRepoInfo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *MockFileFormatter_status_FormatRepoInfo_Call) Return(_a0 string) *MockFileFormatter_status_FormatRepoInfo_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockFileFormatter_status_FormatRepoInfo_Call) RunAndReturn(run func(string, string) string) *MockFileFormatter_status_FormatRepoInfo_Call {
	_c.Call.Return(run)
	return _c
}

// FormatSectionHeader provides a mock function with given fields: path
func (_m *MockFileFormatter_status) FormatSectionHeader(path string) string {
	ret := _m.Called(path)

	if len(ret) == 0 {
		panic("no return value specified for FormatSectionHeader")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockFileFormatter_status_FormatSectionHeader_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FormatSectionHeader'
type MockFileFormatter_status_FormatSectionHeader_Call struct {
	*mock.Call
}

// FormatSectionHeader is a helper method to define mock.On call
//   - path string
func (_e *MockFileFormatter_status_Expecter) FormatSectionHeader(path interface{}) *MockFileFormatter_status_FormatSectionHeader_Call {
	return &MockFileFormatter_status_FormatSectionHeader_Call{Call: _e.mock.On("FormatSectionHeader", path)}
}

func (_c *MockFileFormatter_status_FormatSectionHeader_Call) Run(run func(path string)) *MockFileFormatter_status_FormatSectionHeader_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockFileFormatter_status_FormatSectionHeader_Call) Return(_a0 string) *MockFileFormatter_status_FormatSectionHeader_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockFileFormatter_status_FormatSectionHeader_Call) RunAndReturn(run func(string) string) *MockFileFormatter_status_FormatSectionHeader_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockFileFormatter_status creates a new instance of MockFileFormatter_status. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockFileFormatter_status(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockFileFormatter_status {
	mock := &MockFileFormatter_status{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
