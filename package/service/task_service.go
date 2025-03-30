package service

import (
	"bytes"
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
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TaskService interface {
	// AssignToGroups assigns a task to multiple groups.
	AssignToGroups(tx *gorm.DB, current_user schemas.User, taskId int64, groupIds []int64) error
	// AssignToUsers assigns a task to multiple users.
	AssignToUsers(tx *gorm.DB, current_user schemas.User, taskId int64, userIds []int64) error
	// Create creates a new task.
	Create(tx *gorm.DB, current_user schemas.User, task *schemas.Task) (int64, error)
	// CreateInputOutput creates input and output files for a task.
	CreateInputOutput(tx *gorm.DB, taskId int64, archivePath string) error
	// Delete deletes a task.
	Delete(tx *gorm.DB, current_user schemas.User, taskId int64) error
	// Edit edits an existing task.
	Edit(tx *gorm.DB, currentUser schemas.User, taskId int64, updateInfo *schemas.EditTask) error
	// GetAll retrieves all tasks based on query parameters.
	GetAll(tx *gorm.DB, current_user schemas.User, queryParams map[string]any) ([]schemas.Task, error)
	// GetAllAssigned retrieves all tasks assigned to the current user.
	GetAllAssigned(tx *gorm.DB, current_user schemas.User, queryParams map[string]any) ([]schemas.Task, error)
	// GetAllCreated retrieves all tasks created by the current user.
	GetAllCreated(tx *gorm.DB, current_user schemas.User, queryParams map[string]any) ([]schemas.Task, error)
	// GetAllForGroup retrieves all tasks for a specific group.
	GetAllForGroup(tx *gorm.DB, current_user schemas.User, groupId int64, queryParams map[string]any) ([]schemas.Task, error)
	// Get retrieves a detailed view of a specific task.
	Get(tx *gorm.DB, current_user schemas.User, taskId int64) (*schemas.TaskDetailed, error)
	// GetByTitle retrieves a task by its title.
	GetByTitle(tx *gorm.DB, title string) (*schemas.Task, error)
	// ParseInputOutput parses the input and output files from an archive.
	ParseInputOutput(archivePath string) (int, error)
	// ProcessAndUpload processes and uploads input and output files for a task.
	ProcessAndUpload(tx *gorm.DB, current_user schemas.User, taskId int64, archivePath string) error
	// UnassignFromGroups unassigns a task from multiple groups.
	UnassignFromGroups(tx *gorm.DB, current_user schemas.User, taskId int64, groupIds []int64) error
	// UnassignFromUsers unassigns a task from multiple users.
	UnassignFromUsers(tx *gorm.DB, current_user schemas.User, taskId int64, userId []int64) error
}

type taskService struct {
	fileStorageUrl        string
	groupRepository       repository.GroupRepository
	inputOutputRepository repository.InputOutputRepository
	taskRepository        repository.TaskRepository
	userRepository        repository.UserRepository

	logger *zap.SugaredLogger
}

func (ts *taskService) Create(tx *gorm.DB, current_user schemas.User, task *schemas.Task) (int64, error) {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return 0, err
	}
	// Create a new task
	_, err = ts.GetByTitle(tx, task.Title)
	if err != nil && err != errors.ErrTaskNotFound {
		ts.logger.Errorf("Error getting task by title: %v", err.Error())
		return 0, err
	} else if err == nil {
		return 0, errors.ErrTaskExists
	}

	author, err := ts.userRepository.Get(tx, current_user.Id)
	if err != nil {
		ts.logger.Errorf("Error getting user: %v", err.Error())
		return 0, err
	}

	model := models.Task{
		Title:     task.Title,
		CreatedBy: task.CreatedBy,
		Author:    *author,
	}
	taskId, err := ts.taskRepository.Create(tx, &model)
	if err != nil {
		ts.logger.Errorf("Error creating task: %v", err.Error())
		return 0, err
	}

	return taskId, nil
}

func (ts *taskService) GetByTitle(tx *gorm.DB, title string) (*schemas.Task, error) {
	task, err := ts.taskRepository.GetByTitle(tx, title)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrTaskNotFound
		}
		ts.logger.Errorf("Error getting task by title: %v", err.Error())
		return nil, err
	}

	return TaskToSchema(task), nil
}

func (ts *taskService) GetAll(tx *gorm.DB, current_user schemas.User, queryParams map[string]any) ([]schemas.Task, error) {
	// err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleAdmin})
	// if err != nil {
	// 	ts.logger.Errorf("Error validating user role: %v", err.Error())
	// 	return nil, err
	// }
	limit := queryParams["limit"].(uint64)
	offset := queryParams["offset"].(uint64)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = "created_at:desc"
	}

	// Get all tasks
	tasks, err := ts.taskRepository.GetAll(tx, int(limit), int(offset), sort)
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

func (ts *taskService) GetAllForGroup(tx *gorm.DB, current_user schemas.User, groupId int64, queryParams map[string]any) ([]schemas.Task, error) {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return nil, errors.ErrNotAuthorized
	}
	limit := queryParams["limit"].(uint64)
	offset := queryParams["offset"].(uint64)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = "created_at:desc"
	}

	// Get all tasks
	tasks, err := ts.taskRepository.GetAllForGroup(tx, groupId, int(limit), int(offset), sort)
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

func (ts *taskService) Get(tx *gorm.DB, current_user schemas.User, taskId int64) (*schemas.TaskDetailed, error) {
	// Get the task
	task, err := ts.taskRepository.Get(tx, taskId)
	if err != nil {
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return nil, err
	}

	// switch types.UserRole(current_user.Role) {
	// case types.UserRoleStudent:
	// 	// Check if the task is assigned to the user
	// 	isAssigned, err := ts.taskRepository.IsTaskAssignedToUser(tx, taskId, current_user.Id)
	// 	if err != nil {
	// 		ts.logger.Errorf("Error checking if task is assigned to user: %v", err.Error())
	// 		return nil, err
	// 	}
	// 	if !isAssigned {
	// 		return nil, errors.ErrNotAuthorized
	// 	}
	// case types.UserRoleTeacher:
	// 	// Check if the task is created by the user
	// 	if task.CreatedBy != current_user.Id {
	// 		return nil, errors.ErrNotAuthorized
	// 	}
	// }

	// Convert the model to schema
	groups := make([]int64, len(task.Groups))
	for i, group := range task.Groups {
		groups[i] = group.Id
	}
	result := &schemas.TaskDetailed{
		Id:             task.Id,
		Title:          task.Title,
		DescriptionURL: fmt.Sprintf("%s/getTaskDescription?taskID=%d", ts.fileStorageUrl, task.Id),
		CreatedBy:      task.CreatedBy,
		CreatedByName:  task.Author.Name,
		CreatedAt:      task.CreatedAt,
		GroupIds:       groups,
	}

	return result, nil
}

func (ts *taskService) GetAllAssigned(tx *gorm.DB, current_user schemas.User, queryParams map[string]any) ([]schemas.Task, error) {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleStudent})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return nil, err
	}
	limit := queryParams["limit"].(uint64)
	offset := queryParams["offset"].(uint64)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = "created_at:desc"
	}

	tasks, err := ts.taskRepository.GetAllAssigned(tx, current_user.Id, int(limit), int(offset), sort)
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

func (ts *taskService) GetAllCreated(tx *gorm.DB, current_user schemas.User, queryParams map[string]any) ([]schemas.Task, error) {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return nil, err
	}

	limit := queryParams["limit"].(uint64)
	offset := queryParams["offset"].(uint64)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = "created_at desc"
	}
	tasks, err := ts.taskRepository.GetAllCreated(tx, current_user.Id, int(limit), int(offset), sort)
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

func (ts *taskService) AssignToUsers(tx *gorm.DB, current_user schemas.User, taskId int64, userIds []int64) error {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return err
	}

	task, err := ts.taskRepository.Get(tx, taskId)
	if err != nil {
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return err
	}
	if current_user.Role == types.UserRoleTeacher && task.CreatedBy != current_user.Id {
		return errors.ErrNotAuthorized
	}

	for _, userId := range userIds {
		_, err := ts.userRepository.Get(tx, userId)
		if err != nil {
			ts.logger.Errorf("Error getting user: %v", err.Error())
			return err
		}

		isAssigned, err := ts.taskRepository.IsAssignedToUser(tx, taskId, userId)
		if err != nil {
			ts.logger.Errorf("Error checking if task is assigned to user: %v", err.Error())
			return err
		}
		if isAssigned {
			ts.logger.Errorf("Error task already assigned to user %d", userId)
			continue
		}

		err = ts.taskRepository.AssignToUser(tx, taskId, userId)
		if err != nil {
			ts.logger.Errorf("Error assigning task to user: %v", err.Error())
			return err
		}
	}

	return nil
}

func (ts *taskService) AssignToGroups(tx *gorm.DB, current_user schemas.User, taskId int64, groupIds []int64) error {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return err
	}

	task, err := ts.taskRepository.Get(tx, taskId)
	if err != nil {
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return err
	}
	if current_user.Role == types.UserRoleTeacher && task.CreatedBy != current_user.Id {
		return errors.ErrNotAuthorized
	}

	for _, groupId := range groupIds {
		_, err := ts.groupRepository.Get(tx, groupId)
		if err != nil {
			ts.logger.Errorf("Error getting group: %v", err.Error())
			return err
		}

		isAssigned, err := ts.taskRepository.IsAssignedToGroup(tx, taskId, groupId)
		if err != nil {
			ts.logger.Errorf("Error checking if task is assigned to group: %v", err.Error())
			return err
		}
		if isAssigned {
			ts.logger.Errorf("Error task already assigned to group %d", groupId)
			continue
		}

		err = ts.taskRepository.AssignToGroup(tx, taskId, groupId)
		if err != nil {
			ts.logger.Errorf("Error assigning task to group: %v", err.Error())
			return err
		}
	}

	return nil
}

func (ts *taskService) UnassignFromUsers(tx *gorm.DB, current_user schemas.User, taskId int64, userId []int64) error {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return err
	}

	task, err := ts.taskRepository.Get(tx, taskId)
	if err != nil {
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return err
	}
	if current_user.Role == types.UserRoleTeacher && task.CreatedBy != current_user.Id {
		return errors.ErrNotAuthorized
	}

	for _, userId := range userId {

		isAssigned, err := ts.taskRepository.IsAssignedToUser(tx, taskId, userId)
		if err != nil {
			ts.logger.Errorf("Error checking if task is assigned to user: %v", err.Error())
			return err
		}
		if !isAssigned {
			return errors.ErrTaskNotAssignedToUser
		}

		err = ts.taskRepository.UnassignFromUser(tx, taskId, userId)
		if err != nil {
			ts.logger.Errorf("Error unassigning task from user: %v", err.Error())
			return err
		}
	}

	return nil
}

func (ts *taskService) UnassignFromGroups(tx *gorm.DB, current_user schemas.User, taskId int64, groupIds []int64) error {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return err
	}

	task, err := ts.taskRepository.Get(tx, taskId)
	if err != nil {
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return err
	}
	if current_user.Role == types.UserRoleTeacher && task.CreatedBy != current_user.Id {
		return errors.ErrNotAuthorized
	}

	for _, groupId := range groupIds {

		isAssigned, err := ts.taskRepository.IsAssignedToGroup(tx, taskId, groupId)
		if err != nil {
			ts.logger.Errorf("Error checking if task is assigned to group: %v", err.Error())
			return err
		}
		if !isAssigned {
			return errors.ErrTaskNotAssignedToGroup
		}

		err = ts.taskRepository.UnassignFromGroup(tx, taskId, groupId)
		if err != nil {
			ts.logger.Errorf("Error unassigning task from group: %v", err.Error())
			return err
		}
	}

	return nil
}

func (ts *taskService) Delete(tx *gorm.DB, current_user schemas.User, taskId int64) error {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return err
	}

	task, err := ts.taskRepository.Get(tx, taskId)
	if err != nil {
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return err
	}
	if current_user.Role == types.UserRoleTeacher && task.CreatedBy != current_user.Id {
		return errors.ErrNotAuthorized
	}

	err = ts.taskRepository.Delete(tx, taskId)
	if err != nil {
		ts.logger.Errorf("Error deleting task: %v", err.Error())
		return err
	}

	return nil
}

func (ts *taskService) Edit(tx *gorm.DB, currentUser schemas.User, taskId int64, updateInfo *schemas.EditTask) error {
	currentTask, err := ts.taskRepository.Get(tx, taskId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrTaskNotFound
		}
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return err
	}
	if currentUser.Role == types.UserRoleTeacher && currentTask.CreatedBy != currentUser.Id || currentUser.Role == types.UserRoleStudent {
		return errors.ErrNotAuthorized
	}

	ts.updateModel(currentTask, updateInfo)

	// Update the task
	err = ts.taskRepository.Edit(tx, taskId, currentTask)
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
		return -1, errors.ErrFileOpen
	}
	defer archive.Close()
	temp_dir, err := os.MkdirTemp(os.TempDir(), "task-upload-archive")
	if err != nil {
		return -1, errors.ErrTempDirCreate
	}
	defer os.RemoveAll(temp_dir)
	err = utils.DecompressArchive(archive, temp_dir)
	if err != nil {
		ts.logger.Errorf("Error decompressing archive %s: %v", archivePath, err)
		return -1, errors.ErrDecompressArchive
	}
	entries, err := os.ReadDir(temp_dir)
	if err != nil {
		return -1, fmt.Errorf("failed to read temp directory: %v", err)
	}
	// Sometime archive contain a single directory which was archived, use it
	if len(entries) == 1 {
		if entries[0].IsDir() {
			temp_dir = temp_dir + "/" + entries[0].Name()
		} else {
			return -1, fmt.Errorf("archive contains a single file, expected a single directory or [input/ output/ description.pdf]")
		}
	}

	inputFiles, err := os.ReadDir(temp_dir + "/input")
	if err != nil {
		return -1, errors.ErrNoInputDirectory
	}
	outputFiles, err := os.ReadDir(temp_dir + "/output")
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
		filename_list := strings.Split(file.Name(), ".")
		if len(filename_list) != 2 {
			return -1, fmt.Errorf("input file name is not formatted correctly. Expected format: <filename>.<extension> but got %s", file.Name())
		}
		filename := filename_list[0]
		ext := filename_list[1]
		if ext != "in" {
			return -1, errors.ErrInvalidInExtention
		}
		found := false
		for _, output_file := range outputFiles {
			if file.IsDir() {
				return -1, errors.ErrOutputContainsDir
			}
			output_filename_list := strings.Split(output_file.Name(), ".")
			if len(output_filename_list) != 2 {
				return -1, fmt.Errorf("output file name is not formatted correctly. Expected format: <filename>.<extension> but got %s", output_file.Name())
			}
			output_filename := output_filename_list[0]
			if output_filename_list[1] != "out" {
				return -1, errors.ErrInvalidOutExtention
			}
			if filename == output_filename {
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

func (ts *taskService) CreateInputOutput(tx *gorm.DB, taskId int64, archivePath string) error {
	_, err := ts.taskRepository.Get(tx, taskId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound
		}
		return err
	}
	num_files, err := ts.ParseInputOutput(archivePath)
	if err != nil {
		return fmt.Errorf("failed to parse input and output files: %v", err)
	}
	// Remove existing input output files
	err = ts.inputOutputRepository.DeleteAll(tx, taskId)
	if err != nil {
		return fmt.Errorf("failed to delete existing input output files: %v", err)
	}

	for i := 1; i <= num_files; i++ {
		io := &models.InputOutput{
			TaskId:      taskId,
			Order:       i,
			TimeLimit:   1, // Hardcode for now
			MemoryLimit: 1, // Hardcode for now
		}
		err = ts.inputOutputRepository.Create(tx, io)
		if err != nil {
			return fmt.Errorf("failed to create input output: %v", err)
		}
	}
	return nil
}

func (ts *taskService) ProcessAndUpload(tx *gorm.DB, current_user schemas.User, taskId int64, archivePath string) error {
	if current_user.Role == types.UserRoleStudent {
		return errors.ErrNotAllowed
	}

	err := ts.CreateInputOutput(tx, taskId, archivePath)
	if err != nil {
		return fmt.Errorf("failed to create input output: %v", err)
	}

	// Create a multipart writer for the HTTP request to FileStorage service
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	err = writer.WriteField("taskID", fmt.Sprintf("%d", taskId))
	if err != nil {
		return errors.ErrWriteTaskID
	}
	err = writer.WriteField("overwrite", strconv.FormatBool(true))
	if err != nil {
		return errors.ErrWriteOverwrite
	}

	// Create a form file field and copy the uploaded file to it
	part, err := writer.CreateFormFile("archive", "Task.zip")
	if err != nil {
		return errors.ErrCreateFormFile
	}
	file, err := os.Open(archivePath)
	if err != nil {
		return errors.ErrFileOpen
	}
	defer file.Close()
	if _, err := io.Copy(part, file); err != nil {
		return errors.ErrCopyFile
	}
	writer.Close()

	// Send the request to FileStorage service
	client := &http.Client{}
	resp, err := client.Post(ts.fileStorageUrl+"/createTask", writer.FormDataContentType(), body)
	if err != nil {
		return errors.ErrSendRequest
	}
	defer resp.Body.Close()

	// Handle response from FileStorage
	buffer := make([]byte, resp.ContentLength)
	bytesRead, err := resp.Body.Read(buffer)
	if err != nil && bytesRead == 0 {
		return errors.ErrReadResponse
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s", errors.ErrResponseFromFileStorage, string(buffer))
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
		Id:        model.Id,
		Title:     model.Title,
		CreatedBy: model.CreatedBy,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

func NewTaskService(fileStorageUrl string, taskRepository repository.TaskRepository, inputOutputRepository repository.InputOutputRepository, userRepository repository.UserRepository, groupRepository repository.GroupRepository) TaskService {
	log := utils.NewNamedLogger("task_service")
	return &taskService{
		fileStorageUrl:        fileStorageUrl,
		taskRepository:        taskRepository,
		userRepository:        userRepository,
		groupRepository:       groupRepository,
		inputOutputRepository: inputOutputRepository,
		logger:                log,
	}
}
