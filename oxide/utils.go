package oxide

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/template"
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
	t, err := template.New("url").Parse(u.Path)
	if err != nil {
		return fmt.Errorf("parsing template for url path %q failed: %v", u.Path, err)
	}
	var b bytes.Buffer
	if err := t.Execute(&b, expansions); err != nil {
		return fmt.Errorf("executing template for url path failed: %v", err)
	}

	// set the parameters
	u.Path = b.String()

	// escape the expansions
	for k, v := range expansions {
		expansions[k] = url.QueryEscape(v)
	}

	var bt bytes.Buffer
	if err := t.Execute(&bt, expansions); err != nil {
		return fmt.Errorf("executing template for url path failed: %v", err)
	}

	// set the parameters
	u.RawPath = bt.String()

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
