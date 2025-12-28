package configs

import (
	"os"

	"gopkg.in/yaml.v3"
)


func  NewConfigFromYaml(path string)(error,*IngestionConfig){
	out:=IngestionConfig{}
    configFile, err := os.ReadFile(path)
		if err!=nil{
			return err,nil
		}
		if err:=yaml.Unmarshal(configFile,&out);err!=nil {
		return err,nil;
	}
	return nil,&out;
}
