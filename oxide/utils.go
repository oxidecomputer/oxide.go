package oxide

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"text/template"
)

// resolveRelative combines a url base with a relative path.
func resolveRelative(basestr, relstr string) string {
	u, _ := url.Parse(basestr)
	rel, _ := url.Parse(relstr)
	u = u.ResolveReference(rel)
	us := u.String()
	us = strings.Replace(us, "%7B", "{", -1)
	us = strings.Replace(us, "%7D", "}", -1)
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
