{{template "description" .}}func (c *Client) {{.FunctionName}}(ctx context.Context, {{.ParamsString}}) error { {{if .HasParams}}
    if err := params.Validate(); err != nil {
		return err
	}{{end}}
    // Create the request
    req, err := c.buildRequest(
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
		return fmt.Errorf("error building request: %v", err)
	}

    // Send the request.
    resp, err := c.client.Do(req)
    if err != nil {
        return fmt.Errorf("error sending request: %v", err)
    }
    defer resp.Body.Close()

    // Create and return an HTTPError when an error response code is received.
    if err := NewHTTPError(resp); err != nil {
        return err
    }

    return nil
}

