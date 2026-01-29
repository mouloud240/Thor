package ingestion

import (
	"fmt"
	"log"
	"net"
)
func RunServer(port int,connHandler func(net.Conn,chan <-error))error{
	//Error channel 
	//the number of bounding is hardcoded rn until some more iterations
	errChan:=make(chan(error),5)
	defer close(errChan)
	ln,err:=net.Listen("tcp4",fmt.Sprintf(":%d",port))
	if err!=nil{
		return err;
	}
	log.Print("Server Running on port: ",port)
	//Accept loop
	for {
		conn,err:=ln.Accept()
		if err!=nil{
			return err
		}
		//Current setups spins go routine for each connection , later this will be converted into a bounded worker pool pattern if need be
		log.Print("Established new connection" ,conn.LocalAddr())
		go connHandler(conn,errChan)
		go func(){
			for err :=range errChan{
			conn.Close()
			//Add better alert system later
			log.Fatal(err.Error())
		}
	}()
	}
}
