package main

import (
	"flag"

	"github.com/byeoru/kania/init/cmd"
	_ "github.com/lib/pq"
)

/*
[command line에서 file path 유동적 설정 가능]
ex) config file이 cmd 폴더 안에 있을 경우

	go run main.go -config=./cmd/config.toml
*/
var configPathFlag = flag.String("config", "init/config.toml", "config file")

func main() {
	flag.Parse()
	cmd.NewCmd(*configPathFlag)
}
