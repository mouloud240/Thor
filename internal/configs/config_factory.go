package configs

import (
	"os"

	"gopkg.in/yaml.v3"
)


//TODO :add defaults
func  NewConfigFromYaml(path string)(error,*IngestionConfig){
	out:=IngestionConfig{}
    configFile, err := os.ReadFile(path)
		if err!=nil{
			return err,nil
		}
		if err:=yaml.Unmarshal(configFile,&out);err!=nil {
		return err,nil;
	}
	out.Storage.SegmentSize=out.Storage.SegmentSize*1000 //Convert Kb to bytes
	return nil,&out;
}
