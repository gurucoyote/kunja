package api

import (
	"fmt"
	"bytes"
	"encoding/json"
	"errors"
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

	if resp.StatusCode != http.StatusOK {
		var result map[string]string
		body, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(body, &result)
		return "", fmt.Errorf("status code: %d, message: %s", resp.StatusCode, result["message"])
	}

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
	if resp.StatusCode != http.StatusOK {
		var result map[string]string
		body, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(body, &result)
		return "", fmt.Errorf("status code: %d, message: %s", resp.StatusCode, result["message"])
	}
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
	if resp.StatusCode != http.StatusOK {
		var result map[string]string
		body, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(body, &result)
		return "", fmt.Errorf("status code: %d, message: %s", resp.StatusCode, result["message"])
	}
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
	if resp.StatusCode != http.StatusOK {
		var result map[string]string
		body, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(body, &result)
		return "", fmt.Errorf("status code: %d, message: %s", resp.StatusCode, result["message"])
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

/*

OpenAPI spec for /login:
        "/login": {
            "post": {
                "description": "Logs a user in. Returns a JWT-Token to authenticate further requests.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Login",
                "parameters": [
                    {
                        "description": "The login credentials",
                        "name": "credentials",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/user.Login"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/auth.Token"
                        }
                    },
                    "400": {
                        "description": "Invalid user password model.",
                        "schema": {
                            "$ref": "#/definitions/models.Message"
                        }
                    },
                    "403": {
                        "description": "Invalid username or password.",
                        "schema": {
                            "$ref": "#/definitions/models.Message"
                        }
                    },
                    "412": {
                        "description": "Invalid totp passcode.",
                        "schema": {
                            "$ref": "#/definitions/models.Message"
                        }
                    }
                }
            }
        },
*/
func (client *ApiClient) Login(username string, password string, totp_passcode string) (string, error) {
	payload := map[string]string{
		"username": username,
		"password": password,
		"totp_passcode": totp_passcode,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	response, err := client.Post("/loginP", string(payloadBytes))
	if err != nil {
		return "", err
	}
	var result map[string]string
	err = json.Unmarshal([]byte(response), &result)
	if err != nil {
		return "", err
	}
	token, ok := result["token"]
	if !ok {
		// Print out the original json response as indented string
		formattedResponse, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(formattedResponse))
		return "", errors.New("token not found in response")
	}
	client.Token = token
	return token, nil
}
