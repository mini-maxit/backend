.PHONY: docs mocks generate migrate-diff migrate-apply
# Generate docs from comments
docs: $(wildcard internal/api/http/routes/*.go)
	swag init --dir ./cmd/app,./internal/api/http/httputils,./package/domain/schemas,./internal/api/http/routes,./package/domain/types -o ./docs --ot yaml

# Generate mocks using mockgen
mocks: $(shell find package -name '*.go' 2>/dev/null)
	mockgen -destination package/service/mocks/mockgen.go ./package/service AccessControlService,ContestService,UserService,TaskService,AuthService,GroupService,SubmissionService,LanguageService,JWTService,QueueService,WorkerService
	mockgen -destination package/repository/mocks/mockgen.go ./package/repository AccessControlRepository,SubmissionRepository,GroupRepository,TestCaseRepository,LanguageRepository,QueueMessageRepository,SubmissionResultRepository,TaskRepository,TestRepository,UserRepository,File,ContestRepository
	mockgen -destination ./package/filestorage/mockgen.go -package filestorage github.com/mini-maxit/backend/package/filestorage Decompressor,ArchiveValidator

# Regenerate everything
generate: docs mocks

# Generate a migration from GORM model changes: make migrate-diff name=add_foo
migrate-diff:
	atlas migrate diff --env gorm $(name)

# Apply pending migrations: make migrate-apply DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=disable
# Uses the migrations dir directly so it also works in the runtime image (no Go toolchain there).
migrate-apply:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is required" && exit 1)
	atlas migrate apply --dir file://migrations --url "$(DATABASE_URL)"
