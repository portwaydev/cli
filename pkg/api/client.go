package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/spf13/viper"
)

func NewAPIKeyClientWithResponses(server string, apiKey string) (*ClientWithResponses, error) {
	server = strings.TrimSuffix(server, "/")
	return NewClientWithResponses(server,
		WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("X-API-Key", apiKey)
			return nil
		}),
	)
}

func NewViperClientWithResponses() (*ClientWithResponses, error) {
	server := viper.GetString("url")
	apiKey := viper.GetString("token")
	return NewAPIKeyClientWithResponses(server, apiKey)
}
