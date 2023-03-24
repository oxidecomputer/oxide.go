{{template "description" .}}func (c *Client) {{.FunctionName}}({{.ParamsString}}) (*{{.ResponseType}}, error) {
	var allPages {{.ResponseType}}
	params.PageToken = ""
	params.Limit = 100
	for {
		page, err := c.{{.WrappedFunction}}(params)
		if err != nil {
			return nil, err
		}
		allPages = append(allPages, page.Items...)
		if page.NextPage == "" || page.NextPage == params.PageToken {
			break
		}
		params.PageToken = page.NextPage
	}

	return &allPages, nil
}

