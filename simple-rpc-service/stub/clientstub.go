package stub

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"math"
)

const (
	DefaultHost = "localhost"
	DefaultPort = "5001"
)

const (
	initBuffSize 	= 20
	lengthByteSize  = 4
	integerByteSize = 8
)

type ClientStubber interface {
	Init(host string, port string) error
	Invoke(method string, a int64, b int64)
}

type RPCClientStub struct {
	conn net.Conn
	host string
	port string
}

func (c *RPCClientStub) Init(host string, port string) (error) {
	conn,err := net.Dial("tcp",host + ":" + port );
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *RPCClientStub) Invoke(method string, a int64, b int64) (int64,error){

	if (b > 0 && a > (math.MaxInt - b)) || (a > 0 && b > (math.MaxInt - b)) {
		return -1,fmt.Errorf("Integer overflow 💥")
	}

	//[length][string bytes][int64][int64]
	buff := make([]byte, initBuffSize + uint32(len(method)))
	offset := 0

	binary.BigEndian.PutUint32(buff[:lengthByteSize], uint32(len(method)))
	offset += lengthByteSize

	copy(buff[offset:offset + len(method)], []byte(method))
	offset += len(method)

	binary.PutVarint(buff[offset: offset + integerByteSize], a)
	offset += integerByteSize

	binary.PutVarint(buff[offset: offset + integerByteSize], b)
	offset += integerByteSize

	n,err := c.conn.Write(buff)

	if err != nil {
		//graceful shutdown
		return -1, fmt.Errorf("Error writing to buffer %s 💥", err)
	}
	if n != offset{
		//retry n times in the case of a network problems
		log.Println("Did not write all bytes from buffer...retrying 🔄")
		return -1, fmt.Errorf("Fatal: could not write all bytes...wrote %v bytes 💥", n)
	}

	//recieve data back
	buff = make([]byte, 8)
	n, err = c.conn.Read(buff)
	if n != 8 {
		//retry read
		log.Println("Did not read all bytes from buffer...retrying 🔄")
		return -1, fmt.Errorf("Fatal: could not read all bytes...read %v bytes 💥", n)
	}

	if err != nil {
		return -1, fmt.Errorf("Error occured reading from buffer: %s 💥", err)
	}

	ret := int64(binary.BigEndian.Uint64(buff))

	return ret, nil
}
