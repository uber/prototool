package grpc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetNetworkAddress(t *testing.T) {
	tests := []struct {
		desc          string
		address       string
		expectNetwork string
		expectAddress string
	}{
		{
			desc:          "no prefix defaults to tcp",
			address:       "127.0.0.1:1234",
			expectNetwork: "tcp",
			expectAddress: "127.0.0.1:1234",
		},
		{
			desc:          "tcp prefix",
			address:       "tcp://127.0.0.1:1234",
			expectNetwork: "tcp",
			expectAddress: "127.0.0.1:1234",
		},
		{
			desc:          "unix prefix",
			address:       "unix:///foo",
			expectNetwork: "unix",
			expectAddress: "/foo",
		},
	}

	for _, tt := range tests {
		net, addr := getNetworkAddress(tt.address)
		require.Equal(t, tt.expectNetwork, net)
		require.Equal(t, tt.expectAddress, addr)
	}
}
