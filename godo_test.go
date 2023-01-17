package godo

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

var (
	mux *http.ServeMux

	ctx = context.TODO()

	client *Client

	server *httptest.Server
)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client = NewClient(nil)
	url, _ := url.Parse(server.URL)
	client.BaseURL = url
}

func teardown() {
	server.Close()
}

func testMethod(t *testing.T, r *http.Request, expected string) {
	if expected != r.Method {
		t.Errorf("Request method = %v, expected %v", r.Method, expected)
	}
}

type values map[string]string

func testFormValues(t *testing.T, r *http.Request, values values) {
	expected := url.Values{}
	for k, v := range values {
		expected.Add(k, v)
	}

	err := r.ParseForm()
	if err != nil {
		t.Fatalf("parseForm(): %v", err)
	}

	if !reflect.DeepEqual(expected, r.Form) {
		t.Errorf("Request parameters = %v, expected %v", r.Form, expected)
	}
}

func testURLParseError(t *testing.T, err error) {
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
	if err, ok := err.(*url.Error); !ok || err.Op != "parse" {
		t.Errorf("Expected URL parse error, got %+v", err)
	}
}

func testClientServices(t *testing.T, c *Client) {
	services := []string{
		"Account",
		"Actions",
		"Balance",
		"BillingHistory",
		"CDNs",
		"Domains",
		"Droplets",
		"DropletActions",
		"Images",
		"ImageActions",
		"Invoices",
		"Keys",
		"Monitoring",
		"Regions",
		"Sizes",
		"FloatingIPs",
		"FloatingIPActions",
		"ReservedIPs",
		"ReservedIPActions",
		"Tags",
	}

	cp := reflect.ValueOf(c)
	cv := reflect.Indirect(cp)

	for _, s := range services {
		if cv.FieldByName(s).IsNil() {
			t.Errorf("c.%s shouldn't be nil", s)
		}
	}
}

func testClientDefaultBaseURL(t *testing.T, c *Client) {
	if c.BaseURL == nil || c.BaseURL.String() != defaultBaseURL {
		t.Errorf("NewClient BaseURL = %v, expected %v", c.BaseURL, defaultBaseURL)
	}
}

func testClientDefaultUserAgent(t *testing.T, c *Client) {
	if c.UserAgent != userAgent {
		t.Errorf("NewClient UserAgent = %v, expected %v", c.UserAgent, userAgent)
	}
}

func testClientDefaults(t *testing.T, c *Client) {
	testClientDefaultBaseURL(t, c)
	testClientDefaultUserAgent(t, c)
	testClientServices(t, c)
}

func TestNewClient(t *testing.T) {
	c := NewClient(nil)
	testClientDefaults(t, c)
}

func TestNewFromToken(t *testing.T) {
	c := NewFromToken("myToken")
	testClientDefaults(t, c)
}

func TestNewFromToken_cleaned(t *testing.T) {
	testTokens := []string{"myToken ", " myToken", " myToken ", "'myToken'", " 'myToken' "}
	expected := "Bearer myToken"

	setup()
	defer teardown()

	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for _, tt := range testTokens {
		t.Run(tt, func(t *testing.T) {
			c := NewFromToken(tt)
			req, _ := c.NewRequest(ctx, http.MethodGet, server.URL+"/foo", nil)
			resp, err := c.Do(ctx, req, nil)
			if err != nil {
				t.Fatalf("Do(): %v", err)
			}

			authHeader := resp.Request.Header.Get("Authorization")
			if authHeader != expected {
				t.Errorf("Authorization header = %v, expected %v", authHeader, expected)
			}
		})
	}
}

func TestNew(t *testing.T) {
	c, err := New(nil)

	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	testClientDefaults(t, c)
}

func TestNewRequest(t *testing.T) {
	c := NewClient(nil)

	inURL, outURL := "/foo", defaultBaseURL+"foo"
	inBody, outBody := &DropletCreateRequest{Name: "l"},
		`{"name":"l","region":"","size":"","image":0,`+
			`"ssh_keys":null,"backups":false,"ipv6":false,`+
			`"private_networking":false,"monitoring":false,"tags":null}`+"\n"
	req, _ := c.NewRequest(ctx, http.MethodPost, inURL, inBody)

	// test relative URL was expanded
	if req.URL.String() != outURL {
		t.Errorf("NewRequest(%v) URL = %v, expected %v", inURL, req.URL, outURL)
	}

	// test body was JSON encoded
	body, _ := ioutil.ReadAll(req.Body)
	if string(body) != outBody {
		t.Errorf("NewRequest(%v)Body = %v, expected %v", inBody, string(body), outBody)
	}

	// test default user-agent is attached to the request
	userAgent := req.Header.Get("User-Agent")
	if c.UserAgent != userAgent {
		t.Errorf("NewRequest() User-Agent = %v, expected %v", userAgent, c.UserAgent)
	}
}

func TestNewRequest_get(t *testing.T) {
	c := NewClient(nil)

	inURL, outURL := "/foo", defaultBaseURL+"foo"
	req, _ := c.NewRequest(ctx, http.MethodGet, inURL, nil)

	// test relative URL was expanded
	if req.URL.String() != outURL {
		t.Errorf("NewRequest(%v) URL = %v, expected %v", inURL, req.URL, outURL)
	}

	// test the content-type header is not set
	if contentType := req.Header.Get("Content-Type"); contentType != "" {
		t.Errorf("NewRequest() Content-Type = %v, expected empty string", contentType)
	}

	// test default user-agent is attached to the request
	userAgent := req.Header.Get("User-Agent")
	if c.UserAgent != userAgent {
		t.Errorf("NewRequest() User-Agent = %v, expected %v", userAgent, c.UserAgent)
	}
}

func TestNewRequest_withUserData(t *testing.T) {
	c := NewClient(nil)

	inURL, outURL := "/foo", defaultBaseURL+"foo"
	inBody, outBody := &DropletCreateRequest{Name: "l", UserData: "u"},
		`{"name":"l","region":"","size":"","image":0,`+
			`"ssh_keys":null,"backups":false,"ipv6":false,`+
			`"private_networking":false,"monitoring":false,"user_data":"u","tags":null}`+"\n"
	req, _ := c.NewRequest(ctx, http.MethodPost, inURL, inBody)

	// test relative URL was expanded
	if req.URL.String() != outURL {
		t.Errorf("NewRequest(%v) URL = %v, expected %v", inURL, req.URL, outURL)
	}

	// test body was JSON encoded
	body, _ := ioutil.ReadAll(req.Body)
	if string(body) != outBody {
		t.Errorf("NewRequest(%v)Body = %v, expected %v", inBody, string(body), outBody)
	}

	// test default user-agent is attached to the request
	userAgent := req.Header.Get("User-Agent")
	if c.UserAgent != userAgent {
		t.Errorf("NewRequest() User-Agent = %v, expected %v", userAgent, c.UserAgent)
	}
}

func TestNewRequest_withDropletAgent(t *testing.T) {
	c := NewClient(nil)

	boolVal := true
	inURL, outURL := "/foo", defaultBaseURL+"foo"
	inBody, outBody := &DropletCreateRequest{Name: "l", WithDropletAgent: &boolVal},
		`{"name":"l","region":"","size":"","image":0,`+
			`"ssh_keys":null,"backups":false,"ipv6":false,`+
			`"private_networking":false,"monitoring":false,"tags":null,"with_droplet_agent":true}`+"\n"
	req, _ := c.NewRequest(ctx, http.MethodPost, inURL, inBody)

	// test relative URL was expanded
	if req.URL.String() != outURL {
		t.Errorf("NewRequest(%v) URL = %v, expected %v", inURL, req.URL, outURL)
	}

	// test body was JSON encoded
	body, _ := ioutil.ReadAll(req.Body)
	if string(body) != outBody {
		t.Errorf("NewRequest(%v)Body = %v, expected %v", inBody, string(body), outBody)
	}

	// test default user-agent is attached to the request
	userAgent := req.Header.Get("User-Agent")
	if c.UserAgent != userAgent {
		t.Errorf("NewRequest() User-Agent = %v, expected %v", userAgent, c.UserAgent)
	}
}

func TestNewRequest_badURL(t *testing.T) {
	c := NewClient(nil)
	_, err := c.NewRequest(ctx, http.MethodGet, ":", nil)
	testURLParseError(t, err)
}

func TestNewRequest_withCustomUserAgent(t *testing.T) {
	ua := "testing/0.0.1"
	c, err := New(nil, SetUserAgent(ua))

	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	req, _ := c.NewRequest(ctx, http.MethodGet, "/foo", nil)

	expected := fmt.Sprintf("%s %s", ua, userAgent)
	if got := req.Header.Get("User-Agent"); got != expected {
		t.Errorf("New() UserAgent = %s; expected %s", got, expected)
	}
}

func TestNewRequest_withCustomHeaders(t *testing.T) {
	expectedIdentity := "identity"
	expectedCustom := "x_test_header"

	c, err := New(nil, SetRequestHeaders(map[string]string{
		"Accept-Encoding": expectedIdentity,
		"X-Test-Header":   expectedCustom,
	}))
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	req, _ := c.NewRequest(ctx, http.MethodGet, "/foo", nil)

	if got := req.Header.Get("Accept-Encoding"); got != expectedIdentity {
		t.Errorf("New() Custom Accept Encoding Header = %s; expected %s", got, expectedIdentity)
	}
	if got := req.Header.Get("X-Test-Header"); got != expectedCustom {
		t.Errorf("New() Custom Accept Encoding Header = %s; expected %s", got, expectedCustom)
	}
}

func TestDo(t *testing.T) {
	setup()
	defer teardown()

	type foo struct {
		A string
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if m := http.MethodGet; m != r.Method {
			t.Errorf("Request method = %v, expected %v", r.Method, m)
		}
		fmt.Fprint(w, `{"A":"a"}`)
	})

	req, _ := client.NewRequest(ctx, http.MethodGet, "/", nil)
	body := new(foo)
	_, err := client.Do(context.Background(), req, body)
	if err != nil {
		t.Fatalf("Do(): %v", err)
	}

	expected := &foo{"a"}
	if !reflect.DeepEqual(body, expected) {
		t.Errorf("Response body = %v, expected %v", body, expected)
	}
}

func TestDo_httpError(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad Request", 400)
	})

	req, _ := client.NewRequest(ctx, http.MethodGet, "/", nil)
	_, err := client.Do(context.Background(), req, nil)

	if err == nil {
		t.Error("Expected HTTP 400 error.")
	}
}

// Test handling of an error caused by the internal http client's Do()
// function.
func TestDo_redirectLoop(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusFound)
	})

	req, _ := client.NewRequest(ctx, http.MethodGet, "/", nil)
	_, err := client.Do(context.Background(), req, nil)

	if err == nil {
		t.Error("Expected error to be returned.")
	}
	if err, ok := err.(*url.Error); !ok {
		t.Errorf("Expected a URL error; got %#v.", err)
	}
}

func TestCheckResponse(t *testing.T) {
	testHeaders := make(http.Header, 1)
	testHeaders.Set("x-request-id", "dead-beef")

	tests := []struct {
		title    string
		input    *http.Response
		expected *ErrorResponse
	}{
		{
			title: "default (no request_id)",
			input: &http.Response{
				Request:    &http.Request{},
				StatusCode: http.StatusBadRequest,
				Body: ioutil.NopCloser(strings.NewReader(`{"message":"m",
			"errors": [{"resource": "r", "field": "f", "code": "c"}]}`)),
			},
			expected: &ErrorResponse{
				Message: "m",
			},
		},
		{
			title: "request_id in body",
			input: &http.Response{
				Request:    &http.Request{},
				StatusCode: http.StatusBadRequest,
				Body: ioutil.NopCloser(strings.NewReader(`{"message":"m", "request_id": "dead-beef",
			"errors": [{"resource": "r", "field": "f", "code": "c"}]}`)),
			},
			expected: &ErrorResponse{
				Message:   "m",
				RequestID: "dead-beef",
			},
		},
		{
			title: "request_id in header",
			input: &http.Response{
				Request:    &http.Request{},
				StatusCode: http.StatusBadRequest,
				Header:     testHeaders,
				Body: ioutil.NopCloser(strings.NewReader(`{"message":"m",
			"errors": [{"resource": "r", "field": "f", "code": "c"}]}`)),
			},
			expected: &ErrorResponse{
				Message:   "m",
				RequestID: "dead-beef",
			},
		},
		// This tests that the ID in the body takes precedence to ensure we maintain the current
		// behavior. In practice, the IDs in the header and body should always be the same.
		{
			title: "request_id in both",
			input: &http.Response{
				Request:    &http.Request{},
				StatusCode: http.StatusBadRequest,
				Header:     testHeaders,
				Body: ioutil.NopCloser(strings.NewReader(`{"message":"m", "request_id": "dead-beef-body",
			"errors": [{"resource": "r", "field": "f", "code": "c"}]}`)),
			},
			expected: &ErrorResponse{
				Message:   "m",
				RequestID: "dead-beef-body",
			},
		},
		// ensure that we properly handle API errors that do not contain a
		// response body
		{
			title: "no body",
			input: &http.Response{
				Request:    &http.Request{},
				StatusCode: http.StatusBadRequest,
				Body:       ioutil.NopCloser(strings.NewReader("")),
			},
			expected: &ErrorResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			err := CheckResponse(tt.input).(*ErrorResponse)
			if err == nil {
				t.Fatalf("Expected error response.")
			}
			tt.expected.Response = tt.input

			if !reflect.DeepEqual(err, tt.expected) {
				t.Errorf("Error = %#v, expected %#v", err, tt.expected)
			}
		})
	}
}

func TestErrorResponse_Error(t *testing.T) {
	res := &http.Response{Request: &http.Request{}}
	err := ErrorResponse{Message: "m", Response: res}
	if err.Error() == "" {
		t.Errorf("Expected non-empty ErrorResponse.Error()")
	}
}

func TestDo_rateLimit(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(headerRateLimit, "60")
		w.Header().Add(headerRateRemaining, "59")
		w.Header().Add(headerRateReset, "1372700873")
	})

	var expected int

	if expected = 0; client.Rate.Limit != expected {
		t.Errorf("Client rate limit = %v, expected %v", client.Rate.Limit, expected)
	}
	if expected = 0; client.Rate.Remaining != expected {
		t.Errorf("Client rate remaining = %v, got %v", client.Rate.Remaining, expected)
	}
	if !client.Rate.Reset.IsZero() {
		t.Errorf("Client rate reset not initialized to zero value")
	}
	if client.Rate != client.GetRate() {
		t.Errorf("Client rate is not the same as client.GetRate()")
	}

	req, _ := client.NewRequest(ctx, http.MethodGet, "/", nil)
	_, err := client.Do(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("Do(): %v", err)
	}

	if expected = 60; client.Rate.Limit != expected {
		t.Errorf("Client rate limit = %v, expected %v", client.Rate.Limit, expected)
	}
	if expected = 59; client.Rate.Remaining != expected {
		t.Errorf("Client rate remaining = %v, expected %v", client.Rate.Remaining, expected)
	}
	reset := time.Date(2013, 7, 1, 17, 47, 53, 0, time.UTC)
	if client.Rate.Reset.UTC() != reset {
		t.Errorf("Client rate reset = %v, expected %v", client.Rate.Reset, reset)
	}
	if client.Rate != client.GetRate() {
		t.Errorf("Client rate is not the same as client.GetRate()")
	}
}

func TestDo_rateLimitRace(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(headerRateLimit, "60")
		w.Header().Add(headerRateRemaining, "59")
		w.Header().Add(headerRateReset, "1372700873")
	})

	var (
		wg    sync.WaitGroup
		wait  = make(chan struct{})
		count = 100
	)
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			<-wait
			req, _ := client.NewRequest(ctx, http.MethodGet, "/", nil)
			_, err := client.Do(context.Background(), req, nil)
			if err != nil {
				t.Errorf("Do(): %v", err)
			}
			wg.Done()
		}()
	}
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			<-wait
			_ = client.GetRate()
			wg.Done()
		}()
	}

	close(wait)
	wg.Wait()
}

func TestDo_rateLimit_errorResponse(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(headerRateLimit, "60")
		w.Header().Add(headerRateRemaining, "59")
		w.Header().Add(headerRateReset, "1372700873")
		http.Error(w, `{"message":"bad request"}`, 400)
	})

	var expected int

	req, _ := client.NewRequest(ctx, http.MethodGet, "/", nil)
	_, _ = client.Do(context.Background(), req, nil)

	if expected = 60; client.Rate.Limit != expected {
		t.Errorf("Client rate limit = %v, expected %v", client.Rate.Limit, expected)
	}
	if expected = 59; client.Rate.Remaining != expected {
		t.Errorf("Client rate remaining = %v, expected %v", client.Rate.Remaining, expected)
	}
	reset := time.Date(2013, 7, 1, 17, 47, 53, 0, time.UTC)
	if client.Rate.Reset.UTC() != reset {
		t.Errorf("Client rate reset = %v, expected %v", client.Rate.Reset, reset)
	}
}

func checkCurrentPage(t *testing.T, resp *Response, expectedPage int) {
	links := resp.Links
	p, err := links.CurrentPage()
	if err != nil {
		t.Fatal(err)
	}

	if p != expectedPage {
		t.Fatalf("expected current page to be '%d', was '%d'", expectedPage, p)
	}
}

func checkNextPageToken(t *testing.T, resp *Response, expectedNextPageToken string) {
	t.Helper()
	links := resp.Links
	pageToken, err := links.NextPageToken()
	if err != nil {
		t.Fatal(err)
	}

	if pageToken != expectedNextPageToken {
		t.Fatalf("expected next page token to be '%s', was '%s'", expectedNextPageToken, pageToken)
	}
}

func checkPreviousPageToken(t *testing.T, resp *Response, expectedPreviousPageToken string) {
	t.Helper()
	links := resp.Links
	pageToken, err := links.PrevPageToken()
	if err != nil {
		t.Fatal(err)
	}

	if pageToken != expectedPreviousPageToken {
		t.Fatalf("expected previous page token to be '%s', was '%s'", expectedPreviousPageToken, pageToken)
	}
}

func TestDo_completion_callback(t *testing.T) {
	setup()
	defer teardown()

	type foo struct {
		A string
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if m := http.MethodGet; m != r.Method {
			t.Errorf("Request method = %v, expected %v", r.Method, m)
		}
		fmt.Fprint(w, `{"A":"a"}`)
	})

	req, _ := client.NewRequest(ctx, http.MethodGet, "/", nil)
	body := new(foo)
	var completedReq *http.Request
	var completedResp string
	client.OnRequestCompleted(func(req *http.Request, resp *http.Response) {
		completedReq = req
		b, err := httputil.DumpResponse(resp, true)
		if err != nil {
			t.Errorf("Failed to dump response: %s", err)
		}
		completedResp = string(b)
	})
	_, err := client.Do(context.Background(), req, body)
	if err != nil {
		t.Fatalf("Do(): %v", err)
	}
	if !reflect.DeepEqual(req, completedReq) {
		t.Errorf("Completed request = %v, expected %v", completedReq, req)
	}
	expected := `{"A":"a"}`
	if !strings.Contains(completedResp, expected) {
		t.Errorf("expected response to contain %v, Response = %v", expected, completedResp)
	}
}

func TestAddOptions(t *testing.T) {
	cases := []struct {
		name     string
		path     string
		expected string
		opts     *ListOptions
		isErr    bool
	}{
		{
			name:     "add options",
			path:     "/action",
			expected: "/action?page=1",
			opts:     &ListOptions{Page: 1},
			isErr:    false,
		},
		{
			name:     "add options with existing parameters",
			path:     "/action?scope=all",
			expected: "/action?page=1&scope=all",
			opts:     &ListOptions{Page: 1},
			isErr:    false,
		},
	}

	for _, c := range cases {
		got, err := addOptions(c.path, c.opts)
		if c.isErr && err == nil {
			t.Errorf("%q expected error but none was encountered", c.name)
			continue
		}

		if !c.isErr && err != nil {
			t.Errorf("%q unexpected error: %v", c.name, err)
			continue
		}

		gotURL, err := url.Parse(got)
		if err != nil {
			t.Errorf("%q unable to parse returned URL", c.name)
			continue
		}

		expectedURL, err := url.Parse(c.expected)
		if err != nil {
			t.Errorf("%q unable to parse expected URL", c.name)
			continue
		}

		if g, e := gotURL.Path, expectedURL.Path; g != e {
			t.Errorf("%q path = %q; expected %q", c.name, g, e)
			continue
		}

		if g, e := gotURL.Query(), expectedURL.Query(); !reflect.DeepEqual(g, e) {
			t.Errorf("%q query = %#v; expected %#v", c.name, g, e)
			continue
		}
	}
}

func TestCustomUserAgent(t *testing.T) {
	ua := "testing/0.0.1"
	c, err := New(nil, SetUserAgent(ua))

	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	expected := fmt.Sprintf("%s %s", ua, userAgent)
	if got := c.UserAgent; got != expected {
		t.Errorf("New() UserAgent = %s; expected %s", got, expected)
	}
}

func TestCustomBaseURL(t *testing.T) {
	baseURL := "http://localhost/foo"
	c, err := New(nil, SetBaseURL(baseURL))

	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	expected := baseURL
	if got := c.BaseURL.String(); got != expected {
		t.Errorf("New() BaseURL = %s; expected %s", got, expected)
	}
}

func TestSetStaticRateLimit(t *testing.T) {
	rps := float64(5)
	c, err := New(nil, SetStaticRateLimit(rps))
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	expected := rate.NewLimiter(rate.Limit(rps), 1)
	if got := c.rateLimiter; *got != *expected {
		t.Errorf("rateLimiter = %+v; expected %+v", got, expected)
	}
}

func TestCustomBaseURL_badURL(t *testing.T) {
	baseURL := ":"
	_, err := New(nil, SetBaseURL(baseURL))

	testURLParseError(t, err)
}
