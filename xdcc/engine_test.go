package xdcc

import (
	"animuxd/irc"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fakeIrcEngine struct {
	Channels     []string
	SentMessages []string
	PacketsChan  chan irc.Packet
	ctx          context.Context
}

func (e *fakeIrcEngine) IRCPacketsChann() chan irc.Packet {
	if e.PacketsChan == nil {
		e.PacketsChan = make(chan irc.Packet)
	}

	return e.PacketsChan
}

func (e *fakeIrcEngine) Join(ctx context.Context, channelName string) <-chan bool {
	r := make(chan bool)

	go func() {
		defer close(r)
		e.Channels = append(e.Channels, channelName)
		r <- true
	}()

	return r
}

func (e *fakeIrcEngine) ChannelsOfUser(ctx context.Context, nick string) chan []string {
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

func (e *fakeIrcEngine) Context() context.Context {
	if e.ctx == nil {
		e.ctx = context.Background()
	}

	return e.ctx
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

	dialer := func(e *Engine, payload irc.PrivMsgDccSendPayload) (io.ReadCloser, error) {
		holder.frc = &FakeReadCloser{}

		return holder.frc, nil
	}

	prepareFakeWriter := func(engine *Engine, payload irc.PrivMsgDccSendPayload) (io.Writer, io.Closer, error) {
		holder.fw = &FakeWriter{}

		return holder.fw, holder.fw, nil
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
	assert.Equal(t, Waiting, download.Status)
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
	assert.Equal(t, Done, download.Status)
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
	assert.Equal(t, Done, download.Status)
}

func TestHandleDccSendDialErr(t *testing.T) {
	ircEngine := &fakeIrcEngine{}
	packetsChann := ircEngine.IRCPacketsChann()

	engine := &Engine{}
	_, prepareWriter, _ := PrepareFakes()
	dial := func(*Engine, irc.PrivMsgDccSendPayload) (io.ReadCloser, error) {
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
	assert.Equal(t, Failed, download.Status)
}

func TestHandleDccSendOpenWriterErr(t *testing.T) {
	ircEngine := &fakeIrcEngine{}
	packetsChann := ircEngine.IRCPacketsChann()

	engine := &Engine{}
	dial, _, _ := PrepareFakes()
	prepareWriter := func(engine *Engine, payload irc.PrivMsgDccSendPayload) (io.Writer, io.Closer, error) {
		return nil, nil, errors.New("")
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
	assert.Equal(t, Failed, download.Status)
}

func TestHandleDccSendCopyErr(t *testing.T) {
	ircEngine := &fakeIrcEngine{}
	packetsChann := ircEngine.IRCPacketsChann()

	_, prepareWriter, _ := PrepareFakes()
	dial := func(*Engine, irc.PrivMsgDccSendPayload) (io.ReadCloser, error) {
		return &ErrReader{}, nil
	}
	engine := &Engine{}
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
	assert.Equal(t, Failed, download.Status)
}

func TestDownloadsJSON(t *testing.T) {
	ircEngine := &fakeIrcEngine{}
	packetsChann := ircEngine.IRCPacketsChann()

	dial, prepareWriter, _ := PrepareFakes()
	engine := &Engine{}
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

	buff := new(bytes.Buffer)
	err := engine.DownloadsJSON(buff)
	json := buff.String()

	assert.Nil(t, err)
	assert.Regexp(t, regexp.MustCompile(`"FileName":"foo.bar"`), json)
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf(`"Status":%d`, Done)), json)
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf(`"Size":%d`, 50)), json)
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf(`"Downloaded":%d`, 50)), json)
	assert.Regexp(t, regexp.MustCompile(`"CurrentSpeed":0`), json)
	assert.Regexp(t, regexp.MustCompile(`"AvgSpeed":[1-9]([0-9]*)?`), json)
}

func TestContextAndStop(t *testing.T) {
	ircEngine := &fakeIrcEngine{}
	dial, prepareWriter, _ := PrepareFakes()
	engine := &Engine{}
	engine.Start(ircEngine, dial, prepareWriter, false)
	ctx := engine.Context()
	engine.Stop()

	<-ctx.Done()
}

func TestRestartResumesDownloads(t *testing.T) {
	ircEngine := &fakeIrcEngine{}
	dial, prepareWriter, _ := PrepareFakes()
	engine := &Engine{}

	engine.Start(ircEngine, dial, prepareWriter, false)

	engine.Downloads = map[string]*Download{
		"foo.mkv": &Download{
			Status:    Waiting,
			BotNick:   "b0t",
			PackageNo: 1,
			Size:      1000,
		},
		"bar.mkv": &Download{
			Status:    Downloading,
			BotNick:   "b0t",
			PackageNo: 2,
			Size:      2000,
		},
		"baz.mkv": &Download{
			Status:    Failed,
			BotNick:   "b0t",
			PackageNo: 3,
			Size:      3000,
		},
		"x.mkv": &Download{
			Status:    Done,
			BotNick:   "b0t",
			PackageNo: 4,
			Size:      4000,
		},
	}

	engine.Restart(ircEngine)

	assert.Equal(t, engine.Downloads["foo.mkv"].Status, Waiting)
	assert.Contains(t, ircEngine.SentMessages, "XDCC SEND 1")

	assert.Equal(t, engine.Downloads["bar.mkv"].Status, Waiting)
	assert.Contains(t, ircEngine.SentMessages, "XDCC SEND 2")

	assert.Equal(t, engine.Downloads["baz.mkv"].Status, Waiting)
	assert.Contains(t, ircEngine.SentMessages, "XDCC SEND 3")

	assert.Equal(t, engine.Downloads["x.mkv"].Status, Done)
	assert.NotContains(t, ircEngine.SentMessages, "XDCC SEND 4")
}
