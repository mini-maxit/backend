package testutils

import (
	"context"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	myErrors "github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"gorm.io/gorm"
)

type MockUserRepository struct {
	repository.UserRepository
	users    map[string]*models.User
	counter  int64
	failNext bool
}

func (ur *MockUserRepository) FailNext() {
	ur.failNext = true
}

func (ur *MockUserRepository) Create(tx *gorm.DB, user *models.User) (int64, error) {
	if ur.failNext {
		ur.failNext = false
		return 0, gorm.ErrInvalidDB
	}
	if tx == nil {
		return 0, gorm.ErrInvalidDB
	}
	ur.users[user.Email] = user
	ur.counter++
	user.ID = ur.counter
	return ur.counter, nil
}

func (ur *MockUserRepository) Get(_ *gorm.DB, userID int64) (*models.User, error) {
	if ur.failNext {
		ur.failNext = false
		return nil, gorm.ErrInvalidDB
	}
	for _, user := range ur.users {
		if user.ID == userID {
			return user, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (ur *MockUserRepository) GetByEmail(tx *gorm.DB, email string) (*models.User, error) {
	if tx == nil {
		return nil, gorm.ErrInvalidDB
	}
	if user, ok := ur.users[email]; ok {
		return user, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (ur *MockUserRepository) GetAll(tx *gorm.DB, limit, _ int, _ string) ([]models.User, error) {
	if tx == nil {
		return nil, gorm.ErrInvalidDB
	}
	users := make([]models.User, 0, len(ur.users))
	for _, user := range ur.users {
		users = append(users, *user)
		if len(users) == limit {
			break
		}
	}
	return users, nil
}

func (ur *MockUserRepository) Edit(tx *gorm.DB, user *models.User) error {
	if tx == nil {
		return gorm.ErrInvalidDB
	}
	var userModel *models.User
	for _, u := range ur.users {
		if u.ID == user.ID {
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
	userModel.Role = user.Role
	return nil
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:   make(map[string]*models.User),
		counter: 0,
	}
}

type MockSessionRepository struct {
	repository.SessionRepository
	sessions map[string]*models.Session
	failNext bool
}

func (sr *MockSessionRepository) FailNext() {
	sr.failNext = true
}

func (sr *MockSessionRepository) Create(_ *gorm.DB, session *models.Session) error {
	sr.sessions[session.ID] = session
	return nil
}

func (sr *MockSessionRepository) Get(_ *gorm.DB, sessionID string) (*models.Session, error) {
	if sr.failNext {
		sr.failNext = false
		return nil, gorm.ErrInvalidDB
	}
	if session, ok := sr.sessions[sessionID]; ok {
		return session, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (sr *MockSessionRepository) GetByUserID(_ *gorm.DB, userID int64) (*models.Session, error) {
	if sr.failNext {
		log.Printf("Failing getsession by user id")
		sr.failNext = false
		return nil, gorm.ErrInvalidDB
	}
	for _, session := range sr.sessions {
		if session.UserID == userID {
			return session, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (sr *MockSessionRepository) UpdateExpiration(_ *gorm.DB, sessionID string, expiresAt time.Time) error {
	if session, ok := sr.sessions[sessionID]; ok {
		session.ExpiresAt = expiresAt
		return nil
	}
	return gorm.ErrRecordNotFound
}

func (sr *MockSessionRepository) Delete(_ *gorm.DB, sessionID string) error {
	delete(sr.sessions, sessionID)
	return nil
}

func NewMockSessionRepository() *MockSessionRepository {
	return &MockSessionRepository{
		sessions: make(map[string]*models.Session),
	}
}

type MockTaskRepository struct {
	tasks       map[int64]*models.Task
	tasksCouter int64

	taskGroups []*models.TaskGroup

	taskUsers []*models.TaskUser

	groupRepository *MockGroupRepository
}

func (tr *MockTaskRepository) Create(_ *gorm.DB, task *models.Task) (int64, error) {
	tr.tasksCouter++
	task.ID = tr.tasksCouter
	tr.tasks[task.ID] = task
	return task.ID, nil
}

func (tr *MockTaskRepository) Get(_ *gorm.DB, taskID int64) (*models.Task, error) {
	if task, ok := tr.tasks[taskID]; ok {
		return task, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (tr *MockTaskRepository) GetAll(_ *gorm.DB, _, _ int, _ string) ([]models.Task, error) {
	tasks := make([]models.Task, 0, len(tr.tasks))
	for _, task := range tr.tasks {
		tasks = append(tasks, *task)
	}
	return tasks, nil
}

func (tr *MockTaskRepository) GetAllForUser(_ *gorm.DB, _ int64, _, _ int, _ string) ([]models.Task, error) {
	panic("implement me")
	// var tasks []models.Task
	// for _, task := range tr.tasks {
	// 	if task.CreatedBy == userID {
	// 		tasks = append(tasks, *task)
	// 	}
	// }
	// return tasks, nil
}

func (tr *MockTaskRepository) GetAllForGroup(
	_ *gorm.DB,
	groupID int64,
	_, _ int,
	_ string,
) ([]models.Task, error) {
	var tasks []models.Task
	for _, task := range tr.tasks {
		for _, group := range task.Groups {
			if group.ID == groupID {
				tasks = append(tasks, *task)
				break
			}
		}
	}
	return tasks, nil
}

func (tr *MockTaskRepository) GetByTitle(_ *gorm.DB, title string) (*models.Task, error) {
	for _, task := range tr.tasks {
		if task.Title == title {
			return task, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (tr *MockTaskRepository) GetTimeLimits(_ *gorm.DB, _ int64) ([]int64, error) {
	panic("implement me")
	// if task, ok := tr.tasks[taskID]; ok {
	// 	return task.TimeLimits, nil
	// }
	// return nil, gorm.ErrRecordNotFound
}

func (tr *MockTaskRepository) GetMemoryLimits(_ *gorm.DB, _ int64) ([]int64, error) {
	panic("implement me")
	// if task, ok := tr.tasks[taskID]; ok {
	// 	return task.MemoryLimits, nil
	// }
	// return nil, gorm.ErrRecordNotFound
}

func (tr *MockTaskRepository) Clear() {
	tr.tasks = make(map[int64]*models.Task)
	tr.tasksCouter = 0
}

func (tr *MockTaskRepository) Edit(_ *gorm.DB, taskID int64, task *models.Task) error {
	if _, ok := tr.tasks[taskID]; ok {
		tr.tasks[taskID] = task
		return nil
	}
	return gorm.ErrRecordNotFound
}

func (tr *MockTaskRepository) Delete(_ *gorm.DB, taskID int64) error {
	if _, ok := tr.tasks[taskID]; ok {
		delete(tr.tasks, taskID)
		return nil
	}
	return gorm.ErrRecordNotFound
}

func (tr *MockTaskRepository) AssignToUser(_ *gorm.DB, taskID, userID int64) error {
	if _, ok := tr.tasks[taskID]; !ok {
		return gorm.ErrRecordNotFound
	}

	for _, taskUser := range tr.taskUsers {
		if taskUser.TaskID == taskID && taskUser.UserID == userID {
			return myErrors.ErrTaskAlreadyAssigned
		}
	}

	tr.taskUsers = append(tr.taskUsers, &models.TaskUser{
		TaskID: taskID,
		UserID: userID,
	})
	return nil
}

func (tr *MockTaskRepository) AssignToGroup(_ *gorm.DB, taskID, groupID int64) error {
	for _, taskGroup := range tr.taskGroups {
		if taskGroup.TaskID == taskID && taskGroup.GroupID == groupID {
			return nil
		}
	}
	tr.taskGroups = append(tr.taskGroups, &models.TaskGroup{
		TaskID:  taskID,
		GroupID: groupID,
	})
	for _, task := range tr.tasks {
		if task.ID == taskID {
			task.Groups = append(task.Groups, models.Group{
				ID: groupID,
			})
		}
	}
	return nil
}

func (tr *MockTaskRepository) UnassignFromUser(_ *gorm.DB, taskID, userID int64) error {
	for i, taskUser := range tr.taskUsers {
		if taskUser.TaskID == taskID && taskUser.UserID == userID {
			tr.taskUsers = append(tr.taskUsers[:i], tr.taskUsers[i+1:]...)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (tr *MockTaskRepository) UnassignFromGroup(_ *gorm.DB, taskID, groupID int64) error {
	for i, taskGroup := range tr.taskGroups {
		if taskGroup.TaskID == taskID && taskGroup.GroupID == groupID {
			tr.taskGroups = append(tr.taskGroups[:i], tr.taskGroups[i+1:]...)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (tr *MockTaskRepository) GetAllAssigned(_ *gorm.DB, userID int64, _, _ int, _ string) ([]models.Task, error) {
	var tasks []models.Task
	for _, taskUser := range tr.taskUsers {
		if taskUser.UserID == userID {
			if task, ok := tr.tasks[taskUser.TaskID]; ok {
				tasks = append(tasks, *task)
			}
		}
	}

	var userGroups []int64
	for _, userGroup := range tr.groupRepository.userGroups {
		for _, ug := range userGroup {
			if ug.UserID == userID {
				userGroups = append(userGroups, ug.GroupID)
			}
		}
	}

	for _, taskGroup := range tr.taskGroups {
		for _, ug := range userGroups {
			if taskGroup.GroupID == ug {
				if task, ok := tr.tasks[taskGroup.TaskID]; ok {
					tasks = append(tasks, *task)
				}
			}
		}
	}

	return tasks, nil
}

func (tr *MockTaskRepository) GetAllCreated(_ *gorm.DB, userID int64, _, _ int, _ string) ([]models.Task, error) {
	var tasks []models.Task
	for _, task := range tr.tasks {
		if task.CreatedBy == userID {
			tasks = append(tasks, *task)
		}
	}
	return tasks, nil
}

func (tr *MockTaskRepository) IsAssignedToGroup(_ *gorm.DB, taskID, groupID int64) (bool, error) {
	for _, taskGroup := range tr.taskGroups {
		if taskGroup.TaskID == taskID && taskGroup.GroupID == groupID {
			return true, nil
		}
	}
	return false, nil
}

func (tr *MockTaskRepository) IsAssignedToUser(_ *gorm.DB, taskID, userID int64) (bool, error) {
	for _, taskUser := range tr.taskUsers {
		if taskUser.TaskID == taskID && taskUser.UserID == userID {
			return true, nil
		}
	}

	var userGroups []int64
	for _, userGroup := range tr.groupRepository.userGroups {
		for _, ug := range userGroup {
			if ug.UserID == userID {
				userGroups = append(userGroups, ug.GroupID)
			}
		}
	}

	for _, taskGroup := range tr.taskGroups {
		for _, ug := range userGroups {
			if taskGroup.GroupID == ug {
				if taskGroup.TaskID == taskID {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func NewMockTaskRepository(groupRepository *MockGroupRepository) *MockTaskRepository {
	return &MockTaskRepository{
		tasks:       make(map[int64]*models.Task),
		tasksCouter: 0,

		taskGroups: make([]*models.TaskGroup, 0),

		taskUsers: make([]*models.TaskUser, 0),

		groupRepository: groupRepository,
	}
}

// type MockSubmissionRepository struct {
// 	submissions map[int64]*models.Submission
// 	counter     int64
// }

// func (sr *MockSubmissionRepository) Create(tx *gorm.DB, submission *models.Submission) (int64, error) {
// 	sr.counter++
// 	submission.ID = sr.counter
// 	sr.submissions[submission.ID] = submission
// 	return submission.ID, nil
// }

// func (sr *MockSubmissionRepository) GetSubmission(tx *gorm.DB, submissionID int64) (*models.Submission, error) {
// 	if submission, ok := sr.submissions[submissionID]; ok {
// 		return submission, nil
// 	}
// 	return nil, gorm.ErrRecordNotFound
// }

// func (sr *MockSubmissionRepository) MarkSubmissionProcessing(tx *gorm.DB, submissionID int64) error {
// 	if submission, ok := sr.submissions[submissionID]; ok {
// 		submission.Status = "processing"
// 		return nil
// 	}
// 	return gorm.ErrRecordNotFound
// }

// func (sr *MockSubmissionRepository) MarkSubmissionComplete(tx *gorm.DB, submissionID int64) error {
// 	if submission, ok := sr.submissions[submissionID]; ok {
// 		submission.Status = "completed"
// 		return nil
// 	}
// 	return gorm.ErrRecordNotFound
// }

// func (sr *MockSubmissionRepository) MarkSubmissionFailed(tx *gorm.DB, submissionID int64, errorMsg string) error {
// 	if submission, ok := sr.submissions[submissionID]; ok {
// 		submission.Status = "failed"
// 		submission.StatusMessage = errorMsg
// 		return nil
// 	}
// 	return gorm.ErrRecordNotFound
// }

// func NewMockSubmissionRepository() *MockSubmissionRepository {
// 	return &MockSubmissionRepository{
// 		submissions: make(map[int64]*models.Submission),
// 		counter:     0,
// 	}
// }

type MockGroupRepository struct {
	groups        map[int64]*models.Group
	groupsCounter int64

	userGroups        map[int64][]*models.UserGroup
	userGroupsCounter int64

	userRepository *MockUserRepository
}

func (gr *MockGroupRepository) Create(_ *gorm.DB, group *models.Group) (int64, error) {
	gr.groupsCounter++
	group.ID = gr.groupsCounter
	gr.groups[group.ID] = group
	return group.ID, nil
}

func (gr *MockGroupRepository) Delete(_ *gorm.DB, groupID int64) error {
	if _, ok := gr.groups[groupID]; ok {
		delete(gr.groups, groupID)
		return nil
	}
	return gorm.ErrRecordNotFound
}

func (gr *MockGroupRepository) Edit(_ *gorm.DB, groupID int64, group *models.Group) (*models.Group, error) {
	if _, ok := gr.groups[groupID]; ok {
		gr.groups[groupID] = group
		return group, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (gr *MockGroupRepository) GetAll(*gorm.DB, int, int, string) ([]models.Group, error) {
	groups := make([]models.Group, 0, len(gr.groups))
	for _, group := range gr.groups {
		groups = append(groups, *group)
	}
	return groups, nil
}

func (gr *MockGroupRepository) Get(_ *gorm.DB, groupID int64) (*models.Group, error) {
	if group, ok := gr.groups[groupID]; ok {
		return group, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (gr *MockGroupRepository) AddUsersToGroup(_ *gorm.DB, groupID int64, userID []int64) error {
	for _, id := range userID {
		gr.userGroupsCounter++
		gr.userGroups[gr.userGroupsCounter] = append(gr.userGroups[gr.userGroupsCounter], &models.UserGroup{
			GroupID: groupID,
			UserID:  id,
		})
	}
	return nil
}

func (gr *MockGroupRepository) AddUser(_ *gorm.DB, groupID int64, userID int64) error {
	gr.userGroupsCounter++
	gr.userGroups[gr.userGroupsCounter] = append(gr.userGroups[gr.userGroupsCounter], &models.UserGroup{
		GroupID: groupID,
		UserID:  userID,
	})
	return nil
}

func (gr *MockGroupRepository) DeleteUser(_ *gorm.DB, groupID int64, userID int64) error {
	for i, userGroup := range gr.userGroups {
		for j, ug := range userGroup {
			if ug.GroupID == groupID && ug.UserID == userID {
				gr.userGroups[i] = slices.Delete(gr.userGroups[i], j, j+1)
				return nil
			}
		}
	}
	return gorm.ErrRecordNotFound
}

func (gr *MockGroupRepository) GetAllForTeacher(_ *gorm.DB, userID int64, _, _ int, _ string) ([]models.Group, error) {
	var groups []models.Group
	for _, group := range gr.groups {
		if group.CreatedBy == userID {
			groups = append(groups, *group)
		}
	}
	return groups, nil
}

func (gr *MockGroupRepository) GetUsers(tx *gorm.DB, groupID int64) ([]models.User, error) {
	var users []models.User
	for _, userGroup := range gr.userGroups {
		for _, ug := range userGroup {
			if ug.GroupID == groupID {
				user, err := gr.userRepository.Get(tx, ug.UserID)
				if err == nil {
					users = append(users, *user)
				}
			}
		}
	}
	return users, nil
}

func (gr *MockGroupRepository) UserBelongsTo(_ *gorm.DB, groupID int64, userID int64) (bool, error) {
	for _, userGroup := range gr.userGroups {
		for _, ug := range userGroup {
			if ug.GroupID == groupID && ug.UserID == userID {
				return true, nil
			}
		}
	}
	return false, nil
}

func (gr *MockGroupRepository) GetTasks(_ *gorm.DB, _ int64) ([]models.Task, error) {
	tasks := make([]models.Task, 0)
	return tasks, nil
}

func NewMockGroupRepository(userRepo *MockUserRepository) *MockGroupRepository {
	return &MockGroupRepository{
		groups:        make(map[int64]*models.Group),
		groupsCounter: 0,

		userGroups:        make(map[int64][]*models.UserGroup),
		userGroupsCounter: 0,

		userRepository: userRepo,
	}
}

type MockDatabase struct {
	invalid bool
}

func (db *MockDatabase) BeginTransaction() (*gorm.DB, error) {
	if db.invalid {
		return nil, gorm.ErrInvalidDB
	}
	return &gorm.DB{}, nil
}

func (db *MockDatabase) NewSession() database.Database {
	return &MockDatabase{}
}

func (db *MockDatabase) Commit() error {
	return nil
}

func (db *MockDatabase) Rollback() {
}

func (db *MockDatabase) ShouldRollback() bool {
	return false
}

func (db *MockDatabase) Invalidate() {
	db.invalid = true
}

func (db *MockDatabase) Vaildate() {
	db.invalid = false
}

func (db *MockDatabase) DB() *gorm.DB {
	return &gorm.DB{}
}

func MockDatabaseMiddleware(next http.Handler, db database.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
