package main

import (
	"fmt"
	"log"
	"net/http"

	"hr-system/internal/config"
	"hr-system/internal/database"
	"hr-system/internal/handlers"
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

	// Repositories
	userRepo := repository.NewUserRepository()
	roleRepo := repository.NewRoleRepository()
	deptRepo := repository.NewDepartmentRepository()
	posRepo := repository.NewPositionRepository()
	empRepo := repository.NewEmployeeRepository()
	docRepo := repository.NewEmployeeDocumentRepository()
	ecRepo := repository.NewEmergencyContactRepository()

	// Services
	roleService := services.NewRoleService(roleRepo)
	userService := services.NewUserService(userRepo, roleRepo)
	deptService := services.NewDepartmentService(deptRepo)
	posService := services.NewPositionService(posRepo, deptRepo)
	empService := services.NewEmployeeService(empRepo, deptRepo, posRepo)
	docService := services.NewEmployeeDocumentService(docRepo, empRepo)
	ecService := services.NewEmergencyContactService(ecRepo, empRepo)

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

	// Handlers
	authHandler := handlers.NewAuthHandler(userService)
	deptHandler := handlers.NewDepartmentHandler(deptService)
	posHandler := handlers.NewPositionHandler(posService)
	empHandler := handlers.NewEmployeeHandler(empService)
	docHandler := handlers.NewEmployeeDocumentHandler(docService)
	ecHandler := handlers.NewEmergencyContactHandler(ecService)

	// Routes
	routes.RegisterRoutes(authHandler, deptHandler, posHandler, empHandler, docHandler, ecHandler)

	addr := ":" + cfg.ServerPort
	log.Printf("HR System running on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
