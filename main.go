package main

import (
	"eco/services/server"
	"services/centroDeCostoConstructoras/api"
)

func main() {
	eco := server.New("centroDeCostoConstructoras")
	api.Register(eco)
	eco.Start()
}
