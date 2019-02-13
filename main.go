package goeureka

import (
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
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

var logger = logrus.New()

func Init(config EurekaConfig) EurekaClient {
	logger.Out = os.Stdout
	logger.SetLevel(logrus.DebugLevel)

	config.IpAddress = getOutboundIP().String()
	config.VipAddress = getOutboundIP().String()
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

func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		logger.Warn(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func handleSigterm(config EurekaConfig) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		deregister(config)
		os.Exit(1)
	}()
}

func CombineRoutes(routes Routes, eurekaRouts Routes) Routes {
	for _, route := range eurekaRouts {
		routes = append(routes, route)
	}
	return routes
}

func startWebServer(routes Routes, port string) {
	router := http.NewServeMux()
	router = BuildRoutes(routes, router)
	logger.Info("Server is up and listening on ", port)
	http.ListenAndServe(port, router)
}
