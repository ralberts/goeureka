package goeureka

import (
	"fmt"
	"github.com/labstack/echo/middleware"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
)

type EurekaClient struct {
	Client Eureka
	Routes Routes
}

type EurekaConfig struct {
	Name        string
	Url         string
	VipAddress  string
	IpAddress   string
	HostName    string
	Port        string
	SecurePort  string
	RestService bool
}

func Init(config EurekaConfig) EurekaClient {
	handleSigterm(config) // Graceful shutdown on Ctrl+C or kill
	routes := routes
	go Register(config) // Performs Eureka registration
	// start server and Block if not a rest service...
	if !config.RestService {
		go startWebServer(routes, config.Port)
		wg := sync.WaitGroup{} // Use a WaitGroup to block main() exit
		wg.Add(1)
		wg.Wait()
	}

	var e Eureka
	return EurekaClient{Client: e, Routes: routes}
}

func handleSigterm(config EurekaConfig) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		Deregister(config)
		os.Exit(1)
	}()
}

func CombineRoutes(routes Routes, eurekaRouts Routes) Routes {

	for _, route := range eurekaRouts {
		routes = append(routes, route)
	}
	return routes
}

func PrintRoutes(e *echo.Echo) {
	for _, route := range e.Routes() {
		fmt.Println(fmt.Sprintf("Mapped (%s) with method (%s) to %s", route.Path, route.Method, route.Name))
	}
}

func startWebServer(router Routes, port string) {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Logger.SetLevel(log.INFO)

	e = BuildRoutes(routes, e)
	//log.Println("Starting HTTP service at " + port)
	e.Logger.Fatal(e.Start(":" + port))
}
