package main

import (
	"fmt"
	"io"
	"os"

	"github.com/thatnerdjosh/mindfulness/internal/interfaces/cli"
)

var exit = os.Exit

func main() {
	exit(run(os.Args, os.Stdout, os.Stderr))
}

func run(args []string, out io.Writer, errOut io.Writer) int {
	if err := cli.Run(args, out, errOut); err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}
	return 0
}
