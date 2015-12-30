package server_test

import (
	"bufio"
	"fmt"
	"github.com/SuperLimitBreak/socketServerMq/server"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

func getTcpConn(t *testing.T) (net.Conn, error) {
	conn, err := net.Dial("tcp", ":9872")

	assert.Nil(t, err)
	assert.NotNil(t, conn)

	return conn, err
}

func TestServer(t *testing.T) {
	go server.StartAll()

	//wait to prevent race hazard
	<-time.After(2 * time.Second)
}

func TestTcpConnect(t *testing.T) {
	conn, err := getTcpConn(t)
	if err != nil {
		t.FailNow()
	}

	err = conn.Close()
	assert.Nil(t, err)
}

func TestTcpSimpleSend(t *testing.T) {
	conn, err := getTcpConn(t)
	if err != nil {
		t.FailNow()
	}

	msg := `{"action":"message","data":[{"foo":"bar"}]}` + "\n"

	n, err := fmt.Fprintf(conn, msg)
	assert.Nil(t, err)
	assert.Equal(t, len([]byte(msg)), n)

	rcv, err := bufio.NewReader(conn).ReadString('\n')

	assert.Nil(t, err)
	assert.Equal(t,
		msg,
		rcv,
	)
}

func TestTcpSimpleSubscribe(t *testing.T) {
	conn, err := getTcpConn(t)
	if err != nil {
		t.FailNow()
	}

	sub := `{"action":"subscribe","data":["key1"]}` + "\n"
	fmt.Fprintf(conn, sub)

	//allow time for subscription
	<-time.After(time.Second)

	msgKey2 := `{"action":"message","data":[{"deviceid":"key2","foo":"bar"}]}` + "\n"
	msgKey1 := `{"action":"message","data":[{"deviceid":"key1","baz":"bang"}]}` + "\n"

	fmt.Fprintf(conn, msgKey2)

	//allow time for msg to propgate through mq
	<-time.After(time.Second)

	fmt.Fprintf(conn, msgKey1)

	rcv, err := bufio.NewReader(conn).ReadString('\n')

	assert.Nil(t, err)
	assert.Equal(t,
		`{"action":"message","data":[{"deviceid":"key1","baz":"bang"}]}`+"\n",
		rcv,
	)
}

func TestTcpMultiMessage(t *testing.T) {
	conn, err := getTcpConn(t)
	if err != nil {
		t.FailNow()
	}

	p1 := `{"foo":"bar"}`
	p2 := `{"baz":"bat"}`
	p3 := `{"tum":"thun"}`

	msg := `{"action":"message","data":[` + p1 + "," + p2 + "," + p3 + `]}` + "\n"
	fmt.Fprintf(conn, msg)

	r := bufio.NewReader(conn)

	rcv1, err := r.ReadString('\n')
	assert.Nil(t, err)

	rcv2, err := bufio.NewReader(conn).ReadString('\n')
	assert.Nil(t, err)

	rcv3, err := bufio.NewReader(conn).ReadString('\n')
	assert.Nil(t, err)

	assert.NotEqual(t, rcv1, rcv2)
	assert.NotEqual(t, rcv1, rcv3)
	assert.NotEqual(t, rcv2, rcv3)

	wrap := func(s string) string {
		return `{"action":"message","data":[` + s + `]}` + "\n"
	}

	payloads := map[string]bool{
		wrap(p1): false,
		wrap(p2): false,
		wrap(p3): false,
	}

	payloads[rcv1] = true
	payloads[rcv2] = true
	payloads[rcv3] = true

	if !assert.True(t, payloads[p1]) || !assert.True(t, payloads[p2]) || !assert.True(t, payloads[p3]) {
		for k, v := range payloads {
			fmt.Printf("%v: %v\n", k, v)
		}
	}
}
