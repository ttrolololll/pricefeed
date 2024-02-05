// Code generated by mockery v2.40.1. DO NOT EDIT.

package coindesk

import mock "github.com/stretchr/testify/mock"

// MockIClient is an autogenerated mock type for the IClient type
type MockIClient struct {
	mock.Mock
}

// CurrentPrice provides a mock function with given fields: target
func (_m *MockIClient) CurrentPrice(target string) (*CurrentPriceResponse, error) {
	ret := _m.Called(target)

	if len(ret) == 0 {
		panic("no return value specified for CurrentPrice")
	}

	var r0 *CurrentPriceResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*CurrentPriceResponse, error)); ok {
		return rf(target)
	}
	if rf, ok := ret.Get(0).(func(string) *CurrentPriceResponse); ok {
		r0 = rf(target)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*CurrentPriceResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(target)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewMockIClient creates a new instance of MockIClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockIClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockIClient {
	mock := &MockIClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}