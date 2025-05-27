package core

import "time"

// CalculateUrgency computes the urgency score of the task according to the
// Vikunja formula.  It is attached to *Task here (the type owner) so that the
// method set is available through the api.Task alias as well.
func (task *Task) CalculateUrgency() {
	if task.Done {
		task.Urgency = 0
		return
	}

	dueDateScore := float64(task.getDueDateScore())
	priorityScore := float64(task.Priority)

	favoriteScore := 0.0
	if task.IsFavorite {
		favoriteScore = 1
	}

	// Base 1.0 is added so that unfinished, no-priority tasks are still sorted
	// ahead of completed ones (which are set to 0).
	task.Urgency = 1 + dueDateScore + priorityScore + favoriteScore
}

// getDueDateScore converts the remaining days until due-date into the score
// defined by Vikunja’s web client.
func (task *Task) getDueDateScore() int {
	if task.DueDate.IsZero() {
		return 0
	}

	dueDays := int(task.DueDate.Sub(time.Now()).Hours() / 24)

	switch {
	case dueDays < 0:
		return 6    // overdue
	case dueDays == 0:
		return 5
	case dueDays == 1:
		return 4
	case dueDays <= 2:
		return 3
	case dueDays <= 5:
		return 2
	case dueDays <= 10:
		return 1
	case dueDays > 14:
		return -1   // “Someday”
	default: // 11-14 days
		return 0
	}
}
