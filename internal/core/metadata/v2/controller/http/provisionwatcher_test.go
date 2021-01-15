//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	v2MetadataContainer "github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/infrastructure/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	contractsV2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var testProvisionWatcherName = "TestProvisionWatcher"
var testProvisionWatcherLabels = []string{"test", "temp"}
var testProvisionWatcherIdentifiers = map[string]string{
	"address": "localhost",
	"port":    "3[0-9]{2}",
}
var testProvisionWatcherBlockingIdentifiers = map[string][]string{
	"port": {"397", "398", "399"},
}
var testProvisionWatcherAutoEvents = []dtos.AutoEvent{
	{Resource: "TestResource", Frequency: "300ms", OnChange: true},
}

func buildTestAddProvisionWatcherRequest() requests.AddProvisionWatcherRequest {
	return requests.AddProvisionWatcherRequest{
		BaseRequest: common.BaseRequest{
			RequestId: ExampleUUID,
		},
		ProvisionWatcher: dtos.ProvisionWatcher{
			Id:                  ExampleUUID,
			Name:                testProvisionWatcherName,
			Labels:              testProvisionWatcherLabels,
			Identifiers:         testProvisionWatcherIdentifiers,
			BlockingIdentifiers: testProvisionWatcherBlockingIdentifiers,
			ProfileName:         TestDeviceProfileName,
			ServiceName:         TestDeviceServiceName,
			AdminState:          models.Unlocked,
			AutoEvents:          testProvisionWatcherAutoEvents,
		},
	}
}

func buildTestUpdateProvisionWatcherRequest() requests.UpdateProvisionWatcherRequest {
	testUUID := ExampleUUID
	testName := testProvisionWatcherName
	testServiceName := TestDeviceServiceName
	testProfileName := TestDeviceProfileName
	testAdminState := models.Unlocked

	var testUpdateProvisionWatcherReq = requests.UpdateProvisionWatcherRequest{
		BaseRequest: common.BaseRequest{
			RequestId: ExampleUUID,
		},
		ProvisionWatcher: dtos.UpdateProvisionWatcher{
			Id:                  &testUUID,
			Name:                &testName,
			Labels:              testProvisionWatcherLabels,
			Identifiers:         testProvisionWatcherIdentifiers,
			BlockingIdentifiers: testProvisionWatcherBlockingIdentifiers,
			ServiceName:         &testServiceName,
			ProfileName:         &testProfileName,
			AdminState:          &testAdminState,
			AutoEvents:          testProvisionWatcherAutoEvents,
		},
	}

	return testUpdateProvisionWatcherReq
}

func TestProvisionWatcherController_AddProvisionWatcher_Created(t *testing.T) {
	validReq := buildTestAddProvisionWatcherRequest()
	pwModel := requests.AddProvisionWatcherReqToProvisionWatcherModels([]requests.AddProvisionWatcherRequest{validReq})[0]
	expectedRequestId := ExampleUUID

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("AddProvisionWatcher", pwModel).Return(pwModel, nil)
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewProvisionWatcherController(dic)
	assert.NotNil(t, controller)

	reqWithNoID := validReq
	reqWithNoID.RequestId = ""

	tests := []struct {
		name    string
		Request []requests.AddProvisionWatcherRequest
	}{
		{"Valid - AddProvisionWatcherRequest", []requests.AddProvisionWatcherRequest{validReq}},
		{"Valid - no RequestId", []requests.AddProvisionWatcherRequest{reqWithNoID}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.Request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, contractsV2.ApiProvisionWatcherRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddProvisionWatcher)
			handler.ServeHTTP(recorder, req)
			var res []common.BaseWithIdResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, contractsV2.ApiVersion, res[0].ApiVersion, "API version not as expected")
			if res[0].RequestId != "" {
				assert.Equal(t, expectedRequestId, res[0].RequestId, "RequestID not as expected")
			}
			assert.Equal(t, http.StatusCreated, res[0].StatusCode, "BaseResonse status code not as expected")
			assert.Empty(t, res[0].Message, "Message should be empty when it's successful")
		})
	}
}

func TestProvisionWatcherController_AddProvisionWatcher_BadRequest(t *testing.T) {
	dic := mockDic()

	controller := NewProvisionWatcherController(dic)
	assert.NotNil(t, controller)

	provisionWatcher := buildTestAddProvisionWatcherRequest()
	badRequestId := provisionWatcher
	badRequestId.RequestId = "niv3sl"
	noName := provisionWatcher
	noName.ProvisionWatcher.Name = ""

	tests := []struct {
		name    string
		Request []requests.AddProvisionWatcherRequest
	}{
		{"Invalid - Bad requestId", []requests.AddProvisionWatcherRequest{badRequestId}},
		{"Invalid - Bad name", []requests.AddProvisionWatcherRequest{noName}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.Request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, contractsV2.ApiProvisionWatcherRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddProvisionWatcher)
			handler.ServeHTTP(recorder, req)

			// Assert
			assert.Equal(t, http.StatusBadRequest, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.NotEmpty(t, string(recorder.Body.Bytes()), "Message is empty")
		})
	}
}

func TestProvisionWatcherController_AddProvisionWatcher_Duplicated(t *testing.T) {
	expectedRequestId := ExampleUUID

	duplicateIdRequest := buildTestAddProvisionWatcherRequest()
	duplicateIdModel := requests.AddProvisionWatcherReqToProvisionWatcherModels([]requests.AddProvisionWatcherRequest{duplicateIdRequest})[0]
	duplicateIdDBError := errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("provision watcher id %s already exists", duplicateIdModel.Id), nil)

	duplicateNameRequest := buildTestAddProvisionWatcherRequest()
	duplicateNameRequest.ProvisionWatcher.Id = "" // The infrastructure layer will generate id when the id field is empty
	duplicateNameModel := requests.AddProvisionWatcherReqToProvisionWatcherModels([]requests.AddProvisionWatcherRequest{duplicateNameRequest})[0]
	duplicateNameDBError := errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("provision watcher name %s already exists", duplicateNameModel.Name), nil)

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("AddProvisionWatcher", duplicateNameModel).Return(duplicateNameModel, duplicateNameDBError)
	dbClientMock.On("AddProvisionWatcher", duplicateIdModel).Return(duplicateIdModel, duplicateIdDBError)
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewProvisionWatcherController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name          string
		request       []requests.AddProvisionWatcherRequest
		expectedError errors.CommonEdgeX
	}{
		{"duplicate id", []requests.AddProvisionWatcherRequest{duplicateIdRequest}, duplicateIdDBError},
		{"duplicate name", []requests.AddProvisionWatcherRequest{duplicateNameRequest}, duplicateNameDBError},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, contractsV2.ApiProvisionWatcherRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddProvisionWatcher)
			handler.ServeHTTP(recorder, req)
			var res []common.BaseWithIdResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, contractsV2.ApiVersion, res[0].ApiVersion, "API Version not as expected")
			assert.Equal(t, expectedRequestId, res[0].RequestId, "RequestID not as expected")
			assert.Equal(t, http.StatusConflict, res[0].StatusCode, "BaseResponse status code not as expected")
			assert.Contains(t, res[0].Message, testCase.expectedError.Message(), "Message not as expected")
		})
	}
}

func TestProvisionWatcherController_ProvisionWatcherByName(t *testing.T) {
	provisionWatcher := dtos.ToProvisionWatcherModel(buildTestAddProvisionWatcherRequest().ProvisionWatcher)
	emptyName := ""
	notFoundName := "notFoundName"

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ProvisionWatcherByName", provisionWatcher.Name).Return(provisionWatcher, nil)
	dbClientMock.On("ProvisionWatcherByName", notFoundName).Return(models.ProvisionWatcher{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "provision watcher doesn't exist in the database", nil))
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewProvisionWatcherController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name                 string
		provisionWatcherName string
		errorExpected        bool
		expectedStatusCode   int
	}{
		{"Valid - find provision watcher by name", provisionWatcher.Name, false, http.StatusOK},
		{"Invalid - name parameter is empty", emptyName, true, http.StatusBadRequest},
		{"Invalid - provision watcher not found by name", notFoundName, true, http.StatusNotFound},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s", contractsV2.ApiProvisionWatcherByNameRoute, testCase.provisionWatcherName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{contractsV2.Name: testCase.provisionWatcherName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.ProvisionWatcherByName)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.ProvisionWatcherResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.provisionWatcherName, res.ProvisionWatcher.Name, "Name not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestProvisionWatcherController_ProvisionWatchersByServiceName(t *testing.T) {
	provisionWatcher := dtos.ToProvisionWatcherModel(buildTestAddProvisionWatcherRequest().ProvisionWatcher)
	testServiceA := "testServiceA"
	testServiceB := "testServiceB"
	pw1WithServiceA := provisionWatcher
	pw1WithServiceA.ServiceName = testServiceA
	pw2WithServiceA := provisionWatcher
	pw2WithServiceA.ServiceName = testServiceA
	pw3WithServiceB := provisionWatcher
	pw3WithServiceB.ServiceName = testServiceB

	provisionWatchers := []models.ProvisionWatcher{pw1WithServiceA, pw2WithServiceA, pw3WithServiceB}

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ProvisionWatchersByServiceName", 0, 5, testServiceA).Return([]models.ProvisionWatcher{provisionWatchers[0], provisionWatchers[1]}, nil)
	dbClientMock.On("ProvisionWatchersByServiceName", 1, 1, testServiceA).Return([]models.ProvisionWatcher{provisionWatchers[1]}, nil)
	dbClientMock.On("ProvisionWatchersByServiceName", 4, 1, testServiceB).Return([]models.ProvisionWatcher{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewProvisionWatcherController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		serviceName        string
		errorExpected      bool
		expectedCount      int
		expectedStatusCode int
	}{
		{"Valid - get provision watchers with serviceName", "0", "5", testServiceA, false, 2, http.StatusOK},
		{"Valid - get provision watchers with offset and limit", "1", "1", testServiceA, false, 1, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", testServiceB, true, 0, http.StatusNotFound},
		{"Invalid - get provision watchers without serviceName", "0", "10", "", true, 0, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, contractsV2.ApiProvisionWatcherByServiceNameRoute, http.NoBody)
			query := req.URL.Query()
			query.Add(contractsV2.Offset, testCase.offset)
			query.Add(contractsV2.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{contractsV2.Name: testCase.serviceName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.ProvisionWatchersByServiceName)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiProvisionWatchersResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.ProvisionWatchers), "ProvisionWatcher count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestProvisionWatcherController_ProvisionWatchersByProfileName(t *testing.T) {
	provisionWatcher := dtos.ToProvisionWatcherModel(buildTestAddProvisionWatcherRequest().ProvisionWatcher)
	testProfileA := "testProfileA"
	testProfileB := "testProfileB"
	pw1WithProfileA := provisionWatcher
	pw1WithProfileA.ProfileName = testProfileA
	pw2WithProfileA := provisionWatcher
	pw2WithProfileA.ProfileName = testProfileA
	pw3WithProfileB := provisionWatcher
	pw3WithProfileB.ProfileName = testProfileB

	provisionWatchers := []models.ProvisionWatcher{pw1WithProfileA, pw2WithProfileA, pw3WithProfileB}

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ProvisionWatchersByProfileName", 0, 5, testProfileA).Return([]models.ProvisionWatcher{provisionWatchers[0], provisionWatchers[1]}, nil)
	dbClientMock.On("ProvisionWatchersByProfileName", 1, 1, testProfileA).Return([]models.ProvisionWatcher{provisionWatchers[1]}, nil)
	dbClientMock.On("ProvisionWatchersByProfileName", 4, 1, testProfileB).Return([]models.ProvisionWatcher{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewProvisionWatcherController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		profileName        string
		errorExpected      bool
		expectedCount      int
		expectedStatusCode int
	}{
		{"Valid - get provision watchers with profileName", "0", "5", testProfileA, false, 2, http.StatusOK},
		{"Valid - get provision watchers with offset and limit", "1", "1", testProfileA, false, 1, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", testProfileB, true, 0, http.StatusNotFound},
		{"Invalid - get provision watchers without profileName", "0", "10", "", true, 0, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, contractsV2.ApiProvisionWatcherByProfileNameRoute, http.NoBody)
			query := req.URL.Query()
			query.Add(contractsV2.Offset, testCase.offset)
			query.Add(contractsV2.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{contractsV2.Name: testCase.profileName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.ProvisionWatchersByProfileName)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiProvisionWatchersResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.ProvisionWatchers), "ProvisionWatcher count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestProvisionWatcherController_AllProvisionWatchers(t *testing.T) {
	provisionWatcher := dtos.ToProvisionWatcherModel(buildTestAddProvisionWatcherRequest().ProvisionWatcher)
	provisionWatchers := []models.ProvisionWatcher{provisionWatcher, provisionWatcher, provisionWatcher}

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("AllProvisionWatchers", 0, 10, []string(nil)).Return(provisionWatchers, nil)
	dbClientMock.On("AllProvisionWatchers", 0, 5, testProvisionWatcherLabels).Return([]models.ProvisionWatcher{provisionWatchers[0], provisionWatchers[1]}, nil)
	dbClientMock.On("AllProvisionWatchers", 1, 2, []string(nil)).Return([]models.ProvisionWatcher{provisionWatchers[1], provisionWatchers[2]}, nil)
	dbClientMock.On("AllProvisionWatchers", 4, 1, testProvisionWatcherLabels).Return([]models.ProvisionWatcher{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewProvisionWatcherController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		labels             string
		errorExpected      bool
		expectedCount      int
		expectedStatusCode int
	}{
		{"Valid - get provision watchers without labels", "0", "10", "", false, 3, http.StatusOK},
		{"Valid - get provision watchers with labels", "0", "5", strings.Join(testProvisionWatcherLabels, ","), false, 2, http.StatusOK},
		{"Valid - get provision watchers with offset and no labels", "1", "2", "", false, 2, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", strings.Join(testProvisionWatcherLabels, ","), true, 0, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, contractsV2.ApiAllProvisionWatcherRoute, http.NoBody)
			query := req.URL.Query()
			query.Add(contractsV2.Offset, testCase.offset)
			query.Add(contractsV2.Limit, testCase.limit)
			if len(testCase.labels) > 0 {
				query.Add(contractsV2.Labels, testCase.labels)
			}
			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AllProvisionWatchers)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiProvisionWatchersResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.ProvisionWatchers), "ProvisionWatcher count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestProvisionWatcherController_DeleteProvisionWatcherByName(t *testing.T) {
	provisionWatcher := dtos.ToProvisionWatcherModel(buildTestAddProvisionWatcherRequest().ProvisionWatcher)
	noName := ""
	notFoundName := "notFoundName"

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("DeleteProvisionWatcherByName", provisionWatcher.Name).Return(nil)
	dbClientMock.On("DeleteProvisionWatcherByName", notFoundName).Return(errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "provision watcher doesn't exist in the database", nil))
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewProvisionWatcherController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name                 string
		provisionWatcherName string
		expectedStatusCode   int
	}{
		{"Valid - delete provision watcher by name", provisionWatcher.Name, http.StatusOK},
		{"Invalid - name parameter is empty", noName, http.StatusBadRequest},
		{"Invalid - provision watcher not found by name", notFoundName, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s", contractsV2.ApiProvisionWatcherByNameRoute, testCase.provisionWatcherName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{contractsV2.Name: testCase.provisionWatcherName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeleteProvisionWatcherByName)
			handler.ServeHTTP(recorder, req)
			var res common.BaseResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)

			// Assert
			assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
			assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
			if testCase.expectedStatusCode == http.StatusOK {
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			} else {
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			}
		})
	}
}

func TestProvisionWatcherController_PatchProvisionWatcher(t *testing.T) {
	expectedRequestId := ExampleUUID
	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	testReq := buildTestUpdateProvisionWatcherRequest()
	pwModels := models.ProvisionWatcher{
		Id:                  *testReq.ProvisionWatcher.Id,
		Name:                *testReq.ProvisionWatcher.Name,
		Labels:              testReq.ProvisionWatcher.Labels,
		Identifiers:         testReq.ProvisionWatcher.Identifiers,
		BlockingIdentifiers: testReq.ProvisionWatcher.BlockingIdentifiers,
		AdminState:          models.AdminState(*testReq.ProvisionWatcher.AdminState),
		ServiceName:         *testReq.ProvisionWatcher.ServiceName,
		ProfileName:         *testReq.ProvisionWatcher.ProfileName,
		AutoEvents:          dtos.ToAutoEventModels(testReq.ProvisionWatcher.AutoEvents),
	}

	valid := testReq
	dbClientMock.On("DeviceServiceNameExists", *valid.ProvisionWatcher.ServiceName).Return(true, nil)
	dbClientMock.On("DeviceProfileNameExists", *valid.ProvisionWatcher.ProfileName).Return(true, nil)
	dbClientMock.On("ProvisionWatcherByName", *valid.ProvisionWatcher.Name).Return(pwModels, nil)
	dbClientMock.On("DeleteProvisionWatcherByName", *valid.ProvisionWatcher.Name).Return(nil)
	dbClientMock.On("AddProvisionWatcher", mock.Anything).Return(pwModels, nil)
	validWithNoReqID := testReq
	validWithNoReqID.RequestId = ""
	validWithNoId := testReq
	validWithNoId.ProvisionWatcher.Id = nil
	dbClientMock.On("ProvisionWatcherByName", *validWithNoId.ProvisionWatcher.Name).Return(pwModels, nil)
	validWithNoName := testReq
	validWithNoName.ProvisionWatcher.Name = nil
	dbClientMock.On("ProvisionWatcherById", *validWithNoName.ProvisionWatcher.Id).Return(pwModels, nil)

	invalidId := testReq
	invalidUUID := "invalidUUID"
	invalidId.ProvisionWatcher.Id = &invalidUUID

	emptyString := ""
	emptyId := testReq
	emptyId.ProvisionWatcher.Id = &emptyString
	emptyName := testReq
	emptyName.ProvisionWatcher.Id = nil
	emptyName.ProvisionWatcher.Name = &emptyString

	invalidNoIdAndName := testReq
	invalidNoIdAndName.ProvisionWatcher.Id = nil
	invalidNoIdAndName.ProvisionWatcher.Name = nil

	invalidNotFoundId := testReq
	invalidNotFoundId.ProvisionWatcher.Name = nil
	notFoundId := "12345678-1111-1234-5678-de9dac3fb9bc"
	invalidNotFoundId.ProvisionWatcher.Id = &notFoundId
	notFoundIdError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", notFoundId), nil)
	dbClientMock.On("ProvisionWatcherById", *invalidNotFoundId.ProvisionWatcher.Id).Return(pwModels, notFoundIdError)

	invalidNotFoundName := testReq
	invalidNotFoundName.ProvisionWatcher.Name = nil
	notFoundName := "notFoundName"
	invalidNotFoundName.ProvisionWatcher.Name = &notFoundName
	notFoundNameError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", notFoundName), nil)
	dbClientMock.On("ProvisionWatcherByName", *invalidNotFoundName.ProvisionWatcher.Name).Return(pwModels, notFoundNameError)

	notFountServiceName := "notFoundService"
	notFoundService := testReq
	notFoundService.ProvisionWatcher.ServiceName = &notFountServiceName
	dbClientMock.On("DeviceServiceNameExists", *notFoundService.ProvisionWatcher.ServiceName).Return(false, nil)
	notFountProfileName := "notFoundProfile"
	notFoundProfile := testReq
	notFoundProfile.ProvisionWatcher.ProfileName = &notFountProfileName
	dbClientMock.On("DeviceProfileNameExists", *notFoundProfile.ProvisionWatcher.ProfileName).Return(false, nil)

	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewProvisionWatcherController(dic)
	require.NotNil(t, controller)
	tests := []struct {
		name                 string
		request              []requests.UpdateProvisionWatcherRequest
		expectedStatusCode   int
		expectedResponseCode int
	}{
		{"Valid", []requests.UpdateProvisionWatcherRequest{valid}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no requestId", []requests.UpdateProvisionWatcherRequest{validWithNoReqID}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no id", []requests.UpdateProvisionWatcherRequest{validWithNoId}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no name", []requests.UpdateProvisionWatcherRequest{validWithNoName}, http.StatusMultiStatus, http.StatusOK},
		{"Invalid - invalid id", []requests.UpdateProvisionWatcherRequest{invalidId}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - empty id", []requests.UpdateProvisionWatcherRequest{emptyId}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - empty name", []requests.UpdateProvisionWatcherRequest{emptyName}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - not found id", []requests.UpdateProvisionWatcherRequest{invalidNotFoundId}, http.StatusMultiStatus, http.StatusNotFound},
		{"Invalid - not found name", []requests.UpdateProvisionWatcherRequest{invalidNotFoundName}, http.StatusMultiStatus, http.StatusNotFound},
		{"Invalid - no id and name", []requests.UpdateProvisionWatcherRequest{invalidNoIdAndName}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - not found service", []requests.UpdateProvisionWatcherRequest{notFoundService}, http.StatusMultiStatus, http.StatusNotFound},
		{"Invalid - not found profile", []requests.UpdateProvisionWatcherRequest{notFoundProfile}, http.StatusMultiStatus, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, contractsV2.ApiProvisionWatcherRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.PatchProvisionWatcher)
			handler.ServeHTTP(recorder, req)

			if testCase.expectedStatusCode == http.StatusMultiStatus {
				var res []common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, contractsV2.ApiVersion, res[0].ApiVersion, "API Version not as expected")
				if res[0].RequestId != "" {
					assert.Equal(t, expectedRequestId, res[0].RequestId, "RequestID not as expected")
				}
				assert.Equal(t, testCase.expectedResponseCode, res[0].StatusCode, "BaseResponse status code not as expected")
				if testCase.expectedResponseCode == http.StatusOK {
					assert.Empty(t, res[0].Message, "Message should be empty when it is successful")
				} else {
					assert.NotEmpty(t, res[0].Message, "Response message doesn't contain the error message")
				}
			} else {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedResponseCode, res.StatusCode, "BaseResponse status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			}
		})
	}
}
