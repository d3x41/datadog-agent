// Code generated by MockGen. DO NOT EDIT.
// Source: client.go

// The mock gen fails to carry this over, make sure to add it back
//go:build ec2

// Package aws is a generated GoMock package.
package aws

import (
	context "context"
	reflect "reflect"

	rds "github.com/aws/aws-sdk-go-v2/service/rds"
	gomock "github.com/golang/mock/gomock"
)

// MockRdsClient is a mock of RdsClient interface.
type MockRdsClient struct {
	ctrl     *gomock.Controller
	recorder *MockRdsClientMockRecorder
}

// MockRdsClientMockRecorder is the mock recorder for MockRdsClient.
type MockRdsClientMockRecorder struct {
	mock *MockRdsClient
}

// NewMockRdsClient creates a new mock instance.
func NewMockRdsClient(ctrl *gomock.Controller) *MockRdsClient {
	mock := &MockRdsClient{ctrl: ctrl}
	mock.recorder = &MockRdsClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRdsClient) EXPECT() *MockRdsClientMockRecorder {
	return m.recorder
}

// GetAuroraClusterEndpoints mocks base method.
func (m *MockRdsClient) GetAuroraClusterEndpoints(ctx context.Context, dbClusterIdentifiers []string, dbmTag string) (map[string]*AuroraCluster, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAuroraClusterEndpoints", ctx, dbClusterIdentifiers, dbmTag)
	ret0, _ := ret[0].(map[string]*AuroraCluster)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAuroraClusterEndpoints indicates an expected call of GetAuroraClusterEndpoints.
func (mr *MockRdsClientMockRecorder) GetAuroraClusterEndpoints(ctx, dbClusterIdentifiers, dbmTag interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAuroraClusterEndpoints", reflect.TypeOf((*MockRdsClient)(nil).GetAuroraClusterEndpoints), ctx, dbClusterIdentifiers, dbmTag)
}

// GetAuroraClustersFromTags mocks base method.
func (m *MockRdsClient) GetAuroraClustersFromTags(ctx context.Context, tags []string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAuroraClustersFromTags", ctx, tags)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAuroraClustersFromTags indicates an expected call of GetAuroraClustersFromTags.
func (mr *MockRdsClientMockRecorder) GetAuroraClustersFromTags(ctx, tags interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAuroraClustersFromTags", reflect.TypeOf((*MockRdsClient)(nil).GetAuroraClustersFromTags), ctx, tags)
}

// GetRdsInstancesFromTags mocks base method.
func (m *MockRdsClient) GetRdsInstancesFromTags(ctx context.Context, tags []string, dbmTag string) ([]Instance, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRdsInstancesFromTags", ctx, tags, dbmTag)
	ret0, _ := ret[0].([]Instance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRdsInstancesFromTags indicates an expected call of GetRdsInstancesFromTags.
func (mr *MockRdsClientMockRecorder) GetRdsInstancesFromTags(ctx, tags, dbmTag interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRdsInstancesFromTags", reflect.TypeOf((*MockRdsClient)(nil).GetRdsInstancesFromTags), ctx, tags, dbmTag)
}

// MockrdsService is a mock of rdsService interface.
type MockrdsService struct {
	ctrl     *gomock.Controller
	recorder *MockrdsServiceMockRecorder
}

// MockrdsServiceMockRecorder is the mock recorder for MockrdsService.
type MockrdsServiceMockRecorder struct {
	mock *MockrdsService
}

// NewMockrdsService creates a new mock instance.
func NewMockrdsService(ctrl *gomock.Controller) *MockrdsService {
	mock := &MockrdsService{ctrl: ctrl}
	mock.recorder = &MockrdsServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockrdsService) EXPECT() *MockrdsServiceMockRecorder {
	return m.recorder
}

// DescribeDBClusters mocks base method.
func (m *MockrdsService) DescribeDBClusters(ctx context.Context, params *rds.DescribeDBClustersInput, optFns ...func(*rds.Options)) (*rds.DescribeDBClustersOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, params}
	for _, a := range optFns {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DescribeDBClusters", varargs...)
	ret0, _ := ret[0].(*rds.DescribeDBClustersOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeDBClusters indicates an expected call of DescribeDBClusters.
func (mr *MockrdsServiceMockRecorder) DescribeDBClusters(ctx, params interface{}, optFns ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, params}, optFns...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeDBClusters", reflect.TypeOf((*MockrdsService)(nil).DescribeDBClusters), varargs...)
}

// DescribeDBInstances mocks base method.
func (m *MockrdsService) DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, params}
	for _, a := range optFns {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DescribeDBInstances", varargs...)
	ret0, _ := ret[0].(*rds.DescribeDBInstancesOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeDBInstances indicates an expected call of DescribeDBInstances.
func (mr *MockrdsServiceMockRecorder) DescribeDBInstances(ctx, params interface{}, optFns ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, params}, optFns...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeDBInstances", reflect.TypeOf((*MockrdsService)(nil).DescribeDBInstances), varargs...)
}
