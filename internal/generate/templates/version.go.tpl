// version is the Oxide Go SDK version. This is used to dynamically populate the
// user agent for [Client].
const version = "{{ .SDKVersion }}"

// APIVersion is the OpenAPI specification version the Oxide Go SDK was
// generated from. This is used to populate the 'API-Version' header for
// [Client].
const APIVersion = "{{ .OpenAPIVersion }}"
