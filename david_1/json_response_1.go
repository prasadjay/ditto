/*
 *
 *
 *
 */

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//This defines where config.json file resides
const CONFIG_PATH = "./"

func main() {

	//fetch configurations
	config := GetConfig()
	bearer_token := config["bearer_token"].(string)
	namespace := config["namespace"].(string)
	recursive := config["recursive"].(string)
	log_path := config["log_path"].(string)
	//fmt.Println(config)
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

	req, err := http.NewRequest("GET", "https://api.console.aporeto.com/statsqueries?startRelative=20m&measurement=flows&recursive="+recursive, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("X-Namespace", namespace)
	bearerToken := "Bearer " + bearer_token
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", bearerToken)
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		if resp.StatusCode == 401 {
			fmt.Println("UNAUTHORIZED : Bearer token has been expired. Please update a valid token in config.json file.")
			os.Exit(1)
		} else {
			fmt.Println("ERROR : Unknown Error Occured. Refer the service.")
			os.Exit(1)
		}
	}

	if err != nil {
		panic(err)
	}

	bodyObject := make([]interface{}, 0)
	err = json.Unmarshal(body, &bodyObject)

	if err != nil {
		panic(err)
	}

	rowObject := bodyObject[0].(map[string]interface{})["results"].([]interface{})[0].(map[string]interface{})["rows"].([]interface{})[0].(map[string]interface{})
	columnsRaw := rowObject["columns"].([]interface{})

	//generating column values
	columnString := ""
	columnArray := make([]string, 0)

	for x := 0; x < len(columnsRaw); x++ {
		columnString += columnsRaw[x].(string) + ","
		columnArray = append(columnArray, columnsRaw[x].(string))
	}

	columnString = strings.TrimRight(columnString, ",")

	//generating value strings
	valuesRaw := rowObject["values"].([]interface{})

	valueStringArray := columnString
	valueStringArrayRequested := ""

	for x := 0; x < len(valuesRaw); x++ {
		//iterating through each array block
		tempString := ""
		tempStringRequested := ""
		tempArray := valuesRaw[x].([]interface{})

		for y := 0; y < len(tempArray); y++ {
			//for csv file
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

			//For formatted file
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
		valueStringArray += "\n" + tempString

		tempStringRequested = strings.TrimRight(tempStringRequested, ", ")
		valueStringArrayRequested += tempStringRequested + "\n"
	}

	//printing the request for audit purpose
	err = ioutil.WriteFile(log_path+"json_response.txt", body, 0777)
	//printing the csv file required by task
	err = ioutil.WriteFile(log_path+"json_response.csv", []byte(valueStringArray), 0777)
	//printing the given sample file required by task
	err = ioutil.WriteFile(log_path+"json_response_formatted.txt", []byte(valueStringArrayRequested), 0777)

}

func GetStringValForInterface(input interface{}) (output string) {
	status := false
	if output, status = input.(string); !status {
		if floatVal, status := input.(float64); status {
			output = strconv.Itoa(int(floatVal))
		} else if input == nil {
			output = "null"
		} else {
			//handle all other exceptions here.. currently setting for null
			output = "null"
		}
	}

	return
}

func GetConfig() (config map[string]interface{}) {
	config = make(map[string]interface{})

	content, err := ioutil.ReadFile(CONFIG_PATH + "config.json")

	if err != nil {
		//config file not Available// regenerate it...
		fmt.Println("WARNING : Configuration file not found. Generating an empty configuration file. Please update token and namespace.")
		config["bearer_token"] = ""
		config["namespace"] = ""
		config["recursive"] = "true"
		config["log_path"] = "./"
		byteArray, _ := json.Marshal(config)
		_ = ioutil.WriteFile(CONFIG_PATH+"config.json", byteArray, 0777)
	} else {
		json.Unmarshal(content, &config)
	}

	return
}
