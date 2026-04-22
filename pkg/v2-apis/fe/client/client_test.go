package client

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"enc-server-go/pkg/utils"
)

// Test Constants
const serverAddr = "localhost:7777"

const idStr = "JTH"
const idHexStr = "4a5448"

const badClientMessage = "MakeClient missing configuration serverAddr"

// Test Variables
var (
	goodClientConfig = map[string]string{
		"serverAddr": serverAddr}

	badClientConfig = map[string]string{
		"foo": "bar"}

	goodClient = &clientImpl{
		serverAddr: serverAddr,
		httpClient: &http.Client{},
	}
)

// Mock RoundTripper for HTTP mocking
type mockRoundTripper struct {
	fn func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.fn(req)
}

// Helper function to create a mock HTTP client
func createMockClient(fn func(req *http.Request) (*http.Response, error)) *http.Client {
	return &http.Client{
		Transport: &mockRoundTripper{fn: fn},
	}
}

// MakeClient() - Test Method
func TestClient_MakeClient(t *testing.T) {

	type fields struct{}
	type args struct {
		config map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    utils.ClientFE
		wantErr error
	}{
		{
			name: "should run successfully",
			args: args{goodClientConfig},
			want: goodClient,
		},
		{
			name:    "should fail loading configuration",
			args:    args{badClientConfig},
			wantErr: errors.New(badClientMessage),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := MakeClient(test.args.config)
			assert.Equal(t, test.want, got)
			assert.Equal(t, test.wantErr, err)
		})
	}
}

// StoreRecord() - Test Method
func TestClient_StoreRecord(t *testing.T) {
	tests := []struct {
		name        string
		id          []byte
		data        []byte
		mockFn      func(req *http.Request) (*http.Response, error)
		wantKey     []byte
		wantErr     bool
		errContains string
	}{
		{
			name: "should store record successfully",
			id:   []byte("test-id"),
			data: []byte("test-data"),
			mockFn: func(req *http.Request) (*http.Response, error) {
				// Verify request details
				assert.Equal(t, "POST", req.Method)
				assert.True(t, strings.HasPrefix(req.URL.String(), "http://localhost:7777/records"))
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

				// Return successful response
				responseRecord := record{
					ID:   hex.EncodeToString([]byte("test-id")),
					Key:  hex.EncodeToString([]byte("test-key")),
					Data: hex.EncodeToString([]byte("test-data")),
				}
				respBody, _ := json.Marshal(responseRecord)

				return &http.Response{
					StatusCode: http.StatusCreated,
					Body:       io.NopCloser(strings.NewReader(string(respBody))),
					Header:     make(http.Header),
				}, nil
			},
			wantKey: []byte("test-key"),
			wantErr: false,
		},
		{
			name: "should fail when request returns non-201 status",
			id:   []byte("test-id"),
			data: []byte("test-data"),
			mockFn: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(strings.NewReader("Server error")),
					Header:     make(http.Header),
					Status:     "500 Internal Server Error",
				}, nil
			},
			wantErr:     true,
			errContains: "Bad status making POST request",
		},
		{
			name: "should fail on request error",
			id:   []byte("test-id"),
			data: []byte("test-data"),
			mockFn: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("connection refused")
			},
			wantErr:     true,
			errContains: "Error making POST request",
		},
		{
			name: "should fail on invalid response JSON",
			id:   []byte("test-id"),
			data: []byte("test-data"),
			mockFn: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusCreated,
					Body:       io.NopCloser(strings.NewReader("invalid json")),
					Header:     make(http.Header),
				}, nil
			},
			wantErr:     true,
			errContains: "Error unmarshalling record",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &clientImpl{
				serverAddr: serverAddr,
				httpClient: createMockClient(test.mockFn),
			}

			got, err := client.StoreRecord(test.id, test.data)

			if test.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.errContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.wantKey, got)
			}
		})
	}
}


// RetrieveRecord() - Test Method
func TestClient_RetrieveRecord(t *testing.T) {
	tests := []struct {
		name        string
		id          []byte
		key         []byte
		mockFn      func(req *http.Request) (*http.Response, error)
		wantData    []byte
		wantErr     bool
		errContains string
	}{
		{
			name: "should retrieve record successfully",
			id:   []byte("test-id"),
			key:  []byte("test-key"),
			mockFn: func(req *http.Request) (*http.Response, error) {
				// Verify request details
				assert.Equal(t, "GET", req.Method)
				assert.True(t, strings.Contains(req.URL.String(), "/records/"))
				assert.True(t, strings.Contains(req.URL.RawQuery, "key="))

				// Return successful response
				responseRecord := record{
					ID:   hex.EncodeToString([]byte("test-id")),
					Key:  hex.EncodeToString([]byte("test-key")),
					Data: hex.EncodeToString([]byte("test-data")),
				}
				respBody, _ := json.Marshal(responseRecord)

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(string(respBody))),
					Header:     make(http.Header),
				}, nil
			},
			wantData: []byte(hex.EncodeToString([]byte("test-data"))),
			wantErr:  false,
		},
		// {
		// 	name: "should fail when request returns non-200 status",
		// 	id:   []byte("test-id"),
		// 	key:  []byte("test-key"),
		// 	mockFn: func(req *http.Request) (*http.Response, error) {
		// 		return &http.Response{
		// 			StatusCode: http.StatusNotFound,
		// 			Body:       io.NopCloser(strings.NewReader("Not found")),
		// 			Header:     make(http.Header),
		// 			Status:     "404 Not Found",
		// 		}, nil
		// 	},
		// 	wantErr:     true,
		// 	errContains: "Bad status making GET request",
		// },
		// {
		// 	name: "should fail on request error",
		// 	id:   []byte("test-id"),
		// 	key:  []byte("test-key"),
		// 	mockFn: func(req *http.Request) (*http.Response, error) {
		// 		return nil, errors.New("connection refused")
		// 	},
		// 	wantErr:     true,
		// 	errContains: "Error making GET request",
		// },
		// {
		// 	name: "should fail on invalid response JSON",
		// 	id:   []byte("test-id"),
		// 	key:  []byte("test-key"),
		// 	mockFn: func(req *http.Request) (*http.Response, error) {
		// 		return &http.Response{
		// 			StatusCode: http.StatusOK,
		// 			Body:       io.NopCloser(strings.NewReader("invalid json")),
		// 			Header:     make(http.Header),
		// 		}, nil
		// 	},
		// 	wantErr:     true,
		// 	errContains: "Error unmarshalling record",
		// },
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &clientImpl{
				serverAddr: serverAddr,
				httpClient: createMockClient(test.mockFn),
			}

			got, err := client.RetrieveRecord(test.id, test.key)

			if test.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.errContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.wantData, got)
			}
		})
	}
}
