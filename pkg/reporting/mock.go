package reporters

import "github.com/stretchr/testify/mock"

type MockPrometheus struct {
	mock.Mock
}

func (mp *MockPrometheus) ReportAttempt(bucket string) {
	mp.Called(bucket)
}

func (mp *MockPrometheus) ReportSuccess(bucket string) {
	mp.Called(bucket)
}

func (mp *MockPrometheus) ReportFailure(bucket string) {
	mp.Called(bucket)
}

func (mp *MockPrometheus) Observe(bucket string, value float64) {
	mp.Called(bucket, value)
}
