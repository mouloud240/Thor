package main

import (
	"log"
	"net"

	"mouloud.com/thor/internal/configs"
	"mouloud.com/thor/internal/server"
)
func main (){

	//Load init config
	err,config:=configs.NewConfigFromYaml("config.yaml")
	if err!=nil{
		panic(err.Error())
	}
	//Run tcp server
  if err:=server.RunServer(config.Server.TcpPort,func(conn net.Conn){
		log.Print("Received new Connection")
	}) ;err!=nil{
		panic(err.Error())
	}
}

