package main

import (
	"github.com/eqr/transferit/app/cmd"
	"github.com/eqr/transferit/app/config"
)

var cfg *config.Config

func init() {
	cmd.Build()
}

func main() {
	cmd.Execute()
}
