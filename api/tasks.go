package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-querystring/query"
	"strconv"
)

func (client *ApiClient) GetAllTasks(ctx context.Context, params GetAllTasksParams) ([]Task, error) {
	queryParams, _ := query.Values(params)
	response, err := client.getCtx(ctx, "/tasks/all?"+queryParams.Encode())
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


// UpdateTask updates a task. This includes marking it as done.
// Assignees you pass will be updated, see their individual endpoints for more details on how this is done.
// To update labels, see the description of the endpoint.
func (client *ApiClient) UpdateTask(ctx context.Context, ID int, task Task) (Task, error) {
	taskJson, err := json.Marshal(task)
	if err != nil {
		return Task{}, err
	}
	response, err := client.postCtx(ctx, "/tasks/"+strconv.Itoa(ID), string(taskJson))
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
func (client *ApiClient) DeleteTask(ctx context.Context, ID int) (string, error) {
	response, err := client.deleteCtx(ctx, "/tasks/"+strconv.Itoa(ID))
	if err != nil {
		return "", err
	}
	return response, nil
}

// GetTask returns one task by its ID
func (client *ApiClient) GetTask(ctx context.Context, ID int) (Task, error) {
	response, err := client.getCtx(ctx, "/tasks/"+strconv.Itoa(ID)+"?include=project,label_objects,assignees")
	if err != nil {
		return Task{}, err
	}
	var task Task
	err = json.Unmarshal([]byte(response), &task)
	if err != nil {
		return Task{}, err
	}
	task.CalculateUrgency()
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
func (client *ApiClient) CreateTask(ctx context.Context, projectID int, task Task) (Task, error) {
	taskJson, err := json.Marshal(task)
	if err != nil {
		return Task{}, err
	}
	response, err := client.putCtx(ctx, "/projects/"+strconv.Itoa(projectID), string(taskJson))
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

// AssignUserToTask assigns a user to a task by making a PUT request to the /tasks/{taskID}/assignees endpoint.
func (client *ApiClient) AssignUserToTask(ctx context.Context, taskID int, userID int) (string, error) {
	// Construct the API endpoint with the taskID
	apiEndpoint := fmt.Sprintf("/tasks/%d/assignees", taskID)

	// Create the payload with the userID
	payload := map[string]int{"user_id": userID}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// Make the PUT request
	response, err := client.putCtx(ctx, apiEndpoint, string(payloadBytes))
	if err != nil {
		return "", err
	}

	return response, nil
}

// GetTaskAssignees retrieves all assignees for a given task.
func (client *ApiClient) GetTaskAssignees(ctx context.Context, taskID int) ([]User, error) {
	apiEndpoint := fmt.Sprintf("/tasks/%d/assignees", taskID)

	response, err := client.getCtx(ctx, apiEndpoint)
	if err != nil {
		return nil, err
	}

	var assignees []User
	err = json.Unmarshal([]byte(response), &assignees)
	if err != nil {
		return nil, err
	}

	return assignees, nil
}
