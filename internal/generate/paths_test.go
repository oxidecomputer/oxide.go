package main

// TODO: This test is generating a messed up file ignore for now
//func Test_generatePaths(t *testing.T) {
//	var testMin float64 = 1
//	pathDelDesc := "successful deletion"
//
//	pathsSpec := &openapi3.T{
//		Paths: openapi3.Paths{
//			"/ip-pools": &openapi3.PathItem{
//				Get: &openapi3.Operation{
//					Tags:        []string{"ip-pools"},
//					OperationID: "ip-pool-list",
//					Parameters: openapi3.Parameters{
//						&openapi3.ParameterRef{
//							Value: &openapi3.Parameter{
//								In:   "query",
//								Name: "limit",
//								Schema: &openapi3.SchemaRef{
//									Value: &openapi3.Schema{
//										Nullable: true,
//										Type:     "integer",
//										Format:   "uint32",
//										Min:      &testMin,
//									},
//								},
//								Style: "form",
//							},
//						},
//						&openapi3.ParameterRef{
//							Value: &openapi3.Parameter{
//								In:   "query",
//								Name: "page_token",
//								Schema: &openapi3.SchemaRef{
//									Value: &openapi3.Schema{
//										Nullable: true,
//										Type:     "string",
//									},
//								},
//								Style: "form",
//							},
//						},
//					},
//					Responses: openapi3.Responses{
//						"200": &openapi3.ResponseRef{Value: &openapi3.Response{
//							Content: openapi3.Content{
//								"application/json": &openapi3.MediaType{
//									Schema: &openapi3.SchemaRef{
//										Ref:   "#/components/schemas/IpPoolResultsPage",
//										Value: &openapi3.Schema{},
//									},
//								},
//							},
//						}},
//						"4XX": &openapi3.ResponseRef{
//							Ref:   "#/components/responses/Error",
//							Value: &openapi3.Response{},
//						}},
//				},
//				Post: &openapi3.Operation{
//					Tags:        []string{"ip-pools"},
//					OperationID: "ip-pool-create",
//					RequestBody: &openapi3.RequestBodyRef{
//						Value: &openapi3.RequestBody{
//							Content: openapi3.Content{
//								"application/json": &openapi3.MediaType{
//									Schema: &openapi3.SchemaRef{
//										Ref:   "#/components/schemas/IpPoolCreate",
//										Value: &openapi3.Schema{},
//									},
//								},
//							},
//							Required: true,
//						},
//					},
//					Responses: openapi3.Responses{
//						"201": &openapi3.ResponseRef{Value: &openapi3.Response{
//							Content: openapi3.Content{
//								"application/json": &openapi3.MediaType{
//									Schema: &openapi3.SchemaRef{
//										Ref:   "#/components/schemas/IpPool",
//										Value: &openapi3.Schema{},
//									},
//								},
//							},
//						}},
//						"4XX": &openapi3.ResponseRef{
//							Ref:   "#/components/responses/Error",
//							Value: &openapi3.Response{},
//						}},
//				},
//			},
//			"/ip-pools{pool_name}": &openapi3.PathItem{
//				Get: &openapi3.Operation{
//					Tags:        []string{"ip-pools"},
//					OperationID: "ip-pool-view",
//					Parameters: openapi3.Parameters{
//						&openapi3.ParameterRef{
//							Value: &openapi3.Parameter{
//								In:       "path",
//								Name:     "pool_name",
//								Required: true,
//								Schema: &openapi3.SchemaRef{
//									Ref:   "#/components/schemas/Name",
//									Value: &openapi3.Schema{},
//								},
//								Style: "simple",
//							},
//						},
//					},
//					Responses: openapi3.Responses{
//						"200": &openapi3.ResponseRef{Value: &openapi3.Response{
//							Content: openapi3.Content{
//								"application/json": &openapi3.MediaType{
//									Schema: &openapi3.SchemaRef{
//										Ref:   "#/components/schemas/IpPool",
//										Value: &openapi3.Schema{},
//									},
//								},
//							},
//						}},
//						"4XX": &openapi3.ResponseRef{
//							Ref:   "#/components/responses/Error",
//							Value: &openapi3.Response{},
//						}},
//				},
//				Put: &openapi3.Operation{
//					Tags:        []string{"ip-pools"},
//					OperationID: "ip-pool-update",
//					Parameters: openapi3.Parameters{
//						&openapi3.ParameterRef{
//							Value: &openapi3.Parameter{
//								In:       "path",
//								Name:     "pool_name",
//								Required: true,
//								Schema: &openapi3.SchemaRef{
//									Ref:   "#/components/schemas/Name",
//									Value: &openapi3.Schema{},
//								},
//								Style: "simple",
//							},
//						},
//					},
//					RequestBody: &openapi3.RequestBodyRef{
//						Value: &openapi3.RequestBody{
//							Content: openapi3.Content{
//								"application/json": &openapi3.MediaType{
//									Schema: &openapi3.SchemaRef{
//										Ref:   "#/components/schemas/IpPoolUpdate",
//										Value: &openapi3.Schema{},
//									},
//								},
//							},
//							Required: true,
//						},
//					},
//					Responses: openapi3.Responses{
//						"201": &openapi3.ResponseRef{Value: &openapi3.Response{
//							Content: openapi3.Content{
//								"application/json": &openapi3.MediaType{
//									Schema: &openapi3.SchemaRef{
//										Ref:   "#/components/schemas/IpPool",
//										Value: &openapi3.Schema{},
//									},
//								},
//							},
//						}},
//						"4XX": &openapi3.ResponseRef{
//							Ref:   "#/components/responses/Error",
//							Value: &openapi3.Response{},
//						},
//					},
//				},
//				Delete: &openapi3.Operation{
//					Tags:        []string{"ip-pools"},
//					OperationID: "ip-pool-delete",
//					Parameters: openapi3.Parameters{
//						&openapi3.ParameterRef{
//							Value: &openapi3.Parameter{
//								In:       "path",
//								Name:     "pool_name",
//								Required: true,
//								Schema: &openapi3.SchemaRef{
//									Ref:   "#/components/schemas/Name",
//									Value: &openapi3.Schema{},
//								},
//								Style: "simple",
//							},
//						},
//					},
//					Responses: openapi3.Responses{
//						"204": &openapi3.ResponseRef{Value: &openapi3.Response{
//							Description: &pathDelDesc,
//						}},
//						"4XX": &openapi3.ResponseRef{
//							Ref:   "#/components/responses/Error",
//							Value: &openapi3.Response{},
//						}},
//				},
//			},
//		},
//	}
//
//	type args struct {
//		file string
//		spec *openapi3.T
//	}
//	tests := []struct {
//		name    string
//		args    args
//		wantErr string
//	}{
//		{
//			name:    "fail on non-existent file",
//			args:    args{"sdf/gdsf", pathsSpec},
//			wantErr: "no such file or directory",
//		},
//		{
//			name: "success",
//			args: args{"test_utils/paths_output", pathsSpec},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			// TODO: For now the test is not properly generating the "ListAll" methods
//			// This is because there is a separate check to the response type. The way this works
//			// should be changed
//			if err := generatePaths(tt.args.file, tt.args.spec); err != nil {
//				assert.ErrorContains(t, err, tt.wantErr)
//				return
//			}
//
//			if err := compareFiles("test_utils/paths_output_expected", tt.args.file); err != nil {
//				t.Error(err)
//			}
//		})
//	}
//}
//
