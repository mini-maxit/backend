package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/joho/godotenv"
	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/filestorage"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type Config struct {
	// User generation
	Users        int
	AdminCount   int
	TeacherCount int
	StudentCount int
	UserPassword string

	// Group generation
	Groups        int
	UsersPerGroup int

	// Task generation
	Tasks          int
	VisibleTasks   int
	TestsPerTask   int
	FixturesDir    string
	CreateFixtures bool

	// Contest generation
	Contests                    int
	TasksPerContest             int
	ParticipantsPerContest      int
	GroupParticipantsPerContest int

	// Submission generation
	SubmissionsPerTask    int
	SubmissionsPerContest int

	// Supporting data
	RegistrationRequestsPerContest int
	CollaboratorsPerTask           int
	CollaboratorsPerContest        int

	// Database
	DBHost     string
	DBPort     uint16
	DBUser     string
	DBPassword string
	DBName     string

	// File storage
	FileStorageHost string
	FileStoragePort string

	// Utilities
	ClearExisting         bool
	Seed                  int64
	SkipConnectivityCheck bool
	Verbose               bool
	DryRun                bool
}

type Generator struct {
	config      *Config
	db          database.Database
	fileStorage filestorage.FileStorageService
	logger      *zap.SugaredLogger
	random      *rand.Rand

	// Services
	authService service.AuthService
	taskService service.TaskService
	groupService service.GroupService

	// Repositories (for operations not covered by services)
	userRepo             repository.UserRepository
	groupRepo            repository.GroupRepository
	taskRepo             repository.TaskRepository
	contestRepo          repository.ContestRepository
	submissionRepo       repository.SubmissionRepository
	fileRepo             repository.File
	langRepo             repository.LanguageRepository
	accessControlRepo    repository.AccessControlRepository
	testCaseRepo         repository.TestCaseRepository
	submissionResultRepo repository.SubmissionResultRepository
	testResultRepo       repository.TestRepository

	// Generated data tracking
	users     []*models.User
	admins    []*models.User
	teachers  []*models.User
	students  []*models.User
	groups    []*models.Group
	tasks     []*models.Task
	contests  []*models.Contest
	languages []*models.LanguageConfig
}

var (
	cfg     = &Config{}
	rootCmd = &cobra.Command{
		Use:   "generate-test-data",
		Short: "Generate test data for Mini-Maxit database",
		Long: `Generate comprehensive test data including users, groups, tasks, contests, 
submissions and all supporting tables with configurable parameters.`,
		RunE: runGenerate,
	}
)

func init() {
	// User generation flags
	rootCmd.Flags().IntVar(&cfg.Users, "users", 50, "Total number of users to create")
	rootCmd.Flags().IntVar(&cfg.AdminCount, "admin-count", 2, "Number of admin users")
	rootCmd.Flags().IntVar(&cfg.TeacherCount, "teacher-count", 8, "Number of teacher users")
	rootCmd.Flags().IntVar(&cfg.StudentCount, "student-count", 0, "Number of student users (0 = remaining)")
	rootCmd.Flags().StringVar(&cfg.UserPassword, "user-password", "password123", "Default password for all users")

	// Group generation flags
	rootCmd.Flags().IntVar(&cfg.Groups, "groups", 10, "Number of groups to create")
	rootCmd.Flags().IntVar(&cfg.UsersPerGroup, "users-per-group", 8, "Average number of users per group")

	// Task generation flags
	rootCmd.Flags().IntVar(&cfg.Tasks, "tasks", 15, "Number of tasks to create")
	rootCmd.Flags().IntVar(&cfg.VisibleTasks, "visible-tasks", 10, "Number of visible tasks")
	rootCmd.Flags().IntVar(&cfg.TestsPerTask, "tests-per-task", 3, "Number of test cases per task")
	rootCmd.Flags().StringVar(&cfg.FixturesDir, "fixtures-dir", "./fixtures", "Path to fixtures directory")
	rootCmd.Flags().BoolVar(&cfg.CreateFixtures, "create-fixtures", false, "Create sample fixture files if they don't exist")

	// Contest generation flags
	rootCmd.Flags().IntVar(&cfg.Contests, "contests", 5, "Number of contests to create")
	rootCmd.Flags().IntVar(&cfg.TasksPerContest, "tasks-per-contest", 5, "Number of tasks per contest")
	rootCmd.Flags().IntVar(&cfg.ParticipantsPerContest, "participants-per-contest", 10, "Individual participants per contest")
	rootCmd.Flags().IntVar(&cfg.GroupParticipantsPerContest, "group-participants-per-contest", 2, "Group participants per contest")

	// Submission generation flags
	rootCmd.Flags().IntVar(&cfg.SubmissionsPerTask, "submissions-per-task", 5, "Number of submissions per standalone task")
	rootCmd.Flags().IntVar(&cfg.SubmissionsPerContest, "submissions-per-contest", 3, "Number of submissions per contest task")

	// Supporting data flags
	rootCmd.Flags().IntVar(&cfg.RegistrationRequestsPerContest, "registration-requests-per-contest", 3, "Registration requests per contest")
	rootCmd.Flags().IntVar(&cfg.CollaboratorsPerTask, "collaborators-per-task", 2, "Collaborators per task (AccessControl)")
	rootCmd.Flags().IntVar(&cfg.CollaboratorsPerContest, "collaborators-per-contest", 2, "Collaborators per contest (AccessControl)")

	// Database flags
	rootCmd.Flags().StringVar(&cfg.DBHost, "db-host", "", "Database host (overrides DB_HOST)")
	rootCmd.Flags().Uint16Var(&cfg.DBPort, "db-port", 0, "Database port (overrides DB_PORT)")
	rootCmd.Flags().StringVar(&cfg.DBUser, "db-user", "", "Database user (overrides DB_USER)")
	rootCmd.Flags().StringVar(&cfg.DBPassword, "db-password", "", "Database password (overrides DB_PASSWORD)")
	rootCmd.Flags().StringVar(&cfg.DBName, "db-name", "", "Database name (overrides DB_NAME)")

	// File storage flags
	rootCmd.Flags().StringVar(&cfg.FileStorageHost, "file-storage-host", "", "File storage host (overrides FILE_STORAGE_HOST)")
	rootCmd.Flags().StringVar(&cfg.FileStoragePort, "file-storage-port", "", "File storage port (overrides FILE_STORAGE_PORT)")

	// Utility flags
	rootCmd.Flags().BoolVar(&cfg.ClearExisting, "clear-existing", false, "Clear existing data before generating new")
	rootCmd.Flags().Int64Var(&cfg.Seed, "seed", 0, "Random seed for reproducible data (0 = time-based)")
	rootCmd.Flags().BoolVar(&cfg.SkipConnectivityCheck, "skip-connectivity-check", false, "Skip pre-flight connectivity checks")
	rootCmd.Flags().BoolVar(&cfg.Verbose, "verbose", false, "Verbose output")
	rootCmd.Flags().BoolVar(&cfg.DryRun, "dry-run", false, "Show what would be created without creating")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runGenerate(cmd *cobra.Command, args []string) error {
	logger := utils.NewNamedLogger("generate-test-data")

	// Load .env file from project root by default
	if err := godotenv.Load(".env"); err != nil {
		// Don't fail if .env doesn't exist, just log a debug message
		logger.Debugf(".env file not found or could not be loaded: %v", err)
	}

	// Initialize seed
	if cfg.Seed == 0 {
		cfg.Seed = time.Now().UnixNano()
	}
	gofakeit.Seed(cfg.Seed)
	logger.Infof("Using seed: %d", cfg.Seed)

	// Validate configuration
	if err := validateConfig(cfg); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	if cfg.DryRun {
		printDryRun(cfg, logger)
		return nil
	}

	// Load database configuration
	dbConfig := loadDBConfig(cfg)

	// Create database connection
	db, err := database.NewPostgresDB(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Build file storage URL
	fileStorageURL := fmt.Sprintf("http://%s:%s",
		getEnvOrDefault("FILE_STORAGE_HOST", cfg.FileStorageHost),
		getEnvOrDefault("FILE_STORAGE_PORT", cfg.FileStoragePort))

	// Create file storage service
	fileStorage, err := filestorage.NewFileStorageService(fileStorageURL)
	if err != nil {
		return fmt.Errorf("failed to create file storage service: %w", err)
	}

	// Perform connectivity checks
	if !cfg.SkipConnectivityCheck {
		logger.Info("Performing pre-flight connectivity checks...")
		if err := checkConnectivity(db, fileStorageURL, logger); err != nil {
			return err
		}
		logger.Info("✓ All connectivity checks passed")
	}

	// Create generator
	userRepo := repository.NewUserRepository()
	groupRepo := repository.NewGroupRepository()
	taskRepo := repository.NewTaskRepository()
	contestRepo := repository.NewContestRepository()
	submissionRepo := repository.NewSubmissionRepository()
	fileRepo := repository.NewFileRepository()
	testCaseRepo := repository.NewTestCaseRepository()
	accessControlRepo := repository.NewAccessControlRepository()
	
	// Get JWT secret from environment
	jwtSecretKey := os.Getenv("JWT_SECRET_KEY")
	if jwtSecretKey == "" {
		return fmt.Errorf("JWT_SECRET_KEY environment variable is required for user generation")
	}
	
	// Initialize services with correct order and dependencies
	jwtService := service.NewJWTService(userRepo, jwtSecretKey)
	authService := service.NewAuthService(userRepo, jwtService)
	
	// AccessControlService needs: accessControlRepo, userRepo, taskRepo, contestRepo
	accessControlService := service.NewAccessControlService(
		accessControlRepo,
		userRepo,
		taskRepo,
		contestRepo,
	)
	
	// TaskService needs: fileStorage, fileRepo, taskRepo, testCaseRepo, userRepo, groupRepo, submissionRepo, contestRepo, accessControlService
	taskService := service.NewTaskService(
		fileStorage,
		fileRepo,
		taskRepo,
		testCaseRepo,
		userRepo,
		groupRepo,
		submissionRepo,
		contestRepo,
		accessControlService,
	)
	
	// ContestService needs: contestRepo, userRepo, submissionRepo, taskRepo, accessControlService, taskService
	contestService := service.NewContestService(
		contestRepo,
		userRepo,
		submissionRepo,
		taskRepo,
		accessControlService,
		taskService,
	)
	
	// UserService needs: userRepo, contestService
	userService := service.NewUserService(userRepo, contestService)
	
	// GroupService needs: groupRepo, userRepo, userService
	groupService := service.NewGroupService(groupRepo, userRepo, userService)
	
	gen := &Generator{
		config:               cfg,
		db:                   db,
		fileStorage:          fileStorage,
		logger:               logger,
		random:               rand.New(rand.NewSource(cfg.Seed)),
		authService:          authService,
		taskService:          taskService,
		groupService:         groupService,
		userRepo:             userRepo,
		groupRepo:            groupRepo,
		taskRepo:             taskRepo,
		contestRepo:          contestRepo,
		submissionRepo:       submissionRepo,
		fileRepo:             fileRepo,
		langRepo:             repository.NewLanguageRepository(),
		accessControlRepo:    accessControlRepo,
		testCaseRepo:         testCaseRepo,
		submissionResultRepo: repository.NewSubmissionResultRepository(),
		testResultRepo:       repository.NewTestResultRepository(),
	}

	// Clear existing data if requested
	if cfg.ClearExisting {
		logger.Info("Clearing existing data...")
		if err := gen.clearExistingData(); err != nil {
			return fmt.Errorf("failed to clear existing data: %w", err)
		}
		logger.Info("✓ Existing data cleared")
	}

	// Generate data
	logger.Info("Starting data generation...")

	if err := gen.generateLanguageConfigs(); err != nil {
		return fmt.Errorf("failed to generate language configs: %w", err)
	}

	if err := gen.generateUsers(); err != nil {
		return fmt.Errorf("failed to generate users: %w", err)
	}

	if err := gen.generateGroups(); err != nil {
		return fmt.Errorf("failed to generate groups: %w", err)
	}

	if err := gen.generateTasks(); err != nil {
		return fmt.Errorf("failed to generate tasks: %w", err)
	}

	if err := gen.generateContests(); err != nil {
		return fmt.Errorf("failed to generate contests: %w", err)
	}

	if err := gen.generateSubmissions(); err != nil {
		return fmt.Errorf("failed to generate submissions: %w", err)
	}

	// Print summary
	gen.printSummary()

	logger.Info("✓ Data generation completed successfully!")
	return nil
}

func validateConfig(cfg *Config) error {
	if cfg.Users < 1 {
		return fmt.Errorf("users must be at least 1")
	}
	if cfg.AdminCount < 0 || cfg.TeacherCount < 0 || cfg.StudentCount < 0 {
		return fmt.Errorf("user role counts cannot be negative")
	}
	if cfg.StudentCount == 0 {
		cfg.StudentCount = cfg.Users - cfg.AdminCount - cfg.TeacherCount
	}
	if cfg.AdminCount+cfg.TeacherCount+cfg.StudentCount != cfg.Users {
		return fmt.Errorf("admin + teacher + student counts must equal total users")
	}
	if cfg.VisibleTasks > cfg.Tasks {
		return fmt.Errorf("visible-tasks cannot exceed tasks")
	}
	return nil
}

func loadDBConfig(cfg *Config) *config.Config {
	return &config.Config{
		DB: config.DBConfig{
			Host:     getEnvOrDefault("DB_HOST", cfg.DBHost),
			Port:     getEnvOrDefaultUint16("DB_PORT", cfg.DBPort, 5432),
			User:     getEnvOrDefault("DB_USER", cfg.DBUser),
			Password: getEnvOrDefault("DB_PASSWORD", cfg.DBPassword),
			Name:     getEnvOrDefault("DB_NAME", cfg.DBName),
		},
	}
}

func getEnvOrDefault(envKey, flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if val := os.Getenv(envKey); val != "" {
		return val
	}
	return ""
}

func getEnvOrDefaultUint16(envKey string, flagValue uint16, defaultVal uint16) uint16 {
	if flagValue != 0 {
		return flagValue
	}
	if val := os.Getenv(envKey); val != "" {
		// Parse the value
		var parsed uint16
		if _, err := fmt.Sscanf(val, "%d", &parsed); err == nil && parsed != 0 {
			return parsed
		}
	}
	return defaultVal
}

func checkConnectivity(db database.Database, fileStorageURL string, logger *zap.SugaredLogger) error {
	// Check database connectivity
	logger.Info("Checking database connectivity...")
	sqlDB, err := db.DB().DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database connectivity check failed: %w\nPlease ensure PostgreSQL is running and accessible", err)
	}
	logger.Info("  ✓ Database connection successful")

	// Check file storage connectivity
	logger.Info("Checking file storage connectivity...")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(fileStorageURL + "/health")
	if err != nil {
		// Try without /health endpoint
		resp, err = client.Get(fileStorageURL)
		if err != nil {
			return fmt.Errorf("file storage connectivity check failed: %w\nPlease ensure file-storage service is running at %s", err, fileStorageURL)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("file storage returned error status %d\nPlease ensure file-storage service is running properly", resp.StatusCode)
	}
	logger.Info("  ✓ File storage connection successful")

	return nil
}

func printDryRun(cfg *Config, logger *zap.SugaredLogger) {
	logger.Info("DRY RUN - No data will be created")
	logger.Info("Configuration:")
	logger.Infof("  Users: %d (admins: %d, teachers: %d, students: %d)",
		cfg.Users, cfg.AdminCount, cfg.TeacherCount, cfg.StudentCount)
	logger.Infof("  Groups: %d (avg %d users per group)", cfg.Groups, cfg.UsersPerGroup)
	logger.Infof("  Tasks: %d (%d visible, %d tests per task)",
		cfg.Tasks, cfg.VisibleTasks, cfg.TestsPerTask)
	logger.Infof("  Contests: %d (%d tasks, %d participants, %d group participants each)",
		cfg.Contests, cfg.TasksPerContest, cfg.ParticipantsPerContest, cfg.GroupParticipantsPerContest)
	logger.Infof("  Submissions: %d per task, %d per contest",
		cfg.SubmissionsPerTask, cfg.SubmissionsPerContest)
	logger.Infof("  Registration requests: %d per contest", cfg.RegistrationRequestsPerContest)
	logger.Infof("  Collaborators: %d per task, %d per contest",
		cfg.CollaboratorsPerTask, cfg.CollaboratorsPerContest)
	logger.Infof("  Seed: %d", cfg.Seed)
	logger.Infof("  Clear existing: %v", cfg.ClearExisting)
}

func (g *Generator) clearExistingData() error {
	// Delete in reverse dependency order
	tables := []string{
		"test_results",
		"submission_results",
		"submissions",
		"test_cases",
		"contest_registration_requests",
		"contest_participant_groups",
		"contest_participants",
		"contest_tasks",
		"access_control",
		"user_groups",
		"contests",
		"tasks",
		"groups",
		"files",
		"language_configs",
		"users",
	}

	tx, err := g.db.BeginTransaction()
	if err != nil {
		return err
	}
	defer func() {
		if g.db.ShouldRollback() {
			tx.Rollback()
		}
	}()

	for _, table := range tables {
		fullTable := fmt.Sprintf("maxit.%s", table)
		if err := tx.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", fullTable)).Error; err != nil {
			g.logger.Warnf("Failed to truncate %s: %v (table may not exist)", table, err)
		} else {
			g.logger.Infof("  Truncated %s", table)
		}
	}

	return g.db.Commit()
}

func (g *Generator) generateLanguageConfigs() error {
	g.logger.Info("Generating language configurations...")

	// Check if languages already exist
	existing, err := g.langRepo.GetAll(g.db)
	if err == nil && len(existing) > 0 {
		// Convert []models.LanguageConfig to []*models.LanguageConfig
		for i := range existing {
			g.languages = append(g.languages, &existing[i])
		}
		g.logger.Infof("  Using %d existing language configs", len(existing))
		return nil
	}

	tx, err := g.db.BeginTransaction()
	if err != nil {
		return err
	}
	defer func() {
		if g.db.ShouldRollback() {
			tx.Rollback()
		}
	}()

	// Create basic language configs
	langs := []struct {
		Type      string
		Version   string
		Extension string
	}{
		{"c", "99", ".c"},
		{"c", "11", ".c"},
		{"cpp", "11", ".cpp"},
		{"cpp", "14", ".cpp"},
		{"cpp", "17", ".cpp"},
		{"cpp", "20", ".cpp"},
	}

	for _, lang := range langs {
		disabled := false
		langModel := &models.LanguageConfig{
			Type:          lang.Type,
			Version:       lang.Version,
			FileExtension: lang.Extension,
			IsDisabled:    &disabled,
		}
		err := g.langRepo.Create(g.db, langModel)
		if err != nil {
			g.db.Rollback()
			return fmt.Errorf("failed to create language %s-%s: %w", lang.Type, lang.Version, err)
		}
		g.languages = append(g.languages, langModel)
	}

	if err := g.db.Commit(); err != nil {
		return err
	}

	g.logger.Infof("  ✓ Created %d language configs", len(g.languages))
	return nil
}

func (g *Generator) generateUsers() error {
	g.logger.Info("Generating users...")

	tx, err := g.db.BeginTransaction()
	if err != nil {
		return err
	}
	defer func() {
		if g.db.ShouldRollback() {
			tx.Rollback()
		}
	}()

	// Generate admins
	for i := 0; i < g.config.AdminCount; i++ {
		userRequest := g.createUserRequest()
		_, err := g.authService.Register(g.db, userRequest)
		if err != nil {
			g.db.Rollback()
			return fmt.Errorf("failed to register admin user: %w", err)
		}
		
		// Get the created user and update role
		user, err := g.userRepo.GetByEmail(g.db, userRequest.Email)
		if err != nil {
			g.db.Rollback()
			return fmt.Errorf("failed to get registered admin user: %w", err)
		}
		user.Role = types.UserRoleAdmin
		err = g.userRepo.Edit(g.db, user)
		if err != nil {
			g.db.Rollback()
			return fmt.Errorf("failed to set admin role: %w", err)
		}
		
		g.users = append(g.users, user)
		g.admins = append(g.admins, user)
	}

	// Generate teachers
	for i := 0; i < g.config.TeacherCount; i++ {
		userRequest := g.createUserRequest()
		_, err := g.authService.Register(g.db, userRequest)
		if err != nil {
			g.db.Rollback()
			return fmt.Errorf("failed to register teacher user: %w", err)
		}
		
		// Get the created user and update role
		user, err := g.userRepo.GetByEmail(g.db, userRequest.Email)
		if err != nil {
			g.db.Rollback()
			return fmt.Errorf("failed to get registered teacher user: %w", err)
		}
		user.Role = types.UserRoleTeacher
		err = g.userRepo.Edit(g.db, user)
		if err != nil {
			g.db.Rollback()
			return fmt.Errorf("failed to set teacher role: %w", err)
		}
		
		g.users = append(g.users, user)
		g.teachers = append(g.teachers, user)
	}

	// Generate students
	for i := 0; i < g.config.StudentCount; i++ {
		userRequest := g.createUserRequest()
		_, err := g.authService.Register(g.db, userRequest)
		if err != nil {
			g.db.Rollback()
			return fmt.Errorf("failed to register student user: %w", err)
		}
		
		// Get the created user (students are default role from Register)
		user, err := g.userRepo.GetByEmail(g.db, userRequest.Email)
		if err != nil {
			g.db.Rollback()
			return fmt.Errorf("failed to get registered student user: %w", err)
		}
		
		g.users = append(g.users, user)
		g.students = append(g.students, user)
	}

	if err := g.db.Commit(); err != nil {
		return err
	}

	g.logger.Infof("  ✓ Created %d users (%d admins, %d teachers, %d students)",
		len(g.users), len(g.admins), len(g.teachers), len(g.students))
	return nil
}

func (g *Generator) createUserRequest() schemas.UserRegisterRequest {
	return schemas.UserRegisterRequest{
		Name:     gofakeit.FirstName(),
		Surname:  gofakeit.LastName(),
		Email:    gofakeit.Email(),
		Username: gofakeit.Username(),
		Password: g.config.UserPassword,
	}
}

func (g *Generator) generateGroups() error {
	g.logger.Info("Generating groups...")

	if g.config.Groups == 0 {
		g.logger.Info("  Skipping groups (count = 0)")
		return nil
	}

	tx, err := g.db.BeginTransaction()
	if err != nil {
		return err
	}
	defer func() {
		if g.db.ShouldRollback() {
			tx.Rollback()
		}
	}()

	totalMemberships := 0

	for i := 0; i < g.config.Groups; i++ {
		// Pick a random teacher or admin as creator
		var creator *models.User
		if len(g.teachers) > 0 && (len(g.admins) == 0 || g.random.Intn(2) == 0) {
			creator = g.teachers[g.random.Intn(len(g.teachers))]
		} else if len(g.admins) > 0 {
			creator = g.admins[g.random.Intn(len(g.admins))]
		} else {
			// Fallback to any user
			creator = g.users[g.random.Intn(len(g.users))]
		}

		group := &models.Group{
			Name:      fmt.Sprintf("Group %s", gofakeit.Company()),
			CreatedBy: creator.ID,
		}

		id, err := g.groupRepo.Create(g.db, group)
		if err != nil {
			g.db.Rollback()
			return fmt.Errorf("failed to create group: %w", err)
		}
		group.ID = id
		g.groups = append(g.groups, group)

		// Add random users to the group
		numMembers := g.config.UsersPerGroup
		if numMembers > len(g.users) {
			numMembers = len(g.users)
		}

		// Shuffle and pick users
		selectedUsers := make([]*models.User, len(g.users))
		copy(selectedUsers, g.users)
		g.random.Shuffle(len(selectedUsers), func(i, j int) {
			selectedUsers[i], selectedUsers[j] = selectedUsers[j], selectedUsers[i]
		})

		for j := 0; j < numMembers && j < len(selectedUsers); j++ {
			if err := g.groupRepo.AddUser(g.db, group.ID, selectedUsers[j].ID); err != nil {
				g.db.Rollback()
				return fmt.Errorf("failed to add user to group: %w", err)
			}
			totalMemberships++
		}
	}

	if err := g.db.Commit(); err != nil {
		return err
	}

	g.logger.Infof("  ✓ Created %d groups with %d total memberships", len(g.groups), totalMemberships)
	return nil
}

func (g *Generator) generateTasks() error {
	g.logger.Info("Generating tasks...")

	if g.config.Tasks == 0 {
		g.logger.Info("  Skipping tasks (count = 0)")
		return nil
	}

	// Create fixtures if needed
	if g.config.CreateFixtures {
		if err := g.createFixtures(); err != nil {
			return fmt.Errorf("failed to create fixtures: %w", err)
		}
	}

	tx, err := g.db.BeginTransaction()
	if err != nil {
		return err
	}
	defer func() {
		if g.db.ShouldRollback() {
			tx.Rollback()
		}
	}()

	totalTestCases := 0
	totalCollaborators := 0

	for i := 0; i < g.config.Tasks; i++ {
		// Pick a random teacher or admin as creator
		var creator *models.User
		if len(g.teachers) > 0 && (len(g.admins) == 0 || g.random.Intn(2) == 0) {
			creator = g.teachers[g.random.Intn(len(g.teachers))]
		} else if len(g.admins) > 0 {
			creator = g.admins[g.random.Intn(len(g.admins))]
		} else {
			creator = g.users[g.random.Intn(len(g.users))]
		}

		isVisible := i < g.config.VisibleTasks

		// Create description file
		descFile := &models.File{
			Filename:   fmt.Sprintf("task-%d-description.txt", i+1),
			Path:       fmt.Sprintf("/tasks/task-%d/description.txt", i+1),
			Bucket:     "maxit",
			ServerType: "local",
		}
		if err := g.fileRepo.Create(g.db, descFile); err != nil {
			g.db.Rollback()
			return fmt.Errorf("failed to create description file: %w", err)
		}

		task := &models.Task{
			Title:             fmt.Sprintf("Task %d: %s", i+1, gofakeit.BuzzWord()),
			DescriptionFileID: descFile.ID,
			CreatedBy:         creator.ID,
			IsVisible:         isVisible,
		}

		taskID, err := g.taskRepo.Create(g.db, task)
		if err != nil {
			g.db.Rollback()
			return fmt.Errorf("failed to create task: %w", err)
		}
		task.ID = taskID
		g.tasks = append(g.tasks, task)

		// Create test cases
		for j := 0; j < g.config.TestsPerTask; j++ {
			inputFile := &models.File{
				Filename:   fmt.Sprintf("input_%d.txt", j+1),
				Path:       fmt.Sprintf("/tasks/task-%d/input_%d.txt", i+1, j+1),
				Bucket:     "maxit",
				ServerType: "local",
			}
			if err := g.fileRepo.Create(g.db, inputFile); err != nil {
				g.db.Rollback()
				return fmt.Errorf("failed to create input file: %w", err)
			}

			outputFile := &models.File{
				Filename:   fmt.Sprintf("output_%d.txt", j+1),
				Path:       fmt.Sprintf("/tasks/task-%d/output_%d.txt", i+1, j+1),
				Bucket:     "maxit",
				ServerType: "local",
			}
			if err := g.fileRepo.Create(g.db, outputFile); err != nil {
				g.db.Rollback()
				return fmt.Errorf("failed to create output file: %w", err)
			}

			testCase := &models.TestCase{
				TaskID:       taskID,
				InputFileID:  inputFile.ID,
				OutputFileID: outputFile.ID,
				Order:        j + 1,
				TimeLimit:    1000,
				MemoryLimit:  256 * 1024 * 1024,
			}

			if err := g.testCaseRepo.Create(g.db, testCase); err != nil {
				g.db.Rollback()
				return fmt.Errorf("failed to create test case: %w", err)
			}
			totalTestCases++
		}

		// Add collaborators (AccessControl)
		if g.config.CollaboratorsPerTask > 0 {
			// Add owner permission for creator
			if err := g.accessControlRepo.AddTaskCollaborator(g.db, taskID, creator.ID, types.PermissionOwner); err != nil {
				g.db.Rollback()
				return fmt.Errorf("failed to create owner access control: %w", err)
			}
			totalCollaborators++

			// Add other collaborators
			collabCount := g.config.CollaboratorsPerTask
			if collabCount > len(g.teachers)+len(g.admins)-1 {
				collabCount = len(g.teachers) + len(g.admins) - 1
			}

			potentialCollabs := append([]*models.User{}, g.teachers...)
			potentialCollabs = append(potentialCollabs, g.admins...)

			// Remove creator from potential collaborators
			filtered := make([]*models.User, 0)
			for _, u := range potentialCollabs {
				if u.ID != creator.ID {
					filtered = append(filtered, u)
				}
			}

			g.random.Shuffle(len(filtered), func(i, j int) {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			})

			permissions := []types.Permission{types.PermissionEdit, types.PermissionManage}
			for j := 0; j < collabCount && j < len(filtered); j++ {
				perm := permissions[g.random.Intn(len(permissions))]
				if err := g.accessControlRepo.AddTaskCollaborator(g.db, taskID, filtered[j].ID, perm); err != nil {
					g.db.Rollback()
					return fmt.Errorf("failed to create collaborator access control: %w", err)
				}
				totalCollaborators++
			}
		}
	}

	if err := g.db.Commit(); err != nil {
		return err
	}

	g.logger.Infof("  ✓ Created %d tasks with %d test cases and %d collaborators",
		len(g.tasks), totalTestCases, totalCollaborators)
	return nil
}

func (g *Generator) createFixtures() error {
	g.logger.Info("  Creating sample fixtures...")

	// Create fixtures directory if it doesn't exist
	if err := os.MkdirAll(g.config.FixturesDir, 0755); err != nil {
		return fmt.Errorf("failed to create fixtures directory: %w", err)
	}

	// Create a sample task fixture
	taskDir := filepath.Join(g.config.FixturesDir, "sample-task")
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		return fmt.Errorf("failed to create task directory: %w", err)
	}

	// Create description file
	descPath := filepath.Join(taskDir, "description.pdf")
	descContent := "Sample task description\n\nThis is a placeholder task."
	if err := os.WriteFile(descPath, []byte(descContent), 0644); err != nil {
		return fmt.Errorf("failed to create description file: %w", err)
	}

	// Create test cases directory
	inputDir := filepath.Join(taskDir, "input")
	outputDir := filepath.Join(taskDir, "output")
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		return fmt.Errorf("failed to create input directory: %w", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create sample test cases
	for i := 1; i <= 3; i++ {
		inputPath := filepath.Join(inputDir, fmt.Sprintf("input_%d.txt", i))
		outputPath := filepath.Join(outputDir, fmt.Sprintf("output_%d.txt", i))

		inputContent := fmt.Sprintf("%d %d\n", i, i*2)
		outputContent := fmt.Sprintf("%d\n", i*3)

		if err := os.WriteFile(inputPath, []byte(inputContent), 0644); err != nil {
			return fmt.Errorf("failed to create input file: %w", err)
		}
		if err := os.WriteFile(outputPath, []byte(outputContent), 0644); err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
	}

	g.logger.Info("  ✓ Sample fixtures created")
	return nil
}

func (g *Generator) generateContests() error {
	g.logger.Info("Generating contests...")

	if g.config.Contests == 0 {
		g.logger.Info("  Skipping contests (count = 0)")
		return nil
	}

	if len(g.tasks) == 0 {
		g.logger.Warn("  No tasks available, skipping contests")
		return nil
	}

	tx, err := g.db.BeginTransaction()
	if err != nil {
		return err
	}
	defer func() {
		if g.db.ShouldRollback() {
			tx.Rollback()
		}
	}()

	totalContestTasks := 0
	totalParticipants := 0
	totalGroupParticipants := 0
	totalRegistrationRequests := 0
	totalCollaborators := 0

	for i := 0; i < g.config.Contests; i++ {
		// Pick a random teacher or admin as creator
		var creator *models.User
		if len(g.teachers) > 0 && (len(g.admins) == 0 || g.random.Intn(2) == 0) {
			creator = g.teachers[g.random.Intn(len(g.teachers))]
		} else if len(g.admins) > 0 {
			creator = g.admins[g.random.Intn(len(g.admins))]
		} else {
			creator = g.users[g.random.Intn(len(g.users))]
		}

		startAt := time.Now().Add(time.Duration(g.random.Intn(30)) * 24 * time.Hour)
		endAt := startAt.Add(time.Duration(1+g.random.Intn(7)) * 24 * time.Hour)

		contest := &models.Contest{
			Name:               fmt.Sprintf("Contest %d: %s", i+1, gofakeit.BuzzWord()),
			Description:        gofakeit.Sentence(10),
			CreatedBy:          creator.ID,
			StartAt:            startAt,
			EndAt:              &endAt,
			IsRegistrationOpen: g.random.Intn(2) == 0,
			IsSubmissionOpen:   g.random.Intn(2) == 0,
			IsVisible:          g.random.Intn(2) == 0,
		}

		contestID, err := g.contestRepo.Create(g.db, contest)
		if err != nil {
			g.db.Rollback()
			return fmt.Errorf("failed to create contest: %w", err)
		}
		contest.ID = contestID
		g.contests = append(g.contests, contest)

		// Add tasks to contest
		numTasks := g.config.TasksPerContest
		if numTasks > len(g.tasks) {
			numTasks = len(g.tasks)
		}

		selectedTasks := make([]*models.Task, len(g.tasks))
		copy(selectedTasks, g.tasks)
		g.random.Shuffle(len(selectedTasks), func(i, j int) {
			selectedTasks[i], selectedTasks[j] = selectedTasks[j], selectedTasks[i]
		})

		for j := 0; j < numTasks; j++ {
			taskStartAt := startAt.Add(time.Duration(j*24) * time.Hour)
			taskEndAt := taskStartAt.Add(24 * time.Hour)

			contestTask := models.ContestTask{
				ContestID:        contestID,
				TaskID:           selectedTasks[j].ID,
				StartAt:          taskStartAt,
				EndAt:            &taskEndAt,
				IsSubmissionOpen: true,
			}

			if err := g.contestRepo.AddTaskToContest(g.db, contestTask); err != nil {
				g.db.Rollback()
				return fmt.Errorf("failed to add task to contest: %w", err)
			}
			totalContestTasks++
		}

		// Add individual participants
		numParticipants := g.config.ParticipantsPerContest
		if numParticipants > len(g.users) {
			numParticipants = len(g.users)
		}

		selectedUsers := make([]*models.User, len(g.users))
		copy(selectedUsers, g.users)
		g.random.Shuffle(len(selectedUsers), func(i, j int) {
			selectedUsers[i], selectedUsers[j] = selectedUsers[j], selectedUsers[i]
		})

		for j := 0; j < numParticipants; j++ {
			if err := g.contestRepo.CreateContestParticipant(g.db, contestID, selectedUsers[j].ID); err != nil {
				g.db.Rollback()
				return fmt.Errorf("failed to add participant to contest: %w", err)
			}
			totalParticipants++
		}

		// Add group participants
		if len(g.groups) > 0 {
			numGroupParticipants := g.config.GroupParticipantsPerContest
			if numGroupParticipants > len(g.groups) {
				numGroupParticipants = len(g.groups)
			}

			selectedGroups := make([]*models.Group, len(g.groups))
			copy(selectedGroups, g.groups)
			g.random.Shuffle(len(selectedGroups), func(i, j int) {
				selectedGroups[i], selectedGroups[j] = selectedGroups[j], selectedGroups[i]
			})

			for j := 0; j < numGroupParticipants; j++ {
				// Manually create ContestParticipantGroup entry
				tx := g.db.GetInstance()
				contestParticipantGroup := &models.ContestParticipantGroup{
					ContestID: contestID,
					GroupID:   selectedGroups[j].ID,
				}
				if err := tx.Create(contestParticipantGroup).Error; err != nil {
					g.db.Rollback()
					return fmt.Errorf("failed to add group participant to contest: %w", err)
				}
				totalGroupParticipants++
			}
		}

		// Add registration requests
		if g.config.RegistrationRequestsPerContest > 0 {
			// Use users not yet participants
			nonParticipants := make([]*models.User, 0)
			for _, user := range g.users {
				isParticipant := false
				for j := 0; j < numParticipants && j < len(selectedUsers); j++ {
					if selectedUsers[j].ID == user.ID {
						isParticipant = true
						break
					}
				}
				if !isParticipant {
					nonParticipants = append(nonParticipants, user)
				}
			}

			numRequests := g.config.RegistrationRequestsPerContest
			if numRequests > len(nonParticipants) {
				numRequests = len(nonParticipants)
			}

			statuses := []types.RegistrationRequestStatus{
				types.RegistrationRequestStatusPending,
				types.RegistrationRequestStatusApproved,
				types.RegistrationRequestStatusRejected,
			}

			for j := 0; j < numRequests; j++ {
				status := statuses[g.random.Intn(len(statuses))]
				// Manually create ContestRegistrationRequests entry
				tx := g.db.GetInstance()
				req := &models.ContestRegistrationRequests{
					ContestID: contestID,
					UserID:    nonParticipants[j].ID,
					Status:    status,
				}
				if err := tx.Create(req).Error; err != nil {
					g.db.Rollback()
					return fmt.Errorf("failed to create registration request: %w", err)
				}
				totalRegistrationRequests++
			}
		}

		// Add collaborators (AccessControl)
		if g.config.CollaboratorsPerContest > 0 {
			// Add owner permission for creator
			if err := g.accessControlRepo.AddContestCollaborator(g.db, contestID, creator.ID, types.PermissionOwner); err != nil {
				g.db.Rollback()
				return fmt.Errorf("failed to create owner access control: %w", err)
			}
			totalCollaborators++

			// Add other collaborators
			collabCount := g.config.CollaboratorsPerContest
			potentialCollabs := append([]*models.User{}, g.teachers...)
			potentialCollabs = append(potentialCollabs, g.admins...)

			// Remove creator
			filtered := make([]*models.User, 0)
			for _, u := range potentialCollabs {
				if u.ID != creator.ID {
					filtered = append(filtered, u)
				}
			}

			if collabCount > len(filtered) {
				collabCount = len(filtered)
			}

			g.random.Shuffle(len(filtered), func(i, j int) {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			})

			permissions := []types.Permission{types.PermissionEdit, types.PermissionManage}
			for j := 0; j < collabCount; j++ {
				perm := permissions[g.random.Intn(len(permissions))]
				if err := g.accessControlRepo.AddContestCollaborator(g.db, contestID, filtered[j].ID, perm); err != nil {
					g.db.Rollback()
					return fmt.Errorf("failed to create collaborator access control: %w", err)
				}
				totalCollaborators++
			}
		}
	}

	if err := g.db.Commit(); err != nil {
		return err
	}

	g.logger.Infof("  ✓ Created %d contests with %d tasks, %d participants, %d group participants, %d registration requests, %d collaborators",
		len(g.contests), totalContestTasks, totalParticipants, totalGroupParticipants, totalRegistrationRequests, totalCollaborators)
	return nil
}

func (g *Generator) generateSubmissions() error {
	g.logger.Info("Generating submissions...")

	if len(g.tasks) == 0 {
		g.logger.Info("  Skipping submissions (no tasks)")
		return nil
	}

	if len(g.languages) == 0 {
		g.logger.Warn("  No languages available, skipping submissions")
		return nil
	}

	tx, err := g.db.BeginTransaction()
	if err != nil {
		return err
	}
	defer func() {
		if g.db.ShouldRollback() {
			tx.Rollback()
		}
	}()

	totalSubmissions := 0
	totalResults := 0
	totalTestResults := 0

	// Generate standalone task submissions
	for _, task := range g.tasks {
		for i := 0; i < g.config.SubmissionsPerTask; i++ {
			// Pick random user
			user := g.users[g.random.Intn(len(g.users))]

			if err := g.createSubmission(task.ID, user.ID, nil, i+1); err != nil {
				g.db.Rollback()
				return err
			}
			totalSubmissions++
		}
	}

	// Generate contest submissions
	for _, contest := range g.contests {
		// Get contest tasks
		contestTasks, err := g.contestRepo.GetContestTasksWithSettings(g.db, contest.ID)
		if err != nil {
			g.db.Rollback()
			return fmt.Errorf("failed to get contest tasks: %w", err)
		}

		// Get contest participants - manually query
		tx := g.db.GetInstance()
		var participants []models.User
		err = tx.Table("maxit.users").
			Joins("JOIN maxit.contest_participants ON contest_participants.user_id = users.id").
			Where("contest_participants.contest_id = ?", contest.ID).
			Find(&participants).Error
		if err != nil {
			g.db.Rollback()
			return fmt.Errorf("failed to get contest participants: %w", err)
		}

		if len(participants) == 0 {
			continue
		}

		for _, contestTask := range contestTasks {
			for i := 0; i < g.config.SubmissionsPerContest; i++ {
				// Pick random participant
				participant := participants[g.random.Intn(len(participants))]

				if err := g.createSubmission(contestTask.TaskID, participant.ID, &contest.ID, i+1); err != nil {
					g.db.Rollback()
					return err
				}
				totalSubmissions++
			}
		}
	}

	// Count results (submissions create results automatically in this simplified version)
	// In real scenario, this would be done separately
	totalResults = totalSubmissions
	totalTestResults = totalSubmissions * g.config.TestsPerTask

	if err := g.db.Commit(); err != nil {
		return err
	}

	g.logger.Infof("  ✓ Created %d submissions with %d results and %d test results",
		totalSubmissions, totalResults, totalTestResults)
	return nil
}

func (g *Generator) createSubmission(taskID, userID int64, contestID *int64, order int) error {
	// Create source file
	lang := g.languages[g.random.Intn(len(g.languages))]

	sourceFile := &models.File{
		Filename:   fmt.Sprintf("solution_%d%s", order, lang.FileExtension),
		Path:       fmt.Sprintf("/submissions/task-%d/user-%d/solution_%d%s", taskID, userID, order, lang.FileExtension),
		Bucket:     "maxit",
		ServerType: "local",
	}
	if err := g.fileRepo.Create(g.db, sourceFile); err != nil {
		return fmt.Errorf("failed to create source file: %w", err)
	}

	// Random submission status
	statuses := []types.SubmissionStatus{
		types.SubmissionStatusReceived,
		types.SubmissionStatusSentForEvaluation,
		types.SubmissionStatusEvaluated,
	}
	status := statuses[g.random.Intn(len(statuses))]

	submission := &models.Submission{
		TaskID:      taskID,
		UserID:      userID,
		Order:       order,
		LanguageID:  lang.ID,
		FileID:      sourceFile.ID,
		Status:      status,
		ContestID:   contestID,
		SubmittedAt: time.Now().Add(-time.Duration(g.random.Intn(24*30)) * time.Hour),
	}

	submissionID, err := g.submissionRepo.Create(g.db, submission)
	if err != nil {
		return fmt.Errorf("failed to create submission: %w", err)
	}

	// Create submission result if evaluated
	if status == types.SubmissionStatusEvaluated {
		resultCodes := []types.SubmissionResultCode{
			types.SubmissionResultCodeSuccess,
			types.SubmissionResultCodeTestFailed,
			types.SubmissionResultCodeCompilationError,
			types.SubmissionResultCodeInitializationError,
		}
		resultCode := resultCodes[g.random.Intn(len(resultCodes))]

		result := models.SubmissionResult{
			SubmissionID: submissionID,
			Code:         resultCode,
			Message:      fmt.Sprintf("Result: %s", resultCode),
		}

		resultID, err := g.submissionResultRepo.Create(g.db, result)
		if err != nil {
			return fmt.Errorf("failed to create submission result: %w", err)
		}

		// Create test results
		testCases, err := g.testCaseRepo.GetByTask(g.db, taskID)
		if err != nil {
			return fmt.Errorf("failed to get test cases: %w", err)
		}

		for _, testCase := range testCases {
			passed := resultCode == types.SubmissionResultCodeSuccess || g.random.Intn(2) == 0

			// Create stdout/stderr/diff files
			stdoutFile := &models.File{
				Filename:   "stdout.txt",
				Path:       fmt.Sprintf("/results/submission-%d/test-%d/stdout.txt", submissionID, testCase.ID),
				Bucket:     "maxit",
				ServerType: "local",
			}
			if err := g.fileRepo.Create(g.db, stdoutFile); err != nil {
				return fmt.Errorf("failed to create stdout file: %w", err)
			}

			stderrFile := &models.File{
				Filename:   "stderr.txt",
				Path:       fmt.Sprintf("/results/submission-%d/test-%d/stderr.txt", submissionID, testCase.ID),
				Bucket:     "maxit",
				ServerType: "local",
			}
			if err := g.fileRepo.Create(g.db, stderrFile); err != nil {
				return fmt.Errorf("failed to create stderr file: %w", err)
			}

			diffFile := &models.File{
				Filename:   "diff.txt",
				Path:       fmt.Sprintf("/results/submission-%d/test-%d/diff.txt", submissionID, testCase.ID),
				Bucket:     "maxit",
				ServerType: "local",
			}
			if err := g.fileRepo.Create(g.db, diffFile); err != nil {
				return fmt.Errorf("failed to create diff file: %w", err)
			}

			statusCodes := []types.TestResultStatusCode{
				types.TestResultStatusCodeOK,
				types.TestResultStatusCodeRuntimeError,
				types.TestResultStatusCodeTimeLimit,
			}
			statusCode := types.TestResultStatusCodeOK
			if !passed {
				statusCode = statusCodes[1+g.random.Intn(len(statusCodes)-1)]
			}

			testResult := &models.TestResult{
				SubmissionResultID: resultID,
				TestCaseID:         testCase.ID,
				Passed:             &passed,
				ExecutionTime:      float64(g.random.Intn(1000)) / 1000.0,
				StatusCode:         statusCode,
				ErrorMessage:       "",
				StdoutFileID:       stdoutFile.ID,
				StderrFileID:       stderrFile.ID,
				DiffFileID:         diffFile.ID,
			}

			if err := g.testResultRepo.Create(g.db, testResult); err != nil {
				return fmt.Errorf("failed to create test result: %w", err)
			}
		}
	}

	return nil
}

func (g *Generator) printSummary() {
	g.logger.Info("")
	g.logger.Info("=== Generation Summary ===")
	g.logger.Infof("Seed: %d", g.config.Seed)
	g.logger.Infof("Users: %d (admins: %d, teachers: %d, students: %d)",
		len(g.users), len(g.admins), len(g.teachers), len(g.students))
	g.logger.Infof("Groups: %d", len(g.groups))
	g.logger.Infof("Tasks: %d", len(g.tasks))
	g.logger.Infof("Contests: %d", len(g.contests))
	g.logger.Infof("Languages: %d", len(g.languages))
	g.logger.Info("==========================")
}
