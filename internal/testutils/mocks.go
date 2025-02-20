package testutils

import (
	"time"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"gorm.io/gorm"
)

type MockUserRepository struct {
	users   map[string]*models.User
	counter int64
}

func (ur *MockUserRepository) CreateUser(tx *gorm.DB, user *models.User) (int64, error) {
	if tx == nil {
		return 0, gorm.ErrInvalidDB
	}
	ur.users[user.Email] = user
	ur.counter++
	user.Id = ur.counter
	return ur.counter, nil
}

func (ur *MockUserRepository) GetUser(tx *gorm.DB, userId int64) (*models.User, error) {
	if tx == nil {
		return nil, gorm.ErrInvalidDB
	}
	for _, user := range ur.users {
		if user.Id == userId {
			return user, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (ur *MockUserRepository) GetUserByEmail(tx *gorm.DB, email string) (*models.User, error) {
	if tx == nil {
		return nil, gorm.ErrInvalidDB
	}
	if user, ok := ur.users[email]; ok {
		return user, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (ur *MockUserRepository) GetAllUsers(tx *gorm.DB, limit, offset int, sort string) ([]models.User, error) {
	if tx == nil {
		return nil, gorm.ErrInvalidDB
	}
	users := make([]models.User, 0, len(ur.users))
	for _, user := range ur.users {
		users = append(users, *user)
	}
	return users, nil
}

func (ur *MockUserRepository) EditUser(tx *gorm.DB, user *schemas.User) error {
	if tx == nil {
		return gorm.ErrInvalidDB
	}
	var userModel *models.User
	for _, u := range ur.users {
		if u.Id == user.Id {
			userModel = u
			break
		}
	}
	if userModel == nil {
		return gorm.ErrRecordNotFound
	}
	userModel.Name = user.Name
	userModel.Surname = user.Surname
	userModel.Email = user.Email
	userModel.Username = user.Username
	return nil
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:   make(map[string]*models.User),
		counter: 0,
	}
}

type MockSessionRepository struct {
	sessions map[string]*models.Session
}

func (sr *MockSessionRepository) CreateSession(tx *gorm.DB, session *models.Session) error {
	sr.sessions[session.Id] = session
	return nil
}

func (sr *MockSessionRepository) GetSession(tx *gorm.DB, sessionId string) (*models.Session, error) {
	if session, ok := sr.sessions[sessionId]; ok {
		return session, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (sr *MockSessionRepository) GetSessionByUserId(tx *gorm.DB, userId int64) (*models.Session, error) {
	for _, session := range sr.sessions {
		if session.UserId == userId {
			return session, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (sr *MockSessionRepository) UpdateExpiration(tx *gorm.DB, sessionId string, expires_at time.Time) error {
	if session, ok := sr.sessions[sessionId]; ok {
		session.ExpiresAt = expires_at
		return nil
	}
	return gorm.ErrRecordNotFound
}

func (sr *MockSessionRepository) DeleteSession(tx *gorm.DB, sessionId string) error {
	delete(sr.sessions, sessionId)
	return nil
}

func NewMockSessionRepository() *MockSessionRepository {
	return &MockSessionRepository{
		sessions: make(map[string]*models.Session),
	}
}

type MockTaskRepository struct {
	tasks   map[int64]*models.Task
	counter int64
}

func (tr *MockTaskRepository) Create(tx *gorm.DB, task *models.Task) (int64, error) {
	tr.counter++
	task.Id = tr.counter
	tr.tasks[task.Id] = task
	return task.Id, nil
}

func (tr *MockTaskRepository) GetTask(tx *gorm.DB, taskId int64) (*models.Task, error) {
	if task, ok := tr.tasks[taskId]; ok {
		return task, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (tr *MockTaskRepository) GetAllTasks(tx *gorm.DB, limit, offset int, sort string) ([]models.Task, error) {
	tasks := make([]models.Task, 0, len(tr.tasks))
	for _, task := range tr.tasks {
		tasks = append(tasks, *task)
	}
	return tasks, nil
}

func (tr *MockTaskRepository) GetAllForUser(tx *gorm.DB, userId int64, limit, offset int, sort string) ([]models.Task, error) {
	panic("implement me")
	// var tasks []models.Task
	// for _, task := range tr.tasks {
	// 	if task.CreatedBy == userId {
	// 		tasks = append(tasks, *task)
	// 	}
	// }
	// return tasks, nil
}

func (tr *MockTaskRepository) GetAllForGroup(tx *gorm.DB, groupId int64, limit, offset int, sort string) ([]models.Task, error) {
	panic("implement me")
	// var tasks []models.Task
	// for _, task := range tr.tasks {
	// 	if task.GroupId == groupId {
	// 		tasks = append(tasks, *task)
	// 	}
	// }
	// return tasks, nil
}

func (tr *MockTaskRepository) GetTaskByTitle(tx *gorm.DB, title string) (*models.Task, error) {
	for _, task := range tr.tasks {
		if task.Title == title {
			return task, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (tr *MockTaskRepository) GetTaskTimeLimits(tx *gorm.DB, taskId int64) ([]float64, error) {
	panic("implement me")
	// if task, ok := tr.tasks[taskId]; ok {
	// 	return task.TimeLimits, nil
	// }
	// return nil, gorm.ErrRecordNotFound
}

func (tr *MockTaskRepository) GetTaskMemoryLimits(tx *gorm.DB, taskId int64) ([]float64, error) {
	panic("implement me")
	// if task, ok := tr.tasks[taskId]; ok {
	// 	return task.MemoryLimits, nil
	// }
	// return nil, gorm.ErrRecordNotFound
}

func (tr *MockTaskRepository) Clear() {
	tr.tasks = make(map[int64]*models.Task)
	tr.counter = 0
}

func (tr *MockTaskRepository) UpdateTask(tx *gorm.DB, taskId int64, task *models.Task) error {
	if _, ok := tr.tasks[taskId]; ok {
		tr.tasks[taskId] = task
		return nil
	}
	return gorm.ErrRecordNotFound
}

func NewMockTaskRepository() *MockTaskRepository {
	return &MockTaskRepository{
		tasks:   make(map[int64]*models.Task),
		counter: 0,
	}
}

type MockSubmissionRepository struct {
	submissions map[int64]*models.Submission
	counter     int64
}

func (sr *MockSubmissionRepository) CreateSubmission(tx *gorm.DB, submission *models.Submission) (int64, error) {
	sr.counter++
	submission.Id = sr.counter
	sr.submissions[submission.Id] = submission
	return submission.Id, nil
}

func (sr *MockSubmissionRepository) GetSubmission(tx *gorm.DB, submissionId int64) (*models.Submission, error) {
	if submission, ok := sr.submissions[submissionId]; ok {
		return submission, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (sr *MockSubmissionRepository) MarkSubmissionProcessing(tx *gorm.DB, submissionId int64) error {
	if submission, ok := sr.submissions[submissionId]; ok {
		submission.Status = "processing"
		return nil
	}
	return gorm.ErrRecordNotFound
}

func (sr *MockSubmissionRepository) MarkSubmissionComplete(tx *gorm.DB, submissionId int64) error {
	if submission, ok := sr.submissions[submissionId]; ok {
		submission.Status = "completed"
		return nil
	}
	return gorm.ErrRecordNotFound
}

func (sr *MockSubmissionRepository) MarkSubmissionFailed(tx *gorm.DB, submissionId int64, errorMsg string) error {
	if submission, ok := sr.submissions[submissionId]; ok {
		submission.Status = "failed"
		submission.StatusMessage = errorMsg
		return nil
	}
	return gorm.ErrRecordNotFound
}

func NewMockSubmissionRepository() *MockSubmissionRepository {
	return &MockSubmissionRepository{
		submissions: make(map[int64]*models.Submission),
		counter:     0,
	}
}
