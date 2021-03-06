// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import mock "github.com/stretchr/testify/mock"
import v1alpha1 "github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1alpha1"

// Deployer is an autogenerated mock type for the Deployer type
type Deployer struct {
	mock.Mock
}

// Deploy provides a mock function with given fields: link, function
func (_m *Deployer) Deploy(link *v1alpha1.Link, function *v1alpha1.Function) error {
	ret := _m.Called(link, function)

	var r0 error
	if rf, ok := ret.Get(0).(func(*v1alpha1.Link, *v1alpha1.Function) error); ok {
		r0 = rf(link, function)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Scale provides a mock function with given fields: link, replicas
func (_m *Deployer) Scale(link *v1alpha1.Link, replicas int) error {
	ret := _m.Called(link, replicas)

	var r0 error
	if rf, ok := ret.Get(0).(func(*v1alpha1.Link, int) error); ok {
		r0 = rf(link, replicas)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Undeploy provides a mock function with given fields: link
func (_m *Deployer) Undeploy(link *v1alpha1.Link) error {
	ret := _m.Called(link)

	var r0 error
	if rf, ok := ret.Get(0).(func(*v1alpha1.Link) error); ok {
		r0 = rf(link)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Update provides a mock function with given fields: link, function, replicas
func (_m *Deployer) Update(link *v1alpha1.Link, function *v1alpha1.Function, replicas int) error {
	ret := _m.Called(link, function, replicas)

	var r0 error
	if rf, ok := ret.Get(0).(func(*v1alpha1.Link, *v1alpha1.Function, int) error); ok {
		r0 = rf(link, function, replicas)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
