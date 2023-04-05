{{template "description" .}}func (c *Client) {{.FunctionName}}({{.ParamsString}}) error {
    // Create the url.
    path := "{{.Path}}"
    uri := resolveRelative(c.server, path)

    pathParams := map[string]string{ {{range .PathParams}}
        {{.}}{{end}}
    }
    queryParams := map[string]string{ {{range .QueryParams}}
        {{.}}{{end}}
    }

    req, err := buildRequest(nil, "{{.HTTPMethod}}", uri, pathParams, queryParams)
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

