package contract_test

import (
	"github.com/stretchr/testify/assert"
	"identification-service/pkg/http/contract"
	"identification-service/pkg/test"
	"testing"
)

func TestCreateNewSuccessResponse(t *testing.T) {
	data := test.RandString(8)

	actualResponse := contract.NewSuccessResponse(data)

	expectedResponse := contract.APIResponse{
		Data:    data,
		Success: true,
		Error:   nil,
	}

	assert.Equal(t, expectedResponse, actualResponse)
}

func TestCreateNewFailureResponse(t *testing.T) {
	description := "some error"

	actualResponse := contract.NewFailureResponse(description)

	expectedResponse := contract.APIResponse{
		Data:    nil,
		Success: false,
		Error: &contract.Error{
			Message: description,
		},
	}

	assert.Equal(t, expectedResponse, actualResponse)
}
