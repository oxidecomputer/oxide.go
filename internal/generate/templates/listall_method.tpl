{{template "description" .}}func (c *Client) {{.FunctionName}}(ctx context.Context, {{.ParamsString}}) ({{.ResponseType}}, error) { {{if .HasParams}}
	if err := params.Validate(); err != nil {
		return nil, err
	}{{end}}
	var allPages {{.ResponseType}}
	params.PageToken = ""
	params.Limit = NewPointer(100)
	for {
		page, err := c.{{.WrappedFunction}}(ctx, params)
		if err != nil {
			return nil, err
		}
		allPages = append(allPages, page.Items...)
		if page.NextPage == "" || page.NextPage == params.PageToken {
			break
		}
		params.PageToken = page.NextPage
	}

	return allPages, nil
}

