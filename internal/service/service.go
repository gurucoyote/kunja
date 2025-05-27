package service

import (
	"context"

	"kunja/api"
)

// AuthService defines authentication related operations.
type AuthService interface {
	Login(ctx context.Context, username, password, totpPasscode string) (string, error)
}

// TaskService defines task related operations.
type TaskService interface {
	GetAllTasks(ctx context.Context, params api.GetAllTasksParams) ([]api.Task, error)
	GetTask(ctx context.Context, id int) (api.Task, error)
	CreateTask(ctx context.Context, projectID int, task api.Task) (api.Task, error)
	UpdateTask(ctx context.Context, id int, task api.Task) (api.Task, error)
	DeleteTask(ctx context.Context, id int) (string, error)
	AssignUserToTask(ctx context.Context, taskID, userID int) (string, error)
	GetTaskAssignees(ctx context.Context, taskID int) ([]api.User, error)
}

// ProjectService defines project related operations.
type ProjectService interface {
	GetAllProjects(ctx context.Context) ([]api.Project, error)
	GetProject(ctx context.Context, id int) (api.Project, error)
	GetProjectUsers(ctx context.Context, projectID int) ([]api.UserWithRight, error)

	CreateProject(ctx context.Context, p api.Project) (api.Project, error)
	DeleteProject(ctx context.Context, id int) (string, error)
}

// UserService defines user related operations.
type UserService interface {
	GetAllUsers(ctx context.Context) ([]api.User, error)
}
