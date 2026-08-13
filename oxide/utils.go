package oxide

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// NewPointer returns a pointer to a given value.
func NewPointer[T any](v T) *T {
	return &v
}

// PointerIntToStr converts an *int into a string.
// If nil, an empty string is returned.
func PointerIntToStr(i *int) string {
	if i == nil {
		return ""
	}

	return strconv.Itoa(*i)
}

// PointerUint64ToStr converts a *uint64 into a string.
// If nil, an empty string is returned.
func PointerUint64ToStr(i *uint64) string {
	if i == nil {
		return ""
	}

	return strconv.FormatUint(*i, 10)
}

// PointerTimeToStr converts a *time.Time into an RFC3339 string.
// If nil, an empty string is returned.
func PointerTimeToStr(t *time.Time) string {
	if t == nil {
		return ""
	}

	return t.Format(time.RFC3339)
}

// resolveRelative combines a url base with a relative path.
func resolveRelative(basestr, relstr string) string {
	u, _ := url.Parse(basestr)
	rel, _ := url.Parse(relstr)
	u = u.ResolveReference(rel)
	us := u.String()
	us = strings.ReplaceAll(us, "%7B", "{")
	us = strings.ReplaceAll(us, "%7D", "}")
	return us
}

// expandURL substitutes any {encoded} strings in the URL passed in using
// the map supplied.
func expandURL(u *url.URL, expansions map[string]string) error {
	t, err := template.New("url").Option("missingkey=error").Parse(u.Path)
	if err != nil {
		return fmt.Errorf("parsing template for url path %q failed: %v", u.Path, err)
	}

	// Render escaped and unescaped versions of the URL.
	//
	// url.URL has two fields to store the path. `Path` stores the decoded
	// (unescaped) form and `RawPath` is an optional field that can be used to
	// hint to consumers how the URL should be escaped and used externally,
	// such as by the `URL.String()` and `URL.RequestURI()` methods.
	//
	// We store both values to indicate that part of the path is user input,
	// and therefore must always be escaped. Without `RawPath`, an input such
	// as `../projects/my-project` would cause the endpoint `/disks/{{.disk}}`
	// to render as `/disks/../projects/my-project`, resulting in the wrong
	// API call.
	var unescaped bytes.Buffer
	if err := t.Execute(&unescaped, expansions); err != nil {
		return fmt.Errorf("executing template for url path failed: %v", err)
	}
	u.Path = unescaped.String()

	// Escape path elements and render the escaped path.
	for k, v := range expansions {
		expansions[k] = url.PathEscape(v)
	}

	var escaped bytes.Buffer
	if err := t.Execute(&escaped, expansions); err != nil {
		return fmt.Errorf("executing template for url path failed: %v", err)
	}
	u.RawPath = escaped.String()

	return nil
}

func addQueries(u *url.URL, query map[string]string) {
	q := u.Query()
	for k, v := range query {
		if v == "" {
			continue
		}

		//escape the string
		query[k] = url.QueryEscape(v)

		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
}
