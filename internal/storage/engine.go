// Storage Engine
package storage

import (
	"bufio"
	"os"
	"strconv"
)
type Engine struct {
	Dir string
  SegemntSize int64 
	SegmentsCount int
}
func NewEngine(dir string,SegemntSize int64)(*Engine,error){
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
	return &Engine{Dir: dir,SegemntSize: SegemntSize,SegmentsCount:segmentCount},nil
}
func (e *Engine)NewSegment()error{
	e.SegmentsCount++
	fileName:=e.GetFileName()
_, err := os.OpenFile(fileName, os.O_CREATE, 0644)		
	return err;
}
func (e *Engine)GetFileName()string{
	return e.Dir+"/segment_"+strconv.Itoa(e.SegmentsCount)+".log"
}
func (e *Engine) shouldCreateNewSegement(log string)bool{
	fileName:=e.GetFileName()
	fileInfo, err := os.Stat(fileName)
	if err!=nil{
		return false
	}
	currentSize:=fileInfo.Size()
	newLogSize:=int64(len([]byte(log+"\n")))
	if currentSize+newLogSize>e.SegemntSize{
		return true
	}
	return false
	
}
func (e *Engine)findOrCreateSegment()(* os.File,error){
fileName:=e.GetFileName()
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)		
	if err!=nil{
		return nil,err
	}
	return file,nil
}
func (e *Engine)StoreLog(log string)error{
	
	if e.shouldCreateNewSegement(log){
		if err:=e.NewSegment();err!=nil{
			return err
		}
	}
	file,err:=e.findOrCreateSegment()
	if err!=nil{
		return err;
	}
	defer file.Close()
	writer:=bufio.NewWriter(file)
	_,err=writer.WriteString(log+"\n")

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

