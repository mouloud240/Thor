package pipeline

import (
	lg "log"
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
	engine,err:=storage.NewEngine(p.Logs_dir,p.Segment_size,p.WorkersNumber)
	if err!=nil{
		return nil,nil,err
	}
	work:=make(chan client.Client ,p.WorkersNumber)
	errChan:=make(chan error,p.WorkersNumber) 


go engine.StartWorkers();
	for range p.WorkersNumber{
go p.startPipeLine(work, errChan,engine)
	}

	return work,errChan,nil
}
func (p *PipeLine) startPipeLine(clients <-chan client.Client,errChan chan<- error,engine *storage.Engine){
	for client:=range clients{
	 parsed,err:=FromString(client.Req)
	 if (err!=nil){
		 			 lg.Printf("Error parsing log: %v\n", err)
		 errChan <- err
		 return;
	 }
	 engine.WorkersChan<-storage.NewWorkerInput(errChan,parsed.ToStorageFormat(),client.ResChan)

	}

}
