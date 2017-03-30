package env

import (
	"mmq/conf"
	"mmq/item"
	"log"
	"os"
)

type Context struct {
	Running			bool
	Store 			item.ItemStore
	Configuration 	conf.Configuration
	Logger			*log.Logger
}
func NewContext() *Context {
	return &Context{Running : true, Logger : log.New(os.Stdout, "-", log.Lshortfile)}
}