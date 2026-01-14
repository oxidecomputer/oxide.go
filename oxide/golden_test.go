// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

// TestGoldenRoundTrip tests that real API responses can be unmarshaled and
// marshaled back to equivalent JSON. This catches mismatches between our
// generated types and the actual API format.
//
// To refresh the fixtures, run:
//
//	go run ./oxide/testdata/main.go
func TestGoldenRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
		test    func(t *testing.T, fixture string)
	}{
		{
			name:    "timeseries_query_response",
			fixture: "testdata/recordings/timeseries_query_response.json",
			test:    testRoundTrip[OxqlQueryResult],
		},
		{
			name:    "disk_list_response",
			fixture: "testdata/recordings/disk_list_response.json",
			test:    testRoundTrip[DiskResultsPage],
		},
		{
			name:    "loopback_addresses_response",
			fixture: "testdata/recordings/loopback_addresses_response.json",
			test:    testRoundTrip[LoopbackAddressResultsPage],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t, tt.fixture)
		})
	}
}

func testRoundTrip[T any](t *testing.T, fixturePath string) {
	data, err := os.ReadFile(fixturePath)
	require.NoError(t, err, "failed to read fixture")

	var typed T
	err = json.Unmarshal(data, &typed)
	require.NoError(t, err, "failed to unmarshal fixture")

	remarshaled, err := json.Marshal(typed)
	require.NoError(t, err, "failed to marshal")

	var expected, actual any
	require.NoError(t, json.Unmarshal(data, &expected))
	require.NoError(t, json.Unmarshal(remarshaled, &actual))

	expected = stripNulls(expected)
	actual = stripNulls(actual)

	if diff := cmp.Diff(expected, actual, timestampComparer()); diff != "" {
		t.Errorf("round-trip mismatch (-fixture +remarshaled):\n%s", diff)
	}
}

// timestampComparer returns a cmp.Option that compares timestamp strings
// by their actual time value, handling precision differences. Rust and go format timestamps
// slightly differently, so we need to normalize to avoid spurious differences in marshalled values.
func timestampComparer() cmp.Option {
	return cmp.Comparer(func(a, b string) bool {
		ta, errA := time.Parse(time.RFC3339Nano, a)
		tb, errB := time.Parse(time.RFC3339Nano, b)
		if errA == nil && errB == nil {
			return ta.Equal(tb)
		}
		return a == b
	})
}

// stripNulls recursively removes null values from JSON-unmarshaled data. We use this workaround
// because the SDK and API don't always handle null fields consistently.
//
// TODO: Investigate options to harmonize null handling across services so that we don't have to
// pre-process the responses here.
func stripNulls(v any) any {
	switch val := v.(type) {
	case map[string]any:
		result := make(map[string]any)
		for k, v := range val {
			if v != nil {
				result[k] = stripNulls(v)
			}
		}
		return result
	case []any:
		result := make([]any, len(val))
		for i, v := range val {
			result[i] = stripNulls(v)
		}
		return result
	default:
		return v
	}
}
