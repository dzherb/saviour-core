package main

import (
	"flag"
	"log"

	"saviour/internal/app"
)

func main() {
	var cfgFilePath string

	flag.StringVar(
		&cfgFilePath,
		"config",
		"",
		"path to .json config file",
	)

	flag.Parse()

	err := app.Run([]string{cfgFilePath})
	if err != nil {
		log.Fatal(err)
	}
}
