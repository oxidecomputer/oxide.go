package main

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/iancoleman/strcase"
)

type methodTemplate struct {
	Description     string
	HTTPMethod      string
	FunctionName    string
	WrappedFunction string
	WrappedParams   string // temporary field
	ResponseType    string
	SignatureParams map[string]string
	Summary         string
	Path            string
	PathParams      []string
	ParamsString    string //temporary field
	IsList          bool
	IsListAll       bool
	HasDescription  bool
	HasParams       bool
	HasBody         bool
	HasSummary      bool
}

var tmpGenFile *os.File

// Generate the paths.go file.
func generatePaths(file string, spec *openapi3.T) error {
	f, err := openGeneratedFile(file)
	if err != nil {
		return err
	}
	defer f.Close()

	// TODO: Remove when swap is over
	// create temp file for swapping over generation to use templates
	//r := rand.Int()
	tmpFile := "./test_utils/tpl_method" //+ fmt.Sprint(r)
	tf, err := os.OpenFile(tmpFile, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer tf.Close()
	tmpGenFile = tf
	// END of temp file

	// Iterate over all the paths in the spec and write the types.
	// We want to ensure we keep the order so the diffs don't look like shit.
	keys := make([]string, 0)
	for k := range spec.Paths {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, path := range keys {
		p := spec.Paths[path]
		if p.Ref != "" {
			fmt.Printf("[WARN] TODO: skipping path for %q, since it is a reference\n", path)
			continue
		}

		str, err := writePath(spec, path, p)
		if err != nil {
			return err
		}
		fmt.Fprint(f, str)
	}

	return nil
}

// writePath writes the given path as an http request to the given file.
func writePath(spec *openapi3.T, path string, p *openapi3.PathItem) (string, error) {
	var pathStr string
	if p.Get != nil {
		str, err := buildGetMethod(spec, path, p.Get, false)
		if err != nil {
			return "", err
		}
		pathStr = pathStr + fmt.Sprint(str)
	}

	if p.Post != nil {
		str, err := writeMethod(spec, http.MethodPost, path, p.Post, false)
		if err != nil {
			return "", err
		}
		pathStr = pathStr + fmt.Sprint(str)
	}

	if p.Put != nil {
		str, err := writeMethod(spec, http.MethodPut, path, p.Put, false)
		if err != nil {
			return "", err
		}
		pathStr = pathStr + fmt.Sprint(str)
	}

	if p.Delete != nil {
		str, err := writeMethod(spec, http.MethodDelete, path, p.Delete, false)
		if err != nil {
			return "", err
		}
		pathStr = pathStr + fmt.Sprint(str)
	}

	if p.Patch != nil {
		str, err := writeMethod(spec, http.MethodPatch, path, p.Patch, false)
		if err != nil {
			return "", err
		}
		pathStr = pathStr + fmt.Sprint(str)
	}

	if p.Head != nil {
		str, err := writeMethod(spec, http.MethodHead, path, p.Head, false)
		if err != nil {
			return "", err
		}
		pathStr = pathStr + fmt.Sprint(str)
	}

	if p.Options != nil {
		str, err := writeMethod(spec, http.MethodOptions, path, p.Options, false)
		if err != nil {
			return "", err
		}
		pathStr = pathStr + fmt.Sprint(str)
	}

	return pathStr, nil
}

func buildGetMethod(spec *openapi3.T, path string, o *openapi3.Operation, isGetAllPages bool) (string, error) {
	var methodStr string

	respType, pagedRespType, err := getSuccessResponseType(o, isGetAllPages)
	if err != nil {
		return "", err
	}

	if len(o.Tags) == 0 {
		fmt.Printf("[WARN] TODO: skipping operation %q, since it has no tag\n", o.OperationID)
		return "", nil
	}
	tag := strcase.ToCamel(o.Tags[0])

	if tag == "Hidden" {
		// return early.
		return "", nil
	}

	fnName := strcase.ToCamel(o.OperationID)

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

		paramName := strcase.ToLowerCamel(p.Value.Name)

		// Check if we have a page result.
		if isPageParam(paramName) {
			pageResult = true
		}

		params[p.Value.Name] = p.Value
		paramsString += fmt.Sprintf("%s %s, ", paramName, convertToValidGoType(p.Value.Name, p.Value.Schema))
		if index == len(o.Parameters)-1 {
			docParamsString += paramName
		} else {
			docParamsString += fmt.Sprintf("%s, ", paramName)
		}
	}

	if pageResult && isGetAllPages && len(pagedRespType) > 0 {
		respType = pagedRespType
	}

	ogFnName := fnName
	ogDocParamsString := docParamsString
	if isGetAllPages {
		fnName += "AllPages"
		paramsString = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(paramsString, "pageToken string", ""), "limit int", ""), ", ,", ""))
		delete(params, "page_token")
		delete(params, "limit")
	}

	isList := pageResult && !isGetAllPages

	// Use little template testing function
	// Only for development
	tmpPath := cleanPath(path)
	if err := descriptionTplWrite(
		fnName,
		ogFnName,
		respType,
		paramsString,
		ogDocParamsString,
		// TODO: Something weird happens with ip-pools path
		tmpPath,
		"GET",
		o,
		params,
		isGetAllPages,
		isList,
		o.RequestBody != nil, // If request body is not nil then it has a request body
	); err != nil {
		return "", err
	}

	// Write the description to the file.
	methodStr = methodStr + descriptionTpl(fnName, ogFnName, o, params, isGetAllPages, isList)

	// Write the method.

	// Presence of a "default" response means there is no response type.
	// No response should be returned in this case
	if respType != "" && o.Responses.Default() == nil {
		methodStr = methodStr + fmt.Sprintf("func (c *Client) %s(%s) (*%s, error) {\n",
			fnName,
			paramsString,
			respType)
	} else {
		methodStr = methodStr + fmt.Sprintf("func (c *Client) %s(%s) (error) {\n",
			fnName,
			paramsString)
	}

	// This only applies to "ListAll" methods
	if isGetAllPages {
		// We want to just recursively call the method for each page.
		methodStr = methodStr + fmt.Sprintf(`
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
		return methodStr, nil
	}

	// Create the url.
	methodStr = methodStr + fmt.Sprintln("// Create the url.")
	methodStr = methodStr + fmt.Sprintf("path := %q\n", cleanPath(path))
	methodStr = methodStr + fmt.Sprintln("uri := resolveRelative(c.server, path)")

	// Create the request.
	methodStr = methodStr + fmt.Sprintln("// Create the request.")
	methodStr = methodStr + fmt.Sprintln("req, err := http.NewRequest(\"GET\", uri, nil)")
	methodStr = methodStr + fmt.Sprintln("if err != nil {")

	// Presence of a "default" response means there is no response type.
	// No response should be returned in this case
	if respType != "" && o.Responses.Default() == nil {
		r := `return nil, fmt.Errorf("error creating request: %v", err)`
		methodStr = methodStr + fmt.Sprintln(r)
	} else {
		r := `return fmt.Errorf("error creating request: %v", err)`
		methodStr = methodStr + fmt.Sprintln(r)
	}
	methodStr = methodStr + fmt.Sprintln("}")

	// Add the parameters to the url.
	if len(params) > 0 {
		methodStr = methodStr + fmt.Sprintln("// Add the parameters to the url.")
		methodStr = methodStr + fmt.Sprintln("if err := expandURL(req.URL, map[string]string{")

		// Iterate over all the paths in the spec and write the types.
		// We want to ensure we keep the order so the diffs don't look like shit.
		keys := make([]string, 0)
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, name := range keys {
			p := params[name]
			t := convertToValidGoType(name, p.Schema)
			n := strcase.ToLowerCamel(name)
			if t == "string" {
				methodStr = methodStr + fmt.Sprintf("	%q: %s,\n", name, n)
			} else if t == "int" {
				methodStr = methodStr + fmt.Sprintf("	%q: strconv.Itoa(%s),\n", name, n)
			} else if t == "*time.Time" {
				methodStr = methodStr + fmt.Sprintf("	%q: %s.String(),\n", name, n)
			} else {
				methodStr = methodStr + fmt.Sprintf("	%q: string(%s),\n", name, n)
			}
		}
		methodStr = methodStr + fmt.Sprintln("}); err != nil {")
		// Presence of a "default" response means there is no response type.
		// No response should be returned in this case
		if respType != "" && o.Responses.Default() == nil {
			r := `return nil, fmt.Errorf("expanding URL with parameters failed: %v", err)`
			methodStr = methodStr + fmt.Sprintln(r)
		} else {
			r := `return fmt.Errorf("expanding URL with parameters failed: %v", err)`
			methodStr = methodStr + fmt.Sprintln(r)
		}
		methodStr = methodStr + fmt.Sprintln("}")
	}

	// Send the request.
	methodStr = methodStr + fmt.Sprintln("// Send the request.")
	methodStr = methodStr + fmt.Sprintln("resp, err := c.client.Do(req)")
	methodStr = methodStr + fmt.Sprintln("if err != nil {")

	if respType != "" && o.Responses.Default() == nil {
		r := `return nil, fmt.Errorf("error sending request: %v", err)`
		methodStr = methodStr + fmt.Sprintln(r)
	} else {
		r := `return fmt.Errorf("error sending request: %v", err)`
		methodStr = methodStr + fmt.Sprintln(r)
	}
	methodStr = methodStr + fmt.Sprintln("}")
	methodStr = methodStr + fmt.Sprintln("defer resp.Body.Close()")

	// Check the response if there were any errors.
	methodStr = methodStr + fmt.Sprintln("// Check the response.")
	methodStr = methodStr + fmt.Sprintln("if err := checkResponse(resp); err != nil {")

	if respType != "" && o.Responses.Default() == nil {
		methodStr = methodStr + fmt.Sprintln("return nil, err")
	} else {
		methodStr = methodStr + fmt.Sprintln("return err")
	}
	methodStr = methodStr + fmt.Sprintln("}")

	if respType != "" && o.Responses.Default() == nil {
		// Decode the body from the response.
		methodStr = methodStr + fmt.Sprintln("// Decode the body from the response.")
		methodStr = methodStr + fmt.Sprintln("if resp.Body == nil {")
		methodStr = methodStr + fmt.Sprintln(`return nil, errors.New("request returned an empty body in the response")`)
		methodStr = methodStr + fmt.Sprintln("}")

		methodStr = methodStr + fmt.Sprintf("var body %s\n", respType)
		methodStr = methodStr + fmt.Sprintln("if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {")
		r := `return nil, fmt.Errorf("error decoding response body: %v", err)`
		methodStr = methodStr + fmt.Sprintln(r)

		methodStr = methodStr + fmt.Sprintln("}")

		// Return the response.
		methodStr = methodStr + fmt.Sprintln("// Return the response.")
		methodStr = methodStr + fmt.Sprintln("return &body, nil")
	} else {
		methodStr = methodStr + fmt.Sprintln("// Return.")
		methodStr = methodStr + fmt.Sprintln("return nil")
	}

	// Close the method.
	methodStr = methodStr + fmt.Sprintln("}")
	methodStr = methodStr + fmt.Sprintln("")

	if pageResult && !isGetAllPages {
		// Run the method again with get all pages.
		str, err := buildGetMethod(spec, path, o, true)
		if err != nil {
			return "", err
		}

		methodStr = methodStr + str
	}

	return methodStr, nil
}

func writeMethod(spec *openapi3.T, method string, path string, o *openapi3.Operation, isGetAllPages bool) (string, error) {
	var methodStr string

	respType, _, err := getSuccessResponseType(o, isGetAllPages)
	if err != nil {
		return "", err
	}

	if len(o.Tags) == 0 {
		fmt.Printf("[WARN] TODO: skipping operation %q, since it has no tag\n", o.OperationID)
		return "", nil
	}
	tag := strcase.ToCamel(o.Tags[0])

	if tag == "Hidden" {
		// return early.
		return "", nil
	}

	fnName := strcase.ToCamel(o.OperationID)

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

		paramName := strcase.ToLowerCamel(p.Value.Name)

		// Check if we have a page result.
		if isPageParam(paramName) && method == http.MethodGet {
			pageResult = true
		}

		params[p.Value.Name] = p.Value
		paramsString += fmt.Sprintf("%s %s, ", paramName, convertToValidGoType(p.Value.Name, p.Value.Schema))
		if index == len(o.Parameters)-1 {
			docParamsString += paramName
		} else {
			docParamsString += fmt.Sprintf("%s, ", paramName)
		}
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

			typeName := convertToValidGoType("", r.Schema)

			paramsString += "j *" + typeName

			if len(docParamsString) > 0 {
				docParamsString += ", "
			}
			docParamsString += "body"

			reqBodyParam = "j"
			break
		}

	}

	// Use little template testing function
	// Only for development
	if err := descriptionTplWrite(
		fnName,
		"",
		respType,
		paramsString,
		"",
		cleanPath(path),
		method,
		o,
		params,
		isGetAllPages,
		false,
		o.RequestBody != nil, // If request body is not nil then it has a request body
	); err != nil {
		return "", err
	}

	methodStr = methodStr + descriptionTpl(fnName, "", o, params, isGetAllPages, false)

	// TODO: Add this to the description template somehow
	if reqBodyDescription != "" && reqBodyParam != "nil" {
		methodStr = methodStr + fmt.Sprintf("//\t`%s`: %s\n", reqBodyParam, strings.ReplaceAll(reqBodyDescription, "\n", "\n// "))
	}

	// Write the method.
	if respType != "" && o.Responses.Default() == nil {
		methodStr = methodStr + fmt.Sprintf("func (c *Client) %s(%s) (*%s, error) {\n",
			fnName,
			paramsString,
			respType)
	} else {
		methodStr = methodStr + fmt.Sprintf("func (c *Client) %s(%s) (error) {\n",
			fnName,
			paramsString)
	}

	// Create the url.
	methodStr = methodStr + fmt.Sprintln("// Create the url.")
	methodStr = methodStr + fmt.Sprintf("path := %q\n", cleanPath(path))
	methodStr = methodStr + fmt.Sprintln("uri := resolveRelative(c.server, path)")

	if o.RequestBody != nil {
		for mt := range o.RequestBody.Value.Content {
			// TODO: Handle other content types
			if mt != "application/json" {
				break
			}

			// We need to encode the request body as json.
			methodStr = methodStr + fmt.Sprintln("// Encode the request body as json.")
			methodStr = methodStr + fmt.Sprintln("b := new(bytes.Buffer)")
			methodStr = methodStr + fmt.Sprintln("if err := json.NewEncoder(b).Encode(j); err != nil {")

			if respType != "" {
				r := `return nil, fmt.Errorf("encoding json body request failed: %v", err)`
				methodStr = methodStr + fmt.Sprintln(r)
			} else {
				r := `return fmt.Errorf("encoding json body request failed: %v", err)`
				methodStr = methodStr + fmt.Sprintln(r)
			}

			methodStr = methodStr + fmt.Sprintln("}")

			reqBodyParam = "b"
			break
		}
	}

	// Create the request.
	methodStr = methodStr + fmt.Sprintln("// Create the request.")
	methodStr = methodStr + fmt.Sprintf("req, err := http.NewRequest(%q, uri, %s)\n", method, reqBodyParam)
	methodStr = methodStr + fmt.Sprintln("if err != nil {")

	if respType != "" && o.Responses.Default() == nil {
		r := `return nil, fmt.Errorf("error creating request: %v", err)`
		methodStr = methodStr + fmt.Sprintln(r)
	} else {
		r := `return fmt.Errorf("error creating request: %v", err)`
		methodStr = methodStr + fmt.Sprintln(r)
	}
	methodStr = methodStr + fmt.Sprintln("}")

	// Add the parameters to the url.
	if len(params) > 0 {
		methodStr = methodStr + fmt.Sprintln("// Add the parameters to the url.")
		methodStr = methodStr + fmt.Sprintln("if err := expandURL(req.URL, map[string]string{")

		// Iterate over all the paths in the spec and write the types.
		// We want to ensure we keep the order so the diffs don't look like shit.
		keys := make([]string, 0)
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, name := range keys {
			p := params[name]
			t := convertToValidGoType(name, p.Schema)
			n := strcase.ToLowerCamel(name)
			if t == "string" {
				methodStr = methodStr + fmt.Sprintf("	%q: %s,\n", name, n)
			} else if t == "int" {
				methodStr = methodStr + fmt.Sprintf("	%q: strconv.Itoa(%s),\n", name, n)
			} else if t == "*time.Time" {
				methodStr = methodStr + fmt.Sprintf("	%q: %s.String(),\n", name, n)
			} else {
				methodStr = methodStr + fmt.Sprintf("	%q: string(%s),\n", name, n)
			}
		}
		methodStr = methodStr + fmt.Sprintln("}); err != nil {")
		if respType != "" && o.Responses.Default() == nil {
			r := `return nil, fmt.Errorf("expanding URL with parameters failed: %v", err)`
			methodStr = methodStr + fmt.Sprintln(r)
		} else {
			r := `return fmt.Errorf("expanding URL with parameters failed: %v", err)`
			methodStr = methodStr + fmt.Sprintln(r)
		}
		methodStr = methodStr + fmt.Sprintln("}")
	}

	// Send the request.
	methodStr = methodStr + fmt.Sprintln("// Send the request.")
	methodStr = methodStr + fmt.Sprintln("resp, err := c.client.Do(req)")
	methodStr = methodStr + fmt.Sprintln("if err != nil {")

	if respType != "" && o.Responses.Default() == nil {
		r := `return nil, fmt.Errorf("error sending request: %v", err)`
		methodStr = methodStr + fmt.Sprintln(r)
	} else {
		r := `return fmt.Errorf("error sending request: %v", err)`
		methodStr = methodStr + fmt.Sprintln(r)
	}
	methodStr = methodStr + fmt.Sprintln("}")
	methodStr = methodStr + fmt.Sprintln("defer resp.Body.Close()")

	// Check the response if there were any errors.
	methodStr = methodStr + fmt.Sprintln("// Check the response.")
	methodStr = methodStr + fmt.Sprintln("if err := checkResponse(resp); err != nil {")

	if respType != "" && o.Responses.Default() == nil {
		methodStr = methodStr + fmt.Sprintln("return nil, err")
	} else {
		methodStr = methodStr + fmt.Sprintln("return err")
	}
	methodStr = methodStr + fmt.Sprintln("}")

	if respType != "" && o.Responses.Default() == nil {
		// Decode the body from the response.
		methodStr = methodStr + fmt.Sprintln("// Decode the body from the response.")
		methodStr = methodStr + fmt.Sprintln("if resp.Body == nil {")
		methodStr = methodStr + fmt.Sprintln(`return nil, errors.New("request returned an empty body in the response")`)
		methodStr = methodStr + fmt.Sprintln("}")

		methodStr = methodStr + fmt.Sprintf("var body %s\n", respType)
		methodStr = methodStr + fmt.Sprintln("if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {")
		r := `return nil, fmt.Errorf("error decoding response body: %v", err)`
		methodStr = methodStr + fmt.Sprintln(r)

		methodStr = methodStr + fmt.Sprintln("}")

		// Return the response.
		methodStr = methodStr + fmt.Sprintln("// Return the response.")
		methodStr = methodStr + fmt.Sprintln("return &body, nil")
	} else {
		methodStr = methodStr + fmt.Sprintln("// Return.")
		methodStr = methodStr + fmt.Sprintln("return nil")
	}

	// Close the method.
	methodStr = methodStr + fmt.Sprintln("}")
	methodStr = methodStr + fmt.Sprintln("")

	if pageResult && !isGetAllPages {
		// Run the method again with get all pages.
		str, err := writeMethod(spec, method, path, o, true)
		if err != nil {
			return "", err
		}

		methodStr = methodStr + str
	}

	return methodStr, nil
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
					getAllPagesType = convertToValidGoType("", items)
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

func descriptionTpl(fnName, ogFnName string, o *openapi3.Operation, params map[string]*openapi3.Parameter, isListAll, isList bool) string {
	var description string

	if o.Summary != "" {
		description = description + fmt.Sprintf("// %s: %s\n", fnName, o.Summary)
	} else {
		description = description + fmt.Sprintf("// %s\n", fnName)
	}
	if o.Description != "" {
		description = description + fmt.Sprintln("//")
		description = description + fmt.Sprintf("// %s\n", strings.ReplaceAll(o.Description, "\n", "\n// "))
	}

	if isListAll {
		description = description + fmt.Sprintf("//\n// This method is a wrapper around the `%s` method.\n", ogFnName)
		description = description + "// This method returns all the pages at once.\n"
	} else if isList {
		description = description + fmt.Sprintf("//\n// To iterate over all pages, use the `%sAllPages` method, instead.\n", fnName)
	}

	if len(params) > 0 {
		description = description + "//\n// Parameters:\n"
		keys := make([]string, 0)
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, name := range keys {
			t := params[name]
			if t.Description != "" {
				description = description + fmt.Sprintf("//   - `%s`: %s\n", strcase.ToLowerCamel(name), strings.ReplaceAll(t.Description, "\n", "\n//      "))
			} else {
				description = description + fmt.Sprintf("//   - `%s`\n", strcase.ToLowerCamel(name))
			}
		}
	}

	return description
}

func descriptionTplWrite(fnName, wrappedFn, respType, pStr, wrappedParams, path, method string, o *openapi3.Operation, params map[string]*openapi3.Parameter, isListAll, isList, hasBody bool) error {
	//r := rand.Int()

	sigParams := make(map[string]string)
	if len(params) > 0 {
		keys := make([]string, 0)
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, name := range keys {
			t := params[name]
			sigParams[strcase.ToLowerCamel(name)] = t.Description
		}
	}

	pathParams := []string{}
	if len(params) > 0 {
		// Iterate over all the paths in the spec and write the types.
		// We want to ensure we keep the order so the diffs don't change.
		keys := make([]string, 0)
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, name := range keys {
			p := params[name]
			t := convertToValidGoType(name, p.Schema)
			n := strcase.ToLowerCamel(name)
			if t == "string" {
				pathParams = append(pathParams, fmt.Sprintf("%q: %s,", name, n))
			} else if t == "int" {
				pathParams = append(pathParams, fmt.Sprintf("%q: strconv.Itoa(%s),", name, n))
			} else if t == "*time.Time" {
				pathParams = append(pathParams, fmt.Sprintf("%q: %s.String(),", name, n))
			} else {
				pathParams = append(pathParams, fmt.Sprintf("%q: string(%s),", name, n))
			}
		}
	}

	config := methodTemplate{
		Description:     o.Description,
		HTTPMethod:      method,
		FunctionName:    fnName,
		WrappedFunction: wrappedFn,
		WrappedParams:   wrappedParams,
		ResponseType:    respType,
		SignatureParams: sigParams,
		Summary:         o.Summary,
		ParamsString:    pStr,
		Path:            path,
		PathParams:      pathParams,
		IsList:          isList,
		IsListAll:       isListAll,
		HasBody:         hasBody,
	}

	if len(params) > 0 {
		config.HasParams = true
	}

	if o.Summary != "" {
		config.HasSummary = true
	}

	if o.Description != "" {
		config.HasDescription = true
	}

	var t *template.Template
	var err error

	if config.IsListAll {
		t, err = template.ParseFiles("./templates/listall_method.tpl", "./templates/description.tpl")
		if err != nil {
			return err
		}
	} else if config.ResponseType == "" && config.HasBody {
		t, err = template.ParseFiles("./templates/no_resptype_body_method.tpl", "./templates/description.tpl")
		if err != nil {
			return err
		}
	} else if config.ResponseType == "" {
		t, err = template.ParseFiles("./templates/no_resptype_method.tpl", "./templates/description.tpl")
		if err != nil {
			return err
		}
	} else if config.HasBody {
		t, err = template.ParseFiles("./templates/resptype_body_method.tpl", "./templates/description.tpl")
		if err != nil {
			return err
		}
	} else {
		t, err = template.ParseFiles("./templates/resptype_method.tpl", "./templates/description.tpl")
		if err != nil {
			return err
		}
	}

	// Prints out an extra final IPPools thing in the end I DONT KNWO WHYYYY?????

	//	file := "./test_utils/tpl_method" + fmt.Sprint(r)
	//	f, err := os.Create(file)
	//	if err != nil {
	//		return err
	//	}
	//	defer f.Close()
	//	err = t.Execute(f, config)
	err = t.Execute(tmpGenFile, config)
	if err != nil {
		return err
	}

	return nil
}
