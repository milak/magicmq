package env

import (
	"mmq/conf"
	"mmq/item"
	"log"
	"os"
)

type Context struct {
	Running			bool
	Store 			*item.ItemStore
	Configuration 	*conf.Configuration
	Logger			*log.Logger
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
	return &Context{Running : true, Logger : logger}
}