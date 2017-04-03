package service

import (
	"io"
	"log"
	"bytes"
	"encoding/json"
	"mmq/conf"
	"mmq/env"
	"net"
	"strconv"
	"time"
)
/**
 * The source part is about synchronisation between instances
 */
// A const for DIEZE value
const dieze byte = byte('#')
/**
 * An internal structure used for the map. It links an instance to a connection
 */
type instanceConnection struct {
	instance 	*conf.Instance	// an instance
	connection	*net.Conn		// a connection opened with the instance
}
/**
 * Constructor for InstanceConnection
 */
func newInstanceConnection (aInstance *conf.Instance, aConnection *net.Conn) *instanceConnection{
	return &instanceConnection{instance : aInstance, connection : aConnection}
}
/**
 * The main class of the source code
 */
type SyncService struct {
	running 	bool							// boolean indicating if the service is running, setting it to false, should stop listening 
	context 	*env.Context					// a reference to the context, usefull to get accès to store, logger and configuration
	listener 	net.Listener					// a reference to the listener when doListen has been called
	port 		string 							// will be obtained via configuration
	logger		*log.Logger						// the logger obtained from context, it is copied here for code readability reason
	connections	map[string]*instanceConnection	// a map that links instances to opened net connection
}
/**
 * Constructor for the SyncService class
 */
func NewSyncService (aContext *env.Context) *SyncService {
	result := &SyncService{running : true, context : aContext, logger : aContext.Logger}
	result.connections = make(map[string]*instanceConnection)
	aContext.AddContextListener(result)
	return result
}
func (this *SyncService) TopicAdded (aTopic *conf.Topic) {
	
}
func (this *SyncService) InstanceRemoved (aInstance *conf.Instance) {
	if aInstance.Connected {
		var instanceConnection = this.connections[aInstance.Name()]
		if instanceConnection != nil {
			(*instanceConnection.connection).Close()
		}
	}
}
/**
 * Listen remote Instances call
 * @param aPort : the listening port
 */
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
				this.logger.Println("Failed to listen",err)
			} else {
				this.logger.Println("caught a call")
				go this.handleConnection(conn)
			}
		}
	}
}
/**
 * Send a command to remote Instance :
 *   #<command>#<size of arguments>#<arguments>
 * @param command    : a command cannot be nul
 * @param arguments  : a data to send cannot be nul, if empty, send []byte{}
 * @param connection : an open connection
 */
func (this *SyncService) sendCommand(command string, arguments []byte, connection net.Conn){
	this.logger.Println("sending",command,string(arguments))
	connection.Write([]byte("#"+command+"#"+strconv.Itoa(len(arguments))+"#"))
	connection.Write(arguments)
}
/**
 * Internal method that splits a command received from remote 
 */
func (this *SyncService) splitCommand(line []byte) (command string, arguments []byte, remain []byte, needMore int) {
	this.logger.Println("Splitting ",string(line))
	// Synchronize with firts DIEZE
	if line[0] != dieze {
		return "",[]byte{},line,0
	}
	// Obtain the COMMAND
	i := 1
	command = ""
	for line[i] != dieze {
		command += string(line[i])
		i++
	}
	this.logger.Println("Splitting - command found ",command)
	// Obtain the LENGTH
	i++
	slength := ""
	for line[i] != dieze {
		slength += string(line[i])
		i++
	}
	i++
	this.logger.Println("Splitting - slength found ",slength)
	length,_ := strconv.Atoi(slength)
	// Check wether the whole data has been received
	remain = []byte{}
	needMore = 0
	tailleRestantALire := len(line) - i
	// Missing the end the data
	if length > tailleRestantALire { // la longueur de la donnée annoncée > tailleRestantALire
		this.logger.Println("Splitting - not finished arguments found ",string(arguments))
		arguments = line[i:]
		needMore = length - tailleRestantALire // la longueur de la donnée annoncée - la tailleRestantALire + un dieze
		this.logger.Println("Splitting - need more",needMore)
	} else { // All the data has been obtained
		arguments = line[i:i+length]
		this.logger.Println("Splitting - arguments found ",string(arguments))
		// check wether a part of a next command is found in the data received
		if (i+length) == len(line) {
			this.logger.Println("Splitting - nothing remain")
		} else {
			remain = line[i+length:]
			this.logger.Println("Splitting - something remain", string(remain))
		}
	}
	return command, arguments, remain, needMore
}
/**
 * Reads from connection until missing data is received (linked to NeedMore detection in splitCommand)
 */
func (this *SyncService) takeMore(connection net.Conn, needMore int, writer io.Writer) (remain []byte){
	buffer := make([]byte,2000)
	var total = 0
	for total < needMore {
		count,err := connection.Read(buffer)
		if err != nil {
			this.logger.Println("Failed to read following needed bytes")
			return
		}
		if count + total > needMore {
			part := needMore - total
			writer.Write(buffer[0:part])
			remain = buffer[part:count]
		} else {
			total += count
			writer.Write(buffer[0:count])
			remain = []byte{}
		}
	}
	return remain
}
/**
 * Method used by both side : caller and called
 * all the commands will be received through this link
 */
func (this *SyncService) keepConnected(aInstanceConnection *instanceConnection){
	var byteBuffer *bytes.Buffer
	buffer := make([]byte,2000)
	connection := (*aInstanceConnection.connection)
	instance := aInstanceConnection.instance
	defer func() {
		connection.Close()
		instance.Connected = false
		delete(this.connections,instance.Name())
	}()
	var command string
	var arguments, remain []byte
	var needMore int
	// while the service is running
	for this.context.Running {
		time.Sleep(1 * time.Second)
		// reintroduce remain from previous command
		if len(remain) > 0 {
			this.logger.Println("Reusing remain",remain)
			buffer = remain
		} else {
			this.logger.Println("Listening to remote")
			count,err := connection.Read(buffer)
			if err != nil {
				this.logger.Println("Lost connection with",instance.Name(),err)
				break
			}
			buffer = buffer[0:count]
		}
		// PArse data received
		command, arguments, remain, needMore = this.splitCommand(buffer)
		if needMore != 0 {
			this.logger.Println("Unfinished",arguments,", need",needMore,"bytes")
			var bufferNeeded bytes.Buffer // Todo in the future use a swap in disk for ITEM command
			remain = this.takeMore(connection,needMore,&bufferNeeded)
			arguments = append(arguments,bufferNeeded.Bytes()...)
			this.logger.Println("Finally got",string(arguments))
		}
		// Process the command
		this.logger.Println("Received command " + command,remain)
		// Received HELLO from called
		if command == "HELLO" { // On est côté appelant, on reçoit la réponse de l'appelé, on lui envoie la configuration
			this.sendConfiguration(aInstanceConnection)
		} else if command == "INSTANCES" { // Receive instance list
			var newInstances []*conf.Instance
			byteBuffer = bytes.NewBuffer(arguments)
			decoder := json.NewDecoder(byteBuffer)
			decoder.Decode(&newInstances)
			for _,newInstance := range newInstances {
				this.logger.Println("Received instance :",newInstance)
				if (newInstance.Host == this.context.Host) && (newInstance.Port == this.port) {
					this.logger.Println("Skipped instance cause it is me :)")
					continue
				} else {
					newInstance.Connected = false // ensure the Instance will not be considered as connected
					if this.context.Configuration.AddInstance(newInstance) {
						this.logger.Println("Added instance :",newInstance)
					}
				}
			}
		} else if command == "TOPICS" { // Receive topic list
			var distributedTopics []*conf.Topic
			byteBuffer = bytes.NewBuffer(arguments)
			decoder := json.NewDecoder(byteBuffer)
			decoder.Decode(&distributedTopics)
			for _,topic := range distributedTopics {
				this.logger.Println("Received topic :",topic)
				existingTopic := this.context.Configuration.GetTopic(topic.Name)
				if existingTopic != nil {
					this.logger.Println("Skipped cause allready known")
				} else {
					this.context.Configuration.AddTopic(topic)
				}
			}
		} else if command == "ERROR" {
			this.logger.Println("Received ERROR :",arguments)
		} else {
			this.logger.Println("Not supported command")
			this.sendCommand("ERROR",[]byte("NOT SUPPORTED COMMAND '"+command+"'"),*aInstanceConnection.connection)
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
	this.sendCommand("INSTANCES",buffer.Bytes(),*aInstanceConnection.connection)
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
		this.sendCommand("TOPICS",buffer.Bytes(),*aInstanceConnection.connection)
	}
}
/**
 * Process the connection when called by a remote node.
 */
func (this *SyncService) handleConnection (aConn net.Conn){
	this.context.Host,_,_ = net.SplitHostPort(aConn.LocalAddr().String())
	this.logger.Println("Processing call")
	buffer := make([]byte,1000)
	count,err := aConn.Read(buffer)
	if err != nil {
		this.logger.Println("Unable to read HELLO from remote",err)
		return
	}
	if count < 10 {
		this.logger.Println("Unable to read HELLO from remote ",buffer[0:count])
		this.sendCommand("ERROR",[]byte("Unable to understand"),aConn)
		return
	}
	command, arguments, remain, needMore := this.splitCommand(buffer[0:count])
	if needMore > 0 {
		this.logger.Println("Unfinished",arguments,", need",needMore,"bytes")
		var bufferNeeded bytes.Buffer
		remain = this.takeMore(aConn,needMore,&bufferNeeded)
		arguments = append(arguments,bufferNeeded.Bytes()...)
		this.logger.Println("Finally got",string(arguments))
	}
	if command == "" {
		this.logger.Println("Unable to read HELLO from remote")
		this.sendCommand("ERROR",[]byte("Unable to understand"),aConn)
		return
	}
	this.logger.Println("Received ",command,"-",arguments,"-",remain)
	if command != "HELLO" {
		this.logger.Println("Unable to read HELLO from remote ")
		this.sendCommand("ERROR",[]byte("Unable to understand"),aConn)
		return
	}
	instance := this.context.Configuration.GetInstance(string(arguments))
	if instance == nil {
		host,port,_ := net.SplitHostPort(string(arguments))
		instance = conf.NewInstance(host,port)
		this.logger.Println("Adding caller as new instance",instance)
		this.context.Configuration.AddInstance(instance)
	}
	instance.Connected = true
	for len(remain) > 0 {
		command, arguments, remain, needMore = this.splitCommand(remain)
		this.logger.Println("Received command " + command,arguments,remain,needMore)
	}
	this.sendCommand("HELLO",[]byte(this.context.Host+":"+this.port),aConn) // TODO échanger leur numéros de version
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
		for _,instance := range this.context.Configuration.Instances {
			if !instance.Connected {
				host := instance.Host+":"+instance.Port
				this.logger.Println("Trying to connect to " + host)
				conn, err := net.Dial("tcp", host)
				if err != nil {
					this.logger.Println("Connection failed ", err)
					continue
				} else {
					this.logger.Println("Connection successful")
					this.context.Host,_,_ = net.SplitHostPort(conn.LocalAddr().String())
					this.sendCommand("HELLO",[]byte(this.context.Host+":"+this.port),conn)
					instance.Connected = true
					instanceConnection := newInstanceConnection(instance,&conn)
					this.connections[instance.Name()] = instanceConnection
					go this.keepConnected(instanceConnection)
				}
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