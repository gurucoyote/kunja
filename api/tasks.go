package api

import (
	"encoding/json"
	"github.com/google/go-querystring/query"
	"strconv"
	"time"
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
	for i := range tasks {
		tasks[i].CalculateUrgency()
	}
	return tasks, nil
}

func (task *Task) CalculateUrgency() {
	if task.Done {
		task.Urgency = 0.0
		return
	}
	dueDateScore := float64(task.getDueDateScore())
	priorityScore := float64(task.Priority)
	favoriteScore := 0.0
	if task.IsFavorite {
		favoriteScore = 1.0
	}
	task.Urgency = 1.0 + dueDateScore + priorityScore + favoriteScore
}

func (task *Task) getDueDateScore() int {
	if task.DueDate.IsZero() {
		return 0
	}
	dueDays := int(task.DueDate.Sub(time.Now()).Hours() / 24)
	switch {
	case dueDays < 0:
		return 6
	case dueDays == 0:
		return 5
	case dueDays == 1:
		return 4
	case dueDays > 1 && dueDays <= 2:
		return 3
	case dueDays > 2 && dueDays <= 5:
		return 2
	case dueDays > 5 && dueDays <= 10:
		return 1
	case dueDays > 14:
		return -1
	default:
		return 0
	}
}

// UpdateTask updates a task. This includes marking it as done.
// Assignees you pass will be updated, see their individual endpoints for more details on how this is done.
// To update labels, see the description of the endpoint.
func (client *ApiClient) UpdateTask(ID int, task Task) (Task, error) {
	taskJson, err := json.Marshal(task)
	if err != nil {
		return Task{}, err
	}
	response, err := client.Post("/tasks/"+strconv.Itoa(ID), string(taskJson))
	if err != nil {
		return Task{}, err
	}
	var updatedTask Task
	err = json.Unmarshal([]byte(response), &updatedTask)
	if err != nil {
		return Task{}, err
	}
	return updatedTask, nil
}

// DeleteTask deletes a task from a project. This does not mean "mark it done".
func (client *ApiClient) DeleteTask(ID int) (string, error) {
	response, err := client.Delete("/tasks/" + strconv.Itoa(ID))
	if err != nil {
		return "", err
	}
	return response, nil
}

// GetTask returns one task by its ID
func (client *ApiClient) GetTask(ID int) (Task, error) {
	response, err := client.Get("/tasks/" + strconv.Itoa(ID))
	if err != nil {
		return Task{}, err
	}
	var task Task
	err = json.Unmarshal([]byte(response), &task)
	if err != nil {
		return Task{}, err
	}
	return task, nil
}
/*
TODO: implement this api method to create a task


        "/projects/{id}/tasks": {
            "put": {
                "security": [
                    {
                        "JWTKeyAuth": []
                    }
                ],
                "description": "Inserts a task into a project.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "task"
                ],
                "summary": "Create a task",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Project ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "The task object",
                        "name": "task",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.Task"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "The created task object.",
                        "schema": {
                            "$ref": "#/definitions/models.Task"
                        }
                    },
                    "400": {
                        "description": "Invalid task object provided.",
                        "schema": {
                            "$ref": "#/definitions/web.HTTPError"
                        }
                    },
                    "403": {
                        "description": "The user does not have access to the project",
                        "schema": {
                            "$ref": "#/definitions/web.HTTPError"
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
func (client *ApiClient) CreateTask(projectID int, task Task) (Task, error) {
	taskJson, err := json.Marshal(task)
	if err != nil {
		return Task{}, err
	}
	response, err := client.Put("/projects/"+strconv.Itoa(projectID), string(taskJson))// +"/tasks"
	if err != nil {
		return Task{}, err
	}
	var createdTask Task
	err = json.Unmarshal([]byte(response), &createdTask)
	if err != nil {
		return Task{}, err
	}
	return createdTask, nil
}
