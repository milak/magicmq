package service

import (
	"log"
	"bytes"
	"encoding/json"
	"mmq/conf"
	"mmq/env"
	"net"
	"strconv"
	"time"
)
const dieze byte = byte('#')
type instanceConnection struct {
	instance 	*conf.Instance
	connection	*net.Conn
}
func newInstanceConnection (aInstance *conf.Instance, aConnection *net.Conn) *instanceConnection{
	return &instanceConnection{instance : aInstance, connection : aConnection}
}
type SyncService struct {
	running 	bool
	context 	*env.Context
	listener 	net.Listener
	port 		string // will be obtained via configuration
	host		string // will be obtained once connected
	logger		*log.Logger
	connections	map[string]*instanceConnection
}
func NewSyncService (aContext *env.Context) *SyncService{
	result := SyncService{running : true, context : aContext, logger : aContext.Logger}
	result.connections = make(map[string]*instanceConnection)
	return &result
}
func (this *SyncService) doListen (aPort string) {
	this.logger.Println("listening on port",aPort,"...")
	var err error
	this.listener, err = net.Listen("tcp", ":"+aPort)
	if err != nil {
		this.logger.Println("listening failed",err)
	} else {
		for this.running {
			conn, err := this.listener.Accept()
			if err != nil {
				// TODO handle error
			} else {
				this.logger.Println("caught a call")
				go this.handleConnection(conn)
			}
		}
	}
}
func sendCommand(command string, arguments []byte, aConn net.Conn){
	aConn.Write([]byte("#"+command+"#"+strconv.Itoa(len(arguments))+"#"))
	aConn.Write(arguments)
	aConn.Write([]byte("#"))
}
func splitCommand(line []byte) (command string ,arguments []byte, remain []byte) {
	if line[0] != dieze {
		return "",[]byte{},line
	}
	i := 1
	command = ""
	for line[i] != dieze {
		command += string(line[i])
		i++
	}
	i++
	slength := ""
	for line[i] != dieze {
		slength += string(line[i])
		i++
	}
	i++
	length,_ := strconv.Atoi(slength)
	arguments = line[i:i+length]
	remain = line[i+length+1:]
	return command, arguments, remain
}
func (this *SyncService) keepConnected(aInstanceConnection *instanceConnection){
	buffer := make([]byte,1000)
	connection := (*aInstanceConnection.connection)
	instance := aInstanceConnection.instance
	defer func() {
		connection.Close()
		instance.Connected = false
		delete(this.connections,instance.Name())
	}()
	for this.context.Running {
		time.Sleep(1 * time.Second)
		count,err := connection.Read(buffer)
		if err != nil {
			this.logger.Println("Lost connection with",instance.Name(),err)
			break
		}
		command, arguments, remain := splitCommand(buffer[0:count])
		this.logger.Println("Received command " + command,arguments,remain)
	}
}
func (this *SyncService) handleConnection (aConn net.Conn){
	this.host,_,_ = net.SplitHostPort(aConn.LocalAddr().String())
	this.logger.Println("Processing call")
	buffer := make([]byte,1000)
	count,err := aConn.Read(buffer)
	if err != nil {
		this.logger.Println("Unable to read HELLO from remote",err)
		return
	}
	if count < 10 {
		this.logger.Println("Unable to read HELLO from remote ",buffer[0:count])
		sendCommand("ERROR",[]byte("Unable to understand"),aConn)
		return
	}
	command, arguments, remain := splitCommand(buffer[0:count])
	if command == "" {
		this.logger.Println("Unable to read HELLO from remote")
		sendCommand("ERROR",[]byte("Unable to understand"),aConn)
		return
	}
	this.logger.Println("Received ",command,"-",arguments,"-",remain)
	if command != "HELLO" {
		this.logger.Println("Unable to read HELLO from remote ")
		sendCommand("ERROR",[]byte("Unable to understand"),aConn)
		return
	}
	instance := this.context.Configuration.GetInstance(string(arguments))
	if instance == nil {
		host,port,_ := net.SplitHostPort(string(arguments))
		instance = conf.NewInstance(host,port)
		this.context.Configuration.AddInstance(instance)
	}
	instance.Connected = true
	sendCommand("HELLO",[]byte(this.host+":"+this.port),aConn)
	instanceConnection := newInstanceConnection(instance,&aConn)
	this.connections[instance.Name()] = instanceConnection
	this.keepConnected(instanceConnection)
	// TODO : gerer le fait que les deux peuvent essayer de se connecter en même temps, il y aura alors deux connections entre eux
}
/**
 * Scan not connected Instances and try to Connect
 */
func (this *SyncService) scanInstances() {
	time.Sleep(4 * time.Second)
	for this.running {
		for i := range this.context.Configuration.Instances {
			instance := this.context.Configuration.Instances[i]
			if !instance.Connected {
				host := instance.Host+":"+instance.Port
				this.logger.Println("Trying to connect to " + host)
				conn, err := net.Dial("tcp", host)
				if err != nil {
					this.logger.Println("Connection failed ", err)
					continue
				}
				this.logger.Println("Connection successful")
				this.host,_,_ = net.SplitHostPort(conn.LocalAddr().String())
				sendCommand("HELLO",[]byte(this.host+":"+this.port),conn)
				var buffer bytes.Buffer
				encoder := json.NewEncoder(&buffer)
				encoder.Encode(this.context.Configuration.Instances)
				sendCommand("INSTANCES",buffer.Bytes(),conn)
				instance.Connected = true
				instanceConnection := newInstanceConnection(instance,&conn)
				this.connections[instance.Name()] = instanceConnection
				this.keepConnected(instanceConnection)
			}
		}
		time.Sleep(10 * time.Second)
	}
}
func (this *SyncService) Start (){
	for s := range this.context.Configuration.Services {
		service := this.context.Configuration.Services[s]
		if !service.Active {
			continue
		}
		if service.Name == "SYNC" {
			found := false
			this.logger.Println("starting...")
			for p := range service.Parameters {
				if service.Parameters[p].Name == "port" {
					this.port = service.Parameters[p].Value
					this.running = true
					found = true
					go this.doListen(this.port)
					go this.scanInstances()
					break
				}
			}
			if !found {
				this.logger.Panic("missing port parameter")
			}
		}
	}
}
func (this *SyncService) Stop (){
	this.running = false
	if this.listener != nil {
		this.listener.Close()
	}
}