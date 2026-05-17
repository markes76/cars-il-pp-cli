package main

import (
	"os"

	"github.com/markes76/cars-il-pp-cli/internal/cli"
)

var version = "1.0.0"

func main() {
	_ = version
	os.Exit(commands.Execute())
}
