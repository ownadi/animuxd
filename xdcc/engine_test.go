package xdcc

import (
	"animuxd/irc"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fakeIrcEngine struct {
	Channels     []string
	SentMessages []string
	PacketsChan  chan irc.Packet
}

func (e *fakeIrcEngine) IRCPacketsChann() chan irc.Packet {
	if e.PacketsChan == nil {
		e.PacketsChan = make(chan irc.Packet)
	}

	return e.PacketsChan
}

func (e *fakeIrcEngine) Join(channelName string, timeout int64) <-chan bool {
	r := make(chan bool)

	go func() {
		defer close(r)
		e.Channels = append(e.Channels, channelName)
		r <- true
	}()

	return r
}

func (e *fakeIrcEngine) ChannelsOfUser(nick string, timeout int64) chan []string {
	r := make(chan []string)

	go func() {
		defer close(r)
		r <- []string{"foo", "bar"}
	}()

	return r
}

func (e *fakeIrcEngine) SendMessage(nick string, body string) {
	e.SentMessages = append(e.SentMessages, body)
}

type FakeReadCloser struct {
	Closed bool
}

func (frc *FakeReadCloser) New() {

}

func (frc *FakeReadCloser) Read(p []byte) (n int, err error) {
	p[0] = 'A'
	return 1, nil
}

func (frc *FakeReadCloser) Close() error {
	frc.Closed = true

	return nil
}

type FakeWriter struct {
	BytesWritten int
	Closed       bool
	Flushed      bool
}

func (fw *FakeWriter) Flush() error {
	fw.Flushed = true

	return nil
}

func (fw *FakeWriter) Write(p []byte) (n int, err error) {
	bytesLen := len(p)

	fw.BytesWritten += bytesLen

	return bytesLen, nil
}

func (fw *FakeWriter) Close() error {
	fw.Closed = true

	return nil
}

type FakeIOs struct {
	frc *FakeReadCloser
	fw  *FakeWriter
}

func PrepareFakes() (Dialer, WriteOpener, *FakeIOs) {
	holder := &FakeIOs{}

	dialer := func(network string, address string) (io.ReadCloser, error) {
		holder.frc = &FakeReadCloser{}

		return holder.frc, nil
	}

	prepareFakeWriter := func(engine *Engine, payload irc.PrivMsgDccSendPayload) (io.WriteCloser, error) {
		holder.fw = &FakeWriter{}

		return holder.fw, nil
	}

	return dialer, prepareFakeWriter, holder
}

type ErrReader struct{}

func (er *ErrReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("")
}

func (er *ErrReader) Close() error {
	return nil
}

func TestRequestFile(t *testing.T) {
	ircEngine := &fakeIrcEngine{}
	engine := &Engine{}

	dial, prepareWriter, _ := PrepareFakes()
	engine.Start(ircEngine, dial, prepareWriter, false)

	requestPromise := engine.RequestFile("b0t", 42, "foo.bar")
	<-requestPromise

	assert.Contains(t, ircEngine.Channels, "foo")
	assert.Contains(t, ircEngine.Channels, "bar")
	assert.Contains(t, ircEngine.SentMessages[0], "XDCC SEND 42")
	download, downloadExists := engine.Downloads["foo.bar"]
	assert.True(t, downloadExists)
	assert.Equal(t, Waiting, download.status)
}

func TestHandleDccSend(t *testing.T) {
	ircEngine := &fakeIrcEngine{}
	packetsChann := ircEngine.IRCPacketsChann()

	engine := &Engine{}
	dial, prepareWriter, fakes := PrepareFakes()
	engine.Start(ircEngine, dial, prepareWriter, false)

	requestPromise := engine.RequestFile("b0t", 42, "foo.bar")
	<-requestPromise

	payload := irc.PrivMsgDccSendPayload{
		FileName:   "foo.bar",
		FileLength: 100,
		IP:         net.ParseIP("127.0.0.1"),
		Port:       1337,
	}
	packetsChann <- irc.Packet{Type: irc.PrivMsgDccSend, Payload: payload}
	time.Sleep(50 * time.Millisecond) // FIXME: Wait for the value with timeout

	download, downloadExists := engine.Downloads["foo.bar"]
	assert.True(t, downloadExists)
	assert.Equal(t, Done, download.status)
	assert.Equal(t, 100, fakes.fw.BytesWritten)
	assert.True(t, fakes.fw.Closed)
	assert.True(t, fakes.fw.Flushed)
	assert.True(t, fakes.frc.Closed)
}

func TestHandleDccSendDoesnExist(t *testing.T) {
	ircEngine := &fakeIrcEngine{}
	packetsChann := ircEngine.IRCPacketsChann()

	engine := &Engine{}
	dial, prepareWriter, _ := PrepareFakes()
	engine.Start(ircEngine, dial, prepareWriter, false)

	payload := irc.PrivMsgDccSendPayload{
		FileName:   "foo.bar",
		FileLength: 50,
		IP:         net.ParseIP("127.0.0.1"),
		Port:       1337,
	}
	packetsChann <- irc.Packet{Type: irc.PrivMsgDccSend, Payload: payload}
	time.Sleep(50 * time.Millisecond)

	_, downloadExists := engine.Downloads["foo.bar"]
	assert.False(t, downloadExists)
}

func TestHandleDccSendDoesnExistUnsafe(t *testing.T) {
	ircEngine := &fakeIrcEngine{}
	packetsChann := ircEngine.IRCPacketsChann()

	engine := &Engine{}
	dial, prepareWriter, _ := PrepareFakes()
	engine.Start(ircEngine, dial, prepareWriter, true)

	payload := irc.PrivMsgDccSendPayload{
		FileName:   "foo.bar",
		FileLength: 50,
		IP:         net.ParseIP("127.0.0.1"),
		Port:       1337,
	}
	packetsChann <- irc.Packet{Type: irc.PrivMsgDccSend, Payload: payload}
	time.Sleep(50 * time.Millisecond)

	download, downloadExists := engine.Downloads["foo.bar"]
	assert.True(t, downloadExists)
	assert.Equal(t, Done, download.status)
}

func TestHandleDccSendDialErr(t *testing.T) {
	ircEngine := &fakeIrcEngine{}
	packetsChann := ircEngine.IRCPacketsChann()

	engine := &Engine{}
	_, prepareWriter, _ := PrepareFakes()
	dial := func(string, string) (io.ReadCloser, error) {
		return nil, errors.New("")
	}
	engine.Start(ircEngine, dial, prepareWriter, false)

	requestPromise := engine.RequestFile("b0t", 42, "foo.bar")
	<-requestPromise

	payload := irc.PrivMsgDccSendPayload{
		FileName:   "foo.bar",
		FileLength: 50,
		IP:         net.ParseIP("127.0.0.1"),
		Port:       1337,
	}
	packetsChann <- irc.Packet{Type: irc.PrivMsgDccSend, Payload: payload}
	time.Sleep(50 * time.Millisecond)

	download, downloadExists := engine.Downloads["foo.bar"]
	assert.True(t, downloadExists)
	assert.Equal(t, Failed, download.status)
}

func TestHandleDccSendOpenWriterErr(t *testing.T) {
	ircEngine := &fakeIrcEngine{}
	packetsChann := ircEngine.IRCPacketsChann()

	engine := &Engine{}
	dial, _, _ := PrepareFakes()
	prepareWriter := func(engine *Engine, payload irc.PrivMsgDccSendPayload) (io.WriteCloser, error) {
		return nil, errors.New("")
	}
	engine.Start(ircEngine, dial, prepareWriter, false)

	requestPromise := engine.RequestFile("b0t", 42, "foo.bar")
	<-requestPromise

	payload := irc.PrivMsgDccSendPayload{
		FileName:   "foo.bar",
		FileLength: 50,
		IP:         net.ParseIP("127.0.0.1"),
		Port:       1337,
	}
	packetsChann <- irc.Packet{Type: irc.PrivMsgDccSend, Payload: payload}
	time.Sleep(50 * time.Millisecond)

	download, downloadExists := engine.Downloads["foo.bar"]
	assert.True(t, downloadExists)
	assert.Equal(t, Failed, download.status)
}

func TestHandleDccSendCopyErr(t *testing.T) {
	ircEngine := &fakeIrcEngine{}
	packetsChann := ircEngine.IRCPacketsChann()

	_, prepareWriter, _ := PrepareFakes()
	dial := func(string, string) (io.ReadCloser, error) {
		return &ErrReader{}, nil
	}
	engine := &Engine{}
	engine.Start(ircEngine, dial, prepareWriter, false)

	requestPromise := engine.RequestFile("b0t", 42, "foo.bar")
	<-requestPromise

	payload := irc.PrivMsgDccSendPayload{
		FileName:   "foo.bar",
		FileLength: 11, // The fake returns err on 101th byte
		IP:         net.ParseIP("127.0.0.1"),
		Port:       1337,
	}
	packetsChann <- irc.Packet{Type: irc.PrivMsgDccSend, Payload: payload}
	time.Sleep(50 * time.Millisecond)

	download, downloadExists := engine.Downloads["foo.bar"]
	assert.True(t, downloadExists)
	assert.Equal(t, Failed, download.status)
}
