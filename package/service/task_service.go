package service

import (
	"fmt"
	"os"
	"strings"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var ErrDatabaseConnection = fmt.Errorf("failed to connect to the database")
var ErrTaskExists = fmt.Errorf("task with this title already exists")
var ErrTaskNotFound = fmt.Errorf("task not found")

type TaskService interface {
	// Create creates a new empty task and returns the task ID
	Create(tx *gorm.DB, task *schemas.Task) (int64, error)
	CreateInputOutput(tx *gorm.DB, taskId int64, archivePath string) error
	GetAll(tx *gorm.DB, queryParams map[string]interface{}) ([]schemas.Task, error)
	GetAllForUser(tx *gorm.DB, userId int64, queryParams map[string]interface{}) ([]schemas.Task, error)
	GetAllForGroup(tx *gorm.DB, groupId int64, queryParams map[string]interface{}) ([]schemas.Task, error)
	GetTask(tx *gorm.DB, taskId int64) (*schemas.TaskDetailed, error)
	GetTaskByTitle(tx *gorm.DB, title string) (*schemas.Task, error)
	UpdateTask(tx *gorm.DB, taskId int64, updateInfo schemas.UpdateTask) error

	ParseInputOutput(archivePath string) (int, error)
	modelToSchema(model *models.Task) *schemas.Task
}

type taskService struct {
	fileStorageUrl        string
	taskRepository        repository.TaskRepository
	inputOutputRepository repository.InputOutputRepository
	logger                *zap.SugaredLogger
}

func (ts *taskService) Create(tx *gorm.DB, task *schemas.Task) (int64, error) {
	// Create a new task
	_, err := ts.GetTaskByTitle(tx, task.Title)
	if err != nil && err != ErrTaskNotFound {
		ts.logger.Errorf("Error getting task by title: %v", err.Error())
		return 0, err
	} else if err == nil {
		return 0, ErrTaskExists
	}

	model := models.Task{
		Title:     task.Title,
		CreatedBy: task.CreatedBy,
	}
	taskId, err := ts.taskRepository.Create(tx, &model)
	if err != nil {
		ts.logger.Errorf("Error creating task: %v", err.Error())
		return 0, err
	}

	return taskId, nil
}

func (ts *taskService) GetTaskByTitle(tx *gorm.DB, title string) (*schemas.Task, error) {
	task, err := ts.taskRepository.GetTaskByTitle(tx, title)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTaskNotFound
		}
		ts.logger.Errorf("Error getting task by title: %v", err.Error())
		return nil, err
	}

	result := ts.modelToSchema(task)

	return result, nil
}

func (ts *taskService) GetAll(tx *gorm.DB, queryParams map[string]interface{}) ([]schemas.Task, error) {

	limit := queryParams["limit"].(uint64)
	offset := queryParams["offset"].(uint64)
	sort := queryParams["sort"].(string)

	// Get all tasks
	tasks, err := ts.taskRepository.GetAllTasks(tx, int(limit), int(offset), sort)
	if err != nil {
		ts.logger.Errorf("Error getting all tasks: %v", err.Error())
		return nil, err
	}

	// Convert the models to schemas
	var result []schemas.Task
	for _, task := range tasks {
		result = append(result, *ts.modelToSchema(&task))
	}

	return result, nil
}

func (ts *taskService) GetAllForUser(tx *gorm.DB, userId int64, queryParams map[string]interface{}) ([]schemas.Task, error) {
	limit := queryParams["limit"].(uint64)
	offset := queryParams["offset"].(uint64)
	sort := queryParams["sort"].(string)

	// Get all tasks
	tasks, err := ts.taskRepository.GetAllForUser(tx, userId, int(limit), int(offset), sort)
	if err != nil {
		ts.logger.Errorf("Error getting all tasks for user: %v", err.Error())
		return nil, err
	}

	// Convert the models to schemas
	var result []schemas.Task
	for _, task := range tasks {
		result = append(result, *ts.modelToSchema(&task))
	}

	return result, nil
}

func (ts *taskService) GetAllForGroup(tx *gorm.DB, groupId int64, queryParams map[string]interface{}) ([]schemas.Task, error) {
	limit := queryParams["limit"].(uint64)
	offset := queryParams["offset"].(uint64)
	sort := queryParams["sort"].(string)

	// Get all tasks
	tasks, err := ts.taskRepository.GetAllForGroup(tx, groupId, int(limit), int(offset), sort)
	if err != nil {
		ts.logger.Error("Error getting all tasks for group")
		return nil, err
	}

	// Convert the models to schemas
	var result []schemas.Task
	for _, task := range tasks {
		result = append(result, *ts.modelToSchema(&task))
	}

	return result, nil
}

func (ts *taskService) GetTask(tx *gorm.DB, taskId int64) (*schemas.TaskDetailed, error) {
	// Get the task
	task, err := ts.taskRepository.GetTask(tx, taskId)
	if err != nil {
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return nil, err
	}

	// Convert the model to schema
	result := &schemas.TaskDetailed{
		Id:             task.Id,
		Title:          task.Title,
		DescriptionURL: fmt.Sprintf("%s/getTaskDescription?taskID=%d", ts.fileStorageUrl, task.Id),
		CreatedBy:      task.CreatedBy,
		CreatedByName:  task.Author.Name,
		CreatedAt:      task.CreatedAt,
	}

	return result, nil
}

func (ts *taskService) UpdateTask(tx *gorm.DB, taskId int64, updateInfo schemas.UpdateTask) error {
	currentTask, err := ts.taskRepository.GetTask(tx, taskId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrTaskNotFound
		}
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return err
	}

	ts.updateModel(currentTask, &updateInfo)

	// Update the task
	err = ts.taskRepository.UpdateTask(tx, taskId, currentTask)
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
		return -1, fmt.Errorf("failed to open archive: %v", err)
	}
	defer archive.Close()
	temp_dir, err := os.MkdirTemp(os.TempDir(), "task-upload-archive")
	if err != nil {
		return -1, fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(temp_dir)
	err = utils.DecompressArchive(archive, temp_dir)
	if err != nil {
		return -1, fmt.Errorf("failed to decompress archive: %v", err)
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
		return -1, fmt.Errorf("failed to read input directory: %v", err)
	}
	outputFiles, err := os.ReadDir(temp_dir + "/output")
	if err != nil {
		return -1, fmt.Errorf("failed to read output directory: %v", err)
	}
	if len(inputFiles) != len(outputFiles) {
		return -1, fmt.Errorf("number of input files does not match number of output files")
	}
	for _, file := range inputFiles {
		if file.IsDir() {
			return -1, fmt.Errorf("input directory contains a subdirectory")
		}
		filename_list := strings.Split(file.Name(), ".")
		if len(filename_list) != 2 {
			return -1, fmt.Errorf("input file name is not formatted correctly. Expected format: <filename>.<extension> but got %s", file.Name())
		}
		filename := filename_list[0]
		found := false
		for _, output_file := range outputFiles {
			output_filename_list := strings.Split(output_file.Name(), ".")
			if len(output_filename_list) != 2 {
				return -1, fmt.Errorf("output file name is not formatted correctly. Expected format: <filename>.<extension> but got %s", output_file.Name())
			}
			output_filename := output_filename_list[0]
			ts.logger.Infof("Comparing %s with %s", filename, output_filename)
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
	num_files, err := ts.ParseInputOutput(archivePath)
	if err != nil {
		return fmt.Errorf("failed to parse input and output files: %v", err)
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

func (ts *taskService) updateModel(currentModel *models.Task, updateInfo *schemas.UpdateTask) {
	if updateInfo.Title != "" {
		currentModel.Title = updateInfo.Title
	}
}

func (ts *taskService) modelToSchema(model *models.Task) *schemas.Task {
	return &schemas.Task{
		Id:        model.Id,
		Title:     model.Title,
		CreatedBy: model.CreatedBy,
		CreatedAt: model.CreatedAt,
	}
}

func NewTaskService(fileStorageUrl string, taskRepository repository.TaskRepository, inputOutputRepository repository.InputOutputRepository) TaskService {
	log := utils.NewNamedLogger("task_service")
	return &taskService{
		fileStorageUrl:        fileStorageUrl,
		taskRepository:        taskRepository,
		inputOutputRepository: inputOutputRepository,
		logger:                log,
	}
}
