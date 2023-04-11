{{template "description" .}}func (c *Client) {{.FunctionName}}({{.ParamsString}}) (*{{.ResponseType}}, error) { {{if .HasParams}}
    if err := params.Validate(); err != nil {
		return nil, err
	}{{end}}{{if .IsAppJSON}}
    // Encode the request body as json.
    b := new(bytes.Buffer)
    if err := json.NewEncoder(b).Encode(params.Body); err != nil {
        return nil, fmt.Errorf("encoding json body request failed: %v", err)
    }{{else}}
    b := params.Body{{end}}

    // Create the request
    req, err := buildRequest(
        b, 
        "{{.HTTPMethod}}", 
        resolveRelative(c.server, "{{.Path}}"), 
        map[string]string{ {{range .PathParams}}
            {{.}}{{end}}
        }, 
        map[string]string{ {{range .QueryParams}}
            {{.}}{{end}}
        },
    )
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

