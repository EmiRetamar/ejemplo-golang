package api

import "eco/services/server"

func Register(eco *server.Eco) {
	eco.AddRoute("/api/1/centroDeCostoConstructoras").Method("GET").Register(getCentrosDeCosto)
	eco.AddRoute("/api/1/centroDeCostoConstructoras/{id}").Method("GET").Register(getCentroDeCosto)
}
