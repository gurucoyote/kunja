package api

import "time"

type Bucket struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	IsDoneBucket bool   `json:"is_done_bucket"`
	Limit        int    `json:"limit"`
	Position     int    `json:"position"`
	CountTasks   int    `json:"count_tasks"`
}

type Label struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type TaskReminder struct {
	Reminder       time.Time `json:"reminder"`
	RelativePeriod int       `json:"relative_period"`
	RelativeTo     string    `json:"relative_to"`
}

type Task struct {
	ID             int            `json:"id"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	Priority       int            `json:"priority"`
	IsFavorite     bool           `json:"is_favorite"`
	DueDate        time.Time      `json:"due_date"` // Changed from string to time.Time
	Reminders      []TaskReminder `json:"reminders"`
	RepeatMode     int            `json:"repeat_mode"`
	RepeatAfter    int            `json:"repeat_after"`
	StartDate      time.Time         `json:"start_date"`
	EndDate        time.Time         `json:"end_date"`
	PercentDone    float64        `json:"percent_done"`
	Done           bool           `json:"done"`
	DoneAt         time.Time      `json:"done_at"`
	Labels         []Label        `json:"labels"`
	ProjectID      int            `json:"project_id,omitempty"`
	Position       float64        `json:"position"`
	BucketID       int            `json:"bucket_id"`
	KanbanPosition float64        `json:"kanban_position"`
	Created        time.Time         `json:"created"`
	Updated        time.Time         `json:"updated"`
	Urgency        float64        `json:"urgency"` // Added Urgency field
}
type GetAllTasksParams struct {
	Page               int    `json:"page,omitempty"`
	PerPage            int    `json:"per_page,omitempty"`
	S                  string `json:"s,omitempty"`
	SortBy             string `json:"sort_by,omitempty"`
	OrderBy            string `json:"order_by,omitempty"`
	FilterBy           string `json:"filter_by,omitempty"`
	FilterValue        string `json:"filter_value,omitempty"`
	FilterComparator   string `json:"filter_comparator,omitempty"`
	FilterConcat       string `json:"filter_concat,omitempty"`
	FilterIncludeNulls string `json:"filter_include_nulls,omitempty"`
}

// Project represents a project in the Vikunja API.
type Project struct {
	ID                    int       `json:"id"`
	Title                 string    `json:"title"`
	Description           string    `json:"description"`
	IsFavorite            bool      `json:"is_favorite"`
	IsArchived            bool      `json:"is_archived"`
	ParentProjectID       int       `json:"parent_project_id"`
	AncestorProjects      []Project `json:"ancestor_projects"`
	Created               string    `json:"created"`
	Updated               string    `json:"updated"`
	Owner                 User      `json:"owner"`
	Position              float64   `json:"position"`
	Identifier            string    `json:"identifier"`
	BackgroundBlurHash    string    `json:"background_blur_hash,omitempty"`
	BackgroundInformation *string   `json:"background_information,omitempty"`
	HexColor              string    `json:"hex_color,omitempty"`
}

// User represents a user in the Vikunja API.
type UserWithRight struct {
	ID    int    `json:"id"`
	Username  string `json:"username"`
	Right int    `json:"right"`
}

type User struct {
	ID               int    `json:"id"`
	Username         string `json:"username"`
	Name             string `json:"name"`
	DefaultProjectID int    `json:"default_project_id"`
}
