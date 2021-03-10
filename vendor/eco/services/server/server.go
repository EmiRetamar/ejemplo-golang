package server

import (
	"eco/services/config"
	"eco/services/halt"
	"eco/services/logger"
	"eco/services/net"
	"eco/services/persistence"
	"eco/services/session"
	"eco/services/session/check"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/ian-kent/go-log/levels"
	"github.com/ian-kent/go-log/log"
	. "github.com/logrusorgru/aurora"
)

/**
 * User: Santiago Vidal
 * Date: 22/05/17
 * Time: 15:47
 */

var LocalServiceName string
var ServiceListenPort int

type Eco struct {
	router *mux.Router
	port   int
}

func (eco *Eco) AddRoute(path string) *EcoRoute {
	return &EcoRoute{
		server:    eco,
		path:      path,
		recursive: false,
	}
}

func (eco *Eco) Start() {

	eco.addPublicHandler()

	strPort := strconv.Itoa(eco.port)
	fmt.Println(Blue("Service listening on port " + strPort).Bold())
	if localMode, _ := config.EcoCfg.Bool("localMode"); localMode {
		log.Fatal(http.ListenAndServe(":" + strPort, eco.router))
	} else {
		log.Fatal(http.ListenAndServeTLS(":"+strPort, "STAR_finneg_com.crt", "finneg.com.key", eco.router))
	}
	
}

func (eco *Eco) SetNotFoundHandler(handler http.Handler) {
	eco.router.NotFoundHandler = handler
}

func (eco *Eco) addPublicHandler() {

	eco.router.PathPrefix("/web").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		path := "./resources/public" + r.URL.Path[4:]
		if f, err := os.Stat(path); err == nil && !f.IsDir() { //no devuelvo directorios
			http.ServeFile(w, r, path)
		} else {
			eco.router.NotFoundHandler.ServeHTTP(w, r)
		}

	})
}

type notFoundHandler struct{}

var ecoNotFoundHandler notFoundHandler

func (h notFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := halt.Errorf("api not found", 404)
	err.HTTPError(w)
}

type EcoRoute struct {
	server    *Eco
	path      string
	method    string
	recursive bool
	security  bool
}

func (url *EcoRoute) Recursive() *EcoRoute {
	url.recursive = true
	return url
}

func (url *EcoRoute) Method(method string) *EcoRoute {
	url.method = method
	return url
}

func (url *EcoRoute) WithSecurity() *EcoRoute {
	url.security = true
	return url
}

func (url *EcoRoute) RegisterNoSession(f http.HandlerFunc) {

	method := url.method
	if len(method) == 0 {
		method = "GET"
	}

	var route *mux.Route
	if url.recursive {
		route = url.server.router.Methods(method).PathPrefix(url.path)
	} else {
		route = url.server.router.Methods(method).Path(url.path)
	}
	route.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer halt.HandlePanics(writer)
		f(writer, request)
	})

	log.Info("Route registered (with no session): " + method + " " + url.path)

}

func (url *EcoRoute) Register(endPoint net.EcoApiHandler) {

	method := url.method
	if len(method) == 0 {
		method = "GET"
	}

	var route *mux.Route
	if url.recursive {
		route = url.server.router.Methods(method).PathPrefix(url.path)
	} else {
		route = url.server.router.Methods(method).Path(url.path)
	}

	//la cadena de mando del entry point
	//apicheck(el token o sessionID)->logueo->seguridad->tipo-de-ejecucion->ejecucion(variable 'endPoint')
	var handler net.EcoApiHandler
	if url.security {

		if strings.Index(url.path, "/{featureID}") == -1 {
			err := fmt.Sprintf("\nNo se puede registrar la API %s '%s' con seguridad por no contener un {featureID} en su URL", url.method, url.path)
			panic(err)
		}

		handler = logger.LogHandler(check.SecurityCheck(check.ExecutionCheck(endPoint)), "default")
	} else {
		handler = logger.LogHandler(check.ExecutionCheck(endPoint), "default")
	}

	if len(url.path) > 5 && url.path[0:6] == "/admin" {
		route.Handler(check.AdminCheck(handler))

	} else if len(url.path) > 3 && url.path[0:4] == "/api" {
		route.Handler(check.ApiCheck(handler))

	} else if len(url.path) > 4 && url.path[0:5] == "/auth" {
		route.Handler(check.ApiCheck(handler))

	} else {
		route.Handler(check.LoginCheck(handler))
	}

	log.Info("Route registered: " + method + " " + url.path)
}

func NewEcoServer(router *mux.Router, port int) *Eco {
	return &Eco{
		router: router,
		port:   port,
	}
}

func New(name string) *Eco {

	LocalServiceName = name

	log.Log(levels.INFO, Red("Starting eco "+name+" service...").Bold())
	if err := config.LoadConfig(); err != nil {
		log.Fatal(err)
		return nil
	}
	logger.Init()
	ServiceListenPort, _ = config.EcoCfg.Int("listenPort")

	if debug, _ := config.EcoCfg.Bool("debug"); debug {
		log.Debug(name + " startup configuration:\n" + config.Info() + "\n")
	} else {
		log.Info(name + " startup configuration:\n" + config.Info() + "\n")
	}

	session.Init()

	//si el servicio no requiere de Redis no inicializo persistencia
	if host, _ := config.EcoCfg.String("database.redis.host"); len(host) > 0 {
		if err := persistence.Init(); err != nil {
			log.Fatal(err)
		}
	}

	router := mux.NewRouter().StrictSlash(true)
	eco := NewEcoServer(router, ServiceListenPort)
	eco.router.NotFoundHandler = ecoNotFoundHandler

	//ping api que tienen todos los servicios para los balanceadores
	eco.AddRoute("/ping").RegisterNoSession(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write([]byte(`{ "status": "ok" }`))
	})

	if localMode, _ := config.EcoCfg.Bool("localMode"); localMode {

		//en localMode guardo en /tmp en que puerto esta corriendo el servicio

		dir := "/tmp/eco-services/"
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			log.Fatal(err)
		}

		if err := ioutil.WriteFile(dir+name, []byte(strconv.Itoa(ServiceListenPort)), os.ModePerm); err != nil {
			log.Fatal(err)
		}

		eco.AddRoute("/who").RegisterNoSession(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, `{ "who": "%s" }`, name)
		})

	}

	return eco
}
