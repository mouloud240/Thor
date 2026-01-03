package main

import (
	"strings"
	"io"
	"log"
	"net"

	"mouloud.com/thor/internal/configs"
	"mouloud.com/thor/internal/pipeline"
	"mouloud.com/thor/internal/server"
)
func main (){

	//Load init config
	err,config:=configs.NewConfigFromYaml("config.yaml")
	if err!=nil{
		panic(err.Error())
	}
	//Run tcp server
  if err:=server.RunServer(config.Server.TcpPort,func(conn net.Conn,ch chan<- error){
		//This is just a test hardcoded function until the time comes to start working on the ingestion pipeline besides parsing
		defer conn.Close()
	  var req strings.Builder

buff:=make([]byte,64)
read_loop:
		for {
		n,err:=conn.Read(buff)
		if err!=nil{
			if err==io.EOF {
				//this is probaly not nessecary but I am paranoid
				break read_loop
			}else {
				ch<-err
				//We would love to just return the error but this handler will be run on a seperate go routine more on that in @internal/server/tcp_server.go
				return
			}
		}
		req.WriteString(string(buff[:n]))
		}
				//Just a test for logs and concurrency
		logEntry,err:=pipeline.FromString(req.String())
		if err!=nil{
			// It is fine if it panics sine it is on the same go routine
			ch<-err
			return
		}
		log.Print(logEntry)

	}) ;err!=nil{
		log.Fatal(err.Error())
	}
}

