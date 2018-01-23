package main

import (
	"fmt"
	"log"
	"os"

	overseer "github.com/expandonline/h2o-steam-overseer"
)

func main() {
	if os.Geteuid() != 0 && os.Getenv("OVERSEER_DEV") != "true" {
		log.Fatalln("h2o-steam-overseer only works as root. Set OVERSEER_DEV=true for development.")
	}
	if len(os.Args) != 2 {
		log.Fatalln("Missing arguments. Usage: h2o-steam-overseer /root/steam")
	}
	path := os.Args[1]
	_, err := os.Stat(path)
	if os.IsPermission(err) {
		log.Fatalln("No permissions on given directory")
	}
	if os.IsNotExist(err) {
		log.Fatalln("Given directory does not exist.")
	}
	if err != nil {
		log.Fatalln(fmt.Sprintf("Unknown error: %s", err))
	}
	overseer.EnsureRunning(path)
}
