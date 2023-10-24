package api

import (
	"encoding/json"
	"github.com/google/go-querystring/query"
)

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
