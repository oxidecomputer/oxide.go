{{template "description" .}}func (c *Client) {{.FunctionName}}({{.ParamsString}}) error {
    // Create the url.
    path := "{{.Path}}"
    uri := resolveRelative(c.server, path){{if .IsAppJSON}}

    // Encode the request body as json.
    b := new(bytes.Buffer)
    if err := json.NewEncoder(b).Encode(params.Body); err != nil {
        return fmt.Errorf("encoding json body request failed: %v", err)
    }{{else}}
    b := params.Body{{end}}

    pathParams := map[string]string{ {{range .PathParams}}
        {{.}}{{end}}
    }
    queryParams := map[string]string{ {{range .QueryParams}}
        {{.}}{{end}}
    }

    req, err := buildRequest(b, "{{.HTTPMethod}}", uri, pathParams, queryParams)
	if err != nil {
		return fmt.Errorf("error building request: %v", err)
	}

    // Send the request.
    resp, err := c.client.Do(req)
    if err != nil {
        return fmt.Errorf("error sending request: %v", err)
    }
    defer resp.Body.Close()

    // Check the response.
    if err := checkResponse(resp); err != nil {
        return err
    }

    return nil
}

