package main

import (
	"os"

	"github.com/mvanhorn/cars-il-pp-cli/internal/cli"
)

func main() {
	os.Exit(commands.Execute())
}
