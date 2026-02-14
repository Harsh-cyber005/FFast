package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"FFast/ffast"
	"FFast/ffast/cacheCode"
	"strings"
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

func main() {
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
