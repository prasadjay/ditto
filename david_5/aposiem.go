package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

/*
 *			Main
 */

func main() {

	/*
	 *		Fetch configurations from the aposiem.json (External) file.
	 */

	config := GetConfig()
	bearer_token := config["bearer_token"].(string)
	namespace := config["namespace"].(string)
	recursive := config["recursive"].(string)
	log_path := config["log_path"].(string)

	fmt.Println("ApoSiem running and generating output to " + log_path)

	/*
	 * 		Perform some basic environment checks
	 */

	if bearer_token == "" {
		fmt.Println("ERROR : Bearer Token cannot be empty. Please set the bearer token value in aposiem.json file.")
		os.Exit(1)
	}
	if namespace == "" {
		fmt.Println("ERROR : Namespace cannot be empty. Please set the namespace value in aposiem.json file.")
		os.Exit(1)
	}
	if recursive == "" {
		fmt.Println("ERROR : Recursive cannot be empty. Please set the recursive value in aposiem.json file.")
		os.Exit(1)
	}
	if log_path == "" {
		fmt.Println("ERROR : log_path cannot be empty. Please set the log_path value in aposiem.json file.")
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
	valueStringArray := columnString + "," + "DNS"
	valueStringArrayRequested := ""

	for x := 0; x < len(valuesRaw); x++ {

		/*
		 *		Iterating through each array block
		 */

		tempString := ""
		tempStringRequested := ""
		tempArray := valuesRaw[x].([]interface{})
		dns := ""

		for y := 0; y < len(tempArray); y++ {

			if y == 17 { // resolve dns
				dns = GetDNSForIp(GetStringValForInterface(tempArray[y]))
				//fmt.Println(dns)
			}
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
				if parsedTime, err := time.Parse(time.RFC3339, stringVal); err != nil {
					fmt.Println(err.Error())
					tempStringRequested += columnArray[y] + ": " + GetStringValForInterface(tempArray[y]) + ", "
				} else {
					tempStringRequested += columnArray[y] + ": " + parsedTime.Format(time.RFC3339) + ", "
				}
			} else {
				tempStringRequested += columnArray[y] + ": " + GetStringValForInterface(tempArray[y]) + ", "
			}
		}

		tempString = strings.TrimRight(tempString, ",")
		valueStringArray += "\n" + tempString + "," + dns

		tempStringRequested = strings.TrimRight(tempStringRequested, ", ")
		valueStringArrayRequested += tempStringRequested + ", " + dns + "\n"
	}

	/*
	 *		Write output.  The output selections may be modified, but by default the CSV format
	 *		is used.append  The other options are a raw JSON response and the second is a key-value
	 *		response.  At this time, the splunk and arcsight integration leverage CSV output
	 */

	err = ioutil.WriteFile(log_path+"aporeto_csv.csv", []byte(valueStringArray), 0777)
	// err = ioutil.WriteFile(log_path+"aporeto_json_response.txt", body, 0777)
	// err = ioutil.WriteFile(log_path+"aporeto_json_response_formatted.txt", []byte(valueStringArrayRequested), 0777)

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
	content, err := ioutil.ReadFile("./aposiem.json")

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

func GetDNSForIp(ip string) (dns string) {
	var b []byte
	var err error

	if ip == "127.0.0.1" || strings.HasPrefix(ip, "172") || strings.HasPrefix(ip, "10.") {
		//localhost, docker ips and internal network ips.. all ignore
		dns = "N/A"
	} else {

		if runtime.GOOS == "windows" {
			b, err = exec.Command("cmd", "/C", "nslookup "+ip+" | grep Name").Output()
		} else {
			cmd := exec.Command("sh", "-c", "nslookup "+ip+" | grep name")
			b, err = cmd.Output()
		}

		if err != nil {
			dns = "N/A"
		} else {
			if runtime.GOOS == "windows" {
				temp := strings.TrimSpace(string(b))
				temp = strings.Replace(temp, "Name", "", -1)
				temp = strings.Replace(temp, ":", "", -1)
				dns = strings.TrimSpace(temp)
			} else {
				temp := strings.TrimSpace(string(b))
				temp = strings.TrimSuffix(temp, ".")
				tempArray := strings.Split(temp, "name")
				dns = strings.Replace(tempArray[1], "=", "", -1)
				dns = strings.TrimSpace(dns)
			}
		}
	}
	return
}

/*
 * =============================================================================
 *		aposiem.go eof.
 *  =============================================================================
 */
