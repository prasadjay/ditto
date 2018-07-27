package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

/*
 *		Vulnerabilities
 */

type Vulnerabilities struct {
	ID          string `json:"ID"`
	Annotations struct {
	} `json:"annotations"`
	AssociatedTags []interface{} `json:"associatedTags"`
	CreateTime     time.Time     `json:"createTime"`
	Description    string        `json:"description"`
	Link           string        `json:"link"`
	Name           string        `json:"name"`
	Namespace      string        `json:"namespace"`
	NormalizedTags []string      `json:"normalizedTags"`
	Protected      bool          `json:"protected"`
	Severity       int           `json:"severity"`
	UpdateTime     time.Time     `json:"updateTime"`
}

const CONFIG_PATH = "./"

var BEARER_TOKEN string
var NAMESPACE string
var RECURSIVE string
var START_RELATIVE string
var JIRA_SERVER string
var JIRA_USERNAME string
var JIRA_PASSWORD string
var JIRA_PROJECT string

func main() {
	config := GetConfig()

	BEARER_TOKEN = config["bearer_token"].(string)
	NAMESPACE = config["namespace"].(string)
	RECURSIVE = config["recursive"].(string)
	START_RELATIVE = config["start_relative"].(string)
	JIRA_SERVER = config["jira_server"].(string)
	JIRA_USERNAME = config["jira_username"].(string)
	JIRA_PASSWORD = config["jira_password"].(string)
	JIRA_PROJECT = config["jira_project"].(string)

	if BEARER_TOKEN == "" {
		fmt.Println("ERROR : bearer_token cannot be empty. Please update in config.json file.")
		os.Exit(1)
	}

	if NAMESPACE == "" {
		fmt.Println("ERROR : namespace cannot be empty. Please update in config.json file.")
		os.Exit(1)
	}

	if RECURSIVE == "" {
		fmt.Println("ERROR : recursive cannot be empty. Please update in config.json file.")
		os.Exit(1)
	}

	if START_RELATIVE == "" {
		fmt.Println("ERROR : start_relative cannot be empty. Please update in config.json file.")
		os.Exit(1)
	}

	if JIRA_SERVER == "" {
		fmt.Println("ERROR : jira_server cannot be empty. Please update in config.json file.")
		os.Exit(1)
	}

	if JIRA_USERNAME == "" {
		fmt.Println("ERROR : jira_username cannot be empty. Please update in config.json file.")
		os.Exit(1)
	}

	if JIRA_PASSWORD == "" {
		fmt.Println("ERROR : jira_password cannot be empty. Please update in config.json file.")
		os.Exit(1)
	}

	if JIRA_PROJECT == "" {
		fmt.Println("ERROR : jira_project cannot be empty. Please update in config.json file.")
		os.Exit(1)
	}

	/*
	 *		Create our HTTP request to the API endpoint
	 */

	vulnerabilitiesReq, err := http.NewRequest("GET", "https://api.console.aporeto.com/vulnerabilities?startRelative="+START_RELATIVE+"&recursive="+RECURSIVE, nil)

	if err != nil {
		fmt.Println(err)
	}

	bearerToken := "Bearer " + BEARER_TOKEN
	fmt.Printf("Making request for vulnerabilities...\\c")
	vulnerabilitiesReq.Header.Set("X-Namespace", NAMESPACE)
	vulnerabilitiesReq.Header.Set("Authorization", bearerToken)
	vulnerabilitiesReq.Header.Set("Accept", "application/json")
	vulnerabilitiesReq.Header.Set("Cache-Control", "no-cache")
	vulnerabilitiesResp, err := http.DefaultClient.Do(vulnerabilitiesReq)
	fmt.Println("Complete")

	/*
	 *		Error Check
	 */

	if err != nil {
		fmt.Println(err)
	}

	defer vulnerabilitiesResp.Body.Close()

	// 	Receive our response data

	vulnerabilityContent, _ := ioutil.ReadAll(vulnerabilitiesResp.Body)

	ioutil.WriteFile("raw.json", vulnerabilityContent, 0666)

	/*
	 *		Error Check
	 */

	if err != nil {
		fmt.Println(err)
	}

	/*
	 *		Reference struct and unmarshall json data
	 */

	var vulnResponseData []Vulnerabilities
	err = json.Unmarshal(vulnerabilityContent, &vulnResponseData)

	/*
	 *		Error Check
	 */

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	/*
	 *		Define our output file and write out JSON response data to it
	 */

	vulnOutputFile, err := os.OpenFile("./apovulns.txt", os.O_WRONLY|os.O_CREATE, 0777)

	/*
	 *		Confirm we can create our output files
	 */

	if err != nil {
		fmt.Println("File does not exist or can not create")
		os.Exit(1)
	}

	/*
	 *		Interate through the JSON response and format for SIEM
	 */

	defer vulnOutputFile.Close()

	vulnerabilitiesOutput := bufio.NewWriter(vulnOutputFile)

	// vulnerabililty output
	for v, singledata := range vulnResponseData {
		if singledata.Severity == 5 && singledata.Name == "CVE-2018-6485" {
			//Create issue in jira
			tagString := strings.Join(singledata.NormalizedTags, " ")
			CreateJiraTask(JIRA_PROJECT, singledata.Name, ("$description=" + singledata.Description + " " + tagString))
		}
		fmt.Fprintf(vulnerabilitiesOutput, "%v, %v, %v, %v, %v, %v, %v\n", vulnResponseData[v].ID, vulnResponseData[v].Link, vulnResponseData[v].Namespace, vulnResponseData[v].Severity, vulnResponseData[v].Protected, vulnResponseData[v].CreateTime, vulnResponseData[v].NormalizedTags)
	}
	vulnerabilitiesOutput.Flush()

	/*
	 * 		End of main
	 */

}

func CreateJiraTask(project, summary, description string) (err error) {
	fmt.Print("Creating issue : " + summary + " ... ")

	if CheckIssueExists(summary) {
		msg := "failed -> Reason : Issue already created. Nothing to do."
		fmt.Println(msg)
		err = errors.New(msg)
	} else {

		url := "http://" + JIRA_SERVER + "/rest/api/2/issue/"

		payload := strings.NewReader("{\n\t\"fields\": {\n\t\t\"project\":\n\t\t{\n\t\t\t\"key\":\"" + project + "\"\n\t\t},\n\t\t\"summary\": \"" + summary + "\",\n\t\t\"description\": \"" + description + "\",\n\t\t\"issuetype\": {\n\t\t\t\"name\": \"Bug\"\n\t\t}\n\t}\n}")

		req, _ := http.NewRequest("POST", url, payload)

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Basic "+EncodeToBase64(JIRA_USERNAME+":"+JIRA_PASSWORD))
		req.Header.Add("Cache-Control", "no-cache")
		req.Header.Add("Postman-Token", "5280cec9-22d7-4c19-a1e5-8c9424e641d5")

		res, _ := http.DefaultClient.Do(req)
		defer res.Body.Close()
		if res.StatusCode == 201 {
			//success
			msg := "success"
			fmt.Println(msg)
		} else {
			msg := "failed"
			fmt.Println(msg)
			err = errors.New(msg)

			//print the error to stdout
			body, _ := ioutil.ReadAll(res.Body)
			fmt.Println(string(body))
			fmt.Println()
		}
	}

	return
}

func CheckIssueExists(summary string) (isExist bool) {
	urll := "http://" + JIRA_SERVER + "/rest/api/2/search?jql=summary~" + url.QueryEscape(summary)
	req, _ := http.NewRequest("GET", urll, nil)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+EncodeToBase64(JIRA_USERNAME+":"+JIRA_PASSWORD))
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Postman-Token", "5280cec9-22d7-4c19-a1e5-8c9424e641d5")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	//fmt.Println(res)
	bodyString := string(body)

	if strings.Contains(bodyString, "\"summary\":\""+summary) {
		isExist = true
	} else {
		isExist = false
	}
	return
}

func GetConfig() (config map[string]interface{}) {
	config = make(map[string]interface{})

	content, err := ioutil.ReadFile(CONFIG_PATH + "config.json")

	if err != nil {
		//config file not Available// regenerate it...
		fmt.Println("WARNING : Configuration file not found. Generating an empty configuration file. Please update correspondint details.")
		fmt.Println()
		config["bearer_token"] = ""
		config["namespace"] = ""
		config["recursive"] = "true"
		config["start_relative"] = "1h"
		config["jira_server"] = ""
		config["jira_username"] = ""
		config["jira_password"] = ""
		config["jira_project"] = ""
		byteArray, _ := json.Marshal(config)
		_ = ioutil.WriteFile(CONFIG_PATH+"config.json", byteArray, 0777)
	} else {
		json.Unmarshal(content, &config)
	}

	return
}

func EncodeToBase64(message string) (retour string) {

	base64Byte := make([]byte, base64.StdEncoding.EncodedLen(len(message)))

	base64.StdEncoding.Encode(base64Byte, []byte(message))

	return string(base64Byte)

}
