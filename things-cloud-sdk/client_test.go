package thingscloud

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func stringVal(str string) *string {
	return &str
}

type fakeResponse struct {
	statusCode int
	file       string
}

func fakeServer(t fakeResponse) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open(fmt.Sprintf("tapes/%s", t.file))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Printf("Unable to open fixture %q path %q", t.file, r.URL.Path)
			return
		}
		defer f.Close()

		content, err := io.ReadAll(f)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Printf("Unable to load fixture %q path %q", t.file, r.URL.Path)
			return
		}
		w.WriteHeader(t.statusCode)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, string(content))
	}))
}

func TestClient_UserAgent(t *testing.T) {
	var capturedHeaders http.Header
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	c := New(ts.URL, "test@example.com", "password")

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.do(req)
	if err != nil {
		t.Fatal(err)
	}

	// Verify User-Agent is the updated value
	got := capturedHeaders.Get("User-Agent")
	want := "ThingsMac/32209501"
	if got != want {
		t.Errorf("User-Agent = %q, want %q", got, want)
	}

	// Verify things-client-info header is set and non-empty
	clientInfo := capturedHeaders.Get("Things-Client-Info")
	if clientInfo == "" {
		t.Error("things-client-info header is missing or empty")
	}
}

// abuseServer answers the first `blocked` requests with the abuse prevention
// response and counts every request it sees.
func abuseServer(blocked int, seen *int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*seen++
		if *seen <= blocked {
			w.Header().Set("Things-Response", "AbusePrevention")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte(`{}`))
	}))
}

func recordWaits(c *Client) *[]time.Duration {
	var waits []time.Duration
	c.abuseWait = func(_ context.Context, d time.Duration) error {
		waits = append(waits, d)
		return nil
	}
	return &waits
}

func TestClient_ReadBacksOffOnAbusePrevention(t *testing.T) {
	seen := 0
	ts := abuseServer(2, &seen)
	defer ts.Close()
	c := New(ts.URL, "test@example.com", "password", WithRequestInterval(0))
	waits := recordWaits(c)

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := c.do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || seen != 3 {
		t.Fatalf("status = %d after %d requests, want 200 after 3", resp.StatusCode, seen)
	}
	if len(*waits) != 2 || (*waits)[0] != time.Second || (*waits)[1] != 2*time.Second {
		t.Fatalf("waits = %v, want [1s 2s]", *waits)
	}
}

func TestClient_ReadGivesUpAfterThreeRetries(t *testing.T) {
	seen := 0
	ts := abuseServer(100, &seen)
	defer ts.Close()
	c := New(ts.URL, "test@example.com", "password", WithRequestInterval(0))
	recordWaits(c)

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := c.do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusTooManyRequests || seen != 4 {
		t.Fatalf("status = %d after %d requests, want 429 after 4", resp.StatusCode, seen)
	}
}

func TestClient_WriteIsNotRetriedOnAbusePrevention(t *testing.T) {
	seen := 0
	ts := abuseServer(1, &seen)
	defer ts.Close()
	c := New(ts.URL, "test@example.com", "password", WithRequestInterval(0))
	waits := recordWaits(c)

	req, _ := http.NewRequest("POST", "/test", nil)
	resp, err := c.do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusTooManyRequests || seen != 1 || len(*waits) != 0 {
		t.Fatalf("status = %d, requests = %d, waits = %v; a write must be sent once", resp.StatusCode, seen, *waits)
	}
}
