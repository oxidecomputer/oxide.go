//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/iancoleman/strcase"
)

var EnumStringTypes map[string][]string = map[string][]string{}

func main() {
	// TODO: actually host the spec here.
	/*uri := "https://api.oxide.computer"
	u, err := url.Parse(uri)
	if err != nil {
		fmt.Printf("error parsing url %q: %v\n", uri, err)
		os.Exit(1)
	}

	// Load the open API spec from the URI.
	doc, err := openapi3.NewLoader().LoadFromURI(u)
	if err != nil {
		fmt.Printf("error loading openAPI spec from %q: %v\n", uri, err)
		os.Exit(1)
	}*/

	// Load the open API spec from the file.
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("error getting current working directory: %v\n", err)
		os.Exit(1)
	}
	p := filepath.Join(wd, "spec.json")
	doc, err := openapi3.NewLoader().LoadFromFile(p)
	if err != nil {
		fmt.Printf("error loading openAPI spec from file %q: %v\n", p, err)
		os.Exit(1)
	}

	// Generate the client.go file.
	generateClient(doc)

	// Generate the types.go file.
	generateTypes(doc)

	// Generate the responses.go file.
	generateResponses(doc)

	// Generate the paths.go file.
	generatePaths(doc)

	clientInfo := `// Create a client with your token and host.
client, err := oxide.NewClient("$OXIDE_TOKEN", "your apps user agent", "$OXIDE_HOST")
if err != nil {
  panic(err)
}

// - OR -

// Create a new client with your token and host parsed from the environment
// variables: OXIDE_TOKEN, OXIDE_HOST.
client, err := oxide.NewClientFromEnv("your apps user agent")
if err != nil {
  panic(err)
}`

	doc.Info.Extensions["x-go"] = map[string]string{
		"install": "go get github.com/oxidecomputer/oxide.go",
		"client":  clientInfo,
	}

	// Write back out the new spec.
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		fmt.Printf("error marshalling openAPI spec: %v\n", err)
		os.Exit(1)
	}
	if err := ioutil.WriteFile(p, out, 0644); err != nil {
		fmt.Printf("error writing openAPI spec to %s: %v\n", p, err)
		os.Exit(1)
	}

}

// Generate the types.go file.
func generateTypes(doc *openapi3.T) {
	f := openGeneratedFile("types.go")
	defer f.Close()

	// Iterate over all the schema components in the spec and write the types.
	// We want to ensure we keep the order so the diffs don't look like shit.
	keys := make([]string, 0)
	for k := range doc.Components.Schemas {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		s := doc.Components.Schemas[name]
		if s.Ref != "" {
			fmt.Printf("[WARN] TODO: skipping type for %q, since it is a reference\n", name)
			continue
		}

		writeSchemaType(f, name, s.Value, "")
	}

	// Iterate over all the enum types and add in the slices.
	// We want to ensure we keep the order so the diffs don't look like shit.
	keys = make([]string, 0)
	for k := range EnumStringTypes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		enums := EnumStringTypes[name]
		// Make the enum a collection of the values.
		// Add a description.
		fmt.Fprintf(f, "// %s is the collection of all %s values.\n", makePlural(name), makeSingular(name))
		fmt.Fprintf(f, "var %s = []%s{\n", makePlural(name), makeSingular(name))
		// We want to keep the values in the same order as the enum.
		sort.Strings(enums)
		for _, enum := range enums {
			// Most likely, the enum values are strings.
			fmt.Fprintf(f, "\t%s,\n", strcase.ToCamel(fmt.Sprintf("%s_%s", makeSingular(name), enum)))
		}
		// Close the enum values.
		fmt.Fprintf(f, "}\n")
	}
}

// Generate the responses.go file.
func generateResponses(doc *openapi3.T) {
	f := openGeneratedFile("responses.go")
	defer f.Close()

	// Iterate over all the responses in the spec and write the types.
	// We want to ensure we keep the order so the diffs don't look like shit.
	keys := make([]string, 0)
	for k := range doc.Components.Responses {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		r := doc.Components.Responses[name]
		if r.Ref != "" {
			fmt.Printf("[WARN] TODO: skipping response for %q, since it is a reference\n", name)
			continue
		}

		writeResponseType(f, name, r.Value)
	}
}

// Generate the client.go file.
func generateClient(doc *openapi3.T) {
	f := openGeneratedFile("client.go")
	defer f.Close()

	fmt.Fprintf(f, `// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.oxide.computer for example.
	server string

	// Client is the *http.Client for performing requests.
	client *http.Client

	// token is the API token used for authentication.
	token string
`)

	for _, tag := range doc.Tags {
		if tag.Description != "" {
			fmt.Fprintf(f, "// %s: %s\n", strcase.ToCamel(tag.Name), tag.Description)
		}
		fmt.Fprintf(f, "%s\t*%sService\n", strcase.ToCamel(tag.Name), strcase.ToCamel(tag.Name))
	}

	// Close the struct.
	fmt.Fprintf(f, "}\n\n")

	for _, tag := range doc.Tags {
		if tag.Description != "" {
			fmt.Fprintf(f, "// %sService: %s\n", strcase.ToCamel(tag.Name), tag.Description)
		}
		fmt.Fprintf(f, "type %sService service\n\n", strcase.ToCamel(tag.Name))
	}
}

// Generate the paths.go file.
func generatePaths(doc *openapi3.T) {
	f := openGeneratedFile("paths.go")
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

		writePath(doc, f, path, p)
	}
}

func openGeneratedFile(filename string) *os.File {
	// Get the current working directory.
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("error getting current working directory: %v\n", err)
		os.Exit(1)
	}

	p := filepath.Join(cwd, filename)

	// Create the types.go file.
	// Open the file for writing.
	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Printf("error creating %q: %v\n", p, err)
		os.Exit(1)
	}

	// Add the header to the package.
	fmt.Fprintf(f, "// Code generated by `%s`. DO NOT EDIT.\n\n", filepath.Base(os.Args[0]))
	fmt.Fprintln(f, "package oxide")
	fmt.Fprintln(f, "")

	return f
}

func cleanFnName(name string, tag string, path string) string {
	name = printProperty(name)

	name = strings.ReplaceAll(name, strcase.ToCamel(tag), "")

	if strings.HasSuffix(tag, "s") {
		tag = strings.TrimSuffix(tag, "s")
		name = strings.ReplaceAll(name, strcase.ToCamel(tag), "")
	}

	name = strings.TrimPrefix(name, "Organization")
	name = strings.TrimPrefix(name, "Project")

	name = strings.ReplaceAll(name, "Vpc", "VPC")

	if strings.HasSuffix(name, "Get") && !strings.HasSuffix(path, "}") {
		name = fmt.Sprintf("%sList", strings.TrimSuffix(name, "Get"))
	}

	if strings.HasSuffix(name, "Post") {
		name = fmt.Sprintf("%sCreate", strings.TrimSuffix(name, "Post"))
	}

	if strings.HasPrefix(name, "s") {
		name = strings.TrimPrefix(name, "s")
	}

	if name == "RacksGetRack" {
		name = "GetRack"
	}

	if name == "SledsGetSled" {
		name = "GetSled"
	}

	if strings.HasPrefix(name, "Hardware") {
		name = strings.TrimPrefix(name, "Hardware")
	}

	// Routes paths wind up weird so clean them up.
	if strings.HasPrefix(name, "rs") {
		name = strings.TrimPrefix(name, "rs")
	}

	return name
}

// printProperty converts an object's property name to a valid Go identifier.
func printProperty(p string) string {
	c := strcase.ToCamel(p)
	if c == "Id" {
		c = "ID"
	} else if c == "Ncpus" {
		c = "NCPUs"
	} else if c == "IpAddress" {
		c = "IPAddress"
	} else if c == "UserId" {
		c = "UserID"
	} else if strings.Contains(c, "IdSortMode") {
		strings.ReplaceAll(c, "IdSortMode", "IDSortMode")
	} else if strings.HasPrefix(c, "Cpu") {
		c = strings.Replace(c, "Cpu", "CPU", 1)
	} else if strings.HasPrefix(c, "Vpc") {
		c = strings.Replace(c, "Vpc", "VPC", 1)
	} else if strings.HasPrefix(c, "Vpn") {
		c = strings.Replace(c, "Vpn", "VPN", 1)
	} else if strings.HasPrefix(c, "Ipv4") {
		c = strings.Replace(c, "Ipv4", "IPv4", 1)
	} else if strings.HasPrefix(c, "Ipv6") {
		c = strings.Replace(c, "Ipv6", "IPv6", 1)
	} else if strings.HasSuffix(c, "Id") {
		c = strings.TrimSuffix(c, "Id") + "ID"
	} else if strings.Contains(c, "Cpu") {
		c = strings.ReplaceAll(c, "Cpu", "CPU")
	} else if strings.HasPrefix(c, "SubnetsIps") {
		c = strings.ReplaceAll(c, "SubnetsIps", "SubnetsIPs")
	}

	// Before we remove the tag from the name, we want to remove any lingering "Vpc" strings.
	c = strings.Replace(c, "VPCFirewallRule", "FirewallRule", -1)
	c = strings.Replace(c, "VPCRouter", "Router", -1)
	c = strings.Replace(c, "VPCSubnet", "Subnet", -1)

	return c
}

func printPropertyLower(p string) string {
	s := strcase.ToLowerCamel(printProperty(p))

	if strings.HasPrefix(s, "vPC") {
		s = "vpc" + strings.TrimPrefix(s, "vPC")
	} else if strings.HasPrefix(s, "cPU") {
		s = "cpu" + strings.TrimPrefix(s, "cPU")
	} else if strings.HasPrefix(s, "iPv4") {
		s = "ipv4" + strings.TrimPrefix(s, "iPv4")
	} else if strings.HasPrefix(s, "iPv6") {
		s = "ipv6" + strings.TrimPrefix(s, "iPv6")
	}

	if s == "iD" {
		s = "id"
	} else if s == "iPAddress" {
		s = "ipAddress"
	} else if s == "iDSortMode" {
		s = "idSortMode"
	}

	return s
}

// printType converts a schema type to a valid Go type.
func printType(property string, r *openapi3.SchemaRef) string {
	s := r.Value
	t := s.Type

	// If we have a reference, just use that.
	if r.Ref != "" {
		return getReferenceSchema(r)
	}

	// See if we have an allOf.
	if s.AllOf != nil {
		if len(s.AllOf) > 1 {
			fmt.Printf("[WARN] TODO: allOf for %q has more than 1 item\n", property)
			return "TODO"
		}

		return printType(property, s.AllOf[0])
	}

	if t == "string" {
		reference := getReferenceSchema(r)
		if reference != "" {
			return reference
		}

		return formatStringType(s)
	} else if t == "integer" {
		return "int"
	} else if t == "number" {
		return "float64"
	} else if t == "boolean" {
		return "bool"
	} else if t == "array" {
		reference := getReferenceSchema(s.Items)
		if reference != "" {
			return fmt.Sprintf("[]%s", reference)
		}

		// TODO: handle if it is not a reference.
		return "[]TODO"
	} else if t == "object" {
		// Most likely this is a local object, we will handle it.
		return strcase.ToCamel(property)
	}

	fmt.Printf("[WARN] TODO: skipping type %q for %q, marking as interface{}\n", t, property)
	return "interface{}"
}

// writePath writes the given path as an http request to the given file.
func writePath(doc *openapi3.T, f *os.File, path string, p *openapi3.PathItem) {
	if p.Get != nil {
		writeMethod(doc, f, http.MethodGet, path, p.Get, false)
	}

	if p.Post != nil {
		writeMethod(doc, f, http.MethodPost, path, p.Post, false)
	}

	if p.Put != nil {
		writeMethod(doc, f, http.MethodPut, path, p.Put, false)
	}

	if p.Delete != nil {
		writeMethod(doc, f, http.MethodDelete, path, p.Delete, false)
	}

	if p.Patch != nil {
		writeMethod(doc, f, http.MethodPatch, path, p.Patch, false)
	}

	if p.Head != nil {
		writeMethod(doc, f, http.MethodHead, path, p.Head, false)
	}

	if p.Options != nil {
		writeMethod(doc, f, http.MethodOptions, path, p.Options, false)
	}
}

func writeMethod(doc *openapi3.T, f *os.File, method string, path string, o *openapi3.Operation, isGetAllPages bool) {
	respType, pagedRespType := getSuccessResponseType(o, isGetAllPages)

	if len(o.Tags) == 0 {
		fmt.Printf("[WARN] TODO: skipping operation %q, since it has no tag\n", o.OperationID)
		return
	}
	tag := strcase.ToCamel(o.Tags[0])

	fnName := cleanFnName(o.OperationID, tag, path)

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
			docParamsString += fmt.Sprintf("%s", paramName)
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
		docParamsString = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(docParamsString, "pageToken", ""), "limit", ""), ", ,", ""))
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
	fmt.Fprintf(f, description.String())

	docInfo := map[string]string{
		"example":     fmt.Sprintf("%s", description.String()),
		"libDocsLink": fmt.Sprintf("https://pkg.go.dev/github.com/oxidecomputer/oxide.go/#%sService.%s", tag, fnName),
	}
	if isGetAllPages {
		og := doc.Paths[path].Get.Extensions["x-go"].(map[string]string)
		docInfo["example"] = fmt.Sprintf("%s\n\n// - OR -\n\n%s", og["example"], docInfo["example"])
		docInfo["libDocsLink"] = fmt.Sprintf("https://pkg.go.dev/github.com/oxidecomputer/oxide.go/#%sService.%s", tag, ogFnName)
	}

	// Write the method.
	if respType != "" {
		fmt.Fprintf(f, "func (s *%sService) %s(%s) (*%s, error) {\n",
			tag,
			fnName,
			paramsString,
			respType)
		docInfo["example"] += fmt.Sprintf("%s, err := client.%s.%s(%s)", strcase.ToLowerCamel(respType), tag, fnName, docParamsString)
	} else {
		fmt.Fprintf(f, "func (s *%sService) %s(%s) (error) {\n",
			tag,
			fnName,
			paramsString)
		docInfo["example"] += fmt.Sprintf(`if err := client.%s.%s(%s); err != nil {
	panic(err)
}`, tag, fnName, docParamsString)
	}

	if method == http.MethodGet {
		doc.Paths[path].Get.Extensions["x-go"] = docInfo
	} else if method == http.MethodPost {
		doc.Paths[path].Post.Extensions["x-go"] = docInfo
	} else if method == http.MethodPut {
		doc.Paths[path].Put.Extensions["x-go"] = docInfo
	} else if method == http.MethodDelete {
		doc.Paths[path].Delete.Extensions["x-go"] = docInfo
	} else if method == http.MethodPatch {
		doc.Paths[path].Patch.Extensions["x-go"] = docInfo
	}

	if len(pagedRespType) > 0 {
		// We want to just recursively call the method for each page.
		fmt.Fprintf(f, `
			var allPages %s
			pageToken := ""
			limit := 100
			for {
				page, err := s.%s(%s)
				if err != nil {
					return nil, err
				}
				allPages = append(allPages, page.Items...)
				if  page.NextPage == "" {
					break
				}
				pageToken = page.NextPage
			}

			return &allPages, nil
		}`, pagedRespType, ogFnName, ogDocParamsString)

		// Return early.
		return
	}

	// Create the url.
	fmt.Fprintln(f, "// Create the url.")
	fmt.Fprintf(f, "path := %q\n", cleanPath(path))
	fmt.Fprintln(f, "uri := resolveRelative(s.client.server, path)")

	if o.RequestBody != nil {
		for mt := range o.RequestBody.Value.Content {
			if mt != "application/json" {
				break
			}

			// We need to encode the request body as json.
			fmt.Fprintln(f, "// Encode the request body as json.")
			fmt.Fprintln(f, "b := new(bytes.Buffer)")
			fmt.Fprintln(f, "if err := json.NewEncoder(b).Encode(j); err != nil {")
			if respType != "" {
				fmt.Fprintln(f, `return nil, fmt.Errorf("encoding json body request failed: %v", err)`)
			} else {
				fmt.Fprintln(f, `return fmt.Errorf("encoding json body request failed: %v", err)`)
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
	if respType != "" {
		fmt.Fprintln(f, `return nil, fmt.Errorf("error creating request: %v", err)`)
	} else {
		fmt.Fprintln(f, `return fmt.Errorf("error creating request: %v", err)`)
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
			} else {
				fmt.Fprintf(f, "	%q: string(%s),\n", name, n)
			}
		}
		fmt.Fprintln(f, "}); err != nil {")
		if respType != "" {
			fmt.Fprintln(f, `return nil, fmt.Errorf("expanding URL with parameters failed: %v", err)`)
		} else {
			fmt.Fprintln(f, `return fmt.Errorf("expanding URL with parameters failed: %v", err)`)
		}
		fmt.Fprintln(f, "}")
	}

	// Send the request.
	fmt.Fprintln(f, "// Send the request.")
	fmt.Fprintln(f, "resp, err := s.client.client.Do(req)")
	fmt.Fprintln(f, "if err != nil {")
	if respType != "" {
		fmt.Fprintln(f, `return nil, fmt.Errorf("error sending request: %v", err)`)
	} else {
		fmt.Fprintln(f, `return fmt.Errorf("error sending request: %v", err)`)
	}
	fmt.Fprintln(f, "}")
	fmt.Fprintln(f, "defer resp.Body.Close()")

	// Check the response if there were any errors.
	fmt.Fprintln(f, "// Check the response.")
	fmt.Fprintln(f, "if err := checkResponse(resp); err != nil {")
	if respType != "" {
		fmt.Fprintln(f, "return nil, err")
	} else {
		fmt.Fprintln(f, "return err")
	}
	fmt.Fprintln(f, "}")

	if respType != "" {
		// Decode the body from the response.
		fmt.Fprintln(f, "// Decode the body from the response.")
		fmt.Fprintln(f, "if resp.Body == nil {")
		fmt.Fprintln(f, `return nil, errors.New("request returned an empty body in the response")`)
		fmt.Fprintln(f, "}")

		fmt.Fprintf(f, "var body %s\n", respType)
		fmt.Fprintln(f, "if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {")
		fmt.Fprintln(f, `return nil, fmt.Errorf("error decoding response body: %v", err)`)
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
}

// cleanPath returns the path as a function we can use for a go template.
func cleanPath(path string) string {
	path = strings.Replace(path, "{", "{{.", -1)
	return strings.Replace(path, "}", "}}", -1)
}

func getSuccessResponseType(o *openapi3.Operation, isGetAllPages bool) (string, string) {
	for name, response := range o.Responses {
		if name == "default" {
			name = "200"
		}

		statusCode, err := strconv.Atoi(strings.ReplaceAll(name, "XX", "00"))
		if err != nil {
			fmt.Printf("error converting %q to an integer: %v\n", name, err)
			os.Exit(1)
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
				return getReferenceSchema(content.Schema), getAllPagesType
			}

			return fmt.Sprintf("%sResponse", strcase.ToCamel(o.OperationID)), getAllPagesType
		}
	}

	return "", ""
}

// writeSchemaType writes a type definition for the given schema.
// The additional parameter is only used as a suffix for the type name.
// This is mostly for oneOf types.
func writeSchemaType(f *os.File, name string, s *openapi3.Schema, additionalName string) {
	otype := s.Type
	fmt.Printf("writing type for schema %q -> %s\n", name, otype)

	name = printProperty(name)
	typeName := strings.TrimSpace(fmt.Sprintf("%s%s", name, printProperty(additionalName)))

	if len(s.Enum) == 0 && s.OneOf == nil {
		// Write the type description.
		writeSchemaTypeDescription(typeName, s, f)
	}

	if otype == "string" {
		// If this is an enum, write the enum type.
		if len(s.Enum) > 0 {
			// Make sure we don't redeclare the enum type.
			if _, ok := EnumStringTypes[makeSingular(typeName)]; !ok {
				// Write the type description.
				writeSchemaTypeDescription(makeSingular(typeName), s, f)

				// Write the enum type.
				fmt.Fprintf(f, "type %s string\n", makeSingular(typeName))

				EnumStringTypes[makeSingular(typeName)] = []string{}
			}

			// Define the enum values.
			fmt.Fprintf(f, "const (\n")
			for _, v := range s.Enum {
				// Most likely, the enum values are strings.
				enum, ok := v.(string)
				if !ok {
					fmt.Printf("[WARN] TODO: enum value is not a string for %q -> %#v\n", name, v)
					continue
				}
				// Write the description of the constant.
				fmt.Fprintf(f, "// %s represents the %s `%q`.\n", strcase.ToCamel(fmt.Sprintf("%s_%s", makeSingular(name), enum)), makeSingular(name), enum)
				fmt.Fprintf(f, "\t%s %s = %q\n", strcase.ToCamel(fmt.Sprintf("%s_%s", makeSingular(name), enum)), makeSingular(name), enum)

				// Add the enum type to the list of enum types.
				EnumStringTypes[makeSingular(typeName)] = append(EnumStringTypes[makeSingular(typeName)], enum)
			}
			// Close the enum values.
			fmt.Fprintf(f, ")\n")

		} else {
			fmt.Fprintf(f, "type %s string\n", name)
		}
	} else if otype == "integer" {
		fmt.Fprintf(f, "type %s int\n", name)
	} else if otype == "number" {
		fmt.Fprintf(f, "type %s float64\n", name)
	} else if otype == "boolean" {
		fmt.Fprintf(f, "type %s bool\n", name)
	} else if otype == "array" {
		fmt.Fprintf(f, "type %s []%s\n", name, s.Items.Value.Type)
	} else if otype == "object" {
		recursive := false
		fmt.Fprintf(f, "type %s struct {\n", typeName)
		// We want to ensure we keep the order so the diffs don't look like shit.
		keys := make([]string, 0)
		for k := range s.Properties {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := s.Properties[k]
			// Check if we need to generate a type for this type.
			typeName := printType(k, v)

			if isLocalEnum(v) {
				recursive = true
				typeName = fmt.Sprintf("%s%s", name, printProperty(k))
			}

			if isLocalObject(v) {
				recursive = true
				fmt.Printf("[WARN] TODO: skipping object for %q -> %#v\n", name, v)
				typeName = fmt.Sprintf("%s%s", name, printProperty(k))
			}

			if v.Value.Description != "" {
				fmt.Fprintf(f, "\t// %s is %s\n", printProperty(k), toLowerFirstLetter(strings.ReplaceAll(v.Value.Description, "\n", "\n// ")))
			}
			fmt.Fprintf(f, "\t%s %s `json:\"%s,omitempty\" yaml:\"%s,omitempty\" tfsdk:\"%s\"`\n", printProperty(k), typeName, k, k, k)
		}
		fmt.Fprintf(f, "}\n")

		if recursive {
			// Add a newline at the end of the type.
			fmt.Fprintln(f, "")

			// Iterate over the properties and write the types, if we need to.
			for k, v := range s.Properties {
				if isLocalEnum(v) {
					writeSchemaType(f, fmt.Sprintf("%s%s", name, printProperty(k)), v.Value, "")
				}

				if isLocalObject(v) {
					writeSchemaType(f, fmt.Sprintf("%s%s", name, printProperty(k)), v.Value, "")
				}
			}
		}
	} else {
		if s.OneOf != nil {
			// We want to convert these to a different data type to be more idiomatic.
			// But first, we need to make sure we have a type for each one.
			var oneOfTypes []string
			var properties []string
			for _, v := range s.OneOf {
				// We want to iterate over the properties of the embedded object
				// and find the type that is a string.
				var typeName string

				// Iterate over all the schema components in the spec and write the types.
				// We want to ensure we keep the order so the diffs don't look like shit.
				keys := make([]string, 0)
				for k := range v.Value.Properties {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, prop := range keys {
					p := v.Value.Properties[prop]
					// We want to collect all the unique properties to create our global oneOf type.
					propertyName := printType(prop, p)

					propertyString := fmt.Sprintf("\t%s %s `json:\"%s,omitempty\" yaml:\"%s,omitempty\"`\n", printProperty(prop), propertyName, prop, prop)
					if !containsMatchFirstWord(properties, propertyString) {
						properties = append(properties, propertyString)
					}

					if p.Value.Type == "string" {
						if p.Value.Enum != nil {
							// We want to get the enum value.
							// Make sure there is only one.
							if len(p.Value.Enum) != 1 {
								fmt.Printf("[WARN] TODO: oneOf for %q -> %q enum %#v\n", name, prop, p.Value.Enum)
								continue
							}

							typeName = printProperty(p.Value.Enum[0].(string))
						}
					}

					if len(typeName) == 0 && len(keys) == 1 && v.Value.Required != nil && len(v.Value.Required) == 1 {
						typeName = printProperty(v.Value.Required[0])
					}
				}

				// Basically all of these will have one type embedded in them that is a
				// string and the type, since these come from a Rust sum type.
				oneOfType := fmt.Sprintf("%s%s", name, typeName)
				writeSchemaType(f, name, v.Value, typeName)
				// Add it to our array.
				oneOfTypes = append(oneOfTypes, oneOfType)
			}

			// Now let's create the global oneOf type.
			// Write the type description.
			writeSchemaTypeDescription(typeName, s, f)
			fmt.Fprintf(f, "type %s struct {\n", typeName)
			// Iterate over the properties and write the types, if we need to.
			for _, p := range properties {
				fmt.Fprintf(f, p)
			}
			// Close the struct.
			fmt.Fprintf(f, "}\n")

		} else if s.AnyOf != nil {
			fmt.Printf("[WARN] TODO: skipping type for %q, since it is a ANYOF\n", name)
		} else if s.AllOf != nil {
			fmt.Printf("[WARN] TODO: skipping type for %q, since it is a ALLOF\n", name)
		}
	}

	// Add a newline at the end of the type.
	fmt.Fprintln(f, "")
}

func isLocalEnum(v *openapi3.SchemaRef) bool {
	return v.Ref == "" && v.Value.Type == "string" && len(v.Value.Enum) > 0
}

func isLocalObject(v *openapi3.SchemaRef) bool {
	return v.Ref == "" && v.Value.Type == "object" && len(v.Value.Properties) > 0
}

// formatStringType converts a string schema to a valid Go type.
func formatStringType(t *openapi3.Schema) string {
	if t.Format == "date-time" {
		return "*time.Time"
	} else if t.Format == "date" {
		return "*time.Time"
	} else if t.Format == "time" {
		return "*time.Time"
	} else if t.Format == "email" {
		return "string"
	} else if t.Format == "hostname" {
		return "string"
	} else if t.Format == "ipv4" {
		return "string"
	} else if t.Format == "ipv6" {
		return "string"
	} else if t.Format == "uri" {
		return "string"
	} else if t.Format == "uuid" {
		return "string"
	} else if t.Format == "uuid3" {
		return "string"
	}

	return "string"
}

// toLowerFirstLetter returns the given string with the first letter converted to lower case.
func toLowerFirstLetter(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

// makeSingular returns the given string but singular.
func makeSingular(s string) string {
	if strings.HasSuffix(s, "Status") {
		return s
	}
	return strings.TrimSuffix(s, "s")
}

// makePlural returns the given string but plural.
func makePlural(s string) string {
	singular := makeSingular(s)
	if strings.HasSuffix(singular, "s") {
		return singular + "es"
	}

	return singular + "s"
}

// writeSchemaTypeDescription writes the description of the given type.
func writeSchemaTypeDescription(name string, s *openapi3.Schema, f *os.File) {
	if s.Description != "" {
		fmt.Fprintf(f, "// %s is %s\n", name, toLowerFirstLetter(strings.ReplaceAll(s.Description, "\n", "\n// ")))
	} else {
		fmt.Fprintf(f, "// %s is the type definition for a %s.\n", name, name)
	}
}

// writeReponseTypeDescription writes the description of the given type.
func writeResponseTypeDescription(name string, r *openapi3.Response, f *os.File) {
	if r.Description != nil {
		fmt.Fprintf(f, "// %s is the response given when %s\n", name, toLowerFirstLetter(
			strings.ReplaceAll(*r.Description, "\n", "\n// ")))
	} else {
		fmt.Fprintf(f, "// %s is the type definition for a %s response.\n", name, name)
	}
}

func getReferenceSchema(v *openapi3.SchemaRef) string {
	if v.Ref != "" {
		ref := strings.TrimPrefix(v.Ref, "#/components/schemas/")
		if len(v.Value.Enum) > 0 {
			return printProperty(makeSingular(ref))
		}

		return printProperty(ref)
	}

	return ""
}

// writeResponseType writes a type definition for the given response.
func writeResponseType(f *os.File, name string, r *openapi3.Response) {
	// Write the type definition.
	for k, v := range r.Content {
		fmt.Printf("writing type for response %q -> `%s`\n", name, k)

		name := fmt.Sprintf("%sResponse", name)

		// Write the type description.
		writeResponseTypeDescription(name, r, f)

		// Print the type definition.
		s := v.Schema
		if s.Ref != "" {
			fmt.Fprintf(f, "type %s %s\n", name, getReferenceSchema(s))
			continue
		}

		writeSchemaType(f, name, s.Value, "")
	}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func trimStringFromSpace(s string) string {
	if idx := strings.Index(s, " "); idx != -1 {
		return s[:idx]
	}
	return s
}

func containsMatchFirstWord(s []string, str string) bool {
	for _, v := range s {
		if trimStringFromSpace(v) == trimStringFromSpace(str) {
			return true
		}
	}

	return false
}

func isPageParam(s string) bool {
	return s == "nextPage" || s == "pageToken" || s == "limit"
}
