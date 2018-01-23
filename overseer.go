package overseer

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
	ps "github.com/mitchellh/go-ps"
)

type Service struct {
	ID        int
	Name      string
	ModelID   int
	Port      int
	ProcessID int
}

func EnsureRunning(path string) {
	db, err := sql.Open("sqlite3", filepath.Join(path, "/var/master/db/steam.db"))
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()
	for _, service := range getServices(db) {
		jettyPath := filepath.Join(path, "var/master/assets/jetty-runner.jar")
		warPath, err := findWar(filepath.Join(path, "var/master/model/", strconv.Itoa(service.ModelID)))
		if err != nil {
			continue
		}
		log.Printf("Starting %s", service.Name)
		cmd := exec.Command("screen", "-d", "-m", "-S", strconv.Itoa(service.ID), "java", "-jar", jettyPath, "--port", strconv.Itoa(service.Port), warPath)
		err = cmd.Start()
		if err != nil {
			log.Fatalln(err)
		}
		_, err = db.Exec("UPDATE service SET process_id = ? WHERE id = ?", getProcessID(service.ID), service.ID)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func getServices(db *sql.DB) []Service {
	rows, err := db.Query("SELECT id, name, model_id, port, process_id FROM service")
	if err != nil {
		log.Fatalln(err)
	}
	defer rows.Close()
	var services []Service
	for rows.Next() {
		service := Service{}
		err = rows.Scan(&service.ID, &service.Name, &service.ModelID, &service.Port, &service.ProcessID)
		if err != nil {
			log.Fatalln(err)
		}
		process, err := ps.FindProcess(service.ProcessID)
		if err != nil {
			log.Fatalln(err)
		}
		if process != nil {
			log.Printf("Process %s already running, skipping", service.Name)
			continue
		}
		services = append(services, service)
	}
	return services
}

func findWar(path string) (string, error) {
	var war string
	filepath.Walk(path, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			r, err := regexp.MatchString(".war", f.Name())
			if err == nil && r {
				war = f.Name()
			}
		}
		return nil
	})
	if war == "" {
		return "", fmt.Errorf("War not found in %s", path)
	}
	return filepath.Join(path, war), nil
}

func getProcessID(id int) int {
	time.Sleep(time.Second * 5)
	cmd := exec.Command("screen", "-ls", strconv.Itoa(id))
	var output bytes.Buffer
	cmd.Stdout = &output
	err := cmd.Run()
	if err != nil {
		return 0
	}
	cmd.Wait()
	re := regexp.MustCompile(`(\d+).` + strconv.Itoa(id))
	result := re.FindAllStringSubmatch(output.String(), -1)
	if len(result) > 0 && len(result[0]) > 1 {
		ID, _ := strconv.Atoi(result[0][1])
		return ID
	}
	return 0
}
