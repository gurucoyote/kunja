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
		task.Urgency = 0
		return
	}
	dueDateScore := task.getDueDateScore()
	priorityScore := task.Priority
	favoriteScore := 0
	if task.IsFavorite {
		favoriteScore = 1
	}
	task.Urgency = 1 + dueDateScore + priorityScore + favoriteScore
}

func (task *Task) getDueDateScore() int {
	if task.DueDate == nil {
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
