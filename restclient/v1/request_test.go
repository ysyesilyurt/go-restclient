package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"log"
	"net/http"
	"net/url"
	"testing"
)

type testBodyDto struct {
	TestId     int    `json:"test_id"`
	TestString string `json:"test_string"`
}

func TestNewRequest(t *testing.T) {
	tbd := testBodyDto{
		TestId:     123,
		TestString: "1234",
	}
	wanted := createTestNewRequestWantArgs(tbd)
	type args struct {
		Scheme         string
		Host           string
		PathComponents []string
		Headers        *http.Header
		Body           interface{}
		QueryParams    *url.Values
	}
	tests := []struct {
		name string
		args args
		want *http.Request
	}{
		{
			name: "Request without Body and pathParams, but has 2 query params",
			args: args{
				Scheme:         "https",
				Host:           "ysyesilyurt.com",
				PathComponents: []string{"assessments", "scroll"},
				Headers:        &http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234", "1234-test"}, "Xsrf-Token": []string{"CSRFToken"}, "X-Xsrf-Token": []string{"ab4f3712-1cd4-4860-9fec-1276866403da"}},
				Body:           nil,
				QueryParams:    &url.Values{"tenantId": []string{"d90c3101-53bc-4c54-94db-21582bab8e17"}, "vectorId": []string{"1"}},
			},
			want: wanted[0],
		},
		{
			name: "Request with Body and pathParams, but has no query params",
			args: args{
				Scheme:         "https",
				Host:           "ysyesilyurt.com",
				PathComponents: []string{"assessments", "scroll", "d90c3101-53bc-4c54-94db-21582bab8e17", "1"},
				Headers:        &http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234", "1234-test"}},
				Body:           tbd,
				QueryParams:    nil,
			},
			want: wanted[1],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ri := NewRequestInfo(tt.args.Scheme, tt.args.Host, tt.args.PathComponents, tt.args.QueryParams, tt.args.Headers, tt.args.Body)
			req, err := NewRequest(ri)
			Convey("Requests should be constructed without err and with proper fields", t, func() {
				So(err, ShouldBeNil)
				So(req.URL.Scheme, ShouldEqual, tt.want.URL.Scheme)
				So(req.URL.Host, ShouldEqual, tt.want.URL.Host)
				So(req.URL.Path, ShouldEqual, tt.want.URL.Path)
				So(req.Header, ShouldResemble, tt.want.Header)
				So(req.Body, ShouldResemble, tt.want.Body)
				So(req.URL.RawQuery, ShouldEqual, tt.want.URL.RawQuery)
			})
		})
	}
}

func createTestNewRequestWantArgs(tbd testBodyDto) []*http.Request {
	// Construct first resulting request
	want1, err := http.NewRequest("", "https://ysyesilyurt.com/assessments/scroll?tenantId=d90c3101-53bc-4c54-94db-21582bab8e17&vectorId=1", nil)
	if err != nil {
		log.Fatalf("Could not construct first request want argument %s", err.Error())
	}
	want1.Header.Set("Content-Type", "application/json")
	want1.Header.Set("Cookie", "test-1234")
	want1.Header.Add("Cookie", "1234-test")
	want1.Header.Set("Xsrf-Token", "CSRFToken")
	want1.Header.Set("X-Xsrf-Token", "ab4f3712-1cd4-4860-9fec-1276866403da")

	// Construct second resulting request
	marshalled, err := json.Marshal(tbd)
	if err != nil {
		log.Fatalf("Could not marshal request body for the second request want argument %s", err.Error())
	}
	want2, err := http.NewRequest("", "https://ysyesilyurt.com/assessments/scroll/d90c3101-53bc-4c54-94db-21582bab8e17/1", bytes.NewReader(marshalled))
	if err != nil {
		log.Fatalf("Could not construct second request want argument %s", err.Error())
	}
	want2.Header.Set("Content-Type", "application/json")
	want2.Header.Set("Cookie", "test-1234")
	want2.Header.Add("Cookie", "1234-test")
	return []*http.Request{want1, want2}
}

func TestNewRequestInfoFromURL(t *testing.T) {
	Convey("TEST Parse RequestInfo From Raw URL", t, func() {
		tbd := testBodyDto{
			TestId:     123,
			TestString: "1234",
		}
		headers := http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}}
		queryParams := url.Values{"tenantId": []string{"d90c3101-53bc-4c54-94db-21582bab8e17"}, "vectorId": []string{"1"}}
		wantRi := NewRequestInfo("https", "ysyesilyurt.com", []string{"assessments", "scroll", "d90c3101-53bc-4c54-94db-21582bab8e17", "1"}, &queryParams, &headers, tbd)
		parsedRi, err := NewRequestInfoFromRawURL(fmt.Sprintf("https://ysyesilyurt.com/assessments/scroll/d90c3101-53bc-4c54-94db-21582bab8e17/1?tenantId=%s&vectorId=%s", "d90c3101-53bc-4c54-94db-21582bab8e17", "1"), &headers, tbd)
		Convey("Parsed request info should be identical to correct request info", func() {
			So(err, ShouldBeNil)
			So(parsedRi.Scheme, ShouldEqual, wantRi.Scheme)
			So(parsedRi.Host, ShouldEqual, wantRi.Host)
			So(parsedRi.PathElements, ShouldResemble, wantRi.PathElements)
			So(parsedRi.QueryParams, ShouldResemble, wantRi.QueryParams)
			So(parsedRi.Headers, ShouldResemble, wantRi.Headers)
			So(parsedRi.Body, ShouldResemble, wantRi.Body)
		})
	})
}
