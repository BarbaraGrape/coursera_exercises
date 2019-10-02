package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

//easy:json
type JSONData struct {
	Browsers []string
	Company  string
	Country  string
	Email    string
	Job      string
	Name     string
	Phone    string
}

func findBrowsers(browsers []string, jsonRow *JSONData) (bool, []string) {
	concurrence := len(browsers)
	br := make([]string, 0, len(browsers))
	for _, browser := range browsers {
		if strings.Contains(strings.Join(jsonRow.Browsers[:], ","), browser) {
			concurrence--
			for _, b := range jsonRow.Browsers {
				if strings.Contains(b, browser) {
					br = append(br, b)
				}
			}
		}
	}
	if concurrence == 0 {
		return true, br
	}
	return false, br
}

func parseJson(filepath string, out io.Writer) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("Cant open filepath")
	}

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("Cant read file")
	}

	lines := strings.Split(string(fileContents), "\n")
	browsers := []string{"Android", "MSIE"}
	seenBrowsers := make(map[string]bool)
	reMail := regexp.MustCompile("@")
	jsonData := JSONData{}
	fmt.Fprintln(out, "found users:")
	for i, line := range lines {
		err = jsonData.UnmarshalJSON([]byte(line))
		if err != nil {
			return fmt.Errorf("Cant decode json")
		}

		boolFound, jsonBrowsers := findBrowsers(browsers, &jsonData)
		//fmt.Println(jsonBrowsers)
		if len(jsonBrowsers) != 0 {
			//fmt.Println(boolFound, jsonBrowsers)
			if boolFound {
				//fmt.Println("Android and MSIE user: ", jsonData.Name, "\t", jsonData.Email)
				fmt.Fprintf(out, "[%d] %s <%s>\n", i, jsonData.Name, reMail.ReplaceAllString(jsonData.Email, " [at] "))
			}

			for _, browser := range jsonBrowsers {
				if _, ok := seenBrowsers[browser]; !ok {
					//fmt.Println("New browser: ", browser, "first seen: ", jsonData.Name)
					seenBrowsers[browser] = true
				}
			}
		}
	}

	fmt.Fprintln(out, "\nTotal unique browsers", len(seenBrowsers))
	//fmt.Println("Found users: ", users)
	return nil
}

func FastSearch(out io.Writer) {
	err := parseJson("./data/users.txt", out)
	if err != nil {
		fmt.Println("Something go wrong")
		return
	}
}

func main() {
	FastSearch(ioutil.Discard)
}
