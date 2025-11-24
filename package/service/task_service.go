package service

import (
	"fmt"
	"os"
	"strings"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/filestorage"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TaskService interface {
	// Create creates a new task.
	Create(tx *gorm.DB, currentUser *schemas.User, task *schemas.Task) (int64, error)
	// CreateTestCase creates input and output files for a task.
	CreateTestCase(tx *gorm.DB, taskID int64, archivePath string) error
	// Delete deletes a task.
	Delete(tx *gorm.DB, currentUser *schemas.User, taskID int64) error
	// Edit edits an existing task.
	Edit(tx *gorm.DB, currentUser *schemas.User, taskID int64, updateInfo *schemas.EditTask) error
	// GetAll retrieves all tasks based on query parameters.
	GetAll(tx *gorm.DB, currentUser *schemas.User, paginationParams schemas.PaginationParams) ([]schemas.Task, error)
	// GetAllCreated retrieves all tasks created by the current user.
	GetAllCreated(tx *gorm.DB, currentUser *schemas.User, paginationParams schemas.PaginationParams) (schemas.PaginatedResult[[]schemas.Task], error)
	// Get retrieves a detailed view of a specific task.
	Get(tx *gorm.DB, currentUser *schemas.User, taskID int64) (*schemas.TaskDetailed, error)
	// GetByTitle retrieves a task by its title.
	GetByTitle(tx *gorm.DB, title string) (*schemas.Task, error)
	// GetLimits retrieves limits associated with each input/output
	GetLimits(tx *gorm.DB, currentUser *schemas.User, taskID int64) ([]schemas.TestCase, error)
	// GetMyLiveTasks retrieves live assigned tasks grouped by contests and non-contest tasks with submission statistics.
	GetMyLiveTasks(tx *gorm.DB, currentUser *schemas.User, paginationParams schemas.PaginationParams) (*schemas.MyTasksResponse, error)
	// ParseTestCase parses the input and output files from an archive.
	ParseTestCase(archivePath string) (int, error)
	// ProcessAndUpload processes and uploads input and output files for a task.
	ProcessAndUpload(tx *gorm.DB, currentUser *schemas.User, taskID int64, archivePath string) error
	// PutLimits updates limits associated with each input/output.
	PutLimits(tx *gorm.DB, currentUser *schemas.User, taskID int64, limits schemas.PutTestCaseLimitsRequest) error
}

const defaultTaskSort = "created_at:desc"

type taskService struct {
	filestorage          filestorage.FileStorageService
	fileRepository       repository.File
	groupRepository      repository.GroupRepository
	testCaseRepository   repository.TestCaseRepository
	taskRepository       repository.TaskRepository
	userRepository       repository.UserRepository
	submissionRepository repository.SubmissionRepository
	contestRepository    repository.ContestRepository
	accessControlService AccessControlService

	logger *zap.SugaredLogger
}

func (ts *taskService) Create(tx *gorm.DB, currentUser *schemas.User, task *schemas.Task) (int64, error) {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return 0, err
	}
	// Create a new task
	_, err = ts.GetByTitle(tx, task.Title)
	if err != nil && !errors.Is(err, errors.ErrTaskNotFound) {
		ts.logger.Errorf("Error getting task by title: %v", err.Error())
		return 0, err
	} else if err == nil {
		return 0, errors.ErrTaskExists
	}

	author, err := ts.userRepository.Get(tx, currentUser.ID)
	if err != nil {
		ts.logger.Errorf("Error getting user: %v", err.Error())
		return 0, err
	}

	model := models.Task{
		Title:     task.Title,
		CreatedBy: task.CreatedBy,
		Author:    *author,
		IsVisible: true, // Default to globally visible
	}
	taskID, err := ts.taskRepository.Create(tx, &model)
	if err != nil {
		ts.logger.Errorf("Error creating task: %v", err.Error())
		return 0, err
	}

	// Automatically grant owner permission to the creator (immutable highest level)
	if err := ts.accessControlService.GrantOwnerAccess(tx, models.ResourceTypeTask, taskID, currentUser.ID); err != nil {
		ts.logger.Warnf("Failed to grant owner permission: %v", err)
		// Don't fail the creation if we can't add owner permission entry
	}

	return taskID, nil
}

func (ts *taskService) GetByTitle(tx *gorm.DB, title string) (*schemas.Task, error) {
	task, err := ts.taskRepository.GetByTitle(tx, title)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrTaskNotFound
		}
		ts.logger.Errorf("Error getting task by title: %v", err.Error())
		return nil, err
	}

	return TaskToSchema(task), nil
}

func (ts *taskService) GetAll(tx *gorm.DB, _ *schemas.User, paginationParams schemas.PaginationParams) ([]schemas.Task, error) {
	// err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin})
	// if err != nil {
	// 	ts.logger.Errorf("Error validating user role: %v", err.Error())
	// 	return nil, err
	// }
	if paginationParams.Sort == "" {
		paginationParams.Sort = defaultTaskSort
	}

	// Get all tasks
	tasks, _, err := ts.taskRepository.GetAll(tx, paginationParams.Limit, paginationParams.Offset, paginationParams.Sort)
	if err != nil {
		ts.logger.Errorf("Error getting all tasks: %v", err.Error())
		return nil, err
	}

	// Convert the models to schemas
	result := make([]schemas.Task, len(tasks))
	for i, task := range tasks {
		result[i] = *TaskToSchema(&task)
	}

	return result, nil
}

func (ts *taskService) Get(tx *gorm.DB, _ *schemas.User, taskID int64) (*schemas.TaskDetailed, error) {
	// Get the task
	task, err := ts.taskRepository.Get(tx, taskID)
	if err != nil {
		ts.logger.Errorf("Error getting task: %v", err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}

	// switch types.UserRole(currentUser.Role) {
	// case types.UserRoleStudent:
	// 	// Check if the task is assigned to the user
	// 	isAssigned, err := ts.taskRepository.IsTaskAssignedToUser(tx, taskID, currentUser.ID)
	// 	if err != nil {
	// 		ts.logger.Errorf("Error checking if task is assigned to user: %v", err.Error())
	// 		return nil, err
	// 	}
	// 	if !isAssigned {
	// 		return nil, errors.ErrNotAuthorized
	// 	}
	// case types.UserRoleTeacher:
	// 	// Check if the task is created by the user
	// 	if task.CreatedBy != currentUser.ID {
	// 		return nil, errors.ErrNotAuthorized
	// 	}
	// }

	result := &schemas.TaskDetailed{
		ID:             task.ID,
		Title:          task.Title,
		DescriptionURL: ts.filestorage.GetFileURL(task.DescriptionFile.Path),
		CreatedBy:      task.CreatedBy,
		CreatedByName:  task.Author.Name,
		CreatedAt:      task.CreatedAt,
	}

	return result, nil
}

func (ts *taskService) GetAllCreated(
	tx *gorm.DB,
	currentUser *schemas.User,
	paginationParams schemas.PaginationParams,
) (schemas.PaginatedResult[[]schemas.Task], error) {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return schemas.PaginatedResult[[]schemas.Task]{}, err
	}

	if paginationParams.Sort == "" {
		paginationParams.Sort = defaultTaskSort
	}
	tasks, totalCount, err := ts.taskRepository.GetAllCreated(tx, currentUser.ID, paginationParams.Offset, paginationParams.Limit, paginationParams.Sort)
	if err != nil {
		ts.logger.Errorf("Error getting all created tasks: %v", err.Error())
		return schemas.PaginatedResult[[]schemas.Task]{}, err
	}

	result := make([]schemas.Task, len(tasks))
	for i, task := range tasks {
		result[i] = *TaskToSchema(&task)
	}

	return schemas.NewPaginatedResult(result, paginationParams.Offset, paginationParams.Limit, totalCount), nil
}

// TODO: remove access control associated with the task
func (ts *taskService) Delete(tx *gorm.DB, currentUser *schemas.User, taskID int64) error {
	err := ts.hasTaskPermission(tx, taskID, currentUser, types.PermissionOwner)
	if err != nil {
		return err
	}
	err = ts.taskRepository.Delete(tx, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrNotFound
		}
		ts.logger.Errorf("Error deleting task: %v", err.Error())
		return err
	}

	return nil
}

func (ts *taskService) Edit(tx *gorm.DB, currentUser *schemas.User, taskID int64, updateInfo *schemas.EditTask) error {
	err := ts.hasTaskPermission(tx, taskID, currentUser, types.PermissionEdit)
	if err != nil {
		return err
	}

	currentTask, err := ts.taskRepository.Get(tx, taskID)
	if err != nil {
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return err
	}

	ts.updateModel(currentTask, updateInfo)

	// Update the task
	err = ts.taskRepository.Edit(tx, taskID, currentTask)
	if err != nil {
		ts.logger.Errorf("Error updating task: %v", err.Error())
		return err
	}
	return nil
}

func (ts *taskService) ParseTestCase(archivePath string) (int, error) {
	// Unzip the archive
	// Extract the input and output files
	// Validate correct number of input and output files and naming
	// Count the number of input and output files
	archive, err := os.Open(archivePath)
	if err != nil {
		return -1, errors.ErrFileOpen
	}
	defer archive.Close()
	tempDir, err := os.MkdirTemp(os.TempDir(), "task-upload-archive")
	if err != nil {
		return -1, errors.ErrTempDirCreate
	}
	defer os.RemoveAll(tempDir)
	err = utils.DecompressArchive(archive, tempDir)
	if err != nil {
		ts.logger.Errorf("Error decompressing archive %s: %v", archivePath, err)
		return -1, errors.ErrDecompressArchive
	}
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return -1, fmt.Errorf("failed to read temp directory: %w", err)
	}
	// Sometime archive contain a single directory which was archived, use it
	if len(entries) == 1 {
		if entries[0].IsDir() {
			tempDir = tempDir + "/" + entries[0].Name()
		} else {
			return -1, errors.ErrInvalidArchive
		}
	}

	inputFiles, err := os.ReadDir(tempDir + "/input")
	if err != nil {
		return -1, errors.ErrNoInputDirectory
	}
	outputFiles, err := os.ReadDir(tempDir + "/output")
	if err != nil {
		return -1, errors.ErrNoOutputDirectory
	}
	if len(inputFiles) != len(outputFiles) {
		return -1, errors.ErrIOCountMismatch
	}
	for _, file := range inputFiles {
		if file.IsDir() {
			return -1, errors.ErrInputContainsDir
		}
		filenameList := strings.Split(file.Name(), ".")
		if len(filenameList) != 2 {
			return -1, fmt.Errorf(
				"input file name is not formatted correctly. Expected format: <filename>.<extension> but got %s",
				file.Name(),
			) // TODO: change this
		}
		filename := filenameList[0]
		ext := filenameList[1]
		if ext != "in" {
			return -1, errors.ErrInvalidInExtention
		}
		found := false
		for _, outputFile := range outputFiles {
			if file.IsDir() {
				return -1, errors.ErrOutputContainsDir
			}
			outputFilenameList := strings.Split(outputFile.Name(), ".")
			if len(outputFilenameList) != 2 {
				return -1, fmt.Errorf(
					"output file name is not formatted correctly. Expected format: <filename>.<extension> but got %s",
					outputFile.Name(),
				) // TODO: create a wrapper error to solve this.
			}
			outputFilename := outputFilenameList[0]
			if outputFilenameList[1] != "out" {
				return -1, errors.ErrInvalidOutExtention
			}
			if filename == outputFilename {
				found = true
				break
			}
		}
		if !found {
			return -1, fmt.Errorf("input file %s does not have a corresponding output file", filename)
		}
	}
	return len(inputFiles), nil
}

func (ts *taskService) CreateTestCase(tx *gorm.DB, taskID int64, archivePath string) error {
	_, err := ts.taskRepository.Get(tx, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrNotFound
		}
		return err
	}
	numFiles, err := ts.ParseTestCase(archivePath)
	if err != nil {
		return fmt.Errorf("failed to parse input and output files: %w", err)
	}
	// Remove existing input output files
	err = ts.testCaseRepository.DeleteAll(tx, taskID)
	if err != nil {
		return fmt.Errorf("failed to delete existing input output files: %w", err)
	}

	for i := 1; i <= numFiles; i++ {
		io := &models.TestCase{
			TaskID:      taskID,
			Order:       i,
			TimeLimit:   1, // Hardcode for now
			MemoryLimit: 1, // Hardcode for now
		}
		err = ts.testCaseRepository.Create(tx, io)
		if err != nil {
			return fmt.Errorf("failed to create input output: %w", err)
		}
	}
	return nil
}

func (ts *taskService) ProcessAndUpload(tx *gorm.DB, currentUser *schemas.User, taskID int64, archivePath string) error {
	// Check permissions using collaborator system - need edit permission
	err := ts.hasTaskPermission(tx, taskID, currentUser, types.PermissionEdit)
	if err != nil {
		return err
	}

	err = ts.filestorage.ValidateArchiveStructure(archivePath)
	if err != nil {
		return fmt.Errorf("failed to validate archive structure: %w", err)
	}

	uploadedTaskFiles, err := ts.filestorage.UploadTask(taskID, archivePath)
	if err != nil {
		return fmt.Errorf("failed to upload task archive: %w", err)
	}

	// Save to database records about uploaded files
	descriptionFile := &models.File{
		Filename:   uploadedTaskFiles.DescriptionFile.Filename,
		Path:       uploadedTaskFiles.DescriptionFile.Path,
		Bucket:     uploadedTaskFiles.DescriptionFile.Bucket,
		ServerType: uploadedTaskFiles.DescriptionFile.ServerType,
	}
	err = ts.fileRepository.Create(tx, descriptionFile)
	if err != nil {
		return fmt.Errorf("failed to save description file: %w", err)
	}
	currentTask, err := ts.taskRepository.Get(tx, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrNotFound
		}
		return fmt.Errorf("failed to get task: %w", err)
	}
	currentTask.DescriptionFileID = descriptionFile.ID
	err = ts.taskRepository.Edit(tx, taskID, currentTask)
	if err != nil {
		return fmt.Errorf("failed to update task with description file: %w", err)
	}

	for i := range uploadedTaskFiles.InputFiles {
		input := uploadedTaskFiles.InputFiles[i]
		output := uploadedTaskFiles.OutputFiles[i]

		inputFile := &models.File{
			Filename:   input.Filename,
			Path:       input.Path,
			Bucket:     input.Bucket,
			ServerType: input.ServerType,
		}
		outputFile := &models.File{
			Filename:   output.Filename,
			Path:       output.Path,
			Bucket:     output.Bucket,
			ServerType: output.ServerType,
		}
		err = ts.fileRepository.Create(tx, inputFile)
		if err != nil {
			return fmt.Errorf("failed to save input file: %w", err)
		}
		err = ts.fileRepository.Create(tx, outputFile)
		if err != nil {
			return fmt.Errorf("failed to save output file: %w", err)
		}
		err = ts.testCaseRepository.Create(tx, &models.TestCase{
			TaskID:       taskID,
			InputFileID:  inputFile.ID,
			OutputFileID: outputFile.ID,
			Order:        i + 1,
			TimeLimit:    1, // Hardcode for now
			MemoryLimit:  1, // Hardcode for now
		})
		if err != nil {
			return fmt.Errorf("failed to save input output: %w", err)
		}
	}

	return nil
}

func (ts *taskService) GetLimits(tx *gorm.DB, currentUser *schemas.User, taskID int64) ([]schemas.TestCase, error) {
	_, err := ts.taskRepository.Get(tx, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}

	testCase, err := ts.testCaseRepository.GetByTask(tx, taskID)
	if err != nil {
		return nil, err
	}

	result := make([]schemas.TestCase, len(testCase))
	for i, io := range testCase {
		result[i] = *TestCaseToSchema(&io)
	}
	return result, nil
}

func (ts *taskService) PutLimits(
	tx *gorm.DB,
	currentUser *schemas.User,
	taskID int64,
	newLimits schemas.PutTestCaseLimitsRequest,
) error {
	// Check permissions using collaborator system - need edit permission
	err := ts.hasTaskPermission(tx, taskID, currentUser, types.PermissionEdit)
	if err != nil {
		return err
	}

	for _, io := range newLimits.Limits {
		ioID, err := ts.testCaseRepository.GetTestCaseID(tx, taskID, io.Order)
		if err != nil {
			return err
		}

		current, err := ts.testCaseRepository.Get(tx, ioID)
		if err != nil {
			return err
		}

		current.MemoryLimit = io.MemoryLimit
		current.TimeLimit = io.TimeLimit

		err = ts.testCaseRepository.Put(tx, current)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ts *taskService) updateModel(currentModel *models.Task, updateInfo *schemas.EditTask) {
	if updateInfo.Title != nil {
		currentModel.Title = *updateInfo.Title
	}
	if updateInfo.IsVisible != nil {
		currentModel.IsVisible = *updateInfo.IsVisible
	}
}

func TaskToSchema(model *models.Task) *schemas.Task {
	return &schemas.Task{
		ID:        model.ID,
		Title:     model.Title,
		CreatedBy: model.CreatedBy,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		IsVisible: model.IsVisible,
	}
}

func TestCaseToSchema(model *models.TestCase) *schemas.TestCase {
	return &schemas.TestCase{
		ID:          model.ID,
		TaskID:      model.TaskID,
		Order:       model.Order,
		TimeLimit:   model.TimeLimit,
		MemoryLimit: model.MemoryLimit,
	}
}

func (ts *taskService) GetMyLiveTasks(
	tx *gorm.DB,
	currentUser *schemas.User,
	paginationParams schemas.PaginationParams,
) (*schemas.MyTasksResponse, error) {
	// Get live tasks in contests
	contestTasksMap, err := ts.taskRepository.GetLiveAssignedTasksGroupedByContest(tx, currentUser.ID, paginationParams.Limit, paginationParams.Offset)
	if err != nil {
		ts.logger.Errorf("Error getting live assigned tasks in contests: %v", err.Error())
		return nil, err
	}

	// Build response structure
	response := &schemas.MyTasksResponse{
		Contests: []schemas.ContestWithTasks{},
	}

	// Process contest tasks
	for contestID, tasks := range contestTasksMap {
		contest, err := ts.contestRepository.Get(tx, contestID)
		if err != nil {
			ts.logger.Errorf("Error getting contest: %v", err.Error())
			continue
		}

		contestWithTasks := schemas.ContestWithTasks{
			ContestID:   contest.ID,
			ContestName: contest.Name,
			StartAt:     contest.StartAt,
			EndAt:       contest.EndAt,
			Tasks:       []schemas.TaskWithAttempts{},
		}

		for _, task := range tasks {
			taskWithAttempts, err := ts.enrichTaskWithAttempts(tx, &task, currentUser.ID)
			if err != nil {
				ts.logger.Errorf("Error enriching task with attempts: %v", err.Error())
				continue
			}
			contestWithTasks.Tasks = append(contestWithTasks.Tasks, *taskWithAttempts)
		}

		response.Contests = append(response.Contests, contestWithTasks)
	}

	return response, nil
}

func (ts *taskService) enrichTaskWithAttempts(
	tx *gorm.DB,
	task *models.Task,
	userID int64,
) (*schemas.TaskWithAttempts, error) {
	// Get attempt count
	attemptCount, err := ts.submissionRepository.GetAttemptCountForTaskByUser(tx, task.ID, userID)
	if err != nil {
		return nil, err
	}

	// Get best score if there are attempts
	var bestScore float64
	if attemptCount > 0 {
		bestScore, err = ts.submissionRepository.GetBestScoreForTaskByUser(tx, task.ID, userID)
		ts.logger.Infof("Best score for task %d and user %d: %v", task.ID, userID, bestScore)
		if err != nil {
			return nil, err
		}
	}

	return &schemas.TaskWithAttempts{
		Task: schemas.Task{
			ID:        task.ID,
			Title:     task.Title,
			CreatedBy: task.CreatedBy,
			CreatedAt: task.CreatedAt,
			UpdatedAt: task.UpdatedAt,
		},
		AttemptsSummary: schemas.AttemptsSummary{
			BestScore:    bestScore,
			AttemptCount: attemptCount,
		},
	}, nil
}

// hasTaskPermission checks that task exists and if the user has the required permission for the task.
func (ts *taskService) hasTaskPermission(tx *gorm.DB, taskID int64, user *schemas.User, requiredPermission types.Permission) error {
	_, err := ts.taskRepository.Get(tx, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrNotFound
		}
		return err
	}
	return ts.accessControlService.CanUserAccess(tx, models.ResourceTypeTask, taskID, user, requiredPermission)
}

func TaskToInfoSchema(model *models.Task) *schemas.TaskInfo {
	return &schemas.TaskInfo{
		ID:    model.ID,
		Title: model.Title,
	}
}

func NewTaskService(
	filestorage filestorage.FileStorageService,
	fileRepository repository.File,
	taskRepository repository.TaskRepository,
	testCaseRepository repository.TestCaseRepository,
	userRepository repository.UserRepository,
	groupRepository repository.GroupRepository,
	submissionRepository repository.SubmissionRepository,
	contestRepository repository.ContestRepository,
	accessControlService AccessControlService,
) TaskService {
	log := utils.NewNamedLogger("task_service")
	return &taskService{
		filestorage:          filestorage,
		fileRepository:       fileRepository,
		taskRepository:       taskRepository,
		userRepository:       userRepository,
		groupRepository:      groupRepository,
		testCaseRepository:   testCaseRepository,
		submissionRepository: submissionRepository,
		contestRepository:    contestRepository,
		accessControlService: accessControlService,
		logger:               log,
	}
}
