package fdmconfig

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
	_ "modernc.org/sqlite"
)

func updateFDMPath(db *sql.DB, newPath string) error {
	absPath, err := filepath.Abs(newPath)
	if err != nil {
		return fmt.Errorf("invalid path format: %v", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Printf("Folder does not exist. Creating: %s\n", absPath)
		err := os.MkdirAll(absPath, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
	}

	queries := []struct {
		name  string
		value string
	}{
		{"FixedDownloadPath", absPath},
		{"DownloadPathDependsOnFileType", "0"},
	}

	for _, q := range queries {
		_, err := db.Exec("INSERT OR REPLACE INTO Settings (Name, Value) VALUES (?, ?)", q.name, q.value)
		if err != nil {
			return err
		}
	}
	return nil
}

func getOrUpdatePath(db *sql.DB) string {
	var existingPath string
	reader := bufio.NewReader(os.Stdin)

	finalPath := ""

	err := db.QueryRow("SELECT Value FROM Settings WHERE Name = 'FixedDownloadPath'").Scan(&existingPath)

	if err != nil {
		fmt.Println("No custom download path found in FDM settings.")
		fmt.Print("Enter the new download location: ")
		inputPath, _ := reader.ReadString('\n')
		inputPath = strings.TrimSpace(inputPath)

		finalPath = inputPath

		err := updateFDMPath(db, inputPath)
		if err != nil {
			log.Fatalf("Update failed: %v", err)
		}
		fmt.Println("Path initialized successfully.")
	} else {
		fmt.Printf("Existing FDM Path: %s\n", existingPath)
		fmt.Print("Would you like to change it? (y/N): ")
		choice, _ := reader.ReadString('\n')
		choice = strings.ToLower(strings.TrimSpace(choice))

		if choice == "y" {
			fmt.Print("Enter the new location: ")
			inputPath, _ := reader.ReadString('\n')
			inputPath = strings.TrimSpace(inputPath)

			finalPath = inputPath

			err := updateFDMPath(db, inputPath)
			if err != nil {
				log.Fatalf("Update failed: %v", err)
			}
			fmt.Println("Path updated successfully.")
			restartFDM()
		} else {
			fmt.Println("Keeping existing path.")
			finalPath = existingPath
		}
	}
	if finalPath[len(finalPath)-1] != '\\' {
		finalPath += "\\"
	}
	return finalPath
}

func getFDMExecutablePath() string {
	keyPath := `Software\Microsoft\Windows\CurrentVersion\Uninstall`
	key, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
	if err == nil {
		defer key.Close()
		subkeys, _ := key.ReadSubKeyNames(-1)
		for _, sk := range subkeys {
			s, err := registry.OpenKey(registry.CURRENT_USER, keyPath+`\`+sk, registry.QUERY_VALUE)
			if err != nil {
				continue
			}
			name, _, _ := s.GetStringValue("DisplayName")
			if strings.Contains(name, "Free Download Manager") {
				loc, _, _ := s.GetStringValue("InstallLocation")
				s.Close()
				if loc != "" {
					return filepath.Join(loc, "fdm.exe")
				}
			}
			s.Close()
		}
	}

	userPath := filepath.Join(os.Getenv("LOCALAPPDATA"), `Programs\Softdeluxe\Free Download Manager\fdm.exe`)
	if _, err := os.Stat(userPath); err == nil {
		return userPath
	}

	return `C:\Program Files\Softdeluxe\Free Download Manager\fdm.exe`
}

func restartFDM() {
	fmt.Println("Restarting FDM in background...")

	exec.Command("taskkill", "/F", "/IM", "fdm.exe", "/T").Run()
	fdmPath := getFDMExecutablePath()
	fmt.Println(fdmPath)
	cmd := exec.Command(fdmPath, "--hidden")
	
	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to restart FDM: %v\n", err)
	} else {
		fmt.Println("FDM is back up and running!")
	}
}

func StartConfig() string {
	appData := os.Getenv("LOCALAPPDATA")
	dbPath := filepath.Join(appData, "Softdeluxe", "Free Download Manager", "db.sqlite")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	return getOrUpdatePath(db)
}