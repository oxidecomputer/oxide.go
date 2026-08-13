// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewIpNet(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		want    IpNet
		wantErr string
	}{
		{
			name:  "new IpNet v4",
			input: "192.168.1.0/24",
			want: IpNet{
				Value: NewPointer(Ipv4Net("192.168.1.0/24")),
			},
		},
		{
			name:  "new IpNet v6",
			input: "fd00::/64",
			want: IpNet{
				Value: NewPointer(Ipv6Net("fd00::/64")),
			},
		},
		{
			name:    "invalid IpNet format",
			input:   `0.0.0.0\/0`,
			wantErr: "invalid IP network",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewIpNet(tc.input)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, got)
			}
		})
	}
}
