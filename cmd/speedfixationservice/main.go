// Package start server with configs
package main

import (
	"flag"

	"github.com/IgorRybak2055/speed-control-service/internal/speedfixationservice"
)

func main() {
	flag.Parse()

	speedfixationservice.Run()
}
