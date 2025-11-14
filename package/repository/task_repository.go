package repository

import (
	"fmt"
	"time"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/utils"
	"gorm.io/gorm"
)

type TaskRepository interface {
	// Create creates a new empty task and returns the task ID.
	Create(tx *gorm.DB, task *models.Task) (int64, error)
	// Delete deletes a task. It does not actually delete the task from the database, but performs a soft delete.
	Delete(tx *gorm.DB, taskID int64) error
	// Edit edits a task, by setting the fields of the task to the fields of the function argument.
	Edit(tx *gorm.DB, taskID int64, task *models.Task) error
	// GetAllCreated returns all tasks created by a user. The tasks are paginated.
	GetAllCreated(tx *gorm.DB, userID int64, offset, limit int, sort string) ([]models.Task, int64, error)
	// GetAll returns all tasks. The tasks are paginated.
	GetAll(tx *gorm.DB, limit, offset int, sort string) ([]models.Task, int64, error)
	// Get returns a task by its ID.
	Get(tx *gorm.DB, taskID int64) (*models.Task, error)
	// GetByTitle returns a task by its title.
	GetByTitle(tx *gorm.DB, title string) (*models.Task, error)
	// GetLiveAssignedTasksGroupedByContest returns live assigned tasks grouped by contest for a user.
	// Live tasks are those with start_at before now and (end_at null or end_at in future).
	GetLiveAssignedTasksGroupedByContest(tx *gorm.DB, userID int64, limit, offset int) (map[int64][]models.Task, error)
}

type taskRepository struct {
}

func (tr *taskRepository) Create(tx *gorm.DB, task *models.Task) (int64, error) {
	err := tx.Model(models.Task{}).Create(&task).Error
	if err != nil {
		return 0, err
	}
	return task.ID, nil
}

func (tr *taskRepository) GetByTitle(tx *gorm.DB, title string) (*models.Task, error) {
	task := &models.Task{}
	err := tx.Model(&models.Task{}).Where("title = ?", title).First(task).Error
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (tr *taskRepository) Get(tx *gorm.DB, taskID int64) (*models.Task, error) {
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
	tx *gorm.DB,
	userID int64,
	offset, limit int,
	sort string,
) ([]models.Task, int64, error) {
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

func (tr *taskRepository) GetAll(tx *gorm.DB, limit, offset int, sort string) ([]models.Task, int64, error) {
	tasks := []models.Task{}
	var totalCount int64

	// Get total count first
	err := tx.Model(&models.Task{}).Where("deleted_at IS NULL").Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	paginatedTx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}
	err = paginatedTx.Model(&models.Task{}).Where("deleted_at IS NULL").Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}
	return tasks, totalCount, nil
}

func (tr *taskRepository) GetTimeLimits(tx *gorm.DB, taskID int64) ([]int64, error) {
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
		timeLimits[testCase.Order-1] = testCase.TimeLimit
	}
	return timeLimits, nil
}

func (tr *taskRepository) GetMemoryLimits(tx *gorm.DB, taskID int64) ([]int64, error) {
	testCases := []models.TestCase{}
	err := tx.Model(&models.TestCase{}).Where("task_id = ?", taskID).Find(&testCases).Error
	if err != nil {
		return nil, err
	}
	// Sort by order
	memoryLimits := make([]int64, len(testCases))
	for _, testCase := range testCases {
		memoryLimits[testCase.Order-1] = testCase.MemoryLimit
	}
	return memoryLimits, nil
}

func (tr *taskRepository) Edit(tx *gorm.DB, taskID int64, task *models.Task) error {
	err := tx.Model(&models.Task{}).Where("id = ?", taskID).Updates(task).Error
	if err != nil {
		return err
	}
	return nil
}

func (tr *taskRepository) Delete(tx *gorm.DB, taskID int64) error {
	err := tx.Model(&models.Task{}).Where("id = ?", taskID).Update("deleted_at", time.Now()).Error
	if err != nil {
		return err
	}
	return nil
}

func (tr *taskRepository) GetLiveAssignedTasksGroupedByContest(
	tx *gorm.DB,
	userID int64,
	limit, offset int,
) (map[int64][]models.Task, error) {
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
		// Limit(limit).
		// Offset(offset).
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
