package vikunja

import (
	"context"

	"kunja/api"
	"kunja/internal/service"
)

// Adapter implements the service.* interfaces by delegating to api.ApiClient.
type Adapter struct {
	client *api.ApiClient
}

// New returns a new Adapter wrapping the given ApiClient.
func New(client *api.ApiClient) *Adapter {
	return &Adapter{client: client}
}

/* ---- AuthService ---- */

func (a *Adapter) Login(ctx context.Context, username, password, totpPasscode string) (string, error) {
	return a.client.Login(ctx, username, password, totpPasscode)
}

/* ---- TaskService ---- */

func (a *Adapter) GetAllTasks(ctx context.Context, params api.GetAllTasksParams) ([]api.Task, error) {
	return a.client.GetAllTasks(params)
}

func (a *Adapter) GetTask(ctx context.Context, id int) (api.Task, error) {
	return a.client.GetTask(id)
}

func (a *Adapter) CreateTask(ctx context.Context, projectID int, task api.Task) (api.Task, error) {
	return a.client.CreateTask(projectID, task)
}

func (a *Adapter) UpdateTask(ctx context.Context, id int, task api.Task) (api.Task, error) {
	return a.client.UpdateTask(id, task)
}

func (a *Adapter) DeleteTask(ctx context.Context, id int) (string, error) {
	return a.client.DeleteTask(id)
}

func (a *Adapter) AssignUserToTask(ctx context.Context, taskID, userID int) (string, error) {
	return a.client.AssignUserToTask(taskID, userID)
}

func (a *Adapter) GetTaskAssignees(ctx context.Context, taskID int) ([]api.User, error) {
	return a.client.GetTaskAssignees(taskID)
}

/* ---- ProjectService ---- */

func (a *Adapter) GetAllProjects(ctx context.Context) ([]api.Project, error) {
	return a.client.GetAllProjects(ctx)
}

func (a *Adapter) GetProject(ctx context.Context, id int) (api.Project, error) {
	return a.client.GetProject(ctx, id)
}

func (a *Adapter) GetProjectUsers(ctx context.Context, projectID int) ([]api.UserWithRight, error) {
	return a.client.GetProjectUsers(ctx, projectID)
}

/* ---- UserService ---- */

func (a *Adapter) GetAllUsers(ctx context.Context) ([]api.User, error) {
	return a.client.GetAllUsers(ctx)
}

// Ensure compile-time interface satisfaction
var (
	_ service.AuthService    = (*Adapter)(nil)
	_ service.TaskService    = (*Adapter)(nil)
	_ service.ProjectService = (*Adapter)(nil)
	_ service.UserService    = (*Adapter)(nil)
)
