{{template "description" .}}func (c *Client) {{.FunctionName}}({{.ParamsString}}) error {
    // Create the url.
    path := "{{.Path}}"
    uri := resolveRelative(c.server, path)

    // Create the request.
    req, err := http.NewRequest("{{.HTTPMethod}}", uri, nil)
    if err != nil {
        return fmt.Errorf("error creating request: %v", err)
    }{{if .HasParams}}

    // Add the parameters to the url.
    if err := expandURL(req.URL, map[string]string{ {{range .PathParams}}
        {{.}}{{end}}
    }); err != nil {
        return fmt.Errorf("expanding URL with parameters failed: %v", err)
    }

    // Add query if any
    if err := addQueries(req.URL, map[string]string{ {{range .QueryParams}}
        {{.}}{{end}}
    }); err != nil {
        return fmt.Errorf("adding queries to URL failed: %v", err)
    }{{end}}

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

