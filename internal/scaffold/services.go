package scaffold

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/InitiatDev/initiat-cli/internal/setup"
)

type ServiceSteps struct {
	Provision   []setup.Step
	PostMessage string
}

func InferServiceSteps(dir string) (ServiceSteps, error) {
	if dir == "" {
		wd, _ := os.Getwd()
		dir = wd
	}
	var out ServiceSteps
	hasCompose := fileExists(dir, "docker-compose.yml") || fileExists(dir, "compose.yml")
	services := scanManifestsForServices(dir)
	if len(services.databases) == 0 && !services.redis && !hasCompose {
		return out, nil
	}
	if hasCompose {
		out.Provision = append(out.Provision, setup.Step{
			Name:  "Verify Docker Compose",
			If:    `!cmd_ok("docker compose version")`,
			Print: "Docker Compose is required. Install Docker Desktop or the docker compose plugin.",
		})
		out.Provision = append(out.Provision, setup.Step{
			Name: "Start services with Docker Compose",
			If:   `file_exists("docker-compose.yml") || file_exists("compose.yml")`,
			Run:  "docker compose up -d",
		})
		out.PostMessage = "Services can be managed with: docker compose up -d / docker compose down"
		return out, nil
	}
	for _, db := range services.databases {
		out.Provision = append(out.Provision, dbVerifyStep(db))
	}
	if services.redis {
		out.Provision = append(out.Provision, setup.Step{
			Name:  "Verify Redis",
			If:    `!cmd_ok("redis-cli --version")`,
			Print: "Redis is required. macOS: brew install redis. Linux: sudo apt-get install redis-server.",
		})
	}
	if len(services.databases) > 1 {
		out.PostMessage = "Multiple DBs detected. Remove the steps you do not use from .initiat/setup.yml."
	}
	return out, nil
}

func dbVerifyStep(db string) setup.Step {
	switch db {
	case "postgres", "postgresql":
		return setup.Step{
			Name:  "Verify PostgreSQL",
			If:    `!cmd_ok("psql --version")`,
			Print: "PostgreSQL required. macOS: brew install postgresql@15. Linux: sudo apt-get install postgresql.",
		}
	case "mysql":
		return setup.Step{
			Name:  "Verify MySQL",
			If:    `!cmd_ok("mysql --version")`,
			Print: "MySQL is required. macOS: brew install mysql. Linux: sudo apt-get install mysql-server. Start the service.",
		}
	case "sqlite":
		return setup.Step{
			Name:  "Verify SQLite",
			If:    `!cmd_ok("sqlite3 --version")`,
			Print: "SQLite3 is required. macOS: brew install sqlite. Linux: sudo apt-get install sqlite3.",
		}
	default:
		return setup.Step{Name: "DB " + db, Print: "Configure database: " + db}
	}
}

type inferredServices struct {
	databases []string
	redis     bool
}

func scanManifestsForServices(dir string) inferredServices {
	var s inferredServices
	seen := make(map[string]bool)
	addDB := func(name string) {
		if !seen[name] {
			seen[name] = true
			s.databases = append(s.databases, name)
		}
	}
	scanGemfile(dir, addDB, &s.redis)
	scanMixExs(dir, addDB, &s.redis)
	scanPackageJSON(dir, addDB, &s.redis)
	scanRequirementsTxt(dir, addDB, &s.redis)
	scanPyproject(dir, addDB, &s.redis)
	return s
}

func scanGemfile(dir string, addDB func(string), redis *bool) {
	// #nosec G304 -- path is project dir + constant manifest filename
	b, err := os.ReadFile(filepath.Join(dir, "Gemfile"))
	if err != nil {
		return
	}
	c := string(b)
	if regexp.MustCompile(`gem\s+['\"]pg['\"]`).MatchString(c) {
		addDB("postgres")
	}
	if regexp.MustCompile(`gem\s+['\"]mysql2['\"]`).MatchString(c) {
		addDB("mysql")
	}
	if regexp.MustCompile(`gem\s+['\"]sqlite3['\"]`).MatchString(c) {
		addDB("sqlite")
	}
	if regexp.MustCompile(`gem\s+['\"]redis['\"]`).MatchString(c) {
		*redis = true
	}
}

func scanMixExs(dir string, addDB func(string), redis *bool) {
	// #nosec G304 -- path is project dir + constant manifest filename
	b, err := os.ReadFile(filepath.Join(dir, "mix.exs"))
	if err != nil {
		return
	}
	c := string(b)
	if regexp.MustCompile(`:postgrex`).MatchString(c) {
		addDB("postgres")
	}
	if regexp.MustCompile(`:myxql`).MatchString(c) {
		addDB("mysql")
	}
	if regexp.MustCompile(`:ecto_sqlite3`).MatchString(c) {
		addDB("sqlite")
	}
	if regexp.MustCompile(`:redix`).MatchString(c) {
		*redis = true
	}
}

func scanPackageJSON(dir string, addDB func(string), redis *bool) {
	// #nosec G304 -- path is project dir + constant manifest filename
	b, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return
	}
	c := string(b)
	if regexp.MustCompile(`"pg"`).MatchString(c) {
		addDB("postgres")
	}
	if regexp.MustCompile(`"mysql2"`).MatchString(c) {
		addDB("mysql")
	}
	if regexp.MustCompile(`"sqlite3"|"better-sqlite3"`).MatchString(c) {
		addDB("sqlite")
	}
	if regexp.MustCompile(`"redis"|"ioredis"`).MatchString(c) {
		*redis = true
	}
}

func scanRequirementsTxt(dir string, addDB func(string), redis *bool) {
	// #nosec G304 -- path is project dir + constant manifest filename
	b, err := os.ReadFile(filepath.Join(dir, "requirements.txt"))
	if err != nil {
		return
	}
	c := string(b)
	if regexp.MustCompile(`psycopg|asyncpg`).MatchString(c) {
		addDB("postgres")
	}
	if regexp.MustCompile(`pymysql|mysqlclient|PyMySQL`).MatchString(c) {
		addDB("mysql")
	}
	if regexp.MustCompile(`redis`).MatchString(c) {
		*redis = true
	}
}

func scanPyproject(dir string, addDB func(string), redis *bool) {
	// #nosec G304 -- path is project dir + constant manifest filename
	b, err := os.ReadFile(filepath.Join(dir, "pyproject.toml"))
	if err != nil {
		return
	}
	c := string(b)
	if regexp.MustCompile(`psycopg|asyncpg`).MatchString(c) {
		addDB("postgres")
	}
	if regexp.MustCompile(`pymysql|mysqlclient|PyMySQL`).MatchString(c) {
		addDB("mysql")
	}
	if regexp.MustCompile(`redis`).MatchString(c) {
		*redis = true
	}
}
