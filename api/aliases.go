package api

import "kunja/internal/core"

// Alias the pure domain models from internal/core so existing
// packages can keep importing `kunja/api` without churn.
// All behaviour now lives in core; api only re-exports the types.
type (
    Bucket            = core.Bucket
    Label             = core.Label
    TaskReminder      = core.TaskReminder
    Task              = core.Task
    GetAllTasksParams = core.GetAllTasksParams
    Project           = core.Project
    UserWithRight     = core.UserWithRight
    User              = core.User
)
