package main

import (
	"github.com/mymaqc/proxy/server"
	"github.com/mymaqc/proxy"
)

func main() {
	s := server.New()
	s.Start()
}
