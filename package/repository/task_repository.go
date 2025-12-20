package repository

import (
	"fmt"
	"time"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/utils"
)

type TaskRepository interface {
	// Create creates a new empty task and returns the task ID.
	Create(db database.Database, task *models.Task) (int64, error)
	// Delete deletes a task. It does not actually delete the task from the database, but performs a soft delete.
	Delete(db database.Database, taskID int64) error
	// Edit edits a task, by updating only the fields provided in the updates map.
	Edit(db database.Database, taskID int64, updates map[string]any) error
	// GetAllCreated returns all tasks created by a user. The tasks are paginated.
	GetAllCreated(db database.Database, userID int64, offset, limit int, sort string) ([]models.Task, int64, error)
	// GetAll returns all tasks. The tasks are paginated.
	GetAll(db database.Database, limit, offset int, sort string) ([]models.Task, int64, error)
	// Get returns a task by its ID.
	Get(db database.Database, taskID int64) (*models.Task, error)
	// GetByTitle returns a task by its title.
	GetByTitle(db database.Database, title string) (*models.Task, error)
	// GetLiveAssignedTasksGroupedByContest returns live assigned tasks grouped by contest for a user.
	// Live tasks are those with start_at before now and (end_at null or end_at in future).
	GetLiveAssignedTasksGroupedByContest(db database.Database, userID int64, limit, offset int) (map[int64][]models.Task, error)
}

type taskRepository struct {
}

func (tr *taskRepository) Create(db database.Database, task *models.Task) (int64, error) {
	tx := db.GetInstance()
	err := tx.Model(models.Task{}).Create(&task).Error
	if err != nil {
		return 0, err
	}
	return task.ID, nil
}

func (tr *taskRepository) GetByTitle(db database.Database, title string) (*models.Task, error) {
	tx := db.GetInstance()
	task := &models.Task{}
	err := tx.Model(&models.Task{}).Where("title = ?", title).First(task).Error
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (tr *taskRepository) Get(db database.Database, taskID int64) (*models.Task, error) {
	tx := db.GetInstance()
	task := &models.Task{}
	err := tx.Preload("Author").Preload("DescriptionFile").Model(&models.Task{}).Where(
		"id = ? AND deleted_at IS NULL",
		taskID,
	).First(task).Error
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (tr *taskRepository) GetAllCreated(
	db database.Database,
	userID int64,
	offset, limit int,
	sort string,
) ([]models.Task, int64, error) {
	tx := db.GetInstance()
	var tasks []models.Task
	var totalCount int64

	// Get total count first
	baseQuery := tx.Model(&models.Task{}).Where("created_by = ?", userID)
	err := baseQuery.Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting to a new query
	paginatedQuery, err := utils.ApplyPaginationAndSort(tx.Model(&models.Task{}), limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	err = paginatedQuery.
		Where("created_by = ?", userID).
		Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}
	return tasks, totalCount, nil
}

func (tr *taskRepository) GetAll(db database.Database, limit, offset int, sort string) ([]models.Task, int64, error) {
	tx := db.GetInstance()
	tasks := []models.Task{}
	var totalCount int64

	// Get total count first (only globally visible tasks)
	err := tx.Model(&models.Task{}).Where("deleted_at IS NULL AND is_visible = ?", true).Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	paginatedTx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}
	err = paginatedTx.Model(&models.Task{}).Where("deleted_at IS NULL AND is_visible = ?", true).Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}
	return tasks, totalCount, nil
}

func (tr *taskRepository) GetTimeLimits(db database.Database, taskID int64) ([]int64, error) {
	tx := db.GetInstance()
	testCases := []models.TestCase{}
	err := tx.Model(&models.TestCase{}).
		Where("task_id = ?", taskID).
		Find(&testCases).Error
	if err != nil {
		return nil, err
	}
	// Sort by order
	timeLimits := make([]int64, len(testCases))
	for _, testCase := range testCases {
		timeLimits[testCase.Order-1] = testCase.TimeLimitMs
	}
	return timeLimits, nil
}

func (tr *taskRepository) GetMemoryLimits(db database.Database, taskID int64) ([]int64, error) {
	tx := db.GetInstance()
	testCases := []models.TestCase{}
	err := tx.Model(&models.TestCase{}).Where("task_id = ?", taskID).Find(&testCases).Error
	if err != nil {
		return nil, err
	}
	// Sort by order
	memoryLimits := make([]int64, len(testCases))
	for _, testCase := range testCases {
		memoryLimits[testCase.Order-1] = testCase.MemoryLimitKB
	}
	return memoryLimits, nil
}

func (tr *taskRepository) Edit(db database.Database, taskID int64, updates map[string]any) error {
	tx := db.GetInstance()
	err := tx.Model(&models.Task{}).Where("id = ?", taskID).Updates(updates).Error
	if err != nil {
		return err
	}
	return nil
}

func (tr *taskRepository) Delete(db database.Database, taskID int64) error {
	tx := db.GetInstance()
	err := tx.Model(&models.Task{}).Where("id = ?", taskID).Update("deleted_at", time.Now()).Error
	if err != nil {
		return err
	}
	return nil
}

func (tr *taskRepository) GetLiveAssignedTasksGroupedByContest(
	db database.Database,
	userID int64,
	limit, offset int,
) (map[int64][]models.Task, error) {
	tx := db.GetInstance()
	type TaskWithContest struct {
		models.Task
		ContestID   int64      `gorm:"column:contest_id"`
		ContestName string     `gorm:"column:contest_name"`
		StartAt     *time.Time `gorm:"column:start_at"`
		EndAt       *time.Time `gorm:"column:end_at"`
	}

	var tasksWithContests []TaskWithContest

	// Get tasks assigned to user through contests they participate in
	// Only tasks that are "live" (started and not ended)
	err := tx.Model(&models.Task{}).
		Select("DISTINCT tasks.*, contest_tasks.contest_id, contests.name as contest_name, contest_tasks.start_at, contest_tasks.end_at").
		Joins(fmt.Sprintf("JOIN %s ON contest_tasks.task_id = tasks.id", database.ResolveTableName(tx, &models.ContestTask{}))).
		Joins(fmt.Sprintf("JOIN %s ON contests.id = contest_tasks.contest_id", database.ResolveTableName(tx, &models.Contest{}))).
		Joins(fmt.Sprintf(`LEFT JOIN %s ON contest_participants.contest_id = contests.id AND contest_participants.user_id = ?`, database.ResolveTableName(tx, &models.ContestParticipant{})), userID).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT DISTINCT cpg.contest_id, ug.user_id
			FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			WHERE ug.user_id = ?
		) as user_group_participants ON contests.id = user_group_participants.contest_id`, database.ResolveTableName(tx, &models.ContestParticipantGroup{}), database.ResolveTableName(tx, &models.UserGroup{})), userID).
		Where("(contest_participants.user_id IS NOT NULL OR user_group_participants.user_id IS NOT NULL)").
		Where("contest_tasks.start_at <= NOW() AND (contest_tasks.end_at IS NULL OR contest_tasks.end_at > NOW())").
		Where("tasks.deleted_at IS NULL").
		Find(&tasksWithContests).Error

	if err != nil {
		return nil, err
	}

	// Group tasks by contest
	result := make(map[int64][]models.Task)
	for _, twc := range tasksWithContests {
		result[twc.ContestID] = append(result[twc.ContestID], twc.Task)
	}

	return result, nil
}

func NewTaskRepository() TaskRepository {
	return &taskRepository{}
}
