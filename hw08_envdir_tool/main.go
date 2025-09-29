package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		log.Println(fmt.Errorf("not enough arguments, at least three required"))
	}

	dirPath := os.Args[1]
	env, err := ReadDir(dirPath)
	if err != nil {
		log.Println(err)
	}

	os.Exit(RunCmd(os.Args[2:], env))
}
