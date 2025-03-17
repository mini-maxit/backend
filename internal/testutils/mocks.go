package testutils

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"golang.org/x/crypto/bcrypt"
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

func (ur *MockUserRepository) CreateUser(tx *gorm.DB, user *models.User) (int64, error) {
	if ur.failNext {
		ur.failNext = false
		return 0, gorm.ErrInvalidDB
	}
	if tx == nil {
		return 0, gorm.ErrInvalidDB
	}
	ur.users[user.Email] = user
	ur.counter++
	user.Id = ur.counter
	return ur.counter, nil
}

func (ur *MockUserRepository) GetUser(tx *gorm.DB, userId int64) (*models.User, error) {
	if ur.failNext {
		ur.failNext = false
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

func (ur *MockUserRepository) EditUser(tx *gorm.DB, user *models.User) error {
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

func (sr *MockSessionRepository) CreateSession(tx *gorm.DB, session *models.Session) error {
	sr.sessions[session.Id] = session
	return nil
}

func (sr *MockSessionRepository) GetSession(tx *gorm.DB, sessionId string) (*models.Session, error) {
	if sr.failNext {
		sr.failNext = false
		return nil, gorm.ErrInvalidDB
	}
	if session, ok := sr.sessions[sessionId]; ok {
		return session, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (sr *MockSessionRepository) GetSessionByUserId(tx *gorm.DB, userId int64) (*models.Session, error) {
	if sr.failNext {
		log.Printf("Failing getsession by user id")
		sr.failNext = false
		return nil, gorm.ErrInvalidDB
	}
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
	tasks       map[int64]*models.Task
	tasksCouter int64

	taskGroups []*models.TaskGroup

	taskUsers []*models.TaskUser

	groupRepository *MockGroupRepository
}

func (tr *MockTaskRepository) Create(tx *gorm.DB, task *models.Task) (int64, error) {
	tr.tasksCouter++
	task.Id = tr.tasksCouter
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
	var tasks []models.Task
	for _, task := range tr.tasks {
		for _, group := range task.Groups {
			if group.Id == groupId {
				tasks = append(tasks, *task)
				break
			}
		}
	}
	return tasks, nil
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
	tr.tasksCouter = 0
}

func (tr *MockTaskRepository) EditTask(tx *gorm.DB, taskId int64, task *models.Task) error {
	if _, ok := tr.tasks[taskId]; ok {
		tr.tasks[taskId] = task
		return nil
	}
	return gorm.ErrRecordNotFound
}

func (tr *MockTaskRepository) DeleteTask(tx *gorm.DB, taskId int64) error {
	if _, ok := tr.tasks[taskId]; ok {
		delete(tr.tasks, taskId)
		return nil
	}
	return gorm.ErrRecordNotFound
}

func (tr *MockTaskRepository) AssignTaskToUser(tx *gorm.DB, taskId, userId int64) error {
	if _, ok := tr.tasks[taskId]; !ok {
		return gorm.ErrRecordNotFound
	}

	for _, taskUser := range tr.taskUsers {
		if taskUser.TaskId == taskId && taskUser.UserId == userId {
			return errors.ErrTaskAlreadyAssigned
		}
	}

	tr.taskUsers = append(tr.taskUsers, &models.TaskUser{
		TaskId: taskId,
		UserId: userId,
	})
	return nil
}

func (tr *MockTaskRepository) AssignTaskToGroup(tx *gorm.DB, taskId, groupId int64) error {
	for _, taskGroup := range tr.taskGroups {
		if taskGroup.TaskId == taskId && taskGroup.GroupId == groupId {
			return nil
		}
	}
	tr.taskGroups = append(tr.taskGroups, &models.TaskGroup{
		TaskId:  taskId,
		GroupId: groupId,
	})
	for _, task := range tr.tasks {
		if task.Id == taskId {
			task.Groups = append(task.Groups, models.Group{
				Id: groupId,
			})
		}
	}
	return nil
}

func (tr *MockTaskRepository) UnAssignTaskFromUser(tx *gorm.DB, taskId, userId int64) error {
	for i, taskUser := range tr.taskUsers {
		if taskUser.TaskId == taskId && taskUser.UserId == userId {
			tr.taskUsers = append(tr.taskUsers[:i], tr.taskUsers[i+1:]...)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (tr *MockTaskRepository) UnAssignTaskFromGroup(tx *gorm.DB, taskId, groupId int64) error {
	for i, taskGroup := range tr.taskGroups {
		if taskGroup.TaskId == taskId && taskGroup.GroupId == groupId {
			tr.taskGroups = append(tr.taskGroups[:i], tr.taskGroups[i+1:]...)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (tr *MockTaskRepository) GetAllAssignedTasks(tx *gorm.DB, userId int64, limit, offset int, sort string) ([]models.Task, error) {
	var tasks []models.Task
	for _, taskUser := range tr.taskUsers {
		if taskUser.UserId == userId {
			if task, ok := tr.tasks[taskUser.TaskId]; ok {
				tasks = append(tasks, *task)
			}
		}
	}

	var userGroups []int64
	for _, userGroup := range tr.groupRepository.userGroups {
		for _, ug := range userGroup {
			if ug.UserId == userId {
				userGroups = append(userGroups, ug.GroupId)
			}
		}
	}

	for _, taskGroup := range tr.taskGroups {
		for _, ug := range userGroups {
			if taskGroup.GroupId == ug {
				if task, ok := tr.tasks[taskGroup.TaskId]; ok {
					tasks = append(tasks, *task)
				}
			}
		}
	}

	return tasks, nil
}

func (tr *MockTaskRepository) GetAllCreatedTasks(tx *gorm.DB, userId int64, limit, offset int, sort string) ([]models.Task, error) {
	var tasks []models.Task
	for _, task := range tr.tasks {
		if task.CreatedBy == userId {
			tasks = append(tasks, *task)
		}
	}
	return tasks, nil
}

func (tr *MockTaskRepository) IsTaskAssignedToGroup(tx *gorm.DB, taskId, groupId int64) (bool, error) {
	for _, taskGroup := range tr.taskGroups {
		if taskGroup.TaskId == taskId && taskGroup.GroupId == groupId {
			return true, nil
		}
	}
	return false, nil
}

func (tr *MockTaskRepository) IsTaskAssignedToUser(tx *gorm.DB, taskId, userId int64) (bool, error) {
	for _, taskUser := range tr.taskUsers {
		if taskUser.TaskId == taskId && taskUser.UserId == userId {
			return true, nil
		}
	}

	var userGroups []int64
	for _, userGroup := range tr.groupRepository.userGroups {
		for _, ug := range userGroup {
			if ug.UserId == userId {
				userGroups = append(userGroups, ug.GroupId)
			}
		}
	}

	for _, taskGroup := range tr.taskGroups {
		for _, ug := range userGroups {
			if taskGroup.GroupId == ug {
				if taskGroup.TaskId == taskId {
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

type MockGroupRepository struct {
	groups        map[int64]*models.Group
	groupsCounter int64

	userGroups        map[int64][]*models.UserGroup
	userGroupsCounter int64

	userRepository *MockUserRepository
}

func (gr *MockGroupRepository) CreateGroup(tx *gorm.DB, group *models.Group) (int64, error) {
	gr.groupsCounter++
	group.Id = gr.groupsCounter
	gr.groups[group.Id] = group
	return group.Id, nil
}

func (gr *MockGroupRepository) DeleteGroup(tx *gorm.DB, groupId int64) error {
	if _, ok := gr.groups[groupId]; ok {
		delete(gr.groups, groupId)
		return nil
	}
	return gorm.ErrRecordNotFound
}

func (gr *MockGroupRepository) Edit(tx *gorm.DB, groupId int64, group *models.Group) (*models.Group, error) {
	if _, ok := gr.groups[groupId]; ok {
		gr.groups[groupId] = group
		return group, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (gr *MockGroupRepository) GetAllGroup(*gorm.DB, int, int, string) ([]models.Group, error) {
	groups := make([]models.Group, 0, len(gr.groups))
	for _, group := range gr.groups {
		groups = append(groups, *group)
	}
	return groups, nil
}

func (gr *MockGroupRepository) GetGroup(tx *gorm.DB, groupId int64) (*models.Group, error) {
	if group, ok := gr.groups[groupId]; ok {
		return group, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (gr *MockGroupRepository) AddUsersToGroup(tx *gorm.DB, groupId int64, userId []int64) error {
	for _, id := range userId {
		gr.userGroupsCounter++
		gr.userGroups[gr.userGroupsCounter] = append(gr.userGroups[gr.userGroupsCounter], &models.UserGroup{
			GroupId: groupId,
			UserId:  id,
		})
	}
	return nil
}

func (gr *MockGroupRepository) AddUserToGroup(tx *gorm.DB, groupId int64, userId int64) error {
	gr.userGroupsCounter++
	gr.userGroups[gr.userGroupsCounter] = append(gr.userGroups[gr.userGroupsCounter], &models.UserGroup{
		GroupId: groupId,
		UserId:  userId,
	})
	return nil
}

func (gr *MockGroupRepository) DeleteUserFromGroup(tx *gorm.DB, groupId int64, userId int64) error {
	for i, userGroup := range gr.userGroups {
		for j, ug := range userGroup {
			if ug.GroupId == groupId && ug.UserId == userId {
				gr.userGroups[i] = append(gr.userGroups[i][:j], gr.userGroups[i][j+1:]...)
				return nil
			}
		}
	}
	return gorm.ErrRecordNotFound
}

func (gr *MockGroupRepository) GetAllGroupForTeacher(tx *gorm.DB, teacherId int64, offset int, limit int, sort string) ([]models.Group, error) {
	var groups []models.Group
	for _, group := range gr.groups {
		if group.CreatedBy == teacherId {
			groups = append(groups, *group)
		}
	}
	return groups, nil
}

func (gr *MockGroupRepository) GetGroupUsers(tx *gorm.DB, groupId int64) ([]models.User, error) {
	var users []models.User
	for _, userGroup := range gr.userGroups {
		for _, ug := range userGroup {
			if ug.GroupId == groupId {
				user, err := gr.userRepository.GetUser(tx, ug.UserId)
				if err == nil {
					users = append(users, *user)
				}
			}
		}
	}
	return users, nil
}

func (gr *MockGroupRepository) UserBelongsTo(tx *gorm.DB, groupId int64, userId int64) (bool, error) {
	for _, userGroup := range gr.userGroups {
		for _, ug := range userGroup {
			if ug.GroupId == groupId && ug.UserId == userId {
				return true, nil
			}
		}
	}
	return false, nil
}

func (gr *MockGroupRepository) GetGroupTasks(tx *gorm.DB, groupId int64) ([]models.Task, error) {
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

type MockUserService struct {
}

func (us *MockUserService) GetUserByEmail(tx *gorm.DB, email string) (*schemas.User, error) {
	panic("implement me")
}
func (us *MockUserService) GetAllUsers(tx *gorm.DB, queryParams map[string]interface{}) ([]schemas.User, error) {
	panic("implement me")
}
func (us *MockUserService) GetUserById(tx *gorm.DB, userId int64) (*schemas.User, error) {
	panic("implement me")
}
func (us *MockUserService) EditUser(tx *gorm.DB, currentUser schemas.User, userId int64, updateInfo *schemas.UserEdit) error {
	panic("implement me")
}
func (us *MockUserService) ChangeRole(tx *gorm.DB, currentUser schemas.User, userId int64, role types.UserRole) error {
	panic("implement me")
}
func (us *MockUserService) ChangePassword(tx *gorm.DB, currentUser schemas.User, userId int64, data *schemas.UserChangePassword) error {
	panic("implement me")
}

func NewMockUserService() *MockUserService {
	return &MockUserService{}
}

type MockAuthService struct {
	users map[string]*models.User
}

func (as *MockAuthService) SetUser(user *models.User) {
	as.users[user.Email] = user
}

func (as *MockAuthService) ClearUsers() {
	as.users = make(map[string]*models.User)
}

func (as *MockAuthService) Login(tx *gorm.DB, request schemas.UserLoginRequest) (*schemas.Session, error) {
	user := as.users[request.Email]
	if user != nil {
		err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password))
		if err == nil {
			return &schemas.Session{
				Id:        user.Email + request.Email,
				UserId:    user.Id,
				ExpiresAt: time.Now().Add(time.Hour),
			}, nil
		} else if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, errors.ErrInvalidCredentials
		} else {
			return nil, err
		}
	}
	return nil, errors.ErrUserNotFound
}

func (as *MockAuthService) Register(tx *gorm.DB, userRegister schemas.UserRegisterRequest) (*schemas.Session, error) {
	user := as.users[userRegister.Email]
	if user != nil {
		return nil, errors.ErrUserAlreadyExists
	}
	hashPass, err := bcrypt.GenerateFromPassword([]byte(userRegister.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user = &models.User{
		Name:         userRegister.Name,
		Surname:      userRegister.Surname,
		Email:        userRegister.Email,
		Username:     userRegister.Username,
		PasswordHash: string(hashPass),
		Role:         types.UserRoleStudent,
	}
	as.users[user.Email] = user
	return &schemas.Session{
		Id:        user.Email + userRegister.Email,
		UserId:    user.Id,
		UserRole:  string(user.Role),
		ExpiresAt: time.Now().Add(time.Hour),
	}, nil
}

func NewMockAuthService() *MockAuthService {
	return &MockAuthService{
		users: make(map[string]*models.User),
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

func (db *MockDatabase) Db() *gorm.DB {
	return &gorm.DB{}
}

func MockDatabaseMiddleware(next http.Handler, db database.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type MockInputOutputRepository struct {
	repository.InputOutputRepository
	inputOutputs map[int64]*models.InputOutput
	counter      int64
}

func (ir *MockInputOutputRepository) Create(tx *gorm.DB, inputOutput *models.InputOutput) error {
	for _, io := range ir.inputOutputs {
		io.Id = ir.counter
		ir.inputOutputs[ir.counter] = io
		ir.counter++
	}
	return nil
}

func (ir *MockInputOutputRepository) GetInputOutputId(tx *gorm.DB, taskId, order int64) (int64, error) {
	panic("implement me")
}

func (ir *MockInputOutputRepository) DeleteAll(tx *gorm.DB, taskId int64) error {
	toDelete := make([]int64, 0)
	for _, io := range ir.inputOutputs {
		if io.TaskId == taskId {
			toDelete = append(toDelete, io.Id)
		}
	}
	for _, id := range toDelete {
		delete(ir.inputOutputs, id)
	}
	return nil
}

func NewMockInputOutputRepository() repository.InputOutputRepository {
	return &MockInputOutputRepository{counter: 1}
}
