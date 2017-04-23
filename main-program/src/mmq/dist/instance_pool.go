package dist

import (
	"mmq/conf"
	"mmq/env"
	"log"
	"net"
	"bytes"
	"encoding/json"
)
/**
 * An internal structure used for the map. It links an instance to a connection
 */
type instanceConnection struct {
	instance   *conf.Instance // an instance
	connection *net.Conn      // a connection opened with the instance
	pool		*InstancePool
}
func (this *instanceConnection) SendItem(aItem ManagedItem) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.Encode(aItem)
	this.pool.protocol.sendCommand("ITEM", buffer.Bytes(), *this.connection)
}
func (this *instanceConnection) Close() {
	(*this.connection).Close()
	this.instance.Connected = false
	this.pool.instanceClosed(this.instance)
}
type InstancePool struct {
	context    			*env.Context
	logger 				*log.Logger
	port 				string 							// will be obtained via configuration
	connections			map[string]*instanceConnection	// a map that links instances to opened net connection
	instancesByGroup 	map[string][]*instanceConnection
	protocol			*protocol
}
func NewInstancePool(aContext *env.Context) *InstancePool {
	result := &InstancePool{context : aContext, protocol : NewProtocol(aContext)}
	result.logger 			= aContext.Logger
	result.connections 		= make(map[string]*instanceConnection)
	result.instancesByGroup = make(map[string][]*instanceConnection)
	service := aContext.Configuration.GetServiceByName(conf.SERVICE_SYNC)
	if service != nil {
		param := service.GetParameterByName(conf.PARAMETER_PORT)
		if param != nil {
			result.port = param.Value
		}
	}
	return result
}
/**
 * Constructor for InstanceConnection
 */
func (this *InstancePool) newInstanceConnection (aInstance *conf.Instance, aConnection *net.Conn) *instanceConnection{
	result := &instanceConnection{instance : aInstance, connection : aConnection, pool : this}
	this.connections[aInstance.Name()] = result
	return result
}
func (this *InstancePool) instanceClosed(aInstance *conf.Instance){
	delete(this.connections,aInstance.Name())
}
func (this *InstancePool) GetInstancesByGroup(aGroupName string) []*instanceConnection {
	return nil
}
func (this *InstancePool) GetInstanceByName(aInstanceName string) *instanceConnection {
	return this.connections[aInstanceName]
}
func (this *InstancePool) Connect(aInstance *conf.Instance) error {
	host := aInstance.Host + ":" + aInstance.Port
	this.logger.Println("Trying to connect to " + host)
	conn, err := net.Dial("tcp", host)
	if err != nil {
		this.logger.Println("WARNING Connection failed ", err)
		return err
	} else {
		this.logger.Println("INFO Connection successful")
		this.context.Host, _, _ = net.SplitHostPort(conn.LocalAddr().String())
		this.protocol.sendCommand("HELLO", []byte(this.context.Host+":"+this.port), conn)
		aInstance.Connected = true
		this.newInstanceConnection(aInstance, &conn)
		go this.protocol.keepConnected(aInstance, &conn)
		return nil
	}
}