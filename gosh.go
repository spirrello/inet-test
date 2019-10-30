package main

import (
	"github.com/codeskyblue/go-sh"
)

func main() {

	x := "7"
	sh.Command("ping", "google.com", "-c", "2").Command("grep", "loss").Command("awk", "{print $"+x+"}").Run()
}
