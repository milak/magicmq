package env

import (
	"mmq/conf"
	"log"
	"os"
	"github.com/milak/network"
)
type Context struct {
	Running			bool
	Configuration 	*conf.Configuration
	Logger			*log.Logger
	Host			string // will be obtained once connected, not sure it is operationnal
}
func NewContext() *Context {
	var logger *log.Logger
	file, err := os.Create("mmq.log")
	if err != nil {
		logger = log.New(os.Stdout, "-", log.Lshortfile)
		logger.Println("Unable to open file ")
	} else {
		logger = log.New(file, "-", log.Lshortfile)
	}
	host,_ := network.GetLocalIP()
	return &Context{Running : true, Logger : logger, Host : host}
}