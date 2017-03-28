package web

import (
	"mmq.types"
)
var running = true
var configuration *types.Configuration
func doListen (aPort string) {
	ln, err := net.Listen("tcp", ":"+aPort)
	if err != nil {
		// handle error
	} else {
		for running {
			conn, err := ln.Accept()
			if err != nil {
				// TODO handle error
			} else {
				go handleConnection(conn)
			}
		}
	}
}
func StartSyncListener (aConfiguration *types.Configuration, aStore *item.ItemStore){
	configuration = aConfiguration
	for s := range configuration.Services {
		service := configuration.Services[s]
		if !service.Active continue
		if service.Name == "SYNC" {
			for p := range service.Parameters {
				if service.Parameters[p].Name == "root" {
					port := &service.Parameters[p].Value
					go doListen(port)
					break
				}
			}
		}
	}
}
func StopSyncListener (){
	running = false
}