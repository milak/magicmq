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
	var byteBuffer *bytes.Buffer
	buffer := make([]byte,1000)
	connection := (*aInstanceConnection.connection)
	instance := aInstanceConnection.instance
	defer func() {
		connection.Close()
		instance.Connected = false
		delete(this.connections,instance.Name())
	}()
	var command string
	var arguments, remain []byte
	for this.context.Running {
		time.Sleep(1 * time.Second)
		if len(remain) > 0 {
			buffer = remain
		} else {
			count,err := connection.Read(buffer)
			if err != nil {
				this.logger.Println("Lost connection with",instance.Name(),err)
				break
			}
			buffer = buffer[0:count]
		}
		command, arguments, remain = splitCommand(buffer)
		this.logger.Println("Received command " + command)
		if command == "HELLO" { // On est côté appelant, on reçoit la réponse de l'appelé, on lui envoie la configuration
			this.sendConfiguration(aInstanceConnection)
		} else if command == "INSTANCES" {
			var newInstances []conf.Instance
			byteBuffer = bytes.NewBuffer(arguments)
			decoder := json.NewDecoder(byteBuffer)
			decoder.Decode(&newInstances)
			for _,instance := range newInstances {
				this.logger.Println("Received instance :",instance)
				if (instance.Host == this.host) && (instance.Port == this.port) {
					this.logger.Println("Skipped instance cause it is me :)")
					continue
				}
				instance.Connected = false // ensure the Instance will not be considered as connected
				this.context.Configuration.AddInstance(&instance)
				this.logger.Println("Added instance :",instance)
			}
		} else if command == "TOPICS" {
			var distributedTopics []conf.Topic
			byteBuffer = bytes.NewBuffer(arguments)
			decoder := json.NewDecoder(byteBuffer)
			decoder.Decode(&distributedTopics)
			for _,topic := range distributedTopics {
				this.logger.Println("Received topic :",topic)
			}
		} else if command == "ERROR" {
			this.logger.Println("Received ERROR :",arguments)
		} else {
			this.logger.Println("Not supported command")
			sendCommand("ERROR",[]byte("NOT SUPPORTED COMMAND"),*aInstanceConnection.connection)
		}
	}
}
/**
 * Send configuration to other side :
 *   * the known instances
 *   * the distributed topics
 */
func (this *SyncService) sendConfiguration(aInstanceConnection *instanceConnection){
	this.logger.Println("Sending configuration")
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.Encode(this.context.Configuration.Instances)
	sendCommand("INSTANCES",buffer.Bytes(),*aInstanceConnection.connection)
	buffer.Reset()
	var distibutedTopics []*conf.Topic
	for _,topic := range this.context.Configuration.Topics {
		if topic.IsDistributed() {
			distibutedTopics = append(distibutedTopics,topic)
		}
	}
	if len (distibutedTopics) > 0 {
		encoder.Encode(distibutedTopics)
		this.logger.Println("Sending topics ", string(buffer.Bytes()))
		sendCommand("TOPICS",buffer.Bytes(),*aInstanceConnection.connection)
	}
}
/**
 * Process the connection when called by a remote node.
 */
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
	for len(remain) > 0 {
		command, arguments, remain = splitCommand(remain)
		this.logger.Println("Received command " + command,arguments,remain)
	}
	sendCommand("HELLO",[]byte(this.host+":"+this.port),aConn) // TODO échanger leur numéros de version
	instanceConnection := newInstanceConnection(instance,&aConn)
	this.connections[instance.Name()] = instanceConnection
	this.sendConfiguration(instanceConnection)
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
				instance.Connected = true
				instanceConnection := newInstanceConnection(instance,&conn)
				this.connections[instance.Name()] = instanceConnection
				go this.keepConnected(instanceConnection)
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