package restclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestHttpRequestBuilder(t *testing.T) {
	var testResp testHttpResponse
	trb := testRequestBody{
		TestId:   123,
		TestName: "1234",
	}
	testAuth := newTestBasicAuthenticator("username", "0123")
	wanted := createTestNewRequestWantArgs(trb)

	type args struct {
		rawUrl         string
		scheme         string
		host           string
		pathComponents []string
		headers        *http.Header
		body           io.Reader
		bodyJson       interface{}
		queryParams    *url.Values
		auth           Authenticator
		responseRef    interface{}
		request        *http.Request
		timeout        time.Duration
		loggingEnabled bool
	}
	tests := []struct {
		name string
		args args
		want HttpRequest
	}{
		{
			name: "Request without Body and pathParams, but has 2 query params",
			args: args{
				scheme:         "https",
				host:           "ysyesilyurt.com",
				pathComponents: []string{"assessments", "scroll"},
				headers:        &http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}, "Xsrf-Token": []string{"CSRFToken"}, "X-Xsrf-Token": []string{"ab4f3712-1cd4-4860-9fec-1276866403da"}},
				bodyJson:       nil,
				queryParams:    &url.Values{"tenantId": []string{"d90c3101-53bc-4c54-94db-21582bab8e17"}, "vectorId": []string{"1"}},
				responseRef:    testResp,
			},
			want: HttpRequest{
				request:        wanted[0],
				auth:           nil,
				respReference:  testResp,
				timeout:        defaultTimeoutDuration,
				loggingEnabled: false,
			},
		},
		{
			name: "Request with Json Body and pathParams, but has no query params",
			args: args{
				scheme:         "https",
				host:           "ysyesilyurt.com",
				pathComponents: []string{"assessments", "scroll", "d90c3101-53bc-4c54-94db-21582bab8e17", "1"},
				headers:        &http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}},
				bodyJson:       trb,
				responseRef:    testResp,
			},
			want: HttpRequest{
				request:        wanted[1],
				auth:           nil,
				respReference:  testResp,
				timeout:        defaultTimeoutDuration,
				loggingEnabled: false,
			},
		},
		{
			name: "Request that is constructed with RawUrl, Reader Body and has other fields",
			args: args{
				rawUrl:  fmt.Sprintf("https://ysyesilyurt.com/assessments/scroll/d90c3101-53bc-4c54-94db-21582bab8e17/1?tenantId=%s&vectorId=%s", "d90c3101-53bc-4c54-94db-21582bab8e17", "1"),
				headers: &http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}},
				body: func() io.Reader {
					m, _ := json.Marshal(trb)
					return bytes.NewReader(m)
				}(),
				auth:           testAuth,
				responseRef:    testResp,
				timeout:        5 * time.Second,
				loggingEnabled: true,
			},
			want: HttpRequest{
				request:        wanted[2],
				auth:           testAuth,
				respReference:  testResp,
				timeout:        5 * time.Second,
				loggingEnabled: true,
			},
		},
		{
			name: "Request that is constructed with another http.Request",
			args: args{
				headers:     &http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}},
				auth:        testAuth,
				responseRef: testResp,
				request: func() *http.Request {
					req, _ := http.NewRequest("", "https://ysyesilyurt.com/assessments/scroll?builtWithAnotherRequest=true", nil)
					req.Header = http.Header{"Header-1": []string{"this header belongs to pre-prepared request"}}
					return req
				}(),
				timeout:        6 * time.Second,
				loggingEnabled: true,
			},
			want: HttpRequest{
				request:        wanted[3],
				auth:           testAuth,
				respReference:  testResp,
				timeout:        6 * time.Second,
				loggingEnabled: true,
			},
		},
	}

	for i, tt := range tests {
		var builder HttpRequestBuilder
		t.Run(tt.name, func(t *testing.T) {
			if i == 2 {
				builder = RequestBuilder().
					RawUrl(tt.args.rawUrl).
					Body(tt.args.body).
					Timeout(tt.args.timeout)
			} else if i == 3 {
				builder = RequestBuilder().
					Request(tt.args.request).
					Timeout(tt.args.timeout)
			} else {
				builder = RequestBuilder().
					Scheme(tt.args.scheme).
					Host(tt.args.host).
					PathElements(tt.args.pathComponents).
					QueryParams(tt.args.queryParams).
					BodyJson(tt.args.bodyJson)
			}

			req, reqErr := builder.
				Header(tt.args.headers).
				Auth(tt.args.auth).
				ResponseReference(tt.args.responseRef).
				LoggingEnabled(tt.args.loggingEnabled).
				Build()

			Convey("Requests should be constructed without err and with proper fields", t, func() {
				So(reqErr, ShouldBeNil)
				So(req.request.URL.Scheme, ShouldEqual, tt.want.request.URL.Scheme)
				So(req.request.URL.Host, ShouldEqual, tt.want.request.URL.Host)
				So(req.request.URL.Path, ShouldEqual, tt.want.request.URL.Path)
				So(req.request.Header, ShouldResemble, tt.want.request.Header)
				So(req.request.Body, ShouldResemble, tt.want.request.Body)
				So(req.request.URL.RawQuery, ShouldEqual, tt.want.request.URL.RawQuery)
				So(req.respReference, ShouldResemble, tt.want.respReference)
				So(req.timeout, ShouldEqual, tt.want.timeout)
				So(req.auth, ShouldResemble, tt.args.auth)
				So(req.loggingEnabled, ShouldEqual, tt.args.loggingEnabled)
			})
		})
	}
}

func createTestNewRequestWantArgs(trb testRequestBody) []*http.Request {
	// Construct first resulting request
	want1, err := http.NewRequest("", "https://ysyesilyurt.com/assessments/scroll?tenantId=d90c3101-53bc-4c54-94db-21582bab8e17&vectorId=1", nil)
	if err != nil {
		log.Fatalf("Could not construct first request want argument %s", err.Error())
	}
	want1.Header.Add("Content-Type", "application/json")
	want1.Header.Add("Cookie", "test-1234")
	want1.Header.Add("Xsrf-Token", "CSRFToken")
	want1.Header.Add("X-Xsrf-Token", "ab4f3712-1cd4-4860-9fec-1276866403da")

	// Construct second resulting request
	marshalled, err := json.Marshal(trb)
	if err != nil {
		log.Fatalf("Could not marshal request body for the second request want argument %s", err.Error())
	}

	want2, err := http.NewRequest("", "https://ysyesilyurt.com/assessments/scroll/d90c3101-53bc-4c54-94db-21582bab8e17/1", bytes.NewReader(marshalled))
	if err != nil {
		log.Fatalf("Could not construct second request want argument %s", err.Error())
	}
	want2.Header.Add("Content-Type", "application/json")
	want2.Header.Add("Cookie", "test-1234")

	// Construct third resulting request
	want3, err := http.NewRequest("", "https://ysyesilyurt.com/assessments/scroll/d90c3101-53bc-4c54-94db-21582bab8e17/1?tenantId=d90c3101-53bc-4c54-94db-21582bab8e17&vectorId=1", bytes.NewReader(marshalled))
	if err != nil {
		log.Fatalf("Could not construct third request want argument %s", err.Error())
	}
	want3.Header.Add("Content-Type", "application/json")
	want3.Header.Add("Cookie", "test-1234")

	// Construct third resulting request
	want4, err := http.NewRequest("", "https://ysyesilyurt.com/assessments/scroll?builtWithAnotherRequest=true", nil)
	if err != nil {
		log.Fatalf("Could not construct fourth request want argument %s", err.Error())
	}
	want4.Header.Add("Content-Type", "application/json")
	want4.Header.Add("Cookie", "test-1234")
	want4.Header.Add("Header-1", "this header belongs to pre-prepared request")

	return []*http.Request{want1, want2, want3, want4}
}
