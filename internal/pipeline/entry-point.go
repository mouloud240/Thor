package pipeline

import (
	"mouloud.com/thor/internal/ingestion/client"
	"mouloud.com/thor/internal/storage"
)
type PipeLine struct{
	Logs_dir string
	Segment_size int64
	WorkersNumber int
}
func  NewPipeLine(Logs_dir string,Segment_size int64,WorkersNumber int) *PipeLine{
	return &PipeLine{Logs_dir: Logs_dir,Segment_size: Segment_size,WorkersNumber: WorkersNumber}
}
func (p *PipeLine) StartWorkers() (chan<- client.Client,<-chan error,error) {
	engine,err:=storage.NewEngine(p.Logs_dir,p.Segment_size)
	if err!=nil{
		return nil,nil,err
	}
	work:=make(chan client.Client ,p.WorkersNumber)
	errChan:=make(chan error,p.WorkersNumber) 
	for range p.WorkersNumber{
go p.startPipeLine(work, errChan,engine)
	}

	return work,errChan,nil
}
func (p *PipeLine) startPipeLine(clients <-chan client.Client,errChan chan<- error,engine *storage.Engine){
	for client:=range clients{
	 parsed,err:=FromString(client.Req)
	 if (err!=nil){
		 errChan <- err
		 return;
	 }
	 go engine.StoreLog(parsed.ToString())
	 	 client.ResChan<-"Saved Succefully\n"

	}

}
