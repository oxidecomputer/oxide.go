{{template "description" .}}func (c *Client) {{.FunctionName}}(ctx context.Context, {{.ParamsString}}) (*{{.ResponseType}}, error) { {{if .HasParams}}
    if err := params.Validate(); err != nil {
		return nil, err
	}{{end}}
    // Create the request
    req, err := buildRequest(
        ctx,
        nil, 
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

    // Create and return an HTTPError when an error response code is received.
    if err := NewHTTPError(resp); err != nil {
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

