package main

import (
	"fmt"

	"mouloud.com/thor/internal/configs"
)
func main (){

	err,config:=configs.NewConfigFromYaml("config.yaml")
	if err!=nil{
		panic(err.Error())
	}
	fmt.Print("Running server with init configs")
	fmt.Print(config.String())
}

