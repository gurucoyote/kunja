package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type ApiClient struct {
	HttpClient *http.Client
	Token      string
	ApiBaseUrl string
	Verbose    bool
}

func NewApiClient(baseURL string, token string) *ApiClient {
	return &ApiClient{
		HttpClient: &http.Client{},
		Token:      token,
		ApiBaseUrl: baseURL,
	}
}

func (client *ApiClient) request(ctx context.Context, method, apiPath string, body io.Reader) ([]byte, int, error) {
	if client.Verbose {
		fmt.Printf("%s %s\n", method, client.ApiBaseUrl+apiPath)
	}
	req, err := http.NewRequestWithContext(ctx, method, client.ApiBaseUrl+apiPath, body)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+client.Token)
	if method == http.MethodPost || method == http.MethodPut {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.HttpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return respBody, resp.StatusCode, nil
}

func (client *ApiClient) getCtx(ctx context.Context, apiPath string) (string, error) {
	respBody, status, err := client.request(ctx, http.MethodGet, apiPath, nil)
	if err != nil {
		return "", err
	}
	if status < 200 || status >= 300 {
		var result map[string]string
		json.Unmarshal(respBody, &result)
		return "", fmt.Errorf("status code: %d, message: %s", status, result["message"])
	}
	return string(respBody), nil
}

// putCtx sends a PUT request with context support.
func (client *ApiClient) putCtx(ctx context.Context, apiPath string, payload string) (string, error) {
	respBody, status, err := client.request(ctx, http.MethodPut, apiPath, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return "", err
	}
	if status < 200 || status >= 300 {
		var result map[string]string
		json.Unmarshal(respBody, &result)
		return "", fmt.Errorf("status code: %d, message: %s", status, result["message"])
	}
	return string(respBody), nil
}

// deleteCtx sends a DELETE request with context support.
func (client *ApiClient) deleteCtx(ctx context.Context, apiPath string) (string, error) {
	respBody, status, err := client.request(ctx, http.MethodDelete, apiPath, nil)
	if err != nil {
		return "", err
	}
	if status < 200 || status >= 300 {
		var result map[string]string
		json.Unmarshal(respBody, &result)
		return "", fmt.Errorf("status code: %d, message: %s", status, result["message"])
	}
	return string(respBody), nil
}

func (client *ApiClient) postCtx(ctx context.Context, apiPath string, payload string) (string, error) {
	respBody, status, err := client.request(ctx, http.MethodPost, apiPath, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return "", err
	}
	if status < 200 || status >= 300 {
		var result map[string]string
		json.Unmarshal(respBody, &result)
		return "", fmt.Errorf("status code: %d, message: %s", status, result["message"])
	}
	return string(respBody), nil
}

func (client *ApiClient) Get(apiPath string) (string, error) {
	if client.Verbose {
		fmt.Printf("GET %s\n", client.ApiBaseUrl+apiPath)
	}
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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
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
	if client.Verbose {
		fmt.Printf("POST %s\n", client.ApiBaseUrl+apiPath)
	}
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
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
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
	if client.Verbose {
		fmt.Printf("PUT %s\n", client.ApiBaseUrl+apiPath)
	}
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
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
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
	if client.Verbose {
		fmt.Printf("DELETE %s\n", client.ApiBaseUrl+apiPath)
	}
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
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
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

func (client *ApiClient) Login(ctx context.Context, username string, password string, totp_passcode string) (string, error) {
	payload := map[string]string{
		"username":      username,
		"password":      password,
		"totp_passcode": totp_passcode,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	response, err := client.postCtx(ctx, "/login", string(payloadBytes))
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

// GetAllUsers retrieves all existing users.
func (client *ApiClient) GetAllUsers(ctx context.Context) ([]User, error) {
	response, err := client.getCtx(ctx, "/users?include=details")
	if err != nil {
		return nil, err
	}
	var users []User
	if client.Verbose {
		var prettyJSON bytes.Buffer
		err = json.Indent(&prettyJSON, []byte(response), "", "  ")
		if err == nil {
			fmt.Println("Raw JSON response:", prettyJSON.String())
		}
	}
	err = json.Unmarshal([]byte(response), &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (client *ApiClient) GetAllProjects(ctx context.Context) ([]Project, error) {
	response, err := client.getCtx(ctx, "/projects")
	if err != nil {
		return nil, err
	}
	var projects []Project
	if client.Verbose {
		var prettyJSON bytes.Buffer
		err = json.Indent(&prettyJSON, []byte(response), "", "  ")
		if err == nil {
			fmt.Println("Raw JSON response:", prettyJSON.String())
		}
	}
	err = json.Unmarshal([]byte(response), &projects)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// GetProjectUsers retrieves all users that a given project is shared with.
func (client *ApiClient) GetProjectUsers(ctx context.Context, projectID int) ([]UserWithRight, error) {
	apiEndpoint := fmt.Sprintf("/projects/%d/users", projectID)

	response, err := client.getCtx(ctx, apiEndpoint)
	if err != nil {
		return nil, err
	}

	var usersWithRights []UserWithRight
	err = json.Unmarshal([]byte(response), &usersWithRights)
	if err != nil {
		return nil, err
	}

	return usersWithRights, nil
}

// GetProject retrieves a single project by its ID.
func (client *ApiClient) GetProject(ctx context.Context, projectID int) (Project, error) {
	apiEndpoint := fmt.Sprintf("/projects/%d", projectID)

	response, err := client.getCtx(ctx, apiEndpoint)
	if err != nil {
		return Project{}, err
	}

	var project Project
	err = json.Unmarshal([]byte(response), &project)
	if err != nil {
		return Project{}, err
	}

	return project, nil
}
