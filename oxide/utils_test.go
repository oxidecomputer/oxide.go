package oxide

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

type expandTest struct {
	in         string
	expansions map[string]string
	want       string
}

const testServerURL = "https://example.com"

var expandTests = []expandTest{
	// no expansions
	{
		"",
		map[string]string{},
		testServerURL,
	},
	// multiple expansions, no escaping
	{
		"file/convert/{{.srcFormat}}/{{.outputFormat}}",
		map[string]string{
			"srcFormat":    "step",
			"outputFormat": "obj",
		},
		testServerURL + "/file/convert/step/obj",
	},
}

func TestExpandURL(t *testing.T) {
	for i, test := range expandTests {
		uri := resolveRelative(testServerURL, test.in)
		u, err := url.Parse(uri)
		if err != nil {
			t.Fatalf("parsing url %q failed: %v", test.in, err)
		}
		expandURL(u, test.expansions)
		got := u.String()
		if got != test.want {
			t.Errorf("got %q expected %q in test %d", got, test.want, i+1)
		}
	}
}

func Test_addQueries(t *testing.T) {
	type args struct {
		query map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "keeps URL the same if no query params supplied",
			args: args{query: map[string]string{}},
			want: "https://example.com",
		},
		{
			name: "keeps URL the same if no query values supplied",
			args: args{query: map[string]string{
				"organization": "",
				"project":      "",
			}},
			want: "https://example.com",
		},
		{
			name: "adds query parameters successfully",
			args: args{query: map[string]string{
				"organization": "myorg",
				"project":      "prod",
			}},
			want: "https://example.com?organization=myorg&project=prod",
		},
	}
	for _, tt := range tests {
		u, err := url.Parse(testServerURL)
		if err != nil {
			t.Fatalf("parsing url %q failed: %v", testServerURL, err)
		}

		t.Run(tt.name, func(t *testing.T) {
			addQueries(u, tt.args.query)
			got := u.String()
			assert.Equal(t, tt.want, got)
		})
	}
}
