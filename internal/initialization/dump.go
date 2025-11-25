package initialization

import (
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/service"
	"go.uber.org/zap"
)

func dump(db database.Database, log *zap.SugaredLogger, authService service.AuthService, userRepository repository.UserRepository) {
	_, err := db.BeginTransaction()
	if err != nil {
		log.Warnf("Failed to connect to database to init dump: %s", err.Error())
	}
	users := []struct {
		Name     string
		Surname  string
		Email    string
		Username string
		Password string
		Role     types.UserRole
	}{

		{
			Name:     "AdminName",
			Surname:  "AdminSurname",
			Email:    "admin@admin.com",
			Username: "admin",
			Password: "adminadmin",
			Role:     types.UserRoleAdmin,
		},
		{
			Name:     "TeacherName",
			Surname:  "TeacherSurname",
			Email:    "teacher@teacher.com",
			Username: "teacher",
			Password: "teacherteacher",
			Role:     types.UserRoleTeacher,
		},
		{
			Name:     "StudentName",
			Surname:  "StudentSurname",
			Email:    "student@student.com",
			Username: "student",
			Password: "studentstudent",
			Role:     types.UserRoleStudent,
		},
	}
	for _, user := range users {
		_, err = authService.Register(db, schemas.UserRegisterRequest{
			Name:     user.Name,
			Surname:  user.Surname,
			Email:    user.Email,
			Username: user.Username,
			Password: user.Password,
		})
		if err != nil {
			log.Warnf("Failed to create %s: %s", user.Username, err.Error())
		}
		registeredUser, err := userRepository.GetByEmail(db, user.Email)
		if err != nil {
			log.Warnf("Failed to get %s user: %s", user.Username, err.Error())
		}
		registeredUser.Role = user.Role
		err = userRepository.Edit(db, registeredUser)
		if err != nil {
			log.Warnf("Failed to set %s role: %s", user.Username, err.Error())
		}
	}

	err = db.Commit()
	if err != nil {
		log.Warnf("Failed to commit transaction after dump: %s", err.Error())
	}
}
