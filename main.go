package main

import (
	"FFast/ffast"
	"FFast/ffast/cacheCode"
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

// Take input from the Interactive Terminal
func takeInput(prompt string, placeholder string, defaultValue string, optional bool) string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print(prompt)
	
	url, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("Could not read the Input")
	}

	input := strings.TrimSpace(url)
	if !optional && input == "" {
		log.Fatal("[ERROR] ",placeholder," cannot be empty")
	}
	if input == "" {
		input = defaultValue
	}
	return input
}

func start() {
	URL := ""
	name := ""
	cURL,cName, err := cacheCode.GetCurrentPasteBinURL()
	if err != nil {
		fmt.Println("No current PasteBin URL found in cache.")
		URL = takeInput("Enter the Fucking-Fast URL: ", "URL", "", false)
		name = takeInput("Enter the name for the download: ", "Name", "", true)
		cacheCode.SaveCurrentPasteBinURL(URL, name)
	} else {
		fmt.Println("Current PasteBin Cache Found : ")
		fmt.Println("Name: ",cName)
		fmt.Println("URL: ", cURL)
		useCache := takeInput("Do you want to use the cached URL? (Y/n): ", "Use Cache", "Y", true)
		if strings.ToUpper(useCache) == "Y" {
			URL = cURL
			name = cName
			cacheCode.SaveCurrentPasteBinURL(URL, name)
		} else {
			URL = takeInput("Enter the Fucking-Fast URL: ", "URL", "", false)
			name = takeInput("Enter the name for the download: ", "Name", "", true)
			cacheCode.SaveCurrentPasteBinURL(URL, name)
		}
	}
	
	ff := ffast.Create(URL,name)
	ff.DecodePrivateBin()
	ff.SelectLinks()
	ff.DownloadParts()
}

func waitForExit(wg *sync.WaitGroup) {
	fmt.Println("Press Enter to exit...")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
	wg.Done()
}

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("Usage: ffast [options], where options can be:")
		fmt.Println("  start, s, --start, -s, --run\t\tStart the interactive terminal")
		fmt.Println("  --clear-cache, -c\tClear the cached PasteBin URL")
		fmt.Println("  --help, -h\t\tShow this help message")
		wg := sync.WaitGroup{}
		wg.Add(1)
		go waitForExit(&wg)
		wg.Wait()
		return
	}
	switch args[1] {
	case "--help", "-h":
		fmt.Println("Usage: ffast [options], where options can be:")
		fmt.Println("  --start, -s\t\tStart the interactive terminal")
		fmt.Println("  --clear-cache, -c\tClear the cached PasteBin URL")
		fmt.Println("  --help, -h\t\tShow this help message")
	case "--start", "-s", "start", "s", "run":
		start()
	case "--clear-cache", "-c":
		err := cacheCode.ClearDownloadStateCache()
		if err != nil {
			log.Fatal("Failed to clear cache:", err)
		}
		fmt.Println("Cache cleared successfully.")
	default:
		fmt.Println("Unknown option:", args[1])
		fmt.Println("Use --help or -h to see the available options.")
	}
}
