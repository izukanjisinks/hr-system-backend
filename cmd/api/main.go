package main

import (
	"fmt"
	"log"
	"net/http"

	"hr-system/internal/config"
	"hr-system/internal/database"
	"hr-system/internal/handlers"
	"hr-system/internal/jobs"
	"hr-system/internal/middleware"
	"hr-system/internal/repository"
	"hr-system/internal/routes"
	"hr-system/internal/services"
)

func main() {
	cfg := config.Load()

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)
	if err := database.Connect(connStr); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()
	log.Println("Database connected")

	// Repositories — Phase 1
	userRepo := repository.NewUserRepository()
	roleRepo := repository.NewRoleRepository()
	deptRepo := repository.NewDepartmentRepository()
	posRepo := repository.NewPositionRepository()
	empRepo := repository.NewEmployeeRepository()
	docRepo := repository.NewEmployeeDocumentRepository()
	ecRepo := repository.NewEmergencyContactRepository()

	// Repositories — Phase 2
	ltRepo := repository.NewLeaveTypeRepository()
	lbRepo := repository.NewLeaveBalanceRepository()
	lrRepo := repository.NewLeaveRequestRepository()
	holidayRepo := repository.NewHolidayRepository()
	attRepo := repository.NewAttendanceRepository()

	// Workflow Repositories
	workflowRepo := repository.NewWorkflowRepository()
	instanceRepo := repository.NewWorkflowInstanceRepository()
	taskRepo := repository.NewAssignedTaskRepository()
	historyRepo := repository.NewWorkflowHistoryRepository()

	// Services — Phase 1
	roleService := services.NewRoleService(roleRepo)
	userService := services.NewUserService(userRepo, roleRepo)
	deptService := services.NewDepartmentService(deptRepo)
	posService := services.NewPositionService(posRepo, deptRepo)
	empService := services.NewEmployeeService(empRepo, deptRepo, posRepo)
	docService := services.NewEmployeeDocumentService(docRepo, empRepo)
	ecService := services.NewEmergencyContactService(ecRepo, empRepo)

	// Services — Phase 2
	ltService := services.NewLeaveTypeService(ltRepo)
	lbService := services.NewLeaveBalanceService(lbRepo, ltRepo, empRepo)
	lrService := services.NewLeaveRequestService(lrRepo, lbService, ltRepo, holidayRepo, empRepo)
	holidayService := services.NewHolidayService(holidayRepo)
	attService := services.NewAttendanceService(attRepo, holidayRepo, empRepo)

	// Workflow Service
	workflowService := services.NewWorkflowService(workflowRepo, instanceRepo, taskRepo, historyRepo, userRepo, empRepo)

	// Seed predefined roles
	if err := roleService.InitializePredefinedRoles(); err != nil {
		log.Fatalf("Failed to initialize roles: %v", err)
	}
	log.Println("Roles initialized")

	// Seed default super admin
	if err := userService.SeedSuperAdmin("admin@hr-system.com", "Admin@123"); err != nil {
		log.Printf("Warning: failed to seed super admin: %v", err)
	} else {
		log.Println("Super admin ready (admin@hr-system.com)")
	}

	// Seed default leave types
	if err := ltService.SeedDefaults(); err != nil {
		log.Printf("Warning: failed to seed leave types: %v", err)
	} else {
		log.Println("Leave types initialized")
	}

	// Handlers — Phase 1
	authHandler := handlers.NewAuthHandler(userService)
	deptHandler := handlers.NewDepartmentHandler(deptService)
	posHandler := handlers.NewPositionHandler(posService)
	empHandler := handlers.NewEmployeeHandler(empService)
	docHandler := handlers.NewEmployeeDocumentHandler(docService)
	ecHandler := handlers.NewEmergencyContactHandler(ecService)

	// Handlers — Phase 2
	ltHandler := handlers.NewLeaveTypeHandler(ltService)
	lbHandler := handlers.NewLeaveBalanceHandler(lbService, empService)
	lrHandler := handlers.NewLeaveRequestHandler(lrService, empService)
	holidayHandler := handlers.NewHolidayHandler(holidayService)
	attHandler := handlers.NewAttendanceHandler(attService, empService)

	// Dashboard
	dashboardService := services.NewDashboardService(empRepo, posRepo, deptRepo, lbRepo, lrRepo)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)

	// Workflow Handler
	workflowHandler := handlers.NewWorkflowHandler(workflowService)

	// Workflow Admin Handler
	workflowAdminHandler := handlers.NewWorkflowAdminHandler(workflowRepo)

	// Background jobs
	jobs.NewMonthlyLeaveAccrualJob(empRepo, lbRepo, ltRepo).Start()
	log.Println("Monthly leave accrual job scheduled")
	jobs.NewYearEndCarryForwardJob(lbRepo, ltRepo).Start()
	log.Println("Year-end carry-forward job scheduled")

	// Register routes
	routes.RegisterRoutes(
		authHandler, deptHandler, posHandler, empHandler, docHandler, ecHandler,
		ltHandler, lbHandler, lrHandler, attHandler, holidayHandler, dashboardHandler,
		workflowHandler, workflowAdminHandler,
	)

	// Apply CORS middleware globally to the default mux
	handler := middleware.CORS(http.DefaultServeMux)

	addr := ":" + cfg.ServerPort
	log.Printf("HR System running on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
