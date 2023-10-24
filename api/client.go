package api

import (
	"fmt"
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/google/go-querystring/query"
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
	response, err := client.Post("/login", string(payloadBytes))
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
/*
OpenAPI spec:

        "/tasks/all": {
            "get": {
                "security": [
                    {
                        "JWTKeyAuth": []
                    }
                ],
                "description": "Returns all tasks on any project the user has access to.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "task"
                ],
                "summary": "Get tasks",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "The page number. Used for pagination. If not provided, the first page of results is returned.",
                        "name": "page",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page.",
                        "name": "per_page",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Search tasks by task text.",
                        "name": "s",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "The sorting parameter. You can pass this multiple times to get the tasks ordered by multiple different parameters, along with `order_by`. Possible values to sort by are `id`, `title`, `description`, `done`, `done_at`, `due_date`, `created_by_id`, `project_id`, `repeat_after`, `priority`, `start_date`, `end_date`, `hex_color`, `percent_done`, `uid`, `created`, `updated`. Default is `id`.",
                        "name": "sort_by",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "The ordering parameter. Possible values to order by are `asc` or `desc`. Default is `asc`.",
                        "name": "order_by",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "The name of the field to filter by. Allowed values are all task properties. Task properties which are their own object require passing in the id of that entity. Accepts an array for multiple filters which will be chanied together, all supplied filter must match.",
                        "name": "filter_by",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "The value to filter for.",
                        "name": "filter_value",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "The comparator to use for a filter. Available values are `equals`, `greater`, `greater_equals`, `less`, `less_equals`, `like` and `in`. `in` expects comma-separated values in `filter_value`. Defaults to `equals`",
                        "name": "filter_comparator",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "The concatinator to use for filters. Available values are `and` or `or`. Defaults to `or`.",
                        "name": "filter_concat",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "If set to true the result will include filtered fields whose value is set to `null`. Available values are `true` or `false`. Defaults to `false`.",
                        "name": "filter_include_nulls",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "The tasks",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.Task"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal error",
                        "schema": {
                            "$ref": "#/definitions/models.Message"
                        }
                    }
                }
            }
        },
	*/
func (client *ApiClient) GetAllTasks(params GetAllTasksParams) ([]Task, error) {
	queryParams, _ := query.Values(params)
	response, err := client.Get("/tasks/all?" + queryParams.Encode())
	if err != nil {
		return nil, err
	}
	var tasks []Task
	err = json.Unmarshal([]byte(response), &tasks)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

