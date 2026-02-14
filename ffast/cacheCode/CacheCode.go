package cacheCode

import (
	"crypto/sha1"
	"fmt"
	"os"
	"strings"
)

type Link struct {
	URL string
	Mode string
}

// HashURL generates a SHA-1 hash of the given URL to use as a cache key.
func HashURL(url string) string {
	h := sha1.Sum([]byte(url))
	return fmt.Sprintf("%x", h)
}

// CacheExists checks if a cache file exists for the given key.
func CacheExists(key string) bool {
	_, err := os.Stat("cache/" + key + ".txt")
	return err == nil
}

// ReadCache reads the cached links from a file corresponding to the given key.
func ReadCache(key string) ([]Link, error) {
	data, err := os.ReadFile("cache/" + key + ".txt")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")

	links := []Link{}
	for _, l := range lines {
		parts := strings.Split(l," ")
		links = append(links, Link{parts[0],parts[1]})
	}
	return links, nil
}

// SaveCache saves the given links to a cache file corresponding to the given key.
func SaveCache(key string, links []Link) error {
	os.MkdirAll("cache", 0755)
	lines := []string{}
	for _, l := range links {
		ls := fmt.Sprintf("%s %s",l.URL,l.Mode)
		lines = append(lines, ls)
	}
	content := strings.Join(lines, "\n")
	return os.WriteFile("cache/"+key+".txt", []byte(content), 0644)
}

func SaveCurrentPasteBinURL(url string, name string) error {
	if _, err := os.Stat("cache"); os.IsNotExist(err) {
		err = os.Mkdir("cache", 0755)
		if err != nil {
			return err
		}
	}
	if _, err := os.Stat("cache/curl.txt"); err == nil {
		err = os.Remove("cache/curl.txt")
		if err != nil {
			return err
		}
	}
	saveString := fmt.Sprintf("%s\n%s", url, name)
	os.WriteFile("cache/curl.txt", []byte(saveString), 0644)
	return nil
}

func GetCurrentPasteBinURL() (string, string, error) {
	if _, err := os.Stat("cache/curl.txt"); os.IsNotExist(err) {
		return "", "", fmt.Errorf("current PasteBin URL not found in cache")
	}
	data, err := os.ReadFile("cache/curl.txt")
	if err != nil {
		return "", "", err
	}
	parts := strings.Split(string(data), "\n")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid cache format for current PasteBin URL")
	}
	return parts[0], parts[1], nil
}

func SaveDownloadStateCache(hash string, optionalChoices []int, selectiveChoices []int, mandatoryCompleted int, optionalCompleted int, selectiveCompleted int) error {
	if _, err := os.Stat("cache"); os.IsNotExist(err) {
		err = os.Mkdir("cache", 0755)
		if err != nil {
			return err
		}
	}
	cacheContent := fmt.Sprintf("%v\n%v\n%d\n%d\n%d", optionalChoices, selectiveChoices, mandatoryCompleted, optionalCompleted, selectiveCompleted)
	return os.WriteFile("cache/"+hash+"_state.txt", []byte(cacheContent), 0644)
}

func GetDownloadStateCache(hash string) ([]int, []int, int, int, int, error) {
	if _, err := os.Stat("cache/" + hash + "_state.txt"); os.IsNotExist(err) {
		return nil, nil, 0, 0, 0, fmt.Errorf("download state cache not found for hash: %s", hash)
	}
	data, err := os.ReadFile("cache/" + hash + "_state.txt")
	if err != nil {
		return nil, nil, 0, 0, 0, err
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) < 5 {
		return nil, nil, 0, 0, 0, fmt.Errorf("invalid download state cache format for hash: %s", hash)
	}
	optionalChoicesStr := strings.Trim(lines[0], "[]")
	selectiveChoicesStr := strings.Trim(lines[1], "[]")
	optionalChoices := []int{}
	selectiveChoices := []int{}
	if optionalChoicesStr != "" {
		for _, s := range strings.Split(optionalChoicesStr, " ") {
			var choice int
			fmt.Sscanf(s, "%d", &choice)
			optionalChoices = append(optionalChoices, choice)
		}
	}
	if selectiveChoicesStr != "" {
		for _, s := range strings.Split(selectiveChoicesStr, " ") {
			var choice int
			fmt.Sscanf(s, "%d", &choice)
			selectiveChoices = append(selectiveChoices, choice)
		}
	}
	mandatoryCompleted := 0
	fmt.Sscanf(lines[2], "%d", &mandatoryCompleted)
	optionalCompleted := 0
	fmt.Sscanf(lines[3], "%d", &optionalCompleted)
	selectiveCompleted := 0
	fmt.Sscanf(lines[4], "%d", &selectiveCompleted)
	return optionalChoices, selectiveChoices, mandatoryCompleted, optionalCompleted, selectiveCompleted, nil
}