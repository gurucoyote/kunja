package api

import "time"

type User struct {
	ID                int    `json:"id"`
	Username          string `json:"username"`
	Name              string `json:"name"`
	DefaultProjectID  int    `json:"default_project_id"`
}

type Project struct {
	ID                int       `json:"id"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	IsFavorite        bool      `json:"is_favorite"`
	IsArchived        bool      `json:"is_archived"`
	ParentProjectID   int       `json:"parent_project_id"`
	AncestorProjects  []Project `json:"ancestor_projects"`
}

type Bucket struct {
	ID             int    `json:"id"`
	Title          string `json:"title"`
	IsDoneBucket   bool   `json:"is_done_bucket"`
	Limit          int    `json:"limit"`
	Position       int    `json:"position"`
	CountTasks     int    `json:"count_tasks"`
}

type Label struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type TaskReminder struct {
	Reminder        time.Time `json:"reminder"`
	RelativePeriod  int       `json:"relative_period"`
	RelativeTo      string    `json:"relative_to"`
}

type Task struct {
	ID              int             `json:"id"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	Priority        int             `json:"priority"`
	IsFavorite      bool            `json:"is_favorite"`
	DueDate         time.Time       `json:"due_date"`
	Reminders       []TaskReminder  `json:"reminders"`
	RepeatMode      int             `json:"repeat_mode"`
	RepeatAfter     time.Duration   `json:"repeat_after"`
	StartDate       time.Time       `json:"start_date"`
	EndDate         time.Time       `json:"end_date"`
	PercentDone     float64         `json:"percent_done"`
	Done            bool            `json:"done"`
	DoneAt          time.Time       `json:"done_at"`
	LabelObjects    []Label         `json:"label_objects"`
	Project         Project         `json:"project"`
	Position        int             `json:"position"`
	BucketID        int             `json:"bucket_id"`
	KanbanPosition  int             `json:"kanban_position"`
	Created         time.Time       `json:"created"`
	Updated         time.Time       `json:"updated"`
	Urgency         float64         `json:"-"`
}
