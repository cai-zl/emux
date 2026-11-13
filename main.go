package main

import (
	"os"

	"easymux/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
