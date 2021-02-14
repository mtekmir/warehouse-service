package server_test

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func checkErr(t *testing.T, r *http.Response, expectedMsg string) {
	t.Helper()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("Unable to read body. %v", err)
	}
	type errRes struct {
		Message string `json:"message"`
	}
	var er errRes
	if err := json.Unmarshal(body, &er); err != nil {
		t.Fatalf("Unable to unmarshal err message. %v", err)
	}
	if er.Message != expectedMsg {
		t.Errorf("Expected err message to be %s, got %s", expectedMsg, er.Message)
	}
}

type reqHeader = struct {
	key   string
	value string
}

func testRequest(
	t *testing.T,
	ts *httptest.Server,
	method,
	path string,
	body interface{},
	headers []reqHeader,
) *http.Response {
	var b io.Reader
	switch body.(type) {
	case string:
		if body != "" {
			b = bytes.NewBuffer([]byte(body.(string)))
		}
	case io.Reader:
		b = body.(io.Reader)
	}
	req, err := http.NewRequest(method, ts.URL+path, b)
	for _, h := range headers {
		req.Header.Set(h.key, h.value)
	}

	if err != nil {
		t.Fatal(err)
		return nil
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil
	}
	return resp
}
