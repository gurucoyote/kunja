package api

import (
	"net/http"
)

type ApiClient struct {
	HttpClient *http.Client
	Token      string
	ApiBaseUrl string
}
