package client

import "log"

type Client struct{
	Req string
	ResChan chan<- string
}
func NewClient(req string,resChan chan<- string)*Client{
	log.Print("Received request: ",req)
	if (req[0:5]=="MULTI"){
		return nil
	}
     return &Client{Req: req,ResChan: resChan}
}

