package cli

import (
	"fmt"
	"gophre/pkg/service"
	"gophre/pkg/web"
	"os"
)

func ParseCommand() {
	if len(os.Args) > 1 && os.Args[1] == "reset" {
		service.Reset()
	} else if len(os.Args) > 1 && os.Args[1] == "update" {
		service.Update()
	} else if len(os.Args) > 1 && os.Args[1] == "serve" {
		web.Serve()
	} else {
		fmt.Println("")
		fmt.Println("Usage: gophre <command>")
		fmt.Println("- gophre update         : Update RSS feeds")
		fmt.Println("- gophre serve <port>   : Run Local Web Server (port is optional)")
		fmt.Println("- gophre reset          : Reset all data")
		fmt.Println("")
	}
}
