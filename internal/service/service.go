package service

import "kunja/api"

// AuthService defines authentication related operations.
type AuthService interface {
	Login(username, password, totpPasscode string) (string, error)
}

// TaskService defines task related operations.
type TaskService interface {
	GetAllTasks(params api.GetAllTasksParams) ([]api.Task, error)
	GetTask(id int) (api.Task, error)
	CreateTask(projectID int, task api.Task) (api.Task, error)
	UpdateTask(id int, task api.Task) (api.Task, error)
	DeleteTask(id int) (string, error)
	AssignUserToTask(taskID, userID int) (string, error)
	GetTaskAssignees(taskID int) ([]api.User, error)
}

// ProjectService defines project related operations.
type ProjectService interface {
	GetAllProjects() ([]api.Project, error)
	GetProject(id int) (api.Project, error)
	GetProjectUsers(projectID int) ([]api.UserWithRight, error)
}

// UserService defines user related operations.
type UserService interface {
	GetAllUsers() ([]api.User, error)
}
