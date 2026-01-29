package pipeline

import (
	"errors"
	"strings"
	"time"
)

type LogLevel int

const (
    DEBUG LogLevel = iota  
    INFO                   
    WARN                   
    ERROR                  
    FATAL                  
		NONE
)

func (l LogLevel) String() string {
    return [...]string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}[l]
}
func LevelFromString(raw string) ( error,LogLevel ){
	 switch raw {
	 case "INFO":
		 return nil,INFO
	case "DEBUG":
		return nil,DEBUG
	case "WARN":
		return  nil,WARN
	case "ERROR":
		return nil,ERROR
	case "FATAL":
		return nil,FATAL
	default:
		return errors.New("Unsupported Log Level"),NONE
	 }
}

type Log struct{
	Level LogLevel
  TimeStamp time.Time
	Service string
	Message string
	Metadata map[string]any
}
func  FromString(input string)(*Log,error){
	parts:=strings.Split(input, "|")
	
	if (len(parts)<4){
		return nil,errors.New("Please Provide the correct log format ")
	}
	err,logLevel:=LevelFromString(parts[1])
	if err!=nil{
		return nil,err
	}
	timeStamp,err:=time.Parse(time.RFC3339,parts[0])
	if err!=nil{
		return nil,err
	}
	//NOTE:Ignoring the Metadata currently until figuring out the best approach of sending
	log:=Log{
		Level: logLevel,
		TimeStamp: timeStamp,
		Message: parts[3],
		Service: parts[2],
	}
	return &log,nil
	}

