// Storage Engine
package storage

import (
	"bufio"
	"log"
	lg "log"
	"os"
	"strconv"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type WorkerInput struct{
	errChan chan<- error
	log string
	resChan chan<- string
}
func NewWorkerInput(errChan chan<- error , log string , resChan chan<- string)WorkerInput{
	return WorkerInput{errChan: errChan,log: log,resChan: resChan}
}
type Engine struct {
	Dir string
  SegemntSize int64 
	SegmentsCount int
	numWorkers int
	WorkersChan chan WorkerInput
	mu sync.RWMutex
}
func NewEngine(dir string,SegemntSize int64,numWorkers int)(*Engine,error){
	workChan:=make(chan WorkerInput,numWorkers)

if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	segmentCount:=0
	for {
		fileName:=dir+"/segment_"+strconv.Itoa(segmentCount)+".log"
		stat,err:=os.Stat(fileName)
		if  os.IsNotExist(err) {
			break
		}
		if stat.Size()<SegemntSize{
			break
		}
	
		segmentCount++
	}
		return &Engine{Dir: dir,SegemntSize: SegemntSize,SegmentsCount:segmentCount,numWorkers: numWorkers,WorkersChan: workChan},nil
}
func (e *Engine) StartWorkers(){
	go func(){

		watcher, err := fsnotify.NewWatcher()
		if err!=nil{
			lg.Fatalln("Failed to start watcher",err.Error())
			return;
		}
		go func(){
for {
	select {
	case event,ok :=<-watcher.Events:
		if !ok{
			return
		}
		if event.Has(fsnotify.Write){
			//Quitely  increment count if needed
			lg.Println("Checking if new segment is needed")

			if e.shouldCreateNewSegement(e.getSegmentCount()){
e.mu.Lock()
e.NewSegment();
			e.mu.Unlock()
			}
			lg.Println("Check done")
					}
	case err,ok:=<-watcher.Errors:
		if !ok{
			return;
		}
		log.Println(err.Error())
	}

		}
		}()
		if err:=watcher.Add("dump");err!=nil{
			lg.Fatalln("Failed to add path",err.Error())
		}
    <-make(chan struct{})
	}()
	for range e.numWorkers{
		go func(){
			for work:=range e.WorkersChan{
				if err:=e.StoreLog(work.log);err!=nil{
					work.errChan<-err
				}
				work.resChan<-"Log Stored Successfully"
			}
	}()
}}
func (e *Engine) getSegmentCount() int{

	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.SegmentsCount

}
func (e *Engine)NewSegment()error{
	lg.Printf("Creating new segment file: %s\n", e.GetFileName(e.SegmentsCount))
	e.SegmentsCount++
	fileName:=e.GetFileName(e.SegmentsCount)
_, err := os.OpenFile(fileName, os.O_CREATE, 0644)		

	return err;
}
func (e *Engine)GetFileName(segment int)string{
	return  e.Dir+"/segment_"+strconv.Itoa(segment)+".log"
}
func (e *Engine) shouldCreateNewSegement(currSegment int)bool{
		fileName:=e.GetFileName(currSegment)
	fileInfo, err := os.Stat(fileName)
	if err!=nil{
		return false
	}
	currentSize:=fileInfo.Size()
	if currentSize>=e.SegemntSize{
		return true;
	}
		return false
	
}
func (e *Engine)findOrCreateSegment()(* os.File,error){
fileName:=e.GetFileName(e.getSegmentCount())
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)		
	if err!=nil{
		return nil,err
	}
	return file,nil
}
func (e *Engine)StoreLog(log string)error{
		file,err:=e.findOrCreateSegment()
	if err!=nil{
		return err;
	}
	defer file.Close()
	writer:=bufio.NewWriter(file)
	_,err=writer.WriteString(log)

	if err!=nil{
		return err
	}
	if err=writer.Flush();err!=nil{
		return err
	}
	err = file.Sync()
	if err != nil {
		return err
	}
	return nil
}

