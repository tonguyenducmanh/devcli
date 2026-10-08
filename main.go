// Điểm khởi động của tm - command line app cá nhân.
package main

import (
	"fmt"
	"os"

	"github.com/tonguyenducmanh/devcli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
