package client

import (
	"errors"
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
	}
)

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
