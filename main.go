package main

import (
	"github.com/codeuk/pout/cmd/render"
)

var Version = "dev"

func main() {
	// Start the GUI rendering routine.
	render.Init()
}
