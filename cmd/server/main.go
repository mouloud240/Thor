package main

import (
	"bufio"
	"log"
	"net"
	"mouloud.com/thor/internal/configs"
	"mouloud.com/thor/internal/ingestion"
	"mouloud.com/thor/internal/pipeline"
)
func main (){
	//Load init config
	err,config:=configs.NewConfigFromYaml("config.yaml")
	if err!=nil{
		panic(err.Error())
	}
	//Run tcp server
  if err:=ingestion.RunServer(config.Server.TcpPort,func(conn net.Conn,ch chan<- error){
	  pipeLine:=pipeline.NewPipeLine(config.Storage.LogDir,config.Storage.SegmentSize,config.Pipeline.NumWorkers)
		workChan,errChan,err:=pipeLine.StartWorkers()
		if err!=nil{
			ch<-err
		}
		go func (){
			for err:= range errChan{
				ch<-err
			}
		}()
		
		defer conn.Close()
scanner := bufio.NewScanner(conn)
if scanner.Scan() {
    req := scanner.Text()
    workChan <- req
}
if err := scanner.Err(); err != nil {
    ch <- err
    return
}		
if err!=nil{
			ch<-err
			return
		}

	}) ;err!=nil{
		log.Fatal(err.Error())
	}
}

