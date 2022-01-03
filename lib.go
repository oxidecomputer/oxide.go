package oxide

//go:generate go run generate/generate.go

// DefaultServerURL is the default server URL for the Oxide API.
const DefaultServerURL = "https://api.oxide.computer"

// TokenEnvVar is the environment variable that contains the token.
const TokenEnvVar = "OXIDE_API_TOKEN"
