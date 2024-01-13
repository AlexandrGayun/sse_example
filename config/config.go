package config

type Config struct {
	OpenAIApiKey string `envconfig:"OPENAI_API_KEY" default:""`
	HttpApiPort  string `envconfig:"HTTP_API_PORT" default:"8080"`
}
