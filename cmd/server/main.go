package main

import (
	"bufio"
	"context"
	"log"
	"net"

	"mouloud.com/thor/internal/configs"
	"mouloud.com/thor/internal/ingestion"
	"mouloud.com/thor/internal/ingestion/client"
	"mouloud.com/thor/utils"
)

func main() {
	//Load init config
	err, config := configs.NewConfigFromYaml("config.yaml")
	if err != nil {
		panic(err.Error())
	}
	//Run tcp server
	if err := ingestion.RunServer(config.Server.TcpPort, config,
		func(conn net.Conn, workChan chan<- client.Client, errChan <-chan error, ctx context.Context) {
			_ = ctx
			defer conn.Close()

			scanner := bufio.NewScanner(conn)
			multiMode := false
			remaining := 1
			firstLine := true

			for scanner.Scan() {
				req := scanner.Text()

				if firstLine {
					firstLine = false
					if n, isMulti := utils.ExtractMultiRequestParams(req); isMulti {
						multiMode = true
						remaining = n
						if remaining <= 0 {
							return
						}
						continue
					}
				}

				res := make(chan string)
				c := client.NewClient(req, res)
				if c == nil {
					log.Print("Skipping invalid request")
					continue
				}

				workChan <- *c

				select {
				case out := <-res:
					if _, wErr := conn.Write([]byte(out)); wErr != nil {
						log.Print(wErr.Error())
						return
					}

					if multiMode {
						remaining--
						if remaining == 0 {
							return
						}
					} else {
						return
					}
				case runErr := <-errChan:
					log.Print(runErr.Error())
					conn.Write([]byte(runErr.Error()))
					return
				}
			}

			if scanErr := scanner.Err(); scanErr != nil {
				conn.Write([]byte(scanErr.Error()))
			}
		}); err != nil {
		panic(err.Error())
	}

}
