package main

import (
	"github.com/mymaqc/proxy/server"
)

func main() {
	s := server.New()
	s.Start()
}
