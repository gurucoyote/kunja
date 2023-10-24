package api

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

type ApiClient struct {
	HttpClient *http.Client
	Token      string
	ApiBaseUrl string
}

func NewApiClient(baseURL string, token string) *ApiClient {
	return &ApiClient{
		HttpClient: &http.Client{},
		Token:      token,
		ApiBaseUrl: baseURL,
	}
}

func (client *ApiClient) Get(apiPath string) (string, error) {
	req, err := http.NewRequest("GET", client.ApiBaseUrl+apiPath, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+client.Token)
	resp, err := client.HttpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (client *ApiClient) Post(apiPath string, payload string) (string, error) {
	req, err := http.NewRequest("POST", client.ApiBaseUrl+apiPath, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+client.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.HttpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (client *ApiClient) Put(apiPath string, payload string) (string, error) {
	req, err := http.NewRequest("PUT", client.ApiBaseUrl+apiPath, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+client.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.HttpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (client *ApiClient) Delete(apiPath string) (string, error) {
	req, err := http.NewRequest("DELETE", client.ApiBaseUrl+apiPath, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+client.Token)
	resp, err := client.HttpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
