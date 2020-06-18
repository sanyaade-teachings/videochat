package main

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/oliveagle/jsonpath"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"io"
	"net/http"
	test "net/http/httptest"
	"nkonev.name/chat/client"
	"nkonev.name/chat/db"
	"nkonev.name/chat/handlers"
	. "nkonev.name/chat/logger"
	"nkonev.name/chat/utils"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	setup()
	retCode := m.Run()
	shutdown()
	os.Exit(retCode)
}

func shutdown() {}

func setup() {
	configFile := utils.InitFlags("./config-dev/config.yml")
	utils.InitViper(configFile, "")

	d, err := configureDb(nil)
	defer d.Close()
	if err != nil {
		Logger.Panicf("Error during getting db connection for test: %v", err)
	} else {
		_, err := d.Exec(`
	DROP SCHEMA IF EXISTS public CASCADE;
	CREATE SCHEMA IF NOT EXISTS public;
    GRANT ALL ON SCHEMA public TO chat;
    GRANT ALL ON SCHEMA public TO public;
    COMMENT ON SCHEMA public IS 'standard public schema';
`)
		if err != nil {
			Logger.Panicf("Error during dropping db: %v", err)
		}
	}
}

func TestExtractAuth(t *testing.T) {
	req := test.NewRequest("GET", "/should-be-secured", nil)
	headers := map[string][]string{
		"X-Auth-Expiresin": {"1590022342295000"},
		"X-Auth-Username":  {"tester"},
		"X-Auth-Userid":    {"1"},
	}
	req.Header = headers

	auth, err := handlers.ExtractAuth(req)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), auth.UserId)
	assert.Equal(t, "tester", auth.UserLogin)
	assert.Equal(t, int64(1590022342), auth.ExpiresAt)
}

func request(method, path string, body io.Reader, e *echo.Echo) (int, string, http.Header) {
	req := test.NewRequest(method, path, body)
	Header := map[string][]string{
		echo.HeaderContentType: {"application/json"},
		"X-Auth-Expiresin":     {"1590022342295000"},
		"X-Auth-Username":      {"tester"},
		"X-Auth-Userid":        {"1"},
	}
	req.Header = Header
	rec := test.NewRecorder()
	e.ServeHTTP(rec, req) // most wanted
	return rec.Code, rec.Body.String(), rec.Result().Header
}

func runTest(t *testing.T, testFunc interface{}) *fxtest.App {
	var s fx.Shutdowner
	app := fxtest.New(
		t,
		fx.Logger(Logger),
		fx.Populate(&s),
		fx.Provide(
			client.NewRestClient,
			handlers.ConfigureCentrifuge,
			configureEcho,
			configureStaticMiddleware,
			handlers.ConfigureAuthMiddleware,
			configureDb,
		),
		fx.Invoke(
			runMigrations,
			//runCentrifuge,
			//runEcho,
			testFunc,
		),
	)
	defer app.RequireStart().RequireStop()
	assert.NoError(t, s.Shutdown(), "error in app shutdown")
	return app
}

func TestGetChats(t *testing.T) {
	runTest(t, func(e *echo.Echo) {
		c, b, _ := request("GET", "/chat", nil, e)
		assert.Equal(t, http.StatusOK, c)
		assert.NotEmpty(t, b)
	})
}

func getJsonPathResult(t *testing.T, body string, jsonpath0 string) interface{} {
	var jsonData interface{}
	assert.Nil(t, json.Unmarshal([]byte(body), &jsonData))
	res, err := jsonpath.JsonPathLookup(jsonData, jsonpath0)
	assert.Nil(t, err)
	assert.NotEmpty(t, res)
	return res
}

func TestGetChatsPaginated(t *testing.T) {
	runTest(t, func(e *echo.Echo) {
		c, b, _ := request("GET", "/chat?page=2&size=3", nil, e)
		assert.Equal(t, http.StatusOK, c)
		assert.NotEmpty(t, b)

		typedTes := getJsonPathResult(t, b, "$.name").([]interface{})

		assert.Equal(t, 3, len(typedTes))

		assert.Equal(t, "sit", typedTes[0])
		assert.Equal(t, "amet", typedTes[1])
		assert.Equal(t, "With collegues", typedTes[2])
	})
}

func TestChatCrud(t *testing.T) {
	runTest(t, func(e *echo.Echo, db db.DB) {
		chatsBefore, _ := db.CountChats()
		c, b, _ := request("POST", "/chat", strings.NewReader(`{"name": "Ultra new chat"}`), e)
		assert.Equal(t, http.StatusCreated, c)

		chatsAfterCreate, _ := db.CountChats()
		assert.Equal(t, chatsBefore+1, chatsAfterCreate)

		idInterface := getJsonPathResult(t, b, "$.id").(interface{})
		idString := fmt.Sprintf("%v", idInterface)
		id, _ := utils.ParseInt64(idString)
		assert.True(t, id > 0)

		c3, b3, _ := request("GET", "/chat/"+idString, nil, e)
		assert.Equal(t, http.StatusOK, c3)
		nameInterface := getJsonPathResult(t, b3, "$.name").(interface{})
		nameString := fmt.Sprintf("%v", nameInterface)
		assert.Equal(t, "Ultra new chat", nameString)

		c2, _, _ := request("PUT", "/chat", strings.NewReader(`{ "id": `+idString+`, "name": "Mega ultra new chat"}`), e)
		assert.Equal(t, http.StatusAccepted, c2)
		row := db.QueryRow("SELECT title FROM chat WHERE id = $1", id)
		var newTitle string
		assert.Nil(t, row.Scan(&newTitle))
		assert.Equal(t, "Mega ultra new chat", newTitle)

		c1, _, _ := request("DELETE", "/chat/"+idString, nil, e)
		assert.Equal(t, http.StatusAccepted, c1)
		chatsAfterDelete, _ := db.CountChats()
		assert.Equal(t, chatsBefore, chatsAfterDelete)
	})
}

func interfaceToString(inter interface{}) string {
	return fmt.Sprintf("%v", inter)
}

func TestGetMessagesPaginated(t *testing.T) {
	runTest(t, func(e *echo.Echo) {
		c, b, _ := request("GET", "/chat/1/message?page=2&size=3", nil, e)
		assert.Equal(t, http.StatusOK, c)
		assert.NotEmpty(t, b)

		typedTes := getJsonPathResult(t, b, "$.text").([]interface{})

		assert.Equal(t, 3, len(typedTes))

		assert.True(t, strings.HasPrefix(interfaceToString(typedTes[0]), "generated_mes____sage5"))
		assert.True(t, strings.HasPrefix(interfaceToString(typedTes[1]), "generated_message6"))
		assert.True(t, strings.HasPrefix(interfaceToString(typedTes[2]), "generated_message7"))
	})
}
