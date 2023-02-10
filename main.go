package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type Server interface {
	Address() 	string
	IsAlive()	bool
	Server(rw http.ResponseWriter, r *http.Request)
}

 type demoServer struct {
	address string 
	proxy *httputil.ReverseProxy
 }


 func newServer(address string) *demoServer  {
	serverUrl, err := url.Parse(address)
	handleErr(err)

	return &demoServer{
		address: address,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
 }


type LoadBalancer struct {
	port 				string
	roundRobinCounter   int
	servers 				[]Server
 }


 func NewLoadBalancer(port string, server []Server) *LoadBalancer {
	return &LoadBalancer{
		port: port,
		roundRobinCounter: 0,
		servers: server ,
	}
	
 }

 func handleErr(err error)  {
	if err!=nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
 }


func (s *demoServer) Address() string { return s.address}
func (s *demoServer) IsAlive() bool { return true}
func (s *demoServer) Server(rw http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(rw, r)
}

func (lb *LoadBalancer) getNextAvailableServer() Server {
	server := lb.servers[lb.roundRobinCounter%len(lb.servers)]
	for !server.IsAlive(){
		lb.roundRobinCounter++
		server = lb.servers[lb.roundRobinCounter%len(lb.servers)]
	}
    lb.roundRobinCounter++
	return server
}

func (lb *LoadBalancer) serverProxy(rw http.ResponseWriter, r *http.Request)  {
	targetServer := lb.getNextAvailableServer()
	fmt.Printf("Forwarding request to address %q\n", targetServer.Address())
	targetServer.Server(rw, r)
}

//MAIN func ------------------------------------------------
func main()  {
	servers := []Server{
		newServer("http://www.bing.com"),
		newServer("http://www.duckduckgo.com"),
		newServer("https://www.google.com"),
	}

	lb := NewLoadBalancer("8000", servers)
	handleRedirect := func (rw http.ResponseWriter, r *http.Request)  {
		lb.serverProxy(rw, r)
	}

	http.HandleFunc("/", handleRedirect)

	fmt.Printf("Serving at 'Localhost:%s'\n", &lb.port)
	http.ListenAndServe(":"+lb.port, nil)
}


 