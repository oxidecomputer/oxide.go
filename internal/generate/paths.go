package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/iancoleman/strcase"
)

// Generate the paths.go file.
func generatePaths(doc *openapi3.T) error {
	f, err := openGeneratedFile("../../oxide/paths.go")
	if err != nil {
		return err
	}
	defer f.Close()

	// Iterate over all the paths in the spec and write the types.
	// We want to ensure we keep the order so the diffs don't look like shit.
	keys := make([]string, 0)
	for k := range doc.Paths {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, path := range keys {
		p := doc.Paths[path]
		if p.Ref != "" {
			fmt.Printf("[WARN] TODO: skipping path for %q, since it is a reference\n", path)
			continue
		}

		if err := writePath(doc, f, path, p); err != nil {
			return err
		}
	}

	return nil
}

// writePath writes the given path as an http request to the given file.
func writePath(doc *openapi3.T, f *os.File, path string, p *openapi3.PathItem) error {
	if p.Get != nil {
		if err := writeMethod(doc, f, http.MethodGet, path, p.Get, false); err != nil {
			return err
		}
	}

	if p.Post != nil {
		if err := writeMethod(doc, f, http.MethodPost, path, p.Post, false); err != nil {
			return err
		}
	}

	if p.Put != nil {
		if err := writeMethod(doc, f, http.MethodPut, path, p.Put, false); err != nil {
			return err
		}
	}

	if p.Delete != nil {
		if err := writeMethod(doc, f, http.MethodDelete, path, p.Delete, false); err != nil {
			return err
		}
	}

	if p.Patch != nil {
		if err := writeMethod(doc, f, http.MethodPatch, path, p.Patch, false); err != nil {
			return err
		}
	}

	if p.Head != nil {
		if err := writeMethod(doc, f, http.MethodHead, path, p.Head, false); err != nil {
			return err
		}
	}

	if p.Options != nil {
		if err := writeMethod(doc, f, http.MethodOptions, path, p.Options, false); err != nil {
			return err
		}
	}

	return nil
}

func writeMethod(doc *openapi3.T, f *os.File, method string, path string, o *openapi3.Operation, isGetAllPages bool) error {
	respType, pagedRespType, err := getSuccessResponseType(o, isGetAllPages)
	if err != nil {
		return err
	}

	if len(o.Tags) == 0 {
		fmt.Printf("[WARN] TODO: skipping operation %q, since it has no tag\n", o.OperationID)
		return nil
	}
	tag := strcase.ToCamel(o.Tags[0])

	if tag == "Hidden" {
		// return early.
		return nil
	}

	fnName := printProperty(o.OperationID)

	pageResult := false

	// Parse the parameters.
	params := map[string]*openapi3.Parameter{}
	paramsString := ""
	docParamsString := ""
	for index, p := range o.Parameters {
		if p.Ref != "" {
			fmt.Printf("[WARN] TODO: skipping parameter for %q, since it is a reference\n", p.Value.Name)
			continue
		}

		paramName := printPropertyLower(p.Value.Name)

		// Check if we have a page result.
		if isPageParam(paramName) && method == http.MethodGet {
			pageResult = true
		}

		params[p.Value.Name] = p.Value
		paramsString += fmt.Sprintf("%s %s, ", paramName, printType(p.Value.Name, p.Value.Schema))
		if index == len(o.Parameters)-1 {
			docParamsString += paramName
		} else {
			docParamsString += fmt.Sprintf("%s, ", paramName)
		}
	}

	if pageResult && isGetAllPages && len(pagedRespType) > 0 {
		respType = pagedRespType
	}

	// Parse the request body.
	reqBodyParam := "nil"
	reqBodyDescription := ""
	if o.RequestBody != nil {
		rb := o.RequestBody

		if rb.Value.Description != "" {
			reqBodyDescription = rb.Value.Description
		}

		if rb.Ref != "" {
			fmt.Printf("[WARN] TODO: skipping request body for %q, since it is a reference: %q\n", path, rb.Ref)
		}

		for mt, r := range rb.Value.Content {
			if mt != "application/json" {
				paramsString += "b io.Reader"
				reqBodyParam = "b"
				break
			}

			typeName := printType("", r.Schema)

			paramsString += "j *" + typeName

			if len(docParamsString) > 0 {
				docParamsString += ", "
			}
			docParamsString += "body"

			reqBodyParam = "j"
			break
		}

	}

	ogFnName := fnName
	ogDocParamsString := docParamsString
	if len(pagedRespType) > 0 {
		fnName += "AllPages"
		paramsString = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(paramsString, "pageToken string", ""), "limit int", ""), ", ,", ""))
		delete(params, "page_token")
		delete(params, "limit")
	}

	fmt.Printf("writing method %q for path %q -> %q\n", method, path, fnName)

	var description bytes.Buffer
	// Write the description for the method.
	if o.Summary != "" {
		fmt.Fprintf(&description, "// %s: %s\n", fnName, o.Summary)
	} else {
		fmt.Fprintf(&description, "// %s\n", fnName)
	}
	if o.Description != "" {
		fmt.Fprintln(&description, "//")
		fmt.Fprintf(&description, "// %s\n", strings.ReplaceAll(o.Description, "\n", "\n// "))
	}
	if pageResult && !isGetAllPages {
		fmt.Fprintf(&description, "//\n// To iterate over all pages, use the `%sAllPages` method, instead.\n", fnName)
	}
	if len(pagedRespType) > 0 {
		fmt.Fprintf(&description, "//\n// This method is a wrapper around the `%s` method.\n", ogFnName)
		fmt.Fprintf(&description, "// This method returns all the pages at once.\n")
	}
	if len(params) > 0 {
		fmt.Fprintf(&description, "//\n// Parameters:\n")
		keys := make([]string, 0)
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, name := range keys {
			t := params[name]
			if t.Description != "" {
				fmt.Fprintf(&description, "//\t- `%s`: %s\n", strcase.ToLowerCamel(name), strings.ReplaceAll(t.Description, "\n", "\n//\t\t"))
			} else {
				fmt.Fprintf(&description, "//\t- `%s`\n", strcase.ToLowerCamel(name))
			}
		}
	}

	if reqBodyDescription != "" && reqBodyParam != "nil" {
		fmt.Fprintf(&description, "//\t`%s`: %s\n", reqBodyParam, strings.ReplaceAll(reqBodyDescription, "\n", "\n// "))
	}

	// Write the description to the file.
	fmt.Fprint(f, description.String())

	// Write the method.
	if respType != "" && respType != "ConsumeCredentialsResponse" && respType != "LoginResponse" {
		fmt.Fprintf(f, "func (c *Client) %s(%s) (*%s, error) {\n",
			fnName,
			paramsString,
			respType)
	} else {
		fmt.Fprintf(f, "func (c *Client) %s(%s) (error) {\n",
			fnName,
			paramsString)
	}

	if len(pagedRespType) > 0 {
		// We want to just recursively call the method for each page.
		fmt.Fprintf(f, `
			var allPages %s
			pageToken := ""
			limit := 100
			for {
				page, err := c.%s(%s)
				if err != nil {
					return nil, err
				}
				allPages = append(allPages, page.Items...)
				if  page.NextPage == "" || page.NextPage == pageToken {
					break
				}
				pageToken = page.NextPage
			}

			return &allPages, nil
		}`, pagedRespType, ogFnName, ogDocParamsString)

		// Return early.
		return nil
	}

	// Create the url.
	fmt.Fprintln(f, "// Create the url.")
	fmt.Fprintf(f, "path := %q\n", cleanPath(path))
	fmt.Fprintln(f, "uri := resolveRelative(c.server, path)")

	if o.RequestBody != nil {
		for mt := range o.RequestBody.Value.Content {
			// TODO: Handle other content types
			if mt != "application/json" {
				break
			}

			// We need to encode the request body as json.
			fmt.Fprintln(f, "// Encode the request body as json.")
			fmt.Fprintln(f, "b := new(bytes.Buffer)")
			fmt.Fprintln(f, "if err := json.NewEncoder(b).Encode(j); err != nil {")
			if respType != "" {
				r := `return nil, fmt.Errorf("encoding json body request failed: %v", err)`
				fmt.Fprintln(f, r)
			} else {
				r := `return fmt.Errorf("encoding json body request failed: %v", err)`
				fmt.Fprintln(f, r)
			}
			fmt.Fprintln(f, "}")
			reqBodyParam = "b"
			break
		}

	}

	// Create the request.
	fmt.Fprintln(f, "// Create the request.")

	fmt.Fprintf(f, "req, err := http.NewRequest(%q, uri, %s)\n", method, reqBodyParam)
	fmt.Fprintln(f, "if err != nil {")
	if respType != "" && respType != "ConsumeCredentialsResponse" && respType != "LoginResponse" {
		r := `return nil, fmt.Errorf("error creating request: %v", err)`
		fmt.Fprintln(f, r)
	} else {
		r := `return fmt.Errorf("error creating request: %v", err)`
		fmt.Fprintln(f, r)
	}
	fmt.Fprintln(f, "}")

	// Add the parameters to the url.
	if len(params) > 0 {
		fmt.Fprintln(f, "// Add the parameters to the url.")
		fmt.Fprintln(f, "if err := expandURL(req.URL, map[string]string{")
		// Iterate over all the paths in the spec and write the types.
		// We want to ensure we keep the order so the diffs don't look like shit.
		keys := make([]string, 0)
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, name := range keys {
			p := params[name]
			t := printType(name, p.Schema)
			n := printPropertyLower(name)
			if t == "string" {
				fmt.Fprintf(f, "	%q: %s,\n", name, n)
			} else if t == "int" {
				fmt.Fprintf(f, "	%q: strconv.Itoa(%s),\n", name, n)
			} else if t == "*time.Time" {
				fmt.Fprintf(f, "	%q: %s.String(),\n", name, n)
			} else {
				fmt.Fprintf(f, "	%q: string(%s),\n", name, n)
			}
		}
		fmt.Fprintln(f, "}); err != nil {")
		if respType != "" && respType != "ConsumeCredentialsResponse" && respType != "LoginResponse" {
			r := `return nil, fmt.Errorf("expanding URL with parameters failed: %v", err)`
			fmt.Fprintln(f, r)
		} else {
			r := `return fmt.Errorf("expanding URL with parameters failed: %v", err)`
			fmt.Fprintln(f, r)
		}
		fmt.Fprintln(f, "}")
	}

	// Send the request.
	fmt.Fprintln(f, "// Send the request.")
	fmt.Fprintln(f, "resp, err := c.client.Do(req)")
	fmt.Fprintln(f, "if err != nil {")
	if respType != "" && respType != "ConsumeCredentialsResponse" && respType != "LoginResponse" {
		r := `return nil, fmt.Errorf("error sending request: %v", err)`
		fmt.Fprintln(f, r)
	} else {
		r := `return fmt.Errorf("error sending request: %v", err)`
		fmt.Fprintln(f, r)
	}
	fmt.Fprintln(f, "}")
	fmt.Fprintln(f, "defer resp.Body.Close()")

	// Check the response if there were any errors.
	fmt.Fprintln(f, "// Check the response.")
	fmt.Fprintln(f, "if err := checkResponse(resp); err != nil {")
	if respType != "" && respType != "ConsumeCredentialsResponse" && respType != "LoginResponse" {
		fmt.Fprintln(f, "return nil, err")
	} else {
		fmt.Fprintln(f, "return err")
	}
	fmt.Fprintln(f, "}")

	if respType != "" && respType != "ConsumeCredentialsResponse" && respType != "LoginResponse" {
		// Decode the body from the response.
		fmt.Fprintln(f, "// Decode the body from the response.")
		fmt.Fprintln(f, "if resp.Body == nil {")
		fmt.Fprintln(f, `return nil, errors.New("request returned an empty body in the response")`)
		fmt.Fprintln(f, "}")

		fmt.Fprintf(f, "var body %s\n", respType)
		fmt.Fprintln(f, "if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {")
		r := `return nil, fmt.Errorf("error decoding response body: %v", err)`
		fmt.Fprintln(f, r)
		fmt.Fprintln(f, "}")

		// Return the response.
		fmt.Fprintln(f, "// Return the response.")
		fmt.Fprintln(f, "return &body, nil")
	} else {
		fmt.Fprintln(f, "// Return.")
		fmt.Fprintln(f, "return nil")
	}

	// Close the method.
	fmt.Fprintln(f, "}")
	fmt.Fprintln(f, "")

	if pageResult && !isGetAllPages {
		// Run the method again with get all pages.
		writeMethod(doc, f, method, path, o, true)
	}

	return nil
}

func getSuccessResponseType(o *openapi3.Operation, isGetAllPages bool) (string, string, error) {
	for name, response := range o.Responses {
		if name == "default" {
			name = "200"
		}

		statusCode, err := strconv.Atoi(strings.ReplaceAll(name, "XX", "00"))
		if err != nil {
			return "", "", fmt.Errorf("error converting %q to an integer: %v", name, err)
		}

		if statusCode < 200 || statusCode >= 300 {
			// Continue early, we just want the successful response.
			continue
		}

		if response.Ref != "" {
			fmt.Printf("[WARN] TODO: skipping response for %q, since it is a reference: %q\n", name, response.Ref)
			continue
		}

		for _, content := range response.Value.Content {
			getAllPagesType := ""
			if isGetAllPages {

				if items, ok := content.Schema.Value.Properties["items"]; ok {
					getAllPagesType = printType("", items)
				} else {
					fmt.Printf("[WARN] TODO: skipping response for %q, since it is a get all pages response and has no `items` property:\n%#v\n", o.OperationID, content.Schema.Value.Properties)
				}
			}
			if content.Schema.Ref != "" {
				return getReferenceSchema(content.Schema), getAllPagesType, nil
			}

			return fmt.Sprintf("%sResponse", strcase.ToCamel(o.OperationID)), getAllPagesType, nil
		}
	}

	return "", "", nil
}

// cleanPath returns the path as a function we can use for a go template.
func cleanPath(path string) string {
	path = strings.Replace(path, "{", "{{.", -1)
	return strings.Replace(path, "}", "}}", -1)
}
