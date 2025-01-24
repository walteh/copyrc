// Code generated by mockery v2.51.0. DO NOT EDIT.

package mockery

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	remote "github.com/walteh/copyrc/cmd/copyrc-next/pkg/remote"
)

// MockProvider_remote is an autogenerated mock type for the Provider type
type MockProvider_remote struct {
	mock.Mock
}

type MockProvider_remote_Expecter struct {
	mock *mock.Mock
}

func (_m *MockProvider_remote) EXPECT() *MockProvider_remote_Expecter {
	return &MockProvider_remote_Expecter{mock: &_m.Mock}
}

// GetRepository provides a mock function with given fields: ctx, name
func (_m *MockProvider_remote) GetRepository(ctx context.Context, name string) (remote.Repository, error) {
	ret := _m.Called(ctx, name)

	if len(ret) == 0 {
		panic("no return value specified for GetRepository")
	}

	var r0 remote.Repository
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (remote.Repository, error)); ok {
		return rf(ctx, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) remote.Repository); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(remote.Repository)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockProvider_remote_GetRepository_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetRepository'
type MockProvider_remote_GetRepository_Call struct {
	*mock.Call
}

// GetRepository is a helper method to define mock.On call
//   - ctx context.Context
//   - name string
func (_e *MockProvider_remote_Expecter) GetRepository(ctx interface{}, name interface{}) *MockProvider_remote_GetRepository_Call {
	return &MockProvider_remote_GetRepository_Call{Call: _e.mock.On("GetRepository", ctx, name)}
}

func (_c *MockProvider_remote_GetRepository_Call) Run(run func(ctx context.Context, name string)) *MockProvider_remote_GetRepository_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockProvider_remote_GetRepository_Call) Return(_a0 remote.Repository, _a1 error) *MockProvider_remote_GetRepository_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockProvider_remote_GetRepository_Call) RunAndReturn(run func(context.Context, string) (remote.Repository, error)) *MockProvider_remote_GetRepository_Call {
	_c.Call.Return(run)
	return _c
}

// Name provides a mock function with no fields
func (_m *MockProvider_remote) Name() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Name")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockProvider_remote_Name_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Name'
type MockProvider_remote_Name_Call struct {
	*mock.Call
}

// Name is a helper method to define mock.On call
func (_e *MockProvider_remote_Expecter) Name() *MockProvider_remote_Name_Call {
	return &MockProvider_remote_Name_Call{Call: _e.mock.On("Name")}
}

func (_c *MockProvider_remote_Name_Call) Run(run func()) *MockProvider_remote_Name_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockProvider_remote_Name_Call) Return(_a0 string) *MockProvider_remote_Name_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockProvider_remote_Name_Call) RunAndReturn(run func() string) *MockProvider_remote_Name_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockProvider_remote creates a new instance of MockProvider_remote. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockProvider_remote(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockProvider_remote {
	mock := &MockProvider_remote{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
