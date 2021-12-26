package restclient

import (
	"encoding/json"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	testDataFormat = "%s - %s"
	testSuccess    = "testSuccess"
	testFailed     = "testFailed"
)

type testBasicAuthenticator struct {
	username, password string
}

func newTestBasicAuthenticator(username, password string) Authenticator {
	return &testBasicAuthenticator{
		username: username,
		password: password,
	}
}

func (b testBasicAuthenticator) Apply(request *http.Request) error {
	request.SetBasicAuth(b.username, b.password)
	return nil
}

type testRequestBody struct {
	TestId   int    `json:"test_id"`
	TestName string `json:"test_name"`
}

type testHttpResponse struct {
	StatusCode int         `json:"status_code"`
	Data       interface{} `json:"data"`
}

func TestHttpClientRequests(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response testHttpResponse
		w.Header().Set("Content-Type", "application/json")

		handleRequest := func(reqMethod string, statusCode int, resultString string) {
			username, password, ok := r.BasicAuth()
			if !ok || username != "username" || password != "0123" {
				statusCode = http.StatusUnauthorized
				resultString = testFailed
			}
			response.StatusCode = statusCode
			response.Data = fmt.Sprintf(testDataFormat, resultString, reqMethod)
			w.WriteHeader(statusCode)
		}

		handleRequestWithBody := func(reqMethod string) {
			status := http.StatusOK
			resultString := testSuccess
			var body testRequestBody
			err := unmarshalRequestBody(r, &body)
			if err != nil || body.TestName != "Testing Request Body" || body.TestId != 123 {
				status = http.StatusBadRequest
				resultString = testFailed
			}
			handleRequest(reqMethod, status, resultString)
		}

		switch r.Method {
		case http.MethodGet:
			handleRequest(http.MethodGet, http.StatusOK, testSuccess)
		case http.MethodPost:
			handleRequestWithBody(http.MethodPost)
		case http.MethodPut:
			handleRequestWithBody(http.MethodPut)
		case http.MethodPatch:
			handleRequestWithBody(http.MethodPatch)
		case http.MethodDelete:
			handleRequest(http.MethodDelete, http.StatusOK, testSuccess)
		default:
			response.StatusCode = http.StatusForbidden
			response.Data = fmt.Sprintf(testDataFormat, testFailed, r.Method)
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			log.Fatalf("httptest server failed to respond with testHttpResponse %v", err.Error())
		}
	}))
	defer ts.Close()
	// e.g. ts URL: http://127.0.0.1:63316
	splittedURL := strings.Split(ts.URL, "://")
	testServerScheme := splittedURL[0]
	testServerHost := splittedURL[1]

	Convey("TEST HTTP GET", t, func() {
		var testResponse testHttpResponse
		auth := newTestBasicAuthenticator("username", "0123")
		req, reqErr := RequestBuilder().
			Scheme(testServerScheme).
			Host(testServerHost).
			Auth(auth).
			ResponseReference(&testResponse).
			Timeout(30 * time.Second).
			LoggingEnabled(true).
			Build()
		if reqErr != nil {
			log.Fatalf("failed to construct testRequest, %v", reqErr)
		}

		reqErr = req.Get()
		Convey("Response should be fetched without any err and with proper Data", func() {
			data, ok := testResponse.Data.(string)
			So(reqErr, ShouldBeNil)
			So(testResponse.StatusCode, ShouldEqual, http.StatusOK)
			So(ok, ShouldBeTrue)
			So(data, ShouldEqual, fmt.Sprintf(testDataFormat, testSuccess, http.MethodGet))
		})
	})

	Convey("TEST HTTP GET with failing authentication credentials", t, func() {
		var testResponse testHttpResponse
		auth := newTestBasicAuthenticator("FAILING", "CREDENTIALS")
		req, reqErr := RequestBuilder().
			Scheme(testServerScheme).
			Host(testServerHost).
			Auth(auth).
			ResponseReference(&testResponse).
			Timeout(30 * time.Second).
			LoggingEnabled(true).
			Build()
		if reqErr != nil {
			log.Fatalf("failed to construct testRequest, %v", reqErr)
		}

		reqErr = req.Get()
		Convey("Response should be fetched without any err and with proper Data", func() {
			_, ok := testResponse.Data.(string)
			So(ok, ShouldBeFalse)
			So(testResponse.StatusCode, ShouldEqual, 0)
			So(reqErr, ShouldNotBeNil)
			So(reqErr.GetTopLevelError(), ShouldEqual, UnauthorizedErr)
			So(reqErr.GetStatusCode(), ShouldEqual, http.StatusUnauthorized)
		})
	})

	Convey("TEST HTTP POST with correct body", t, func() {
		var testResponse testHttpResponse
		testBody := testRequestBody{
			TestId:   123,
			TestName: "Testing Request Body",
		}
		auth := newTestBasicAuthenticator("username", "0123")
		req, reqErr := RequestBuilder().
			Scheme(testServerScheme).
			Host(testServerHost).
			BodyJson(testBody).
			Auth(auth).
			ResponseReference(&testResponse).
			Timeout(30 * time.Second).
			LoggingEnabled(true).
			Build()
		if reqErr != nil {
			log.Fatalf("failed to construct testRequest, %v", reqErr)
		}

		reqErr = req.Post()
		Convey("Response should be fetched without any err and with proper Data", func() {
			data, ok := testResponse.Data.(string)
			So(reqErr, ShouldBeNil)
			So(testResponse.StatusCode, ShouldEqual, http.StatusOK)
			So(ok, ShouldBeTrue)
			So(data, ShouldEqual, fmt.Sprintf(testDataFormat, testSuccess, http.MethodPost))
		})
	})

	Convey("TEST HTTP POST with incorrect body", t, func() {
		var testResponse testHttpResponse
		testBody := "INCORRECT REQUEST BODY"
		auth := newTestBasicAuthenticator("username", "0123")
		req, reqErr := RequestBuilder().
			Scheme(testServerScheme).
			Host(testServerHost).
			BodyJson(testBody).
			Auth(auth).
			ResponseReference(&testResponse).
			Timeout(30 * time.Second).
			LoggingEnabled(true).
			Build()
		if reqErr != nil {
			log.Fatalf("failed to construct testRequest, %v", reqErr)
		}

		reqErr = req.Post()
		Convey("Response should be fetched without any err and with proper Data", func() {
			_, ok := testResponse.Data.(string)
			So(ok, ShouldBeFalse)
			So(testResponse.StatusCode, ShouldEqual, 0)
			So(reqErr, ShouldNotBeNil)
			So(reqErr.GetTopLevelError(), ShouldEqual, ParseErr)
			So(reqErr.GetStatusCode(), ShouldEqual, http.StatusBadRequest)
		})
	})

	Convey("TEST HTTP PUT with correct body", t, func() {
		var testResponse testHttpResponse
		testBody := testRequestBody{
			TestId:   123,
			TestName: "Testing Request Body",
		}
		auth := newTestBasicAuthenticator("username", "0123")
		req, reqErr := RequestBuilder().
			Scheme(testServerScheme).
			Host(testServerHost).
			BodyJson(testBody).
			Auth(auth).
			ResponseReference(&testResponse).
			Timeout(30 * time.Second).
			LoggingEnabled(true).
			Build()
		if reqErr != nil {
			log.Fatalf("failed to construct testRequest, %v", reqErr)
		}

		reqErr = req.Put()
		Convey("Response should be fetched without any err and with proper Data", func() {
			data, ok := testResponse.Data.(string)
			So(reqErr, ShouldBeNil)
			So(testResponse.StatusCode, ShouldEqual, http.StatusOK)
			So(ok, ShouldBeTrue)
			So(data, ShouldEqual, fmt.Sprintf(testDataFormat, testSuccess, http.MethodPut))
		})
	})

	Convey("TEST HTTP PATCH with correct body", t, func() {
		var testResponse testHttpResponse
		testBody := testRequestBody{
			TestId:   123,
			TestName: "Testing Request Body",
		}
		auth := newTestBasicAuthenticator("username", "0123")
		req, reqErr := RequestBuilder().
			Scheme(testServerScheme).
			Host(testServerHost).
			BodyJson(testBody).
			Auth(auth).
			ResponseReference(&testResponse).
			Timeout(30 * time.Second).
			LoggingEnabled(true).
			Build()
		if reqErr != nil {
			log.Fatalf("failed to construct testRequest, %v", reqErr)
		}

		reqErr = req.Patch()
		Convey("Response should be fetched without any err and with proper Data", func() {
			data, ok := testResponse.Data.(string)
			So(reqErr, ShouldBeNil)
			So(testResponse.StatusCode, ShouldEqual, http.StatusOK)
			So(ok, ShouldBeTrue)
			So(data, ShouldEqual, fmt.Sprintf(testDataFormat, testSuccess, http.MethodPatch))
		})
	})

	Convey("TEST HTTP DELETE", t, func() {
		var testResponse testHttpResponse
		auth := newTestBasicAuthenticator("username", "0123")
		req, reqErr := RequestBuilder().
			Scheme(testServerScheme).
			Host(testServerHost).
			Auth(auth).
			ResponseReference(&testResponse).
			Timeout(30 * time.Second).
			LoggingEnabled(true).
			Build()
		if reqErr != nil {
			log.Fatalf("failed to construct testRequest, %v", reqErr)
		}

		reqErr = req.Delete()
		Convey("Response should be fetched without any err and with proper Data", func() {
			data, ok := testResponse.Data.(string)
			So(reqErr, ShouldBeNil)
			So(testResponse.StatusCode, ShouldEqual, http.StatusOK)
			So(ok, ShouldBeTrue)
			So(data, ShouldEqual, fmt.Sprintf(testDataFormat, testSuccess, http.MethodDelete))
		})
	})
}
