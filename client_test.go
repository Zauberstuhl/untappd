package untappd_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/mdlayher/untappd"
	"github.com/nelsam/hel/pers"
)

// TestNewClient tests for all possible errors which can occur during a call
// to NewClient.
func TestNewClient(t *testing.T) {
	var tests = []struct {
		description  string
		clientID     string
		clientSecret string
		expErr       error
	}{
		{"no client ID or client secret", "", "", untappd.ErrNoClientID},
		{"no client ID", "", "bar", untappd.ErrNoClientID},
		{"no client secret", "foo", "", untappd.ErrNoClientSecret},
		{"ok", "foo", "bar", nil},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			if _, err := untappd.NewClient(tt.clientID, tt.clientSecret, nil); err != tt.expErr {
				t.Fatalf("unexpected error for test %q: %v != %v", tt.description, err, tt.expErr)
			}
		})
	}
}

// TestNewAuthenticatedClient tests for all possible errors which can occur
// during a call to NewAuthenticatedClient.
func TestNewAuthenticatedClient(t *testing.T) {
	var tests = []struct {
		description string
		accessToken string
		expErr      error
	}{
		{"no access token", "", untappd.ErrNoAccessToken},
		{"ok", "foo", nil},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			if _, err := untappd.NewAuthenticatedClient(tt.accessToken, nil); err != tt.expErr {
				t.Fatalf("unexpected error for test %q: %v != %v", tt.description, err, tt.expErr)
			}
		})
	}
}

// TestErrorError tests for consistent output from the Error.Error method.
func TestErrorError(t *testing.T) {
	var tests = []struct {
		description string
		code        int
		eType       string
		details     string
		developer   string
		result      string
	}{
		{
			description: "only details",
			code:        500,
			eType:       "auth_failed",
			details:     "authentication failed",
			developer:   "",
			result:      "500 [auth_failed]: authentication failed",
		},
		{
			description: "only developer friendly",
			code:        501,
			eType:       "auth_failed",
			details:     "",
			developer:   "authentication failed due to server error",
			result:      "501 [auth_failed]: authentication failed due to server error",
		},
		{
			description: "both details and developer friendly",
			code:        502,
			eType:       "auth_failed",
			details:     "authentication failed",
			developer:   "authentication failed due to server error",
			result:      "502 [auth_failed]: authentication failed due to server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			err := &untappd.Error{
				Code:              tt.code,
				Detail:            tt.details,
				Type:              tt.eType,
				DeveloperFriendly: tt.developer,
			}

			if res := err.Error(); res != tt.result {
				t.Fatalf("unexpected result string for test %q: %q != %q", tt.description, res, tt.result)
			}
		})
	}
}

// TestClient_requestContainsAPIKeys verifies that both client_id and client_secret
// are always present in API requests.
//func TestClient_requestContainsAPIKeys(t *testing.T) {
//	method := "GET"
//	c, done := testClient(t, func(t *testing.T, w http.ResponseWriter, r *http.Request) {
//		if m := r.Method; m != method {
//			t.Fatalf("unexpected method: %q != %q", m, method)
//		}
//
//		assertParameters(t, r, url.Values{
//			"client_id":     []string{"foo"},
//			"client_secret": []string{"bar"},
//		})
//	})
//	defer done()
//
//	if _, err := c.request(method, "foo", nil, nil, nil); err != nil {
//		t.Fatal(err)
//	}
//}
//
//// TestClient_requestPrefersAccessToken verifies that an authenticated access_token
//// is always preferred for API requests.
//func TestClient_requestPrefersAccessToken(t *testing.T) {
//	method := "GET"
//	c, done := testClient(t, func(t *testing.T, w http.ResponseWriter, r *http.Request) {
//		if m := r.Method; m != method {
//			t.Fatalf("unexpected method: %q != %q", m, method)
//		}
//
//		assertParameters(t, r, url.Values{
//			"access_token":  []string{"foo"},
//			"client_id":     []string{""},
//			"client_secret": []string{""},
//		})
//	})
//	defer done()
//
//	c.accessToken = "foo"
//	if _, err := c.request(method, "foo", nil, nil, nil); err != nil {
//		t.Fatal(err)
//	}
//}
//
//// TestClient_requestContainsRequestBody verifies that all request body items
//// are present in API requests, when HTTP method is POST
//func TestClient_requestContainsRequestBody(t *testing.T) {
//	method := "POST"
//	c, done := testClient(t, func(t *testing.T, w http.ResponseWriter, r *http.Request) {
//		if m := r.Method; m != method {
//			t.Fatalf("unexpected method: %q != %q", m, method)
//		}
//
//		k := "foo"
//		if got, want := r.PostFormValue(k), "bar"; got != want {
//			t.Fatalf("unexpected request body parameter %q: %v != %v", k, got, want)
//		}
//		k = "bar"
//		if got, want := r.PostFormValue(k), "baz"; got != want {
//			t.Fatalf("unexpected request body parameter %q: %v != %v", k, got, want)
//		}
//	})
//	defer done()
//
//	if _, err := c.request(method, "foo", url.Values{
//		"foo": []string{"bar"},
//		"bar": []string{"baz"},
//	}, nil, nil); err != nil {
//		t.Fatal(err)
//	}
//}

// TestClient_requestContainsQueryParameters verifies that all custom query
// parameters are present in API requests.
//func TestClient_requestContainsQueryParameters(t *testing.T) {
//	method := "POST"
//	c, done := testClient(t, func(t *testing.T, w http.ResponseWriter, r *http.Request) {
//		if m := r.Method; m != method {
//			t.Fatalf("unexpected method: %q != %q", m, method)
//		}
//
//		assertParameters(t, r, url.Values{
//			"foo": []string{"bar"},
//			"bar": []string{"baz"},
//		})
//
//		s, ok := r.URL.Query()["baz"]
//		if !ok {
//			t.Fatal("missing query parameter: baz")
//		}
//		for _, ss := range s {
//			if ss != "qux" && ss != "corge" {
//				t.Fatal("did not find \"qux\" or \"corge\" in key \"baz\"")
//			}
//		}
//	})
//	defer done()
//
//	if _, err := c.request(method, "foo", nil, url.Values{
//		"foo": []string{"bar"},
//		"bar": []string{"baz"},
//		"baz": []string{"qux", "corge"},
//	}, nil); err != nil {
//		t.Fatal(err)
//	}
//}
//
//// TestClient_requestContainsHeaders verifies that all typical headers are set
//// by the client during an API request.
//func TestClient_requestContainsHeaders(t *testing.T) {
//	method := "PUT"
//	c, done := testClient(t, func(t *testing.T, w http.ResponseWriter, r *http.Request) {
//		if m := r.Method; m != method {
//			t.Fatalf("unexpected method: %q != %q", m, method)
//		}
//
//		h := r.Header
//
//		if s := h.Get("Accept"); s != untappd.JSONContentType {
//			t.Fatalf("unexpected Accept header: %q != %q", s, untappd.JSONContentType)
//		}
//		if s := h.Get("User-Agent"); s != untappdUserAgent {
//			t.Fatalf("unexpected User-Agent header: %q != %q", s, untappdUserAgent)
//		}
//	})
//	defer done()
//
//	if _, err := c.request(method, "foo", nil, nil, nil); err != nil {
//		t.Fatal(err)
//	}
//}

// TestClient_requestContainsBody verifies that a response body can be
// unmarshaled from JSON following an API request.
//func TestClient_requestContainsBody(t *testing.T) {
//	method := "GET"
//	c, done := testClient(t, func(t *testing.T, w http.ResponseWriter, r *http.Request) {
//		if m := r.Method; m != method {
//			t.Fatalf("unexpected method: %q != %q", m, method)
//		}
//
//		// Use canned JSON with HTTP 500, though the HTTP code here will
//		// return 200, for processing
//		w.Write(apiErrJSON)
//	})
//	defer done()
//
//	var v struct {
//		Meta struct {
//			Code int `json:"code"`
//		} `json:"meta"`
//	}
//
//	if _, err := c.request(method, "foo", nil, nil, &v); err != nil {
//		t.Fatal(err)
//	}
//
//	if c := v.Meta.Code; c != http.StatusInternalServerError {
//		t.Fatalf("unexpected code in response body: %d != %d", c, http.StatusInternalServerError)
//	}
//}

// Test_checkResponseWrongContentType verifies that checkResponse returns an error
// when the Content-Type header does not indicate application/json.
//func Test_checkResponseWrongContentType(t *testing.T) {
//	withHTTPResponse(t, http.StatusOK, "foo/bar", nil, func(t *testing.T, res *http.Response) {
//		if err := checkResponse(res); err.Error() != "expected application/json content type, but received foo/bar" {
//			t.Fatal(err)
//		}
//	})
//}
//
//// Test_checkResponseEOF verifies that checkResponse returns an io.EOF when no
//// JSON body is found in the HTTP response body.
//func Test_checkResponseJSONEOF(t *testing.T) {
//	withHTTPResponse(t, http.StatusInternalServerError, untappd.JSONContentType, nil, func(t *testing.T, res *http.Response) {
//		if err := checkResponse(res); err != io.EOF {
//			t.Fatal(err)
//		}
//	})
//}
//
//// Test_checkResponseEOF verifies that checkResponse returns an io.ErrUnexpectedEOF
//// when a short JSON body is found in the HTTP response body.
//func Test_checkResponseJSONUnexpectedEOF(t *testing.T) {
//	withHTTPResponse(t, http.StatusInternalServerError, untappd.JSONContentType, []byte("{"), func(t *testing.T, res *http.Response) {
//		if err := checkResponse(res); err != io.ErrUnexpectedEOF {
//			t.Fatal(err)
//		}
//	})
//}
//
//// Test_checkResponseEOF verifies that checkResponse returns the appropriate error
//// assuming all sanity checks pass, but the API did return a client-consumable error.
//func Test_checkResponseErrorOK(t *testing.T) {
//	withHTTPResponse(t, http.StatusInternalServerError, untappd.JSONContentType, apiErrJSON, func(t *testing.T, res *http.Response) {
//		apiErr := &untappd.Error{
//			Code:              500,
//			Detail:            "The user has not authorized this application or the token is invalid.",
//			Type:              "invalid_auth",
//			DeveloperFriendly: "The user has not authorized this application or the token is invalid.",
//			Duration:          time.Duration(0 * time.Second),
//		}
//
//		if err := checkResponse(res); err.Error() != apiErr.Error() {
//			t.Fatalf("unexpected API error: %v != %v", err, apiErr)
//		}
//	})
//}

// Test_checkResponseEOF verifies that checkResponse returns no error when HTTP
// status is OK, but response body is empty.
//func Test_checkResponseOKNoBody(t *testing.T) {
//	withHTTPResponse(t, http.StatusOK, untappd.JSONContentType, nil, func(t *testing.T, res *http.Response) {
//		if err := checkResponse(res); err != nil {
//			t.Fatal(err)
//		}
//	})
//}
//
//// Test_checkResponseEOF verifies that checkResponse returns no error when HTTP
//// status is OK, but response body contains data.
//func Test_checkResponseOKWithBody(t *testing.T) {
//	withHTTPResponse(t, http.StatusOK, untappd.JSONContentType, []byte("{}"), func(t *testing.T, res *http.Response) {
//		if err := checkResponse(res); err != nil {
//			t.Fatal(err)
//		}
//	})
//}

// Test_FormatFloat verifies that untappd.FormatFloat produces consistent
// strings from float64 values.
func Test_FormatFloat(t *testing.T) {
	var tests = []struct {
		f float64
		s string
	}{
		{
			f: 0.0,
			s: "0",
		},
		{
			f: 1.5,
			s: "1.5",
		},
		{
			f: 2.2345,
			s: "2.2345",
		},
		{
			f: 03.456789,
			s: "3.456789",
		},
	}

	for i, tt := range tests {
		if s := untappd.FormatFloat(tt.f); s != tt.s {
			t.Fatalf("%02d: unexpected string for %f: %s != %s", i, tt.f, s, tt.s)
		}
	}
}

// withHTTPResponse is a test helper which generates a *http.Response and invokes
// an input closure, used for testing.
func withHTTPResponse(t *testing.T, code int, contentType string, body []byte, fn func(t *testing.T, res *http.Response)) {
	res := &http.Response{
		StatusCode: code,
		Header: http.Header{
			"Content-Type": []string{contentType},
		},
		Body: ioutil.NopCloser(bytes.NewReader(body)),
	}

	fn(t, res)
}

// testClient wires up a new Client with a HTTP test server, allowing for easy
// setup and teardown of repetitive code.  The input closure is invoked in the
// HTTP server, to change the functionality as needed for each test.
func testClient(t *testing.T, fn func(t *testing.T, w http.ResponseWriter, r *http.Request)) (*untappd.Client, func()) {
	t.Helper()
	hClient, done := newMockHTTPClientHandler(t, fn)

	client, err := untappd.NewClient("foo", "bar", hClient)
	if err != nil {
		t.Fatal(err)
	}

	return client, done
}

func newMockHTTPClientHandler(t *testing.T, handle func(t *testing.T, w http.ResponseWriter, r *http.Request)) (*mockHTTPClient, func()) {
	t.Helper()

	if handle == nil {
		t.Fatal("A handler function must be passed in")
	}
	hClient := newMockHTTPClient()
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-hClient.GetCalled:
				// TODO: clean up the tests so that they're able to handle request/response directly,
				// rather than as an http handler.
				rec := httptest.NewRecorder()
				handle(t, rec, httptest.NewRequest("GET", <-hClient.GetInput.Url, nil))
				rec.HeaderMap.Set("Content-Type", untappd.JSONContentType)
				pers.Return(hClient.GetOutput, &http.Response{
					StatusCode: rec.Code,
					Header:     rec.HeaderMap,
					Body:       ioutil.NopCloser(rec.Body),
				}, nil)
			case <-hClient.DoCalled:
				// TODO: same as above wrt cleaning up the tests.
				rec := httptest.NewRecorder()
				req := <-hClient.DoInput.Req
				// This is a client side request.  We need to adjust it to look more like a server side
				// request.
				if req.Method == "" {
					req.Method = "GET"
				}
				req.URL = &url.URL{
					Path:     req.URL.Path,
					RawQuery: req.URL.RawQuery,
				}
				if host := req.Header.Get("Host"); host != "" {
					delete(req.Header, "Host")
					req.Host = host
				}
				if req.Body == nil {
					req.Body = ioutil.NopCloser(bytes.NewBufferString(""))
				}

				handle(t, rec, req)
				rec.HeaderMap.Set("Content-Type", untappd.JSONContentType)
				pers.Return(hClient.DoOutput, &http.Response{
					StatusCode: rec.Code,
					Header:     rec.HeaderMap,
					Body:       ioutil.NopCloser(rec.Body),
				}, nil)
			case <-done:
				return
			}
		}
	}()
	return hClient, func() { close(done) }
}

// assertParameters asserts that query parameters from an HTTP request
// match an expected set of query parameter values.
func assertParameters(t *testing.T, r *http.Request, expected url.Values) {
	q := r.URL.Query()

	for k := range expected {
		if actual, expected := q.Get(k), expected.Get(k); actual != expected {
			t.Fatalf("unexpected parameter %q: %v != %v", k, actual, expected)
		}
	}
}

// assertBodyParameters asserts that body parameters from an HTTP request
// match an expected set of query parameter values.
func assertBodyParameters(t *testing.T, r *http.Request, expected url.Values) {
	for k := range expected {
		if actual, expected := r.PostFormValue(k), expected.Get(k); actual != expected {
			t.Fatalf("unexpected parameter %q: %v != %v", k, actual, expected)
		}
	}
}

// assertInvalidUserErr asserts that an input error was generated from the
// invalidUserErrJSON used in some tests.
func assertInvalidUserErr(t *testing.T, err error) {
	uErr := assertInvalidCommonErr(t, err)

	detail := "There is no user with that username."
	if d := uErr.Detail; d != detail {
		t.Fatalf("unexpected error detail: %q != %q", d, detail)
	}
	eType := "invalid_auth"
	if e := uErr.Type; e != eType {
		t.Fatalf("unexpected error type: %q != %q", e, eType)
	}
}

// assertInvalidBeerErr asserts that an input error was generated from the
// invalidBeerErrJSON used in some tests.
func assertInvalidBeerErr(t *testing.T, err error) {
	uErr := assertInvalidCommonErr(t, err)

	detail := "This Beer ID is invalid."
	if d := uErr.Detail; d != detail {
		t.Fatalf("unexpected error detail: %q != %q", d, detail)
	}
	eType := "invalid_param"
	if e := uErr.Type; e != eType {
		t.Fatalf("unexpected error type: %q != %q", e, eType)
	}
}

// assertInvalidBreweryErr asserts that an input error was generated from the
// invalidBreweryErrJSON used in some tests.
func assertInvalidBreweryErr(t *testing.T, err error) {
	uErr := assertInvalidCommonErr(t, err)

	detail := "This Brewery ID is invalid."
	if d := uErr.Detail; d != detail {
		t.Fatalf("unexpected error detail: %q != %q", d, detail)
	}
	eType := "invalid_param"
	if e := uErr.Type; e != eType {
		t.Fatalf("unexpected error type: %q != %q", e, eType)
	}
}

// assertInvalidVenueErr asserts that an input error was generated from the
// invalidVenueErrJSON used in some tests.
func assertInvalidVenueErr(t *testing.T, err error) {
	uErr := assertInvalidCommonErr(t, err)

	detail := "This Venue ID is invalid."
	if d := uErr.Detail; d != detail {
		t.Fatalf("unexpected error detail: %q != %q", d, detail)
	}
	eType := "invalid_param"
	if e := uErr.Type; e != eType {
		t.Fatalf("unexpected error type: %q != %q", e, eType)
	}
}

// assertInvalidLocalErr asserts that an input error was generated from the
// invalidLocalErrJSON used in some tests.
func assertInvalidLocalErr(t *testing.T, err error) {
	uErr := assertInvalidCommonErr(t, err)

	detail := "Your missing the 'lat' or 'lng' parameter."
	if d := uErr.Detail; d != detail {
		t.Fatalf("unexpected error detail: %q != %q", d, detail)
	}
	eType := "invalid_param"
	if e := uErr.Type; e != eType {
		t.Fatalf("unexpected error type: %q != %q", e, eType)
	}
}

// assertInvalidQueryErr asserts that an input error was generated from the
// invalidQueryErrJSON used in some tests.
func assertInvalidQueryErr(t *testing.T, err error) {
	uErr := assertInvalidCommonErr(t, err)

	detail := "Your missing the 'q' parameter."
	if d := uErr.Detail; d != detail {
		t.Fatalf("unexpected error detail: %q != %q", d, detail)
	}
	eType := "invalid_param"
	if e := uErr.Type; e != eType {
		t.Fatalf("unexpected error type: %q != %q", e, eType)
	}
}

// assertInvalidCheckinErr asserts that an input error was generated from the
// invalidCheckinErrJSON used in some tests.
func assertInvalidCheckinErr(t *testing.T, err error) {
	uErr := assertInvalidCommonErr(t, err)

	detail := "The bid field is required."
	if d := uErr.Detail; d != detail {
		t.Fatalf("unexpected error detail: %q != %q", d, detail)
	}
	eType := "invalid_param"
	if e := uErr.Type; e != eType {
		t.Fatalf("unexpected error type: %q != %q", e, eType)
	}
}

// assertInvalidCommonErr removes some redundant logic from other assert
// test helpers.
func assertInvalidCommonErr(t *testing.T, err error) *untappd.Error {
	if err == nil {
		t.Fatal("error should have occurred, but error is nil")
	}

	uErr, ok := err.(*untappd.Error)
	if !ok {
		t.Fatal("error is not of type *Error")
	}

	if c := uErr.Code; c != http.StatusInternalServerError {
		t.Fatalf("unexpected error code: %d != %d", c, http.StatusNotFound)
	}

	return uErr
}

// JSON taken from Untappd APIv4 documentation: https://untappd.com/api/docs
var apiErrJSON = []byte(`{
  "meta": {
    "code": 500,
    "error_detail": "The user has not authorized this application or the token is invalid.",
    "error_type": "invalid_auth",
    "developer_friendly": "The user has not authorized this application or the token is invalid.",
    "response_time": {
      "time": 0,
      "measure": "seconds"
    }
  }
}`)

// invalidUserErrJSON is canned JSON used to test for invalid user handling
var invalidUserErrJSON = []byte(`{"meta":{"code":500,"error_detail":"There is no user with that username.","error_type":"invalid_auth","response_time":{"time":0,"measure":"seconds"}}}`)

// invalidBeerErrJSON is canned JSON used to test for invalid beer handling
var invalidBeerErrJSON = []byte(`{"meta":{"code":500,"error_detail":"This Beer ID is invalid.","error_type":"invalid_param","response_time":{"time":0,"measure":"seconds"}}}`)

// invalidBreweryErrJSON is canned JSON used to test for invalid brewery handling
var invalidBreweryErrJSON = []byte(`{"meta":{"code":500,"error_detail":"This Brewery ID is invalid.","error_type":"invalid_param","response_time":{"time":0,"measure":"seconds"}}}`)

// invalidVenueErrJSON is canned JSON used to test for invalid venue handling
var invalidVenueErrJSON = []byte(`{"meta":{"code":500,"error_detail":"This Venue ID is invalid.","error_type":"invalid_param","response_time":{"time":0,"measure":"seconds"}}}`)

// invalidLocalErrJSON is canned JSON used to test for invalid query handling
var invalidLocalErrJSON = []byte(`{"meta":{"code":500,"error_detail":"Your missing the 'lat' or 'lng' parameter.","error_type":"invalid_param","response_time":{"time":0,"measure":"seconds"}}}`)

// invalidQueryErrJSON is canned JSON used to test for invalid query handling
var invalidQueryErrJSON = []byte(`{"meta":{"code":500,"error_detail":"Your missing the 'q' parameter.","error_type":"invalid_param","response_time":{"time":0,"measure":"seconds"}}}`)

// invalidCheckinErrJSON is canned JSON used to test for an invalid checkin attempt
var invalidCheckinErrJSON = []byte(`{"meta":{"code":500,"error_detail":"The bid field is required.","error_type":"invalid_param","developer_friendly":"","response_time":{"time":0.036,"measure":"seconds"}},"response":[]}`)
