{{template "description" .}}func (c *Client) {{.FunctionName}}({{.ParamsString}}) (*{{.ResponseType}}, error) {
	var allPages {{.ResponseType}}
	pageToken := ""
	limit := 100
	for {
		page, err := c.{{.WrappedFunction}}({{.WrappedParams}})
		if err != nil {
			return nil, err
		}
		allPages = append(allPages, page.Items...)
		if page.NextPage == "" || page.NextPage == pageToken {
			break
		}
		pageToken = page.NextPage
	}

	return &allPages, nil
}
