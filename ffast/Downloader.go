package ffast

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"FFast/ffast/cacheCode"
	fdmconfig "FFast/ffast/fdmConfig"
	"regexp"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/chromedp/chromedp"
)

type fFastDownloader struct {
	id string
	name string
	url string
	allLinks []cacheCode.Link
	mandatoryLinks []string
	optionalLinks []string
	selectiveLinks []string
	optionalChoices []int
	selectiveChoices []int
	mandatoryCompleted int
	optionalCompleted int
	selectiveCompleted int
	fdmDownloadPath string
}

// Create a FFast Downloader Object
func Create(url string, name string) *fFastDownloader {
	fdmPath := fdmconfig.StartConfig()
	ff := fFastDownloader{ url: url, name: name, fdmDownloadPath: fdmPath }
	optionalChoices, selectiveChoices, mandatoryCompleted, optionalCompleted, selectiveCompleted, err := cacheCode.GetDownloadStateCache(cacheCode.HashURL(url))
	if err == nil {
		ff.optionalChoices = optionalChoices
		ff.selectiveChoices = selectiveChoices
		ff.mandatoryCompleted = mandatoryCompleted
		ff.optionalCompleted = optionalCompleted
		ff.selectiveCompleted = selectiveCompleted
	}
	return &ff
}

// Returns All links for the downloader object
func (ff *fFastDownloader) GetAllLinks() []cacheCode.Link {
	return ff.allLinks
}

// Returns Mandatory links for the downloader object
func (ff *fFastDownloader) GetMandatoryLinks() []string {
	return ff.mandatoryLinks
}

// Returns Optional links for the downloader object
func (ff *fFastDownloader) GetOptionalLinks() []string {
	return ff.optionalLinks
}

// Returns Selective links for the downloader object
func (ff *fFastDownloader) GetSelectiveLinks() []string {
	return ff.selectiveLinks
}

// Returns the URL for the downloader object
func (ff *fFastDownloader) GetURL() string {
	return ff.url
}

// Returns the Name for the downloader object
func (ff *fFastDownloader) GetName() string {
	return ff.name
}

func removeDuplicates(input []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range input {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

func (ff *fFastDownloader) segregateLinks(id string) {
	mandatoryLinks := []string{}
	optionalLinks := []string{}
	selectiveLinks := []string{}
	for _, link := range ff.allLinks {
		switch link.Mode {
		case "optional":
			optionalLinks = append(optionalLinks, link.URL)
		case "selective":
			selectiveLinks = append(selectiveLinks, link.URL)
		default:
			mandatoryLinks = append(mandatoryLinks, link.URL)
		}
	}
	ff.id = id
	ff.mandatoryLinks = mandatoryLinks
	ff.optionalLinks = optionalLinks
	ff.selectiveLinks = selectiveLinks
}

func (ff *fFastDownloader) SelectLinks() {
	selectedOptional := []string{}
	selectedSelective := []string{}

	for _, choice := range ff.optionalChoices {
		if choice >= 0 && choice < len(ff.optionalLinks) {
			selectedOptional = append(selectedOptional, ff.optionalLinks[choice])
		}
	}

	for _, choice := range ff.selectiveChoices {
		if choice >= 0 && choice < len(ff.selectiveLinks) {
			selectedSelective = append(selectedSelective, ff.selectiveLinks[choice])
		}
	}

	useCache := "N"
	if len(selectedOptional) > 0 {
		fmt.Println("These are the cached optional links: ")
		for _, link := range selectedOptional {
			fmt.Println(link)
		}
		useCache = takeInput("Do you want to use the cached optional links? (Y/n): ", "Use Cache", "Y", true)
	}
	if strings.ToUpper(useCache) != "Y" {
		if len(ff.optionalLinks) > 0 {
			prompt := &survey.MultiSelect{
				Message: "Select optional links:",
				Options: ff.optionalLinks,
			}
			var selectedOptional []string
			err := survey.AskOne(prompt, &selectedOptional)
			if err != nil {
				log.Fatal("[ERROR] Could not select optional links: ", err)
			}
			for _, link := range selectedOptional {
				fmt.Println(link)
			}
			ff.optionalChoices = []int{}
			for _, selected := range selectedOptional {
				for i, option := range ff.optionalLinks {
					if selected == option {
						ff.optionalChoices = append(ff.optionalChoices, i)
					}
				}
			}
			ff.optionalLinks = selectedOptional
		}
	} else {
		ff.optionalLinks = selectedOptional
	}

	useCache = "N"
	if len(selectedSelective) > 0 {
		fmt.Println("These are the cached selective links: ")
		for _, link := range selectedSelective {
			fmt.Println(link)
		}
		useCache = takeInput("Do you want to use the cached selective links? (Y/n): ", "Use Cache", "Y", true)
	}
	if strings.ToUpper(useCache) != "Y" {
		if len(ff.selectiveLinks) > 0 {
			prompt := &survey.MultiSelect{
				Message: "Select selective links:",
				Options: ff.selectiveLinks,
			}
			var selectedSelective []string
			err := survey.AskOne(prompt, &selectedSelective)
			if err != nil {
				log.Fatal("[ERROR] Could not select selective links: ", err)
			}
			
			for _, link := range selectedSelective {
				fmt.Println(link)
			}
			ff.selectiveChoices = []int{}
			for _, selected := range selectedSelective {
				for i, option := range ff.selectiveLinks {
					if selected == option {
						ff.selectiveChoices = append(ff.selectiveChoices, i)
					}
				}
			}
			ff.selectiveLinks = selectedSelective
		}
	} else {
		ff.selectiveLinks = selectedSelective
	}
}

func getFinalLink(initialLink string) string {
	resp, err := http.Get(initialLink)
	if err != nil {
		log.Fatal("[ERROR] Could not get the initial link: ", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("[ERROR] Could not read the initial link body: ", err)
	}

	html := string(body)
	
	re := regexp.MustCompile(`window\.open\("([^"]+)"\)`)
	match := re.FindStringSubmatch(html)
	
	if len(match) <= 1 {
		log.Fatal("[ERROR] Could not find the final link in the HTML")
	}

	return match[1]
}

func downloadAPart(url string, fdmDownloadPath string) {
	downloadURL := getFinalLink(url)

	filename := strings.Split(url, "#")[1]
	tempFilePath := fdmDownloadPath+filename+".fdmdownload"
	downloadPath := fdmDownloadPath+filename

	if _, err := os.Stat(downloadPath); err == nil {
		fmt.Println("File already exists: ", downloadPath," skipping download.")
		return
	}

	if _, err := os.Stat(tempFilePath); err == nil {
		err := os.Remove(tempFilePath)
		if err != nil {
			log.Fatal("[ERROR] Could not remove the temporary file: ", err)
		}
	}

	cmd := exec.Command(
		"fdm",
		"-s",
		"-u", downloadURL,
	)
	err := cmd.Run()
	if err != nil {
		log.Fatal("[ERROR] Could not download the file: ", err)
	}
	time.Sleep(1500*time.Millisecond)

	_, err = os.Stat(tempFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			if _, err := os.Stat(downloadPath); err == nil {
				fmt.Println("File already exists: ", downloadPath," skipping download.")
				return
			}
		}
		log.Fatal("[ERROR] Could not get the file info: ", err)
	}

	fmt.Println("Downloading... ", downloadPath)
	for {
		if _, err := os.Stat(downloadPath); err == nil {
			fmt.Println("Download completed: ", downloadPath)
			break
		} else if os.IsNotExist(err) {
			time.Sleep(1*time.Second)
		} else {
			log.Fatal("[ERROR] Could not check the file existence: ", err)
		}
	}
}

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

// Download the files using the final links
func (ff *fFastDownloader) DownloadParts() {
	alreadyDownloadedMandatory := ff.mandatoryCompleted
	alreadyDownloadedOptional := ff.optionalCompleted
	alreadyDownloadedSelective := ff.selectiveCompleted

	alreadyDownloadedMandatoryNextToLastLinkFileName := ""
	alreadyDownloadedOptionalNextToLastLinkFileName := ""
	alreadyDownloadedSelectiveNextToLastLinkFileName := ""
	if len(ff.mandatoryLinks) > alreadyDownloadedMandatory+1 {
		alreadyDownloadedMandatoryNextToLastLinkFileName = strings.Split(ff.mandatoryLinks[alreadyDownloadedMandatory+1], "#")[1]
	}
	for i := range ff.optionalLinks {
		if i == alreadyDownloadedOptional {
			if i < len(ff.optionalLinks)-1 {
				alreadyDownloadedOptionalNextToLastLinkFileName = strings.Split(ff.optionalLinks[i+1], "#")[1]
				break
			} else {
				alreadyDownloadedOptionalNextToLastLinkFileName = strings.Split(ff.optionalLinks[i], "#")[1]
				break
			}
		} else if i > alreadyDownloadedOptional {
			break
		}
	}
	for i := range ff.selectiveLinks {
		if i == alreadyDownloadedSelective {
			if i < len(ff.selectiveLinks)-1 {
				alreadyDownloadedSelectiveNextToLastLinkFileName = strings.Split(ff.selectiveLinks[i+1], "#")[1]
				break
			} else {
				alreadyDownloadedSelectiveNextToLastLinkFileName = strings.Split(ff.selectiveLinks[i], "#")[1]
				break
			}
		} else if i > alreadyDownloadedSelective {
			break
		}
	}

	if alreadyDownloadedMandatoryNextToLastLinkFileName != "" {
		if _, err := os.Stat(ff.fdmDownloadPath + alreadyDownloadedMandatoryNextToLastLinkFileName); err == nil {
			fmt.Println("Mandatory part ", alreadyDownloadedMandatory+1, " already downloaded, skipping...")
			alreadyDownloadedMandatory++
		}
	}
	if alreadyDownloadedOptionalNextToLastLinkFileName != "" {
		if _, err := os.Stat(ff.fdmDownloadPath + alreadyDownloadedOptionalNextToLastLinkFileName); err == nil {
			fmt.Println("Optional part ", alreadyDownloadedOptional+1, " already downloaded, skipping...")
			alreadyDownloadedOptional++
		}
	}
	if alreadyDownloadedSelectiveNextToLastLinkFileName != "" {
		if _, err := os.Stat(ff.fdmDownloadPath + alreadyDownloadedSelectiveNextToLastLinkFileName); err == nil {
			fmt.Println("Selective part ", alreadyDownloadedSelective+1, " already downloaded, skipping...")
			alreadyDownloadedSelective++
		}
	}

	for i, url := range ff.mandatoryLinks {
		if i < alreadyDownloadedMandatory {
			continue
		}
		fmt.Println(strings.Repeat("-", 20))
		fmt.Printf("Downloading part %d/%d\n", i+1, len(ff.mandatoryLinks))
		downloadAPart(url, ff.fdmDownloadPath)
		if i == len(ff.mandatoryLinks)-1 {
			fmt.Println("All mandatory parts downloaded.")
		}
		ff.mandatoryCompleted++
		cacheCode.SaveDownloadStateCache(ff.id, ff.optionalChoices, ff.selectiveChoices, ff.mandatoryCompleted, ff.optionalCompleted, ff.selectiveCompleted)
	}
	for i, url := range ff.optionalLinks {
		if i < alreadyDownloadedOptional {
			continue
		}
		fmt.Println(strings.Repeat("-", 20))
		fmt.Printf("Downloading optional part %d/%d\n", i+1, len(ff.optionalLinks))
		downloadAPart(url, ff.fdmDownloadPath)
		if i == len(ff.optionalLinks)-1 {
			fmt.Println("All optional parts downloaded.")
		}
		ff.optionalCompleted++
		cacheCode.SaveDownloadStateCache(ff.id, ff.optionalChoices, ff.selectiveChoices, ff.mandatoryCompleted, ff.optionalCompleted, ff.selectiveCompleted)
	}
	for i, url := range ff.selectiveLinks {
		if i < alreadyDownloadedSelective {
			continue
		}
		fmt.Println(strings.Repeat("-", 20))
		fmt.Printf("Downloading selective part %d/%d\n", i+1, len(ff.selectiveLinks))
		downloadAPart(url, ff.fdmDownloadPath)
		if i == len(ff.selectiveLinks)-1 {
			fmt.Println("All selective parts downloaded.")
		}
		ff.selectiveCompleted++
		cacheCode.SaveDownloadStateCache(ff.id, ff.optionalChoices, ff.selectiveChoices, ff.mandatoryCompleted, ff.optionalCompleted, ff.selectiveCompleted)
	}
	fmt.Println("########## Game Downloaded Successfully ##########")
}

// Get the downloader object in a pretty format
func (ff *fFastDownloader) PrintMarshalledDownloader() string {
	type fFastDownloader struct {
		Id string
		Name string
		URL string
		AllLinks []cacheCode.Link
		MandatoryLinks []string
		OptionalLinks []string
		SelectiveLinks []string
	}
	FF := fFastDownloader{
		Id: ff.id,
		Name: ff.name,
		URL: ff.url,
		AllLinks: ff.allLinks,
		MandatoryLinks: ff.mandatoryLinks,
		OptionalLinks: ff.optionalLinks,
		SelectiveLinks: ff.selectiveLinks,
	}
	buf, err := json.MarshalIndent(FF,"","	")
	if err != nil {
		log.Fatal("[ERROR] Could not marshal the downloader object: ", err)
	}
	return string(buf)
}

// Decode the PrivateBinURL for download links
func (ff *fFastDownloader) DecodePrivateBin() {
	fmt.Println("Fetching the Links from Private Paste Bin...")

	timeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancelTimeout()
	
	ctx, cancel := chromedp.NewContext(timeoutCtx)
	defer cancel()

	hashKey := cacheCode.HashURL(ff.url)
	cacheExists := cacheCode.CacheExists(hashKey)

	if(cacheExists) {
		links, err := cacheCode.ReadCache(hashKey)
		if err != nil {
			log.Fatal("[ERROR] Could not read from cache: ", err)
		}
		ff.allLinks = links
		ff.segregateLinks(hashKey)
		return
	}
	
	var pageHTML string
	
	err := chromedp.Run(ctx,
		chromedp.Navigate(ff.url),
		chromedp.Sleep(5*time.Second),
		chromedp.OuterHTML("body", &pageHTML),
	)
	if err != nil {
		log.Fatal(err)
	}
	
	re := regexp.MustCompile(`https://fuckingfast\.co/[^\s"<>]+`)
	links := removeDuplicates(re.FindAllString(pageHTML, -1))
	
	clinks := []cacheCode.Link{}
	for _, l := range links {
		mode := "mandatory"
		if strings.Contains(l,"optional"){mode="optional"
		} else if strings.Contains(l,"selective"){mode="selective"}
		clinks = append(clinks, cacheCode.Link{URL: l, Mode: mode})
	}
	ff.allLinks = clinks

	ff.segregateLinks(hashKey)

	err = cacheCode.SaveCache(hashKey, ff.allLinks)
	if err != nil {
		log.Fatal("[ERROR] Could not save to cache: ", err)
	}
}