package main

import (
	"os"

	"github.com/markes76/cars-il-pp-cli/internal/cli"
)

func main() {
	os.Exit(commands.Execute())
}
