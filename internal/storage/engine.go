// Storage Engine
package storage

import (
	"bufio"
	"os"
	"strconv"

)
type Engine struct {
	Dir string
  SegemntSize float64
	SegmentsCount int
}
func NewEngine(dir string,SegemntSize float64)*Engine{
	return &Engine{Dir: dir,SegemntSize: SegemntSize,SegmentsCount:0}
}
func (e *Engine)findOrCreateSegment()(* os.File,error){
fileName:=e.Dir+"/segment_"+strconv.Itoa(e.SegmentsCount)+".log"
  _,err:=os.Open(fileName)
	if err!=nil{
		if os.IsNotExist(err){
			e.SegmentsCount++
		}
	}
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

