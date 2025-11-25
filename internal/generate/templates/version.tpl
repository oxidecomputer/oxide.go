// sdkVersion is the Oxide Go SDK sdkVersion. This is used to dynamically
// populate the user agent for [Client]. It is purposefully unexported to
// prevent external users from reading it. The value of this comes from the
// VERSION file in the root of this repository.
const sdkVersion = "{{ .SDKVersion }}"

// openAPIVersion is the OpenAPI specification version the Oxide Go SDK was
// generated from. This is used to dynamically populate the 'API-Version' header
// for [Client]. It is purposefully unexported to prevent external users from
// reading it. The value of this comes from the OpenAPI specification associated
// with the OMICRON_VERSION file in the root of this repository.
const openAPIVersion = "{{ .OpenAPIVersion }}"
