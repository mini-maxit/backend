package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	myerrors "github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TaskService interface {
	// AssignToGroups assigns a task to multiple groups.
	AssignToGroups(tx *gorm.DB, currentUser schemas.User, taskID int64, groupIDs []int64) error
	// AssignToUsers assigns a task to multiple users.
	AssignToUsers(tx *gorm.DB, currentUser schemas.User, taskID int64, userIDs []int64) error
	// Create creates a new task.
	Create(tx *gorm.DB, currentUser schemas.User, task *schemas.Task) (int64, error)
	// CreateInputOutput creates input and output files for a task.
	CreateInputOutput(tx *gorm.DB, taskID int64, archivePath string) error
	// Delete deletes a task.
	Delete(tx *gorm.DB, currentUser schemas.User, taskID int64) error
	// Edit edits an existing task.
	Edit(tx *gorm.DB, currentUser schemas.User, taskID int64, updateInfo *schemas.EditTask) error
	// GetAll retrieves all tasks based on query parameters.
	GetAll(tx *gorm.DB, currentUser schemas.User, queryParams map[string]any) ([]schemas.Task, error)
	// GetAllAssigned retrieves all tasks assigned to the current user.
	GetAllAssigned(tx *gorm.DB, currentUser schemas.User, queryParams map[string]any) ([]schemas.Task, error)
	// GetAllCreated retrieves all tasks created by the current user.
	GetAllCreated(tx *gorm.DB, currentUser schemas.User, queryParams map[string]any) ([]schemas.Task, error)
	// GetAllForGroup retrieves all tasks for a specific group.
	GetAllForGroup(
		tx *gorm.DB,
		currentUser schemas.User,
		groupID int64,
		queryParams map[string]any,
	) ([]schemas.Task, error)
	// Get retrieves a detailed view of a specific task.
	Get(tx *gorm.DB, currentUser schemas.User, taskID int64) (*schemas.TaskDetailed, error)
	// GetByTitle retrieves a task by its title.
	GetByTitle(tx *gorm.DB, title string) (*schemas.Task, error)
	// ParseInputOutput parses the input and output files from an archive.
	ParseInputOutput(archivePath string) (int, error)
	// ProcessAndUpload processes and uploads input and output files for a task.
	ProcessAndUpload(tx *gorm.DB, currentUser schemas.User, taskID int64, archivePath string) error
	// UnassignFromGroups unassigns a task from multiple groups.
	UnassignFromGroups(tx *gorm.DB, currentUser schemas.User, taskID int64, groupIDs []int64) error
	// UnassignFromUsers unassigns a task from multiple users.
	UnassignFromUsers(tx *gorm.DB, currentUser schemas.User, taskID int64, userID []int64) error
}

const defaultTaskSort = "created_at:desc"

type taskService struct {
	fileStorageURL        string
	groupRepository       repository.GroupRepository
	inputOutputRepository repository.InputOutputRepository
	taskRepository        repository.TaskRepository
	userRepository        repository.UserRepository

	logger *zap.SugaredLogger
}

func (ts *taskService) Create(tx *gorm.DB, currentUser schemas.User, task *schemas.Task) (int64, error) {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return 0, err
	}
	// Create a new task
	_, err = ts.GetByTitle(tx, task.Title)
	if err != nil && !errors.Is(err, myerrors.ErrTaskNotFound) {
		ts.logger.Errorf("Error getting task by title: %v", err.Error())
		return 0, err
	} else if err == nil {
		return 0, myerrors.ErrTaskExists
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
	}
	taskID, err := ts.taskRepository.Create(tx, &model)
	if err != nil {
		ts.logger.Errorf("Error creating task: %v", err.Error())
		return 0, err
	}

	return taskID, nil
}

func (ts *taskService) GetByTitle(tx *gorm.DB, title string) (*schemas.Task, error) {
	task, err := ts.taskRepository.GetByTitle(tx, title)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrTaskNotFound
		}
		ts.logger.Errorf("Error getting task by title: %v", err.Error())
		return nil, err
	}

	return TaskToSchema(task), nil
}

func (ts *taskService) GetAll(tx *gorm.DB, _ schemas.User, queryParams map[string]any) ([]schemas.Task, error) {
	// err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin})
	// if err != nil {
	// 	ts.logger.Errorf("Error validating user role: %v", err.Error())
	// 	return nil, err
	// }
	limit := queryParams["limit"].(int)
	offset := queryParams["offset"].(int)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = defaultTaskSort
	}

	// Get all tasks
	tasks, err := ts.taskRepository.GetAll(tx, limit, offset, sort)
	if err != nil {
		ts.logger.Errorf("Error getting all tasks: %v", err.Error())
		return nil, err
	}

	// Convert the models to schemas
	var result []schemas.Task
	for _, task := range tasks {
		result = append(result, *TaskToSchema(&task))
	}

	return result, nil
}

func (ts *taskService) GetAllForGroup(
	tx *gorm.DB,
	currentUser schemas.User,
	groupID int64,
	queryParams map[string]any,
) ([]schemas.Task, error) {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return nil, myerrors.ErrNotAuthorized
	}
	limit := queryParams["limit"].(int)
	offset := queryParams["offset"].(int)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = defaultTaskSort
	}

	// Get all tasks
	tasks, err := ts.taskRepository.GetAllForGroup(tx, groupID, limit, offset, sort)
	if err != nil {
		ts.logger.Error("Error getting all tasks for group")
		return nil, err
	}

	// Convert the models to schemas
	var result []schemas.Task
	for _, task := range tasks {
		result = append(result, *TaskToSchema(&task))
	}

	return result, nil
}

func (ts *taskService) Get(tx *gorm.DB, _ schemas.User, taskID int64) (*schemas.TaskDetailed, error) {
	// Get the task
	task, err := ts.taskRepository.Get(tx, taskID)
	if err != nil {
		ts.logger.Errorf("Error getting task: %v", err.Error())
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

	// Convert the model to schema
	groups := make([]int64, len(task.Groups))
	for i, group := range task.Groups {
		groups[i] = group.ID
	}
	result := &schemas.TaskDetailed{
		ID:             task.ID,
		Title:          task.Title,
		DescriptionURL: fmt.Sprintf("%s/getTaskDescription?taskID=%d", ts.fileStorageURL, task.ID),
		CreatedBy:      task.CreatedBy,
		CreatedByName:  task.Author.Name,
		CreatedAt:      task.CreatedAt,
		GroupIDs:       groups,
	}

	return result, nil
}

func (ts *taskService) GetAllAssigned(
	tx *gorm.DB,
	currentUser schemas.User,
	queryParams map[string]any,
) ([]schemas.Task, error) {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleStudent})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return nil, err
	}
	limit := queryParams["limit"].(int)
	offset := queryParams["offset"].(int)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = defaultTaskSort
	}

	tasks, err := ts.taskRepository.GetAllAssigned(tx, currentUser.ID, limit, offset, sort)
	if err != nil {
		ts.logger.Errorf("Error getting all assigned tasks: %v", err.Error())
		return nil, err
	}

	var result []schemas.Task

	for task := range tasks {
		result = append(result, *TaskToSchema(&tasks[task]))
	}

	return result, nil
}

func (ts *taskService) GetAllCreated(
	tx *gorm.DB,
	currentUser schemas.User,
	queryParams map[string]any,
) ([]schemas.Task, error) {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return nil, err
	}

	limit := queryParams["limit"].(int)
	offset := queryParams["offset"].(int)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = "created_at desc"
	}
	tasks, err := ts.taskRepository.GetAllCreated(tx, currentUser.ID, limit, offset, sort)
	if err != nil {
		ts.logger.Errorf("Error getting all created tasks: %v", err.Error())
		return nil, err
	}

	var result []schemas.Task

	for task := range tasks {
		result = append(result, *TaskToSchema(&tasks[task]))
	}

	return result, nil
}

func (ts *taskService) AssignToUsers(tx *gorm.DB, currentUser schemas.User, taskID int64, userIDs []int64) error {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return err
	}

	task, err := ts.taskRepository.Get(tx, taskID)
	if err != nil {
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return err
	}
	if currentUser.Role == types.UserRoleTeacher && task.CreatedBy != currentUser.ID {
		return myerrors.ErrNotAuthorized
	}

	for _, userID := range userIDs {
		_, err := ts.userRepository.Get(tx, userID)
		if err != nil {
			ts.logger.Errorf("Error getting user: %v", err.Error())
			return err
		}

		isAssigned, err := ts.taskRepository.IsAssignedToUser(tx, taskID, userID)
		if err != nil {
			ts.logger.Errorf("Error checking if task is assigned to user: %v", err.Error())
			return err
		}
		if isAssigned {
			ts.logger.Errorf("Error task already assigned to user %d", userID)
			continue
		}

		err = ts.taskRepository.AssignToUser(tx, taskID, userID)
		if err != nil {
			ts.logger.Errorf("Error assigning task to user: %v", err.Error())
			return err
		}
	}

	return nil
}

func (ts *taskService) AssignToGroups(tx *gorm.DB, currentUser schemas.User, taskID int64, groupIDs []int64) error {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return err
	}

	task, err := ts.taskRepository.Get(tx, taskID)
	if err != nil {
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return err
	}
	if currentUser.Role == types.UserRoleTeacher && task.CreatedBy != currentUser.ID {
		return myerrors.ErrNotAuthorized
	}

	for _, groupID := range groupIDs {
		_, err := ts.groupRepository.Get(tx, groupID)
		if err != nil {
			ts.logger.Errorf("Error getting group: %v", err.Error())
			return err
		}

		isAssigned, err := ts.taskRepository.IsAssignedToGroup(tx, taskID, groupID)
		if err != nil {
			ts.logger.Errorf("Error checking if task is assigned to group: %v", err.Error())
			return err
		}
		if isAssigned {
			ts.logger.Errorf("Error task already assigned to group %d", groupID)
			continue
		}

		err = ts.taskRepository.AssignToGroup(tx, taskID, groupID)
		if err != nil {
			ts.logger.Errorf("Error assigning task to group: %v", err.Error())
			return err
		}
	}

	return nil
}

func (ts *taskService) UnassignFromUsers(tx *gorm.DB, currentUser schemas.User, taskID int64, userID []int64) error {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return err
	}

	task, err := ts.taskRepository.Get(tx, taskID)
	if err != nil {
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return err
	}
	if currentUser.Role == types.UserRoleTeacher && task.CreatedBy != currentUser.ID {
		return myerrors.ErrNotAuthorized
	}

	for _, userID := range userID {
		isAssigned, err := ts.taskRepository.IsAssignedToUser(tx, taskID, userID)
		if err != nil {
			ts.logger.Errorf("Error checking if task is assigned to user: %v", err.Error())
			return err
		}
		if !isAssigned {
			return myerrors.ErrTaskNotAssignedToUser
		}

		err = ts.taskRepository.UnassignFromUser(tx, taskID, userID)
		if err != nil {
			ts.logger.Errorf("Error unassigning task from user: %v", err.Error())
			return err
		}
	}

	return nil
}

func (ts *taskService) UnassignFromGroups(tx *gorm.DB, currentUser schemas.User, taskID int64, groupIDs []int64) error {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return err
	}

	task, err := ts.taskRepository.Get(tx, taskID)
	if err != nil {
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return err
	}
	if currentUser.Role == types.UserRoleTeacher && task.CreatedBy != currentUser.ID {
		return myerrors.ErrNotAuthorized
	}

	for _, groupID := range groupIDs {
		isAssigned, err := ts.taskRepository.IsAssignedToGroup(tx, taskID, groupID)
		if err != nil {
			ts.logger.Errorf("Error checking if task is assigned to group: %v", err.Error())
			return err
		}
		if !isAssigned {
			return myerrors.ErrTaskNotAssignedToGroup
		}

		err = ts.taskRepository.UnassignFromGroup(tx, taskID, groupID)
		if err != nil {
			ts.logger.Errorf("Error unassigning task from group: %v", err.Error())
			return err
		}
	}

	return nil
}

func (ts *taskService) Delete(tx *gorm.DB, currentUser schemas.User, taskID int64) error {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return err
	}

	task, err := ts.taskRepository.Get(tx, taskID)
	if err != nil {
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return err
	}
	if currentUser.Role == types.UserRoleTeacher && task.CreatedBy != currentUser.ID {
		return myerrors.ErrNotAuthorized
	}

	err = ts.taskRepository.Delete(tx, taskID)
	if err != nil {
		ts.logger.Errorf("Error deleting task: %v", err.Error())
		return err
	}

	return nil
}

func (ts *taskService) Edit(tx *gorm.DB, currentUser schemas.User, taskID int64, updateInfo *schemas.EditTask) error {
	currentTask, err := ts.taskRepository.Get(tx, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrTaskNotFound
		}
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return err
	}
	if currentUser.Role == types.UserRoleTeacher &&
		(currentTask.CreatedBy != currentUser.ID || currentUser.Role == types.UserRoleStudent) {
		return myerrors.ErrNotAuthorized
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

func (ts *taskService) ParseInputOutput(archivePath string) (int, error) {
	// Unzip the archive
	// Extract the input and output files
	// Validate correct number of input and output files and naming
	// Count the number of input and output files
	archive, err := os.Open(archivePath)
	if err != nil {
		return -1, myerrors.ErrFileOpen
	}
	defer archive.Close()
	tempDir, err := os.MkdirTemp(os.TempDir(), "task-upload-archive")
	if err != nil {
		return -1, myerrors.ErrTempDirCreate
	}
	defer os.RemoveAll(tempDir)
	err = utils.DecompressArchive(archive, tempDir)
	if err != nil {
		ts.logger.Errorf("Error decompressing archive %s: %v", archivePath, err)
		return -1, myerrors.ErrDecompressArchive
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
			return -1, myerrors.ErrInvalidArchive
		}
	}

	inputFiles, err := os.ReadDir(tempDir + "/input")
	if err != nil {
		return -1, myerrors.ErrNoInputDirectory
	}
	outputFiles, err := os.ReadDir(tempDir + "/output")
	if err != nil {
		return -1, myerrors.ErrNoOutputDirectory
	}
	if len(inputFiles) != len(outputFiles) {
		return -1, myerrors.ErrIOCountMismatch
	}
	for _, file := range inputFiles {
		if file.IsDir() {
			return -1, myerrors.ErrInputContainsDir
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
			return -1, myerrors.ErrInvalidInExtention
		}
		found := false
		for _, outputFile := range outputFiles {
			if file.IsDir() {
				return -1, myerrors.ErrOutputContainsDir
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
				return -1, myerrors.ErrInvalidOutExtention
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

func (ts *taskService) CreateInputOutput(tx *gorm.DB, taskID int64, archivePath string) error {
	_, err := ts.taskRepository.Get(tx, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrNotFound
		}
		return err
	}
	numFiles, err := ts.ParseInputOutput(archivePath)
	if err != nil {
		return fmt.Errorf("failed to parse input and output files: %w", err)
	}
	// Remove existing input output files
	err = ts.inputOutputRepository.DeleteAll(tx, taskID)
	if err != nil {
		return fmt.Errorf("failed to delete existing input output files: %w", err)
	}

	for i := 1; i <= numFiles; i++ {
		io := &models.InputOutput{
			TaskID:      taskID,
			Order:       i,
			TimeLimit:   1, // Hardcode for now
			MemoryLimit: 1, // Hardcode for now
		}
		err = ts.inputOutputRepository.Create(tx, io)
		if err != nil {
			return fmt.Errorf("failed to create input output: %w", err)
		}
	}
	return nil
}

func (ts *taskService) ProcessAndUpload(tx *gorm.DB, currentUser schemas.User, taskID int64, archivePath string) error {
	if currentUser.Role == types.UserRoleStudent {
		return myerrors.ErrNotAllowed
	}

	err := ts.CreateInputOutput(tx, taskID, archivePath)
	if err != nil {
		return fmt.Errorf("failed to create input output: %w", err)
	}

	// Create a multipart writer for the HTTP request to FileStorage service
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	err = writer.WriteField("taskID", strconv.FormatInt(taskID, 10))
	if err != nil {
		return myerrors.ErrWriteTaskID
	}
	err = writer.WriteField("overwrite", strconv.FormatBool(true))
	if err != nil {
		return myerrors.ErrWriteOverwrite
	}

	// Create a form file field and copy the uploaded file to it
	part, err := writer.CreateFormFile("archive", "Task.zip")
	if err != nil {
		return myerrors.ErrCreateFormFile
	}
	file, err := os.Open(archivePath)
	if err != nil {
		return myerrors.ErrFileOpen
	}
	defer file.Close()
	if _, err := io.Copy(part, file); err != nil {
		return myerrors.ErrCopyFile
	}
	writer.Close()

	// Send the request to FileStorage service
	client := &http.Client{}
	resp, err := client.Post(ts.fileStorageURL+"/createTask", writer.FormDataContentType(), body)
	if err != nil {
		return myerrors.ErrSendRequest
	}
	defer resp.Body.Close()

	// Handle response from FileStorage
	buffer := make([]byte, resp.ContentLength)
	bytesRead, err := resp.Body.Read(buffer)
	if err != nil && bytesRead == 0 {
		return myerrors.ErrReadResponse
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s", myerrors.ErrResponseFromFileStorage, string(buffer))
	}

	return nil
}

func (ts *taskService) updateModel(currentModel *models.Task, updateInfo *schemas.EditTask) {
	if updateInfo.Title != nil {
		currentModel.Title = *updateInfo.Title
	}
}

func TaskToSchema(model *models.Task) *schemas.Task {
	return &schemas.Task{
		ID:        model.ID,
		Title:     model.Title,
		CreatedBy: model.CreatedBy,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

func NewTaskService(
	fileStorageURL string,
	taskRepository repository.TaskRepository,
	inputOutputRepository repository.InputOutputRepository,
	userRepository repository.UserRepository,
	groupRepository repository.GroupRepository,
) TaskService {
	log := utils.NewNamedLogger("task_service")
	return &taskService{
		fileStorageURL:        fileStorageURL,
		taskRepository:        taskRepository,
		userRepository:        userRepository,
		groupRepository:       groupRepository,
		inputOutputRepository: inputOutputRepository,
		logger:                log,
	}
}
