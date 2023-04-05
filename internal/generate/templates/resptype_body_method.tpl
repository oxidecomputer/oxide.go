{{template "description" .}}func (c *Client) {{.FunctionName}}({{.ParamsString}}) (*{{.ResponseType}}, error) {
    // Create the url.
    path := "{{.Path}}"
    uri := resolveRelative(c.server, path){{if .IsAppJSON}}

    // Encode the request body as json.
    b := new(bytes.Buffer)
    if err := json.NewEncoder(b).Encode(params.Body); err != nil {
        return nil, fmt.Errorf("encoding json body request failed: %v", err)
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
		return nil, fmt.Errorf("error building request: %v", err)
	}

    // Send the request.
    resp, err := c.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("error sending request: %v", err)
    }
    defer resp.Body.Close()

    // Check the response.
    if err := checkResponse(resp); err != nil {
        return nil, err
    }

    // Decode the body from the response.
    if resp.Body == nil {
        return nil, errors.New("request returned an empty body in the response")
    }

    var body {{.ResponseType}}
    if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
        return nil, fmt.Errorf("error decoding response body: %v", err)
    }

    // Return the response.
    return &body, nil
}

