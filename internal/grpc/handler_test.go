// Copyright (c) 2021 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

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
		expectError   string
	}{
		{
			desc:          "no scheme defaults to tcp",
			address:       "127.0.0.1:1234",
			expectNetwork: "tcp",
			expectAddress: "127.0.0.1:1234",
		},
		{
			desc:          "tcp scheme",
			address:       "tcp://127.0.0.1:1234",
			expectNetwork: "tcp",
			expectAddress: "127.0.0.1:1234",
		},
		{
			desc:          "unix scheme",
			address:       "unix:///foo",
			expectNetwork: "unix",
			expectAddress: "/foo",
		},
		{
			desc:        "invalid scheme",
			address:     "foo:///bar",
			expectError: "invalid network, only tcp or unix allowed: foo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			net, addr, err := getNetworkAddress(tt.address)
			if tt.expectError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expectNetwork, net)
			require.Equal(t, tt.expectAddress, addr)
		})
	}
}
