package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		env("DB_HOST", "localhost"),
		env("DB_PORT", "5432"),
		env("DB_USER", "postgres"),
		env("DB_PASSWORD", ""),
		env("DB_NAME", "hr_system"),
		env("DB_SSLMODE", "disable"),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	log.Println("Connected to database")

	// ------------------------------------------------------------------ roles
	log.Println("Fetching roles...")
	roles := map[string]uuid.UUID{}
	rows, err := db.Query(`SELECT role_id, name FROM roles`)
	if err != nil {
		log.Fatalf("query roles: %v", err)
	}
	for rows.Next() {
		var id uuid.UUID
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			log.Fatalf("scan role: %v", err)
		}
		roles[name] = id
	}
	rows.Close()
	if len(roles) == 0 {
		log.Fatal("No roles found — run `make migrate-up` first")
	}
	log.Printf("  Found %d roles\n", len(roles))

	// ----------------------------------------------------------- departments
	log.Println("Seeding departments...")
	depts := map[string]uuid.UUID{}

	topDepts := []struct{ name, code, desc string }{
		{"Engineering",        "ENG",  "Software engineering and product development"},
		{"Human Resources",    "HR",   "People operations and HR management"},
		{"Finance",            "FIN",  "Finance and accounting"},
		{"Sales & Marketing",  "SAL",  "Revenue generation and brand marketing"},
		{"Operations",         "OPS",  "Business operations and logistics"},
	}
	for _, d := range topDepts {
		id := insertDept(db, d.name, d.code, d.desc, nil)
		depts[d.code] = id
	}

	// Sub-departments
	subDepts := []struct {
		name, code, desc string
		parent           string
	}{
		{"Backend",         "ENG-BE",  "Backend services and APIs",        "ENG"},
		{"Frontend",        "ENG-FE",  "Web and mobile frontend",           "ENG"},
		{"DevOps",          "ENG-DO",  "Infrastructure and CI/CD",          "ENG"},
		{"Talent Acquisition", "HR-TA", "Recruiting and onboarding",       "HR"},
		{"Payroll",         "HR-PAY",  "Payroll processing",                "HR"},
		{"Accounting",      "FIN-ACC", "Bookkeeping and accounting",        "FIN"},
		{"Treasury",        "FIN-TRE", "Cash management and treasury",      "FIN"},
		{"Sales",           "SAL-SLS", "Direct and indirect sales",         "SAL"},
		{"Marketing",       "SAL-MKT", "Brand and digital marketing",       "SAL"},
	}
	for _, d := range subDepts {
		pid := depts[d.parent]
		id := insertDept(db, d.name, d.code, d.desc, &pid)
		depts[d.code] = id
	}
	log.Printf("  Inserted %d departments\n", len(depts))

	// ------------------------------------------------------------- positions
	log.Println("Seeding positions...")
	pos := map[string]uuid.UUID{}

	positions := []struct {
		title, code string
		dept        string
		grade       string
		min, max    float64
	}{
		// Engineering
		{"Software Engineer I",        "SWE-1",   "ENG-BE", "L1", 50000, 70000},
		{"Software Engineer II",       "SWE-2",   "ENG-BE", "L2", 70000, 95000},
		{"Senior Software Engineer",   "SWE-SR",  "ENG-BE", "L3", 95000, 130000},
		{"Staff Engineer",             "SWE-ST",  "ENG-BE", "L4", 130000, 170000},
		{"Frontend Engineer",          "FE-1",    "ENG-FE", "L2", 65000, 90000},
		{"Senior Frontend Engineer",   "FE-SR",   "ENG-FE", "L3", 90000, 125000},
		{"DevOps Engineer",            "DO-1",    "ENG-DO", "L2", 70000, 100000},
		{"Engineering Manager",        "ENG-MGR", "ENG",    "M1", 140000, 180000},
		// HR
		{"HR Generalist",              "HR-GEN",  "HR-TA",  "L2", 45000, 65000},
		{"Talent Acquisition Specialist", "HR-TAS", "HR-TA","L2", 50000, 70000},
		{"HR Manager",                 "HR-MGR",  "HR",     "M1", 80000, 110000},
		{"Payroll Specialist",         "PAY-SP",  "HR-PAY", "L2", 50000, 70000},
		// Finance
		{"Accountant",                 "FIN-ACC1","FIN-ACC", "L2", 55000, 75000},
		{"Senior Accountant",          "FIN-ACC2","FIN-ACC", "L3", 75000, 100000},
		{"Finance Manager",            "FIN-MGR", "FIN",     "M1", 100000, 140000},
		// Sales & Marketing
		{"Sales Representative",       "SAL-REP", "SAL-SLS", "L1", 40000, 65000},
		{"Account Executive",          "SAL-AE",  "SAL-SLS", "L2", 60000, 90000},
		{"Marketing Specialist",       "MKT-SP",  "SAL-MKT", "L2", 50000, 75000},
		{"Sales Manager",              "SAL-MGR", "SAL",     "M1", 90000, 130000},
		// Operations
		{"Operations Analyst",         "OPS-AN",  "OPS",    "L2", 50000, 70000},
		{"Operations Manager",         "OPS-MGR", "OPS",    "M1", 85000, 120000},
	}
	for _, p := range positions {
		id := insertPosition(db, p.title, p.code, depts[p.dept], p.grade, p.min, p.max)
		pos[p.code] = id
	}
	log.Printf("  Inserted %d positions\n", len(pos))

	// --------------------------------------------------------------- users
	log.Println("Seeding users...")

	// Grab the existing super admin so we can use their user_id as uploaded_by later
	var adminUserID uuid.UUID
	err = db.QueryRow(`SELECT user_id FROM users WHERE email = 'admin@hr-system.com'`).Scan(&adminUserID)
	if err != nil {
		// If not found, insert it
		adminUserID = insertUser(db, "admin@hr-system.com", "Admin@123", roles["super_admin"])
		log.Println("  Created super admin user")
	} else {
		log.Println("  Super admin already exists — skipping")
	}

	users := []struct {
		email    string
		password string
		role     string
	}{
		{"hr.manager@hr-system.com",    "HrManager@123",    "hr_manager"},
		{"eng.manager@hr-system.com",   "Manager@123",      "manager"},
		{"fin.manager@hr-system.com",   "Manager@123",      "manager"},
		{"sales.manager@hr-system.com", "Manager@123",      "manager"},
		{"ops.manager@hr-system.com",   "Manager@123",      "manager"},
		{"alice.smith@hr-system.com",   "Employee@123",     "employee"},
		{"bob.jones@hr-system.com",     "Employee@123",     "employee"},
		{"carol.white@hr-system.com",   "Employee@123",     "employee"},
		{"david.brown@hr-system.com",   "Employee@123",     "employee"},
		{"eve.davis@hr-system.com",     "Employee@123",     "employee"},
		{"frank.miller@hr-system.com",  "Employee@123",     "employee"},
		{"grace.wilson@hr-system.com",  "Employee@123",     "employee"},
		{"henry.moore@hr-system.com",   "Employee@123",     "employee"},
		{"iris.taylor@hr-system.com",   "Employee@123",     "employee"},
		{"jack.anderson@hr-system.com", "Employee@123",     "employee"},
	}
	userIDs := map[string]uuid.UUID{}
	for _, u := range users {
		userIDs[u.email] = insertUser(db, u.email, u.password, roles[u.role])
	}
	log.Printf("  Inserted %d users\n", len(userIDs))

	// ------------------------------------------------------------- employees
	log.Println("Seeding employees...")

	// Insert managers first (no manager_id dependency)
	managers := []struct {
		email, num, first, last, phone, gender, nat, city, country string
		dob                                                          time.Time
		dept, posCode                                                string
		userEmail                                                    string
		hire                                                         time.Time
	}{
		{
			"hr.manager.emp@hr-system.com", "EMP-001", "Sarah", "Connor", "+12025550101",
			"female", "NAT-001", "New York", "USA",
			date(1982, 3, 14), "HR", "HR-MGR", "hr.manager@hr-system.com",
			date(2018, 1, 10),
		},
		{
			"eng.manager.emp@hr-system.com", "EMP-002", "James", "Kirk", "+12025550102",
			"male", "NAT-002", "San Francisco", "USA",
			date(1980, 7, 22), "ENG", "ENG-MGR", "eng.manager@hr-system.com",
			date(2017, 3, 1),
		},
		{
			"fin.manager.emp@hr-system.com", "EMP-003", "Laura", "Palmer", "+12025550103",
			"female", "NAT-003", "Chicago", "USA",
			date(1979, 11, 5), "FIN", "FIN-MGR", "fin.manager@hr-system.com",
			date(2016, 6, 15),
		},
		{
			"sales.manager.emp@hr-system.com", "EMP-004", "Tony", "Stark", "+12025550104",
			"male", "NAT-004", "Austin", "USA",
			date(1984, 5, 29), "SAL", "SAL-MGR", "sales.manager@hr-system.com",
			date(2019, 2, 1),
		},
		{
			"ops.manager.emp@hr-system.com", "EMP-005", "Natasha", "Romanova", "+12025550105",
			"female", "NAT-005", "Boston", "USA",
			date(1986, 9, 18), "OPS", "OPS-MGR", "ops.manager@hr-system.com",
			date(2020, 4, 1),
		},
	}

	mgrEmpIDs := map[string]uuid.UUID{} // dept code -> employee id
	for _, m := range managers {
		uid := userIDs[m.userEmail]
		id := insertEmployee(db, insertEmployeeArgs{
			userID:     &uid,
			num:        m.num,
			first:      m.first,
			last:       m.last,
			email:      m.email,
			phone:      m.phone,
			dob:        m.dob,
			gender:     m.gender,
			natID:      m.nat,
			city:       m.city,
			country:    m.country,
			deptID:     depts[m.dept],
			posID:      pos[m.posCode],
			managerID:  nil,
			hire:       m.hire,
			empType:    "full_time",
			empStatus:  "active",
		})
		mgrEmpIDs[m.dept] = id
	}

	// Regular employees
	type empRow struct {
		num, first, last, email, phone, gender, nat, city, country string
		dob                                                          time.Time
		dept, posCode, mgrDept, userEmail                            string
		hire                                                         time.Time
		empType, empStatus                                           string
	}
	emps := []empRow{
		// Backend
		{"EMP-006", "Alice", "Smith", "alice.smith.emp@hr-system.com", "+12025550106", "female", "NAT-006", "New York", "USA", date(1994, 2, 14), "ENG-BE", "SWE-2", "ENG", "alice.smith@hr-system.com", date(2021, 3, 1), "full_time", "active"},
		{"EMP-007", "Bob", "Jones", "bob.jones.emp@hr-system.com", "+12025550107", "male", "NAT-007", "New York", "USA", date(1992, 6, 30), "ENG-BE", "SWE-SR", "ENG", "bob.jones@hr-system.com", date(2020, 5, 15), "full_time", "active"},
		{"EMP-008", "Carol", "White", "carol.white.emp@hr-system.com", "+12025550108", "female", "NAT-008", "Remote", "USA", date(1996, 10, 3), "ENG-BE", "SWE-1", "ENG", "carol.white@hr-system.com", date(2022, 8, 1), "full_time", "active"},
		// Frontend
		{"EMP-009", "David", "Brown", "david.brown.emp@hr-system.com", "+12025550109", "male", "NAT-009", "San Francisco", "USA", date(1993, 4, 20), "ENG-FE", "FE-SR", "ENG", "david.brown@hr-system.com", date(2021, 1, 10), "full_time", "active"},
		{"EMP-010", "Eve", "Davis", "eve.davis.emp@hr-system.com", "+12025550110", "female", "NAT-010", "Seattle", "USA", date(1997, 8, 8), "ENG-FE", "FE-1", "ENG", "eve.davis@hr-system.com", date(2023, 2, 1), "full_time", "active"},
		// HR
		{"EMP-011", "Frank", "Miller", "frank.miller.emp@hr-system.com", "+12025550111", "male", "NAT-011", "New York", "USA", date(1990, 12, 25), "HR-TA", "HR-TAS", "HR", "frank.miller@hr-system.com", date(2019, 7, 1), "full_time", "active"},
		{"EMP-012", "Grace", "Wilson", "grace.wilson.emp@hr-system.com", "+12025550112", "female", "NAT-012", "Chicago", "USA", date(1995, 3, 17), "HR-PAY", "PAY-SP", "HR", "grace.wilson@hr-system.com", date(2021, 6, 1), "full_time", "active"},
		// Finance
		{"EMP-013", "Henry", "Moore", "henry.moore.emp@hr-system.com", "+12025550113", "male", "NAT-013", "Chicago", "USA", date(1988, 7, 4), "FIN-ACC", "FIN-ACC2", "FIN", "henry.moore@hr-system.com", date(2018, 9, 1), "full_time", "active"},
		// Sales
		{"EMP-014", "Iris", "Taylor", "iris.taylor.emp@hr-system.com", "+12025550114", "female", "NAT-014", "Austin", "USA", date(1998, 1, 29), "SAL-SLS", "SAL-AE", "SAL", "iris.taylor@hr-system.com", date(2022, 3, 15), "full_time", "active"},
		{"EMP-015", "Jack", "Anderson", "jack.anderson.emp@hr-system.com", "+12025550115", "male", "NAT-015", "Dallas", "USA", date(1991, 5, 11), "SAL-MKT", "MKT-SP", "SAL", "jack.anderson@hr-system.com", date(2020, 11, 1), "full_time", "on_leave"},
	}

	empIDs := map[string]uuid.UUID{} // employee_number -> id
	for _, e := range emps {
		uid := userIDs[e.userEmail]
		mgrID := mgrEmpIDs[e.mgrDept]
		id := insertEmployee(db, insertEmployeeArgs{
			userID:    &uid,
			num:       e.num,
			first:     e.first,
			last:      e.last,
			email:     e.email,
			phone:     e.phone,
			dob:       e.dob,
			gender:    e.gender,
			natID:     e.nat,
			city:      e.city,
			country:   e.country,
			deptID:    depts[e.dept],
			posID:     pos[e.posCode],
			managerID: &mgrID,
			hire:      e.hire,
			empType:   e.empType,
			empStatus: e.empStatus,
		})
		empIDs[e.num] = id
	}

	// Merge all employee IDs
	for dept, id := range mgrEmpIDs {
		_ = dept
		_ = id
	}
	allEmpIDs := map[string]uuid.UUID{}
	for k, v := range mgrEmpIDs {
		allEmpIDs["mgr-"+k] = v
	}
	for k, v := range empIDs {
		allEmpIDs[k] = v
	}
	log.Printf("  Inserted %d employees\n", len(allEmpIDs))

	// ---------------------------------------------------- emergency contacts
	log.Println("Seeding emergency contacts...")
	contacts := []struct {
		empNum, name, rel, phone, email string
	}{
		{"EMP-006", "Michael Smith",   "Spouse",  "+12025550201", "michael.smith@gmail.com"},
		{"EMP-007", "Linda Jones",     "Mother",  "+12025550202", "linda.jones@gmail.com"},
		{"EMP-008", "Peter White",     "Father",  "+12025550203", "peter.white@gmail.com"},
		{"EMP-009", "Emma Brown",      "Spouse",  "+12025550204", "emma.brown@gmail.com"},
		{"EMP-010", "Chris Davis",     "Brother", "+12025550205", "chris.davis@gmail.com"},
		{"EMP-011", "Anna Miller",     "Spouse",  "+12025550206", "anna.miller@gmail.com"},
		{"EMP-012", "Tom Wilson",      "Father",  "+12025550207", "tom.wilson@gmail.com"},
		{"EMP-013", "Sandra Moore",    "Spouse",  "+12025550208", "sandra.moore@gmail.com"},
		{"EMP-014", "Kevin Taylor",    "Brother", "+12025550209", "kevin.taylor@gmail.com"},
		{"EMP-015", "Olivia Anderson", "Spouse",  "+12025550210", "olivia.anderson@gmail.com"},
	}
	for _, c := range contacts {
		eid := empIDs[c.empNum]
		insertContact(db, eid, c.name, c.rel, c.phone, c.email)
	}
	log.Printf("  Inserted %d emergency contacts\n", len(contacts))

	// ------------------------------------------------------- documents
	log.Println("Seeding employee documents...")
	docs := []struct {
		empNum, docType, title, fileName string
		verified                          bool
	}{
		{"EMP-006", "contract",     "Employment Contract 2021", "contract_alice_2021.pdf", true},
		{"EMP-006", "id_document",  "National ID",             "id_alice.pdf",             true},
		{"EMP-007", "contract",     "Employment Contract 2020", "contract_bob_2020.pdf",   true},
		{"EMP-007", "certification","AWS Solutions Architect",  "aws_cert_bob.pdf",        true},
		{"EMP-008", "contract",     "Employment Contract 2022", "contract_carol_2022.pdf", false},
		{"EMP-008", "offer_letter", "Offer Letter 2022",        "offer_carol_2022.pdf",    true},
		{"EMP-009", "contract",     "Employment Contract 2021", "contract_david_2021.pdf", true},
		{"EMP-010", "contract",     "Employment Contract 2023", "contract_eve_2023.pdf",   false},
		{"EMP-011", "contract",     "Employment Contract 2019", "contract_frank_2019.pdf", true},
		{"EMP-013", "contract",     "Employment Contract 2018", "contract_henry_2018.pdf", true},
	}
	for _, d := range docs {
		eid := empIDs[d.empNum]
		var verifiedBy *uuid.UUID
		var verifiedAt *time.Time
		if d.verified {
			verifiedBy = &adminUserID
			t := time.Now().Add(-24 * time.Hour)
			verifiedAt = &t
		}
		insertDocument(db, eid, d.docType, d.title, d.fileName, adminUserID, verifiedBy, verifiedAt)
	}
	log.Printf("  Inserted %d documents\n", len(docs))

	log.Println("Seed complete!")
	printCredentials()
}

// ------------------------------------------------------------------ helpers

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func date(y, m, d int) time.Time {
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

func hashpw(pw string) string {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), 10)
	if err != nil {
		log.Fatalf("bcrypt: %v", err)
	}
	return string(b)
}

func insertDept(db *sql.DB, name, code, desc string, parentID *uuid.UUID) uuid.UUID {
	var id uuid.UUID
	err := db.QueryRow(`
		INSERT INTO departments (name, code, description, parent_department_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`,
		name, code, desc, parentID,
	).Scan(&id)
	if err != nil {
		log.Fatalf("insertDept %s: %v", code, err)
	}
	return id
}

func insertPosition(db *sql.DB, title, code string, deptID uuid.UUID, grade string, minS, maxS float64) uuid.UUID {
	var id uuid.UUID
	err := db.QueryRow(`
		INSERT INTO positions (title, code, department_id, grade_level, min_salary, max_salary)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (code) DO UPDATE SET title = EXCLUDED.title
		RETURNING id`,
		title, code, deptID, grade, minS, maxS,
	).Scan(&id)
	if err != nil {
		log.Fatalf("insertPosition %s: %v", code, err)
	}
	return id
}

func insertUser(db *sql.DB, email, password string, roleID uuid.UUID) uuid.UUID {
	var id uuid.UUID
	err := db.QueryRow(`
		INSERT INTO users (email, password, role_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO UPDATE SET role_id = EXCLUDED.role_id
		RETURNING user_id`,
		email, hashpw(password), roleID,
	).Scan(&id)
	if err != nil {
		log.Fatalf("insertUser %s: %v", email, err)
	}
	return id
}

type insertEmployeeArgs struct {
	userID                                                    *uuid.UUID
	num, first, last, email, phone, gender, natID, city, country string
	dob                                                           time.Time
	deptID, posID                                                 uuid.UUID
	managerID                                                     *uuid.UUID
	hire                                                          time.Time
	empType, empStatus                                            string
}

func insertEmployee(db *sql.DB, a insertEmployeeArgs) uuid.UUID {
	var id uuid.UUID
	err := db.QueryRow(`
		INSERT INTO employees (
			user_id, employee_number, first_name, last_name, email,
			phone, date_of_birth, gender, national_id,
			city, country, department_id, position_id, manager_id,
			hire_date, employment_type, employment_status
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		ON CONFLICT (employee_number) DO UPDATE SET first_name = EXCLUDED.first_name
		RETURNING id`,
		a.userID, a.num, a.first, a.last, a.email,
		a.phone, a.dob, a.gender, a.natID,
		a.city, a.country, a.deptID, a.posID, a.managerID,
		a.hire, a.empType, a.empStatus,
	).Scan(&id)
	if err != nil {
		log.Fatalf("insertEmployee %s: %v", a.num, err)
	}
	return id
}

func insertContact(db *sql.DB, empID uuid.UUID, name, rel, phone, email string) {
	_, err := db.Exec(`
		INSERT INTO emergency_contacts (employee_id, name, relationship, phone, email)
		VALUES ($1, $2, $3, $4, $5)`,
		empID, name, rel, phone, email,
	)
	if err != nil {
		log.Fatalf("insertContact %s: %v", name, err)
	}
}

func insertDocument(db *sql.DB, empID uuid.UUID, docType, title, fileName string, uploadedBy uuid.UUID, verifiedBy *uuid.UUID, verifiedAt *time.Time) {
	isVerified := verifiedBy != nil
	_, err := db.Exec(`
		INSERT INTO employee_documents (
			employee_id, document_type, title, file_url, file_name,
			file_size, mime_type, uploaded_by,
			is_verified, verified_by, verified_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		empID, docType, title,
		"https://storage.hr-system.local/docs/"+fileName, fileName,
		102400, "application/pdf", uploadedBy,
		isVerified, verifiedBy, verifiedAt,
	)
	if err != nil {
		log.Fatalf("insertDocument %s: %v", fileName, err)
	}
}

func printCredentials() {
	fmt.Println()
	fmt.Println("======================================")
	fmt.Println("  Seed credentials")
	fmt.Println("======================================")
	fmt.Println("  Super Admin:  admin@hr-system.com        / Admin@123")
	fmt.Println("  HR Manager:   hr.manager@hr-system.com   / HrManager@123")
	fmt.Println("  Eng Manager:  eng.manager@hr-system.com  / Manager@123")
	fmt.Println("  Fin Manager:  fin.manager@hr-system.com  / Manager@123")
	fmt.Println("  Sales Mgr:    sales.manager@hr-system.com/ Manager@123")
	fmt.Println("  Ops Manager:  ops.manager@hr-system.com  / Manager@123")
	fmt.Println("  Employees:    alice/bob/carol/david/eve   / Employee@123")
	fmt.Println("                frank/grace/henry/iris/jack / Employee@123")
	fmt.Println("======================================")
}
