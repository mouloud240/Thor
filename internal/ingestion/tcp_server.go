package ingestion

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"mouloud.com/thor/internal/configs"
	"mouloud.com/thor/internal/ingestion/client"
	"mouloud.com/thor/internal/pipeline"
)
func RunServer(port int,config *configs.IngestionConfig,connHandler func(net.Conn,chan<- client.Client, <-chan error,context.Context))error{
	
	pipeLine:=pipeline.NewPipeLine(config.Storage.LogDir,config.Storage.SegmentSize,config.Pipeline.NumWorkers)
  ch,errCh,err:=pipeLine.StartWorkers()
	if err!=nil{
		return err
	}

	//Error channel 
	//the number of bounding is hardcoded rn until some more iterations
	ln,err:=net.Listen("tcp4",fmt.Sprintf(":%d",port))
	if err!=nil{
		return err;
	}
	log.Print("Server Running on port: ",port)
	//Accept loop
	for {
		ctx,_:=context.WithTimeout(context.Background(),time.Duration(5000))

		conn,err:=ln.Accept()
		if err!=nil{
			return err
		}
		//Current setups spins go routine for each connection , later this will be converted into a bounded worker pool pattern if need be
		log.Print("Established new connection" ,conn.LocalAddr())
		go connHandler(conn,ch,errCh,ctx)
	}
}
