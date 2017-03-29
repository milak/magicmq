package service

import (
	"mmq/conf"
	"mmq/item"
	"net"
	"log"
)
type SyncService struct {
	running bool
	configuration *conf.Configuration
}
func NewSyncService (aConfiguration *conf.Configuration, aStore *item.ItemStore) *SyncService{
	return &SyncService{running : true, configuration : aConfiguration}
}
func (this *SyncService) doListen (aPort string) {
	log.Println("SyncService listening on port",aPort,"...")
	ln, err := net.Listen("tcp", ":"+aPort)
	if err != nil {
		// handle error
	} else {
		for this.running {
			conn, err := ln.Accept()
			if err != nil {
				// TODO handle error
			} else {
				log.Println("SyncService caught a call")
				go this.handleConnection(conn)
			}
		}
	}
}
func (this *SyncService) handleConnection (aConn net.Conn){
	aConn.Write([]byte("HELLO"))
	aConn.Write([]byte("BYE"))
	aConn.Close()
}
func (this *SyncService) Start (){
	for s := range this.configuration.Services {
		service := this.configuration.Services[s]
		if !service.Active {
			continue
		}
		if service.Name == "SYNC" {
			for p := range service.Parameters {
				if service.Parameters[p].Name == "root" {
					port := service.Parameters[p].Value
					this.running = true
					go this.doListen(port)
					break
				}
			}
		}
	}
}
func (this *SyncService) Stop (){
	this.running = false
}