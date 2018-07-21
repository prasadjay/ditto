package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"time"
)

func main() {
	startTime := time.Now()

	var fileName string
	config := GetConfig()
	fileName = config["file_path"].(string)

	if fileName == "" {
		fmt.Println("ERROR : File name cannot be empty. Recheck config.")
		os.Exit(1)
	}

	if _, err := os.Stat("converted_files"); os.IsNotExist(err) {
		os.Mkdir("converted_files", 0666)
	}

	var slash string
	if runtime.GOOS == "windows" {
		slash = "\\"
	} else {
		slash = "/"
	}

	f, fErr := os.Open(fileName)
	if fErr != nil {
		fmt.Println("File not found by name:" + fileName + " Please recheck config file.")
		os.Exit(1)
	}

	defer f.Close()

	r := csv.NewReader(bufio.NewReader(f))
	count := 0
	fmt.Println("Starting to read the file : " + fileName)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}

		if count == 0 {
			count += 1
		} else {
			if len(record) > 1 {
				err := ioutil.WriteFile(("converted_files" + slash + record[0] + ".txt"), []byte(record[1]), 0666)
				if err != nil {
					fmt.Println(err.Error())
				}
			} else {
				fmt.Println("WARNNING : Only one column for record row no: " + strconv.Itoa(count) + " was found. Writing file for line was skipped. Please verify data.")
			}
			count += 1
		}
	}

	fmt.Println("Completed reading and writing...")

	endTime := time.Now()

	fmt.Println("Total Duration : " + endTime.Sub(startTime).String())

}

func GetConfig() (config map[string]interface{}) {
	config = make(map[string]interface{})

	content, err := ioutil.ReadFile("config.json")

	if err != nil {
		//config file generated
		fmt.Println("WARNING : Configuration file not found. Generating an empty configuration file. Please update with a file name.")
		config["file_path"] = "./test.csv"
		byteArray, _ := json.Marshal(config)
		_ = ioutil.WriteFile("config.json", byteArray, 0777)
	} else {
		json.Unmarshal(content, &config)
	}

	return
}
