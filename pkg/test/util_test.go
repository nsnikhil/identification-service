package test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/app"
	"identification-service/pkg/config"
	"identification-service/pkg/database"
	"identification-service/pkg/http/contract"
	"identification-service/pkg/queue"
	"identification-service/pkg/util"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
)

//TODO: PICK FROM CONFIG
const address = "http://127.0.0.1:8089"

type testDeps struct {
	cl  *http.Client
	db  database.SQLDatabase
	cc  *redis.Client
	qu  queue.Queue
	cfg config.Config
	ctx context.Context
}

func setupTest(t *testing.T) testDeps {
	require.NoError(t, os.Setenv("ENV", "test"))

	configFile := "../../local.env"
	cfg := config.NewConfig(configFile)

	db := NewDB(t, cfg)
	cc := NewCache(t, cfg)
	qu := NewQueue(t, cfg.QueueConfig())

	go app.StartHTTPServer(configFile)
	go app.StartWorker(configFile)
	time.Sleep(time.Second * 2)

	return testDeps{
		db:  db,
		qu:  qu,
		cc:  cc,
		cfg: cfg,
		cl:  &http.Client{Timeout: time.Minute},
		ctx: context.Background(),
	}
}

func tearDownTest(t *testing.T, deps testDeps) {
	//deps.cl.CloseIdleConnections()
	require.NoError(t, os.Unsetenv("ENV"))
	//require.NoError(t, deps.ch.Close())
	//require.NoError(t, deps.db.Close())
	//require.NoError(t, deps.cc.Close())
}

func registerClientAndGetHeaders(t *testing.T, cfg config.AuthConfig, cl *http.Client, clientData map[string]interface{}) map[string]string {
	reqBody := getRegisterClientReqBody(clientData)

	clientResp := testRegisterClient(t, cfg, cl, http.StatusCreated, contract.APIResponse{Success: true}, reqBody)

	return map[string]string{
		"CLIENT-ID":     reqBody.Name,
		"CLIENT-SECRET": clientResp.Data.(map[string]interface{})["secret"].(string),
	}
}

func signUpUser(
	t *testing.T,
	cfg config.QueueConfig,
	cl *http.Client,
	qu queue.Queue,
	headers map[string]string,
) contract.CreateUserRequest {

	reqBody := getCreateUserReqBody(map[string]interface{}{})

	expectedRespData := contract.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"message": "user created successfully"},
	}

	testSignUpUser(t, cfg, cl, qu, http.StatusCreated, expectedRespData, headers, reqBody, false)

	return reqBody
}

func loginUser(t *testing.T, cl *http.Client, headers map[string]string, reqBodyOverride map[string]interface{}) {
	reqBody := getLoginReqBody(reqBodyOverride)
	testLogin(t, cl, http.StatusCreated, contract.APIResponse{Success: true}, headers, reqBody)
}

func execRequest(t *testing.T, cl *http.Client, req *http.Request) *http.Response {
	resp, err := cl.Do(req)
	require.NoError(t, err)

	return resp
}

func getData(t *testing.T, expectedCode int, resp *http.Response) contract.APIResponse {
	assert.Equal(t, expectedCode, resp.StatusCode)

	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	var res contract.APIResponse
	err = json.Unmarshal(b, &res)
	require.NoError(t, err)

	return res
}

func verifyResp(
	t *testing.T,
	expectedResponse contract.APIResponse,
	actualResponse contract.APIResponse,
	verifyDataEquality bool,
	dataFunc func(interface{}) []interface{},
) {

	require.Equal(t, expectedResponse.Success, actualResponse.Success)

	if !expectedResponse.Success {
		assert.Equal(t, expectedResponse.Error, actualResponse.Error)
		return
	}

	if verifyDataEquality {
		assert.Equal(t, expectedResponse.Data, actualResponse.Data)
		return
	}

	//TODO: REFACTOR
	data := dataFunc(actualResponse.Data)
	for _, d := range data {
		s, ok := d.(string)
		if ok {
			assert.NotEmpty(t, s)
		} else {
			assert.NotNil(t, d)
		}
	}
}

func newRequest(t *testing.T, method, path string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s", address, path), body)
	require.NoError(t, err)

	return req
}

func testMessageConsume(t *testing.T, queueName string, qu queue.Queue) {
	ch, err := qu.Consume(queueName)
	require.NoError(t, err)

	for {
		select {
		case msg := <-ch:
			assert.NotEmpty(t, string(msg.Body))
			assert.True(t, util.IsValidUUID(string(msg.Body)))
			return
		case <-time.After(time.Second * 2):
			t.Fail()
			return
		}
	}
}

func truncateTables(t *testing.T, ctx context.Context, db database.SQLDatabase, tableNames ...string) {
	for _, tableName := range tableNames {
		_, err := db.ExecContext(ctx, fmt.Sprintf(`truncate %s cascade`, tableName))
		require.NoError(t, err)
	}
}
