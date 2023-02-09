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
	QueryParams     []string
	IsList          bool
	IsListAll       bool
	HasDescription  bool
	HasParams       bool
	HasBody         bool
	HasSummary      bool
	IsAppJSON       bool
}

type paramsInfo struct {
	parameters      map[string]*openapi3.Parameter
	paramsString    string
	docParamsString string
	isPageResult    bool
}

// Generate the paths.go file.
func generatePaths(file string, spec *openapi3.T) error {
	f, err := openGeneratedFile(file)
	if err != nil {
		return err
	}
	defer f.Close()

	// Iterate over all the paths in the spec and write the methods.
	// We want to ensure we keep the order.
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

		err := buildPath(f, spec, path, p)
		if err != nil {
			return err
		}
	}

	return nil
}

// buildPath builds the given path as an http request to the given file.
func buildPath(f *os.File, spec *openapi3.T, path string, p *openapi3.PathItem) error {
	if p.Get != nil {
		err := buildMethod(f, spec, http.MethodGet, path, p.Get, false)
		if err != nil {
			return err
		}
	}

	if p.Post != nil {
		err := buildMethod(f, spec, http.MethodPost, path, p.Post, false)
		if err != nil {
			return err
		}
	}

	if p.Put != nil {
		err := buildMethod(f, spec, http.MethodPut, path, p.Put, false)
		if err != nil {
			return err
		}
	}

	if p.Delete != nil {
		err := buildMethod(f, spec, http.MethodDelete, path, p.Delete, false)
		if err != nil {
			return err
		}
	}

	if p.Patch != nil {
		err := buildMethod(f, spec, http.MethodPatch, path, p.Patch, false)
		if err != nil {
			return err
		}
	}

	if p.Head != nil {
		err := buildMethod(f, spec, http.MethodHead, path, p.Head, false)
		if err != nil {
			return err
		}
	}

	if p.Options != nil {
		err := buildMethod(f, spec, http.MethodOptions, path, p.Options, false)
		if err != nil {
			return err
		}
	}

	return nil
}

func buildMethod(f *os.File, spec *openapi3.T, method string, path string, o *openapi3.Operation, isGetAllPages bool) error {
	respType, pagedRespType, err := getSuccessResponseType(o, isGetAllPages)
	if err != nil {
		return err
	}

	if len(o.Tags) == 0 || o.Tags[0] == "hidden" {
		fmt.Printf("[WARN] TODO: skipping operation %q, since it has no tag or is hidden\n", o.OperationID)
		return nil
	}

	methodName := strcase.ToCamel(o.OperationID)
	pInfo := parseParams(o.Parameters, method)

	// Adapt for ListAll methods
	if pInfo.isPageResult && isGetAllPages && len(pagedRespType) > 0 {
		respType = pagedRespType
	}

	ogmethodName := methodName
	ogDocParamsString := pInfo.docParamsString
	if isGetAllPages {
		methodName += "AllPages"
		pInfo.paramsString = listAllSignature(pInfo.paramsString)
		delete(pInfo.parameters, "page_token")
		delete(pInfo.parameters, "limit")
	}

	isList := pInfo.isPageResult && !isGetAllPages
	// end ListAll specific code

	pInfo = parseRequestBody(o.RequestBody, pInfo, methodName)
	sigParams := buildSignatureParams(pInfo.parameters)
	pathParams := buildPathParams(pInfo.parameters)
	queryParams := buildQueryParams(pInfo.parameters)

	config := methodTemplate{
		Description:     o.Description,
		HTTPMethod:      method,
		FunctionName:    methodName,
		WrappedFunction: ogmethodName,
		WrappedParams:   ogDocParamsString,
		ResponseType:    respType,
		SignatureParams: sigParams,
		Summary:         o.Summary,
		ParamsString:    pInfo.paramsString,
		Path:            cleanPath(path),
		PathParams:      pathParams,
		QueryParams:     queryParams,
		IsList:          isList,
		IsListAll:       isGetAllPages,
		HasBody:         o.RequestBody != nil,
		IsAppJSON:       true,
		HasParams:       len(pInfo.parameters) > 0,
		HasSummary:      o.Summary != "",
		HasDescription:  o.Description != "",
	}

	// TODO: Handle other content types
	if o.RequestBody != nil {
		for mt := range o.RequestBody.Value.Content {
			if mt != "application/json" {
				config.IsAppJSON = false
				break
			}
		}
	}

	// Presence of a "default" response means there is no response type.
	// No response should be returned in this case
	if o.Responses.Default() != nil {
		config.ResponseType = ""
	}

	if err := writeTpl(f, config); err != nil {
		return err
	}

	if pInfo.isPageResult && !isGetAllPages {
		// Run the method again with get all pages for ListAll methods.
		err := buildMethod(f, spec, method, path, o, true)
		if err != nil {
			return err
		}
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

func writeTpl(f *os.File, config methodTemplate) error {
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

	err = t.Execute(f, config)
	if err != nil {
		return err
	}

	return nil
}

func buildSignatureParams(params map[string]*openapi3.Parameter) map[string]string {
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
	return sigParams
}

func buildPathParams(params map[string]*openapi3.Parameter) []string {
	pathParams := make([]string, 0)
	if len(params) > 0 {
		// Iterate over all the paths in the spec and write the types.
		// We want to ensure we keep the order so the diffs don't change.
		keys := make([]string, 0)
		for k, v := range params {
			if v.In == "path" {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		for _, name := range keys {
			p := params[name]
			t := convertToValidGoType(name, p.Schema)
			n := strcase.ToLowerCamel(name)
			if t == "string" {
				pathParams = append(pathParams, fmt.Sprintf("%q: %s,", name, n))
				// TODO: Identify interfaces instead of singling out NameOrId
			} else if t == "NameOrId" {
				pathParams = append(pathParams, fmt.Sprintf("%q: %s.(string),", name, n))
			} else if t == "int" {
				pathParams = append(pathParams, fmt.Sprintf("%q: strconv.Itoa(%s),", name, n))
			} else if t == "*time.Time" {
				pathParams = append(pathParams, fmt.Sprintf("%q: %s.String(),", name, n))
			} else {
				pathParams = append(pathParams, fmt.Sprintf("%q: string(%s),", name, n))
			}
		}
	}
	return pathParams
}

func buildQueryParams(params map[string]*openapi3.Parameter) []string {
	pathParams := make([]string, 0)
	if len(params) > 0 {
		// Iterate over all the paths in the spec and write the types.
		// We want to ensure we keep the order so the diffs don't change.
		keys := make([]string, 0)
		for k, v := range params {
			if v.In == "query" {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		for _, name := range keys {
			p := params[name]
			t := convertToValidGoType(name, p.Schema)
			n := strcase.ToLowerCamel(name)
			if t == "string" {
				pathParams = append(pathParams, fmt.Sprintf("%q: %s,", name, n))
				// TODO: Identify interfaces instead of singling out NameOrId
			} else if t == "NameOrId" {
				pathParams = append(pathParams, fmt.Sprintf("%q: %s.(string),", name, n))
			} else if t == "int" {
				pathParams = append(pathParams, fmt.Sprintf("%q: strconv.Itoa(%s),", name, n))
			} else if t == "*time.Time" {
				pathParams = append(pathParams, fmt.Sprintf("%q: %s.String(),", name, n))
			} else {
				pathParams = append(pathParams, fmt.Sprintf("%q: string(%s),", name, n))
			}
		}
	}
	return pathParams
}

func parseParams(specParams openapi3.Parameters, method string) paramsInfo {
	pInfo := paramsInfo{parameters: make(map[string]*openapi3.Parameter, 0)}
	for index, p := range specParams {
		if p.Ref != "" {
			fmt.Printf("[WARN] TODO: skipping parameter for %q, since it is a reference\n", p.Value.Name)
			continue
		}

		paramName := strcase.ToLowerCamel(p.Value.Name)

		// Check if we have a page result.
		if isPageParam(paramName) && method == http.MethodGet {
			pInfo.isPageResult = true
		}

		pInfo.parameters[p.Value.Name] = p.Value
		pInfo.paramsString += fmt.Sprintf("%s %s, ", paramName, convertToValidGoType(p.Value.Name, p.Value.Schema))
		if index == len(specParams)-1 {
			pInfo.docParamsString += paramName
		} else {
			pInfo.docParamsString += fmt.Sprintf("%s, ", paramName)
		}
	}

	return pInfo
}

func parseRequestBody(reqBody *openapi3.RequestBodyRef, pInfo paramsInfo, methodName string) paramsInfo {
	if reqBody == nil {
		return pInfo
	}

	if reqBody.Ref != "" {
		fmt.Printf("[WARN] TODO: skipping request body for %q, since it is a reference: %q\n", methodName, reqBody.Ref)
	}

	for mt, r := range reqBody.Value.Content {
		if mt != "application/json" {
			pInfo.paramsString += "b io.Reader"
			break
		}

		typeName := convertToValidGoType("", r.Schema)

		pInfo.paramsString += "j *" + typeName

		if len(pInfo.docParamsString) > 0 {
			pInfo.docParamsString += ", "
		}
		pInfo.docParamsString += "body"
		break
	}

	return pInfo
}

func listAllSignature(params string) string {
	// Remove pageToken and limit as we want to list all pages
	params = strings.ReplaceAll(params, "pageToken string,", "")
	params = strings.ReplaceAll(params, "limit int,", "")
	return strings.TrimSpace(params)
}
