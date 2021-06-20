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
	testAuth := newTestBasicAuthenticator("detection", "0123")
	wanted := createTestNewRequestWantArgs(trb)
	type args struct {
		RawUrl         string
		Scheme         string
		Host           string
		PathComponents []string
		Headers        *http.Header
		Body           io.Reader
		BodyJson       interface{}
		QueryParams    *url.Values
		Auth           Authenticator
		ResponseRef    interface{}
		Timeout        time.Duration
		LoggingEnabled bool
	}
	tests := []struct {
		name string
		args args
		want HttpRequest
	}{
		{
			name: "Request without Body and pathParams, but has 2 query params",
			args: args{
				Scheme:         "https",
				Host:           "picus-detection.com",
				PathComponents: []string{"assessments", "scroll"},
				Headers:        &http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}, "Xsrf-Token": []string{"CSRFToken"}, "X-Xsrf-Token": []string{"ab4f3712-1cd4-4860-9fec-1276866403da"}},
				BodyJson:       nil,
				QueryParams:    &url.Values{"tenantId": []string{"d90c3101-53bc-4c54-94db-21582bab8e17"}, "vectorId": []string{"1"}},
				ResponseRef:    testResp,
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
				Scheme:         "https",
				Host:           "picus-detection.com",
				PathComponents: []string{"assessments", "scroll", "d90c3101-53bc-4c54-94db-21582bab8e17", "1"},
				Headers:        &http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}},
				BodyJson:       trb,
				ResponseRef:    testResp,
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
				RawUrl:  fmt.Sprintf("https://picus-detection.com/assessments/scroll/d90c3101-53bc-4c54-94db-21582bab8e17/1?tenantId=%s&vectorId=%s", "d90c3101-53bc-4c54-94db-21582bab8e17", "1"),
				Headers: &http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}},
				Body: func() io.Reader {
					m, _ := json.Marshal(trb)
					return bytes.NewReader(m)
				}(),
				Auth:           testAuth,
				ResponseRef:    testResp,
				Timeout:        5 * time.Second,
				LoggingEnabled: true,
			},
			want: HttpRequest{
				request:        wanted[2],
				auth:           testAuth,
				respReference:  testResp,
				timeout:        5 * time.Second,
				loggingEnabled: true,
			},
		},
	}

	for i, tt := range tests {
		var req *HttpRequest
		var reqErr RequestError
		t.Run(tt.name, func(t *testing.T) {
			if i == 2 {
				req, reqErr = RequestBuilder().
					RawUrl(tt.args.RawUrl).
					Header(tt.args.Headers).
					Body(tt.args.Body).
					Auth(tt.args.Auth).
					ResponseReference(tt.args.ResponseRef).
					Timeout(tt.args.Timeout).
					LoggingEnabled(tt.args.LoggingEnabled).
					Build()
			} else {
				req, reqErr = RequestBuilder().
					Scheme(tt.args.Scheme).
					Host(tt.args.Host).
					PathElements(tt.args.PathComponents).
					QueryParams(tt.args.QueryParams).
					Header(tt.args.Headers).
					BodyJson(tt.args.BodyJson).
					Auth(tt.args.Auth).
					ResponseReference(tt.args.ResponseRef).
					LoggingEnabled(tt.args.LoggingEnabled).
					Build()
			}
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
				So(req.auth, ShouldResemble, tt.args.Auth)
				So(req.loggingEnabled, ShouldEqual, tt.args.LoggingEnabled)
			})
		})
	}
}

func createTestNewRequestWantArgs(trb testRequestBody) []*http.Request {
	// Construct first resulting request
	want1, err := http.NewRequest("", "https://picus-detection.com/assessments/scroll?tenantId=d90c3101-53bc-4c54-94db-21582bab8e17&vectorId=1", nil)
	if err != nil {
		log.Fatalf("Could not construct first request want argument %s", err.Error())
	}
	want1.Header.Set("Content-Type", "application/json")
	want1.Header.Set("Cookie", "test-1234")
	want1.Header.Set("Xsrf-Token", "CSRFToken")
	want1.Header.Set("X-Xsrf-Token", "ab4f3712-1cd4-4860-9fec-1276866403da")

	// Construct second resulting request
	marshalled, err := json.Marshal(trb)
	if err != nil {
		log.Fatalf("Could not marshal request body for the second request want argument %s", err.Error())
	}

	want2, err := http.NewRequest("", "https://picus-detection.com/assessments/scroll/d90c3101-53bc-4c54-94db-21582bab8e17/1", bytes.NewReader(marshalled))
	if err != nil {
		log.Fatalf("Could not construct second request want argument %s", err.Error())
	}
	want2.Header.Set("Content-Type", "application/json")
	want2.Header.Set("Cookie", "test-1234")

	// Construct third resulting request
	want3, err := http.NewRequest("", "https://picus-detection.com/assessments/scroll/d90c3101-53bc-4c54-94db-21582bab8e17/1?tenantId=d90c3101-53bc-4c54-94db-21582bab8e17&vectorId=1", bytes.NewReader(marshalled))
	if err != nil {
		log.Fatalf("Could not construct third request want argument %s", err.Error())
	}
	want3.Header.Set("Content-Type", "application/json")
	want3.Header.Set("Cookie", "test-1234")

	return []*http.Request{want1, want2, want3}
}
