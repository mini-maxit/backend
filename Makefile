# Generate docs from comments
docs: $(wildcard internal/api/http/routes/*.go)
	swag init --dir ./cmd/app,./internal/api/http/httputils,./package/domain/schemas,. -o ./docs --ot yaml --st

# Generate mocks using mockgen
mocks: $(shell find package -name '*.go')
	mockgen -destination package/service/mocks/mockgen.go ./package/service ContestService,UserService,TaskService,AuthService,GroupService,SubmissionService,LanguageService,JWTService,QueueService,WorkerService
	mockgen -destination package/repository/mocks/mockgen.go ./package/repository SubmissionRepository,GroupRepository,TestCaseRepository,LanguageRepository,QueueMessageRepository,SubmissionResultRepository,TaskRepository,TestRepository,UserRepository,File,ContestRepository

# Regenerate everything
generate: docs mocks
