package server

import (
	"fmt"
	"log"
	"net"
)
func RunServer(port int,connHandler func(net.Conn))error{
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
		go connHandler(conn)
	}
}
