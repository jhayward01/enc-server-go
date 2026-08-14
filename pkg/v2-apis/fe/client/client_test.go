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

// Test data
const idStr = "JTH"
const idHexStr = "4a5448"

const testID = "test-id"
const testKey = "test-key"
const testData = "test-data"

// Content types
const contentTypeJSON = "application/json"
const contentTypeHeader = "Content-Type"

// Server paths
const serverAddr = "localhost:7777"
const serverRecordsAddr = "http://localhost:7777/records"
const serverRecordsEndpoint = "/records/"
const serverRecordsQuery = "key="

// HTTP methods
const httpMethodPOST = "POST"
const httpMethodGET = "GET"
const httpMethodDELETE = "DELETE"

// Bad status
const bad404Status = "404 Not Found"
const bad500Status = "500 Internal Server Error"

// Error messages
const errClientMessage = "MakeClient missing configuration serverAddr"
const errConnectionRefused = "connection refused"
const errServerError = "Server error"
const errNotFound = "Not found"
const errInvalidJSON = "invalid json"

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
			wantErr: errors.New(errClientMessage),
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
			id:   []byte(testID),
			data: []byte(testData),
			mockFn: func(req *http.Request) (*http.Response, error) {
				// Verify request details
				assert.Equal(t, httpMethodPOST, req.Method)
				assert.True(t, strings.HasPrefix(req.URL.String(), serverRecordsAddr))
				assert.Equal(t, contentTypeJSON, req.Header.Get(contentTypeHeader))

				// Return successful response
				responseRecord := record{
					ID:   hex.EncodeToString([]byte(testID)),
					Key:  hex.EncodeToString([]byte(testKey)),
					Data: hex.EncodeToString([]byte(testData)),
				}
				respBody, _ := json.Marshal(responseRecord)

				return &http.Response{
					StatusCode: http.StatusCreated,
					Body:       io.NopCloser(strings.NewReader(string(respBody))),
					Header:     make(http.Header),
				}, nil
			},
			wantKey: []byte(testKey),
			wantErr: false,
		},
		{
			name: "should fail when request returns non-201 status",
			id:   []byte(testID),
			data: []byte(testData),
			mockFn: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(strings.NewReader(errServerError)),
					Header:     make(http.Header),
					Status:     bad500Status,
				}, nil
			},
			wantErr:     true,
			errContains: "Bad status making POST request",
		},
		{
			name: "should fail on request error",
			id:   []byte(testID),
			data: []byte(testData),
			mockFn: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New(errConnectionRefused)
			},
			wantErr:     true,
			errContains: "Error making POST request",
		},
		{
			name: "should fail on invalid response JSON",
			id:   []byte(testID),
			data: []byte(testData),
			mockFn: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusCreated,
					Body:       io.NopCloser(strings.NewReader(errInvalidJSON)),
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
			id:   []byte(testID),
			key:  []byte(testKey),
			mockFn: func(req *http.Request) (*http.Response, error) {
				// Verify request details
				assert.Equal(t, httpMethodGET, req.Method)
				assert.True(t, strings.Contains(req.URL.String(), serverRecordsEndpoint))
				assert.True(t, strings.Contains(req.URL.RawQuery, serverRecordsQuery))

				// Return successful response
				responseRecord := record{
					ID:   hex.EncodeToString([]byte(testID)),
					Key:  hex.EncodeToString([]byte(testKey)),
					Data: hex.EncodeToString([]byte(testData)),
				}
				respBody, _ := json.Marshal(responseRecord)

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(string(respBody))),
					Header:     make(http.Header),
				}, nil
			},
			wantData: []byte(hex.EncodeToString([]byte(testData))),
			wantErr:  false,
		},
		{
			name: "should fail when request returns non-200 status",
			id:   []byte(testID),
			key:  []byte(testKey),
			mockFn: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader(errNotFound)),
					Header:     make(http.Header),
					Status:     bad404Status,
				}, nil
			},
			wantErr:     true,
			errContains: "Bad status making GET request",
		},
		{
			name: "should fail on request error",
			id:   []byte(testID),
			key:  []byte(testKey),
			mockFn: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New(errConnectionRefused)
			},
			wantErr:     true,
			errContains: "Error making GET request",
		},
		{
			name: "should fail on invalid response JSON",
			id:   []byte(testID),
			key:  []byte(testKey),
			mockFn: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(errInvalidJSON)),
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

// DeleteRecord() - Test Method
func TestClient_DeleteRecord(t *testing.T) {
	tests := []struct {
		name        string
		id          []byte
		mockFn      func(req *http.Request) (*http.Response, error)
		wantErr     bool
		errContains string
	}{
		{
			name: "should delete record successfully",
			id:   []byte(testID),
			mockFn: func(req *http.Request) (*http.Response, error) {
				// Verify request details
				assert.Equal(t, httpMethodDELETE, req.Method)
				assert.True(t, strings.HasPrefix(req.URL.String(), serverRecordsAddr))

				return &http.Response{
					StatusCode: http.StatusAccepted,
					Body:       io.NopCloser(strings.NewReader("")),
					Header:     make(http.Header),
				}, nil
			},
			wantErr: false,
		},
		{
			name: "should fail when request returns non-202 status",
			id:   []byte(testID),
			mockFn: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(strings.NewReader(errServerError)),
					Header:     make(http.Header),
					Status:     bad500Status,
				}, nil
			},
			wantErr:     true,
			errContains: "Bad status making DELETE request",
		},
		{
			name: "should fail on request error",
			id:   []byte(testID),
			mockFn: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New(errConnectionRefused)
			},
			wantErr:     true,
			errContains: "Error making DELETE request",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &clientImpl{
				serverAddr: serverAddr,
				httpClient: createMockClient(test.mockFn),
			}

			err := client.DeleteRecord(test.id)

			if test.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
