package v1

import (
	"encoding/json"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/ysyesilyurt/go-restclient/restclient"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	TestDataFormat = "%s - %s"
	TestSuccess    = "TestSuccess"
	TestFailed     = "TestFailed"
)

type testBasicAuthenticator struct {
	Username, Password string
}

func NewTestBasicAuthenticator(username, password string) restclient.Authenticator {
	return &testBasicAuthenticator{
		Username: username,
		Password: password,
	}
}

func (b testBasicAuthenticator) Apply(request *http.Request) error {
	request.SetBasicAuth(b.Username, b.Password)
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
			if !ok || username != "ysyesilyurt" || password != "0123" {
				statusCode = http.StatusUnauthorized
				resultString = TestFailed
			}
			response.StatusCode = statusCode
			response.Data = fmt.Sprintf(TestDataFormat, resultString, reqMethod)
			w.WriteHeader(statusCode)
		}

		handleRequestWithBody := func(reqMethod string) {
			status := http.StatusOK
			resultString := TestSuccess
			var body testRequestBody
			err := UnmarshalRequestBody(r, &body)
			if err != nil || body.TestName != "Testing Request Body" || body.TestId != 123 {
				status = http.StatusBadRequest
				resultString = TestFailed
			}
			handleRequest(reqMethod, status, resultString)
		}

		switch r.Method {
		case http.MethodGet:
			handleRequest(http.MethodGet, http.StatusOK, TestSuccess)
		case http.MethodPost:
			handleRequestWithBody(http.MethodPost)
		case http.MethodPut:
			handleRequestWithBody(http.MethodPut)
		case http.MethodPatch:
			handleRequestWithBody(http.MethodPatch)
		case http.MethodDelete:
			handleRequest(http.MethodDelete, http.StatusOK, TestSuccess)
		default:
			response.StatusCode = http.StatusForbidden
			response.Data = fmt.Sprintf(TestDataFormat, TestFailed, r.Method)
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
		testRequestInfo := NewRequestInfo(testServerScheme, testServerHost, nil, nil, nil, nil)
		testRequest, err := NewRequest(testRequestInfo)
		if err != nil {
			log.Fatal("failed to construct testRequest")
		}
		client := NewHttpClient(true, 30*time.Second)
		auth := NewTestBasicAuthenticator("ysyesilyurt", "0123")
		cri := NewDoRequestInfo(testRequest, auth, &testResponse)
		err = client.Get(cri)
		Convey("Response should be fetched without any err and with proper Data", func() {
			data, ok := testResponse.Data.(string)
			So(err, ShouldBeNil)
			So(testResponse.StatusCode, ShouldEqual, http.StatusOK)
			So(ok, ShouldBeTrue)
			So(data, ShouldEqual, fmt.Sprintf(TestDataFormat, TestSuccess, http.MethodGet))
		})
	})

	Convey("TEST HTTP GET with failing authentication credentials", t, func() {
		var testResponse testHttpResponse
		testRequestInfo := NewRequestInfo(testServerScheme, testServerHost, nil, nil, nil, nil)
		testRequest, err := NewRequest(testRequestInfo)
		if err != nil {
			log.Fatal("failed to construct testRequest")
		}
		client := NewHttpClient(true, 30*time.Second)
		auth := NewTestBasicAuthenticator("FAILING", "CREDENTIALS")
		cri := NewDoRequestInfo(testRequest, auth, &testResponse)
		err = client.Get(cri)
		Convey("Response should be fetched without any err and with proper Data", func() {
			_, ok := testResponse.Data.(string)
			So(ok, ShouldBeFalse)
			So(testResponse.StatusCode, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, fmt.Sprintf("{\"status_code\":%d,\"data\":\"%s\"}\n: %s", http.StatusUnauthorized, fmt.Sprintf(TestDataFormat, TestFailed, http.MethodGet), unauthorizedErr))
		})
	})

	Convey("TEST HTTP POST with correct body", t, func() {
		var testResponse testHttpResponse
		testBody := testRequestBody{
			TestId:   123,
			TestName: "Testing Request Body",
		}
		testRequestInfo := NewRequestInfo(testServerScheme, testServerHost, nil, nil, nil, testBody)
		testRequest, err := NewRequest(testRequestInfo)
		if err != nil {
			log.Fatal("failed to construct testRequest")
		}
		client := NewHttpClient(true, 30*time.Second)
		auth := NewTestBasicAuthenticator("ysyesilyurt", "0123")
		cri := NewDoRequestInfo(testRequest, auth, &testResponse)
		err = client.Post(cri)
		Convey("Response should be fetched without any err and with proper Data", func() {
			data, ok := testResponse.Data.(string)
			So(err, ShouldBeNil)
			So(testResponse.StatusCode, ShouldEqual, http.StatusOK)
			So(ok, ShouldBeTrue)
			So(data, ShouldEqual, fmt.Sprintf(TestDataFormat, TestSuccess, http.MethodPost))
		})
	})

	Convey("TEST HTTP POST with incorrect body", t, func() {
		var testResponse testHttpResponse
		testBody := "INCORRECT REQUEST BODY"
		testRequestInfo := NewRequestInfo(testServerScheme, testServerHost, nil, nil, nil, testBody)
		testRequest, err := NewRequest(testRequestInfo)
		if err != nil {
			log.Fatal("failed to construct testRequest")
		}
		client := NewHttpClient(true, 30*time.Second)
		auth := NewTestBasicAuthenticator("ysyesilyurt", "0123")
		cri := NewDoRequestInfo(testRequest, auth, &testResponse)
		err = client.Post(cri)
		Convey("Response should be fetched without any err and with proper Data", func() {
			_, ok := testResponse.Data.(string)
			So(ok, ShouldBeFalse)
			So(testResponse.StatusCode, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, fmt.Sprintf("{\"status_code\":%d,\"data\":\"%s\"}\n: %s", http.StatusBadRequest, fmt.Sprintf(TestDataFormat, TestFailed, http.MethodPost), badRequestErr))
		})
	})

	Convey("TEST HTTP PUT with correct body", t, func() {
		var testResponse testHttpResponse
		testBody := testRequestBody{
			TestId:   123,
			TestName: "Testing Request Body",
		}
		testRequestInfo := NewRequestInfo(testServerScheme, testServerHost, nil, nil, nil, testBody)
		testRequest, err := NewRequest(testRequestInfo)
		if err != nil {
			log.Fatal("failed to construct testRequest")
		}
		client := NewHttpClient(true, 30*time.Second)
		auth := NewTestBasicAuthenticator("ysyesilyurt", "0123")
		cri := NewDoRequestInfo(testRequest, auth, &testResponse)
		err = client.Put(cri)
		Convey("Response should be fetched without any err and with proper Data", func() {
			data, ok := testResponse.Data.(string)
			So(err, ShouldBeNil)
			So(testResponse.StatusCode, ShouldEqual, http.StatusOK)
			So(ok, ShouldBeTrue)
			So(data, ShouldEqual, fmt.Sprintf(TestDataFormat, TestSuccess, http.MethodPut))
		})
	})

	Convey("TEST HTTP PATCH with correct body", t, func() {
		var testResponse testHttpResponse
		testBody := testRequestBody{
			TestId:   123,
			TestName: "Testing Request Body",
		}
		testRequestInfo := NewRequestInfo(testServerScheme, testServerHost, nil, nil, nil, testBody)
		testRequest, err := NewRequest(testRequestInfo)
		if err != nil {
			log.Fatal("failed to construct testRequest")
		}
		client := NewHttpClient(true, 30*time.Second)
		auth := NewTestBasicAuthenticator("ysyesilyurt", "0123")
		cri := NewDoRequestInfo(testRequest, auth, &testResponse)
		err = client.Patch(cri)
		Convey("Response should be fetched without any err and with proper Data", func() {
			data, ok := testResponse.Data.(string)
			So(err, ShouldBeNil)
			So(testResponse.StatusCode, ShouldEqual, http.StatusOK)
			So(ok, ShouldBeTrue)
			So(data, ShouldEqual, fmt.Sprintf(TestDataFormat, TestSuccess, http.MethodPatch))
		})
	})

	Convey("TEST HTTP DELETE", t, func() {
		var testResponse testHttpResponse
		testRequestInfo := NewRequestInfo(testServerScheme, testServerHost, nil, nil, nil, nil)
		testRequest, err := NewRequest(testRequestInfo)
		if err != nil {
			log.Fatal("failed to construct testRequest")
		}
		client := NewHttpClient(true, 30*time.Second)
		auth := NewTestBasicAuthenticator("ysyesilyurt", "0123")
		cri := NewDoRequestInfo(testRequest, auth, &testResponse)
		err = client.Delete(cri)
		Convey("Response should be fetched without any err and with proper Data", func() {
			data, ok := testResponse.Data.(string)
			So(err, ShouldBeNil)
			So(testResponse.StatusCode, ShouldEqual, http.StatusOK)
			So(ok, ShouldBeTrue)
			So(data, ShouldEqual, fmt.Sprintf(TestDataFormat, TestSuccess, http.MethodDelete))
		})
	})
}
