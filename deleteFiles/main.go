package main

import (
    "fmt"
    "os"
    "path/filepath"
)

func main(){
    searchDir := "I:\\Temp"
    _ = filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
            DeleteFolder(path, f)
            return nil
    })
}

func DeleteFolder(path string,f os.FileInfo) {
    defer func() {
        if err := recover(); err != nil {
            fmt.Println(path, err)
        }
    }()

    if f.IsDir() && f.Name() == "node_modules" && f.Name() != "node_modules\\"{
        fmt.Println("Removing : ",path)
        _= os.RemoveAll(path)
    }
}