package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

func (client *ApiClient) Login(username string, password string, totp_passcode string) (string, error) {
	payload := map[string]string{
		"username":      username,
		"password":      password,
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

// GetAllUsers retrieves all existing users.
func (client *ApiClient) GetAllUsers() ([]User, error) {
	response, err := client.Get("/users?include=details")
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

func (client *ApiClient) GetAllProjects() ([]Project, error) {
	response, err := client.Get("/projects")
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
func (client *ApiClient) GetProjectUsers(projectID int) ([]UserWithRight, error) {
	apiEndpoint := fmt.Sprintf("/projects/%d/users", projectID)

	response, err := client.Get(apiEndpoint)
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
func (client *ApiClient) GetProject(projectID int) (Project, error) {
	apiEndpoint := fmt.Sprintf("/projects/%d", projectID)

	response, err := client.Get(apiEndpoint)
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
