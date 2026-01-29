package main

import (
	"bufio"
	"net"
	"mouloud.com/thor/internal/configs"
	"mouloud.com/thor/internal/ingestion"
	"mouloud.com/thor/internal/ingestion/client"
)
func main (){
	//Load init config
	err,config:=configs.NewConfigFromYaml("config.yaml")
	if err!=nil{
		panic(err.Error())
	}
	//Run tcp server
  if err:=ingestion.RunServer(config.Server.TcpPort,config,
	func(conn net.Conn,workChan chan<- client.Client,errChan <-chan error ){
defer conn.Close()
// Create Response Channel
res:=make(chan string)
//Scan request body and start worker 
scanner := bufio.NewScanner(conn)
if scanner.Scan() {
    req := scanner.Text()
		client:=client.NewClient(req,res)
    workChan <- *client
}
if err := scanner.Err(); err != nil {
	conn.Write([]byte(err.Error()))
	return;
}
if err!=nil{
	conn.Write([]byte(err.Error()))
	return;
		}
for {
	select {
	case out:=<-res:
		conn.Write([]byte(out))
		return
	case err:=<-errChan:
		conn.Write([]byte(err.Error()))
		return
	}
}
	});
	// Run server Error
	err!=nil{
		panic(err.Error())
	}
	
}

