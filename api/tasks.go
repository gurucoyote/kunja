package api

import (
	"encoding/json"
	"github.com/google/go-querystring/query"
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
/*
TODO: please implement these missing task related api methods
TODO: add a short descriptive comment in go doc style explaining what each method does and is used for
        "/tasks/{ID}": {
            "post": {
                "security": [
                    {
                        "JWTKeyAuth": []
                    }
                ],
                "description": "Updates a task. This includes marking it as done. Assignees you pass will be updated, see their individual endpoints for more details on how this is done. To update labels, see the description of the endpoint.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "task"
                ],
                "summary": "Update a task",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "The Task ID",
                        "name": "ID",
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
                    "200": {
                        "description": "The updated task object.",
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
                        "description": "The user does not have access to the task (aka its project)",
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
            },
            "delete": {
                "security": [
                    {
                        "JWTKeyAuth": []
                    }
                ],
                "description": "Deletes a task from a project. This does not mean \"mark it done\".",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "task"
                ],
                "summary": "Delete a task",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Task ID",
                        "name": "ID",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "The created task object.",
                        "schema": {
                            "$ref": "#/definitions/models.Message"
                        }
                    },
                    "400": {
                        "description": "Invalid task ID provided.",
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
        "/tasks/{id}": {
            "get": {
                "security": [
                    {
                        "JWTKeyAuth": []
                    }
                ],
                "description": "Returns one task by its ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "task"
                ],
                "summary": "Get one task",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "The task ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "The task",
                        "schema": {
                            "$ref": "#/definitions/models.Task"
                        }
                    },
                    "404": {
                        "description": "Task not found",
                        "schema": {
                            "$ref": "#/definitions/models.Message"
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
// UpdateTask updates a task. This includes marking it as done. 
// Assignees you pass will be updated, see their individual endpoints for more details on how this is done. 
// To update labels, see the description of the endpoint.
func (client *ApiClient) UpdateTask(ID int, task Task) (Task, error) {
	response, err := client.Post("/tasks/"+strconv.Itoa(ID), task)
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
	response, err := client.Delete("/tasks/"+strconv.Itoa(ID))
	if err != nil {
		return "", err
	}
	return response, nil
}

// GetTask returns one task by its ID
func (client *ApiClient) GetTask(ID int) (Task, error) {
	response, err := client.Get("/tasks/"+strconv.Itoa(ID))
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
