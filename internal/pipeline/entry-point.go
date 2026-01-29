package pipeline

import (
	Logger "log"
	"mouloud.com/thor/internal/ingestion/client"
)
type PipeLine struct{
	Logs_dir string
	Segment_size float64
	WorkersNumber int
}
func  NewPipeLine(Logs_dir string,Segment_size float64,WorkersNumber int) *PipeLine{
	return &PipeLine{Logs_dir: Logs_dir,Segment_size: Segment_size,WorkersNumber: WorkersNumber}
}
func (p *PipeLine) StartWorkers() (chan<- client.Client,<-chan error,error) {
	work:=make(chan client.Client ,p.WorkersNumber)
	errChan:=make(chan error,p.WorkersNumber) 
	for range p.WorkersNumber{
go p.startPipeLine(work, errChan)
	}

	return work,errChan,nil
}
func (p *PipeLine) startPipeLine(clients <-chan client.Client,errChan chan<- error){
	for client:=range clients{
		Logger.Print("Got new Log")
	 parsed,err:=FromString(client.Req)
	 if (err!=nil){
		 errChan <- err
		 return;
	 }
	 Logger.Printf("%s", "Received log on Service: "+parsed.Service)
	 Logger.Printf("%s", "With message: "+parsed.Message)
	 client.ResChan<-"Saved Succefully\n"

	}

}
