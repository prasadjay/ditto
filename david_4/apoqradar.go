/*
 *
 */

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	//"log/syslog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

/*
 *			Main
 */

var GLOBAL_CONFIG map[string]interface{}

func main() {

	/*
	 *		Fetch configurations from the config.json (External) file.
	 */
	GLOBAL_CONFIG = make(map[string]interface{})
	config := GetConfig()
	GLOBAL_CONFIG = config
	bearer_token := config["bearer_token"].(string)
	namespace := config["namespace"].(string)
	recursive := config["recursive"].(string)
	log_path := config["log_path"].(string)

	fmt.Println("ApoQradar running and generating output to " + log_path)

	/*
	 * 		Perform some basic environment checks
	 */

	if bearer_token == "" {
		fmt.Println("ERROR : Bearer Token cannot be empty. Please set the bearer token value in config.json file.")
		os.Exit(1)
	}
	if namespace == "" {
		fmt.Println("ERROR : Namespace cannot be empty. Please set the namespace value in config.json file.")
		os.Exit(1)
	}
	if recursive == "" {
		fmt.Println("ERROR : Recursive cannot be empty. Please set the recursive value in config.json file.")
		os.Exit(1)
	}
	if log_path == "" {
		fmt.Println("ERROR : log_path cannot be empty. Please set the log_path value in config.json file.")
		os.Exit(1)
	}

	/*
	 * 		HTTP Request and error check
	 */

	req, err := http.NewRequest("GET", "https://api.console.aporeto.com/statsqueries?startRelative=20m&measurement=flows&recursive="+recursive, nil)
	if err != nil {
		panic(err)
	}

	/*
	 * 		Set our headers for HTTP request
	 */

	req.Header.Set("X-Namespace", namespace)
	bearerToken := "Bearer " + bearer_token
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", bearerToken)
	req.Header.Set("Cache-Control", "no-cache")

	/*
	 *		HTTP Response handling
	 */

	resp, err := http.DefaultClient.Do(req)

	/*
	 * 		Error Checking
	 */

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	/*
	 * 		Check our response codes.  If error, likely a token expiration issue.
	 */

	if resp.StatusCode != 200 {
		if resp.StatusCode == 401 {
			fmt.Println("UNAUTHORIZED : Bearer token has been expired. Please update a valid token in aposiem.json file.")
			os.Exit(1)
		} else {
			fmt.Println("ERROR : Unknown Error Occured. Refer the service.")
			os.Exit(1)
		}
	}

	/*
	 * 		Error Checking
	 */

	if err != nil {
		panic(err)
	}

	bodyObject := make([]interface{}, 0)
	err = json.Unmarshal(body, &bodyObject)

	/*
	 * 		Error Checking
	 */

	if err != nil {
		panic(err)
	}

	rowObject := bodyObject[0].(map[string]interface{})["results"].([]interface{})[0].(map[string]interface{})["rows"].([]interface{})[0].(map[string]interface{})
	columnsRaw := rowObject["columns"].([]interface{})

	/*
	 *		Generating column values
	 */

	columnString := ""
	columnArray := make([]string, 0)

	for x := 0; x < len(columnsRaw); x++ {
		columnString += columnsRaw[x].(string) + ","
		columnArray = append(columnArray, columnsRaw[x].(string))
	}

	columnString = strings.TrimRight(columnString, ",")

	/*
	 *		Generating value strings
	 */

	valuesRaw := rowObject["values"].([]interface{})
	valueStringArray := columnString
	valueStringArrayRequested := ""

	//for headersend json
	objectArray := make([]map[string]string, 0)

	for x := 0; x < len(valuesRaw); x++ {
		singleObject := make(map[string]string)

		/*
		 *		Iterating through each array block
		 */

		tempString := ""
		tempStringRequested := ""
		tempArray := valuesRaw[x].([]interface{})

		for y := 0; y < len(tempArray); y++ {

			/*
			 *	For our CSV output
			 */

			if y == 0 {
				stringVal := GetStringValForInterface(tempArray[y])
				if parsedTime, err := time.Parse(time.RFC3339, stringVal); err != nil {
					fmt.Println(err.Error())
					tempString += GetStringValForInterface(tempArray[y]) + ","
				} else {
					tempString += parsedTime.Format(time.RFC3339) + ","
				}
			} else {
				tempString += GetStringValForInterface(tempArray[y]) + ","
			}

			/*
			 * 	For our formatted file type (key-value pair)
			 */

			if columnArray[y] == "time" {
				stringVal := GetStringValForInterface(tempArray[y])
				singleObject[strings.TrimPrefix(strings.TrimPrefix(columnArray[y], "@"), "$")] = stringVal

				if parsedTime, err := time.Parse(time.RFC3339, stringVal); err != nil {
					fmt.Println(err.Error())
					tempStringRequested += columnArray[y] + ": " + GetStringValForInterface(tempArray[y]) + ", "
				} else {
					tempStringRequested += columnArray[y] + ": " + parsedTime.Format(time.RFC3339) + ", "
				}
			} else {
				tempStringRequested += columnArray[y] + ": " + GetStringValForInterface(tempArray[y]) + ", "
				singleObject[strings.TrimPrefix(strings.TrimPrefix(columnArray[y], "@"), "$")] = GetStringValForInterface(tempArray[y])
			}

		}

		objectArray = append(objectArray, singleObject)

		tempString = strings.TrimRight(tempString, ",")
		valueStringArray += "\n" + tempString

		tempStringRequested = strings.TrimRight(tempStringRequested, ", ")
		valueStringArrayRequested += tempStringRequested + "\n"
	}

	//Begin transforming to send in headers
	PublishToIbm(objectArray)

	/*
	 *		Write output.  The output selections may be modified, but by default the CSV format
	 *		is used.append  The other options are a raw JSON response and the second is a key-value
	 *		response
	 */

	err = ioutil.WriteFile(log_path+"aporeto_csv.csv", []byte(valueStringArray), 0777)
	err = ioutil.WriteFile(log_path+"aporeto_json_response.txt", body, 0777)
	err = ioutil.WriteFile(log_path+"aporeto_json_response_formatted.txt", []byte(valueStringArrayRequested), 0777)

}

/*
 *		Get String Value for Interface
 */

func GetStringValForInterface(input interface{}) (output string) {
	status := false
	if output, status = input.(string); !status {
		if floatVal, status := input.(float64); status {
			output = strconv.Itoa(int(floatVal))
		} else if input == nil {
			output = "null"
		} else {

			/*
			 *		Handle all other exceptions here.. currently setting for null
			 */

			output = "null"
		}
	}

	return
}

/*
 *		Get the configuration information.  If no config file is found,
 *		we create a config file in the working directory with required
 *		fields.
 */

func GetConfig() (config map[string]interface{}) {
	config = make(map[string]interface{})
	//content, err := ioutil.ReadFile("./aposiem.json")
	content, err := ioutil.ReadFile("./apoqradar.json")

	if err != nil {
		fmt.Println("WARNING : Configuration file not found. Generating an empty configuration file. Please update token and namespace.")
		config["bearer_token"] = ""
		config["namespace"] = ""
		config["recursive"] = "true"
		config["log_path"] = "./"
		byteArray, _ := json.Marshal(config)
		_ = ioutil.WriteFile("./aposiem.json", byteArray, 0777)
	} else {
		json.Unmarshal(content, &config)
	}

	return
}

//Get file content
//reformat
//make header

func PublishToIbm(objects []map[string]string) {
	url := GLOBAL_CONFIG["ibm_server_ip"].(string) + ":" + GLOBAL_CONFIG["ibm_server_port"].(string)

	fmt.Print("Initiating service to SYSLOG server (" + url + ")... ")
	/*log, err := syslog.Dial("tcp", url, syslog.LOG_INFO, "apoqradar")

	isOkay := true

	if err != nil {
		isOkay = false
		fmt.Println("failed. Reason : " + err.Error())
		fmt.Println()
	} else {
		fmt.Println("success")
		fmt.Println()
	}*/

	newTemplateFileString := ""
	for x := 0; x < len(objects); x++ {
		object := objects[x]
		headerString := GLOBAL_CONFIG["ibm_leef_header"].(string) + "\ttime=" + object["time"] + "\tnamespace=" + object["namespace"] + "\taction=" + object["action"] + "\tdestid=" + object["destid"] + "\tdestip=" + object["destip"] + "\tdestport=" + object["destport"] + "\tdesttype=" + object["desttype"] + "\tencrypted=" + object["encrypted"] + "\tl4proto=" + object["l4proto"] + "\toaction=" + object["oaction"] + "\tobserved=" + object["observed"] + "\topolicyid=" + object["opolicyid"] + "\tpolicyid=" + object["policyid"] + "\tpolicyns=" + object["policyns"] + "\treason=" + object["reason"] + "\tsrcid=" + object["srcid"] + "\tsrcip=" + object["srcip"] + "\tsrctype=" + object["srctype"] + "\tsrvid=" + object["srvid"] + "\tsrvtype=" + object["srvtype"] + "\turi=" + object["uri"] + "\tvalue=" + object["value"]
		newTemplateFileString += headerString + "\n"

		// if isOkay {
		// 	if err := log.Info(headerString); err != nil {
		// 		fmt.Println("Error sending syslog. Reason : " + err.Error())
		// 	}
		// }

	}
	//write to file
	_ = ioutil.WriteFile(GLOBAL_CONFIG["log_path"].(string)+"aporeto_new_format.txt", []byte(newTemplateFileString), 0777)

}

/*
 * =============================================================================
 *		apoqradar.go eof.
 *  =============================================================================
 */
