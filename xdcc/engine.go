package xdcc

import (
	"animuxd/irc"
	"fmt"
	"io"
	"sync"
)

const timeoutMsec = 2000

type DownloadStatus int

const (
	Waiting DownloadStatus = iota
	Downloading
	Done
	Failed
)

// Dialer is a function that connects somewhere and retturns IO.
type Dialer func(string, string) (io.ReadCloser, error)

// WriteOpener is a function that prepares and returns IO
// that requested files will be written to.
type WriteOpener func(engine *Engine, payload irc.PrivMsgDccSendPayload) (io.WriteCloser, error)

// Download describes current status and other metadata.
type Download struct {
	status DownloadStatus
}

// An Engine represents that part of the app which is responsible
// for handling XDCC download method. Works on top of irc.Engine.
type Engine struct {
	ircEngine      irc.IRCEngine
	dialer         Dialer
	openWriter     WriteOpener
	UnsafeMode     bool
	Downloads      map[string]*Download
	downloadsMutex *sync.Mutex
}

// Start initializes an engine.
func (e *Engine) Start(ircEngine irc.IRCEngine, dialer Dialer, writeOpener WriteOpener, unsafe bool) {
	e.ircEngine = ircEngine
	e.dialer = dialer
	e.openWriter = writeOpener
	e.UnsafeMode = unsafe
	e.Downloads = make(map[string]*Download)
	e.downloadsMutex = &sync.Mutex{}

	packets := e.ircEngine.IRCPacketsChann()

	go func() {
		for packet := range packets {
			if packet.Type == irc.PrivMsgDccSend {
				go func(dccSendPacket irc.Packet) {
					e.handleDccSendPacket(dccSendPacket)
				}(packet)
			}
		}
	}()
}

// joinBotChannels joins all channels that bot under given nick
// is present on. Returns promise channel.
func (e *Engine) joinBotChannels(botNick string) chan bool {
	r := make(chan bool)

	go func() {
		defer close(r)

		channelsPromise := e.ircEngine.ChannelsOfUser(botNick, timeoutMsec)
		channels := <-channelsPromise

		joinPromises := make([]<-chan bool, len(channels))
		for idx, channelName := range channels {
			joinPromises[idx] = e.ircEngine.Join(channelName, timeoutMsec)
		}
		for _, joinPromise := range joinPromises {
			<-joinPromise
		}

		r <- true
	}()

	return r
}

// RequestFile sends and memoizes download request.
func (e *Engine) RequestFile(botNick string, packageNo int, fileName string) <-chan bool {
	r := make(chan bool)

	go func() {
		defer close(r)

		joinPromise := e.joinBotChannels(botNick)
		<-joinPromise

		e.ircEngine.SendMessage(botNick, fmt.Sprintf("XDCC SEND %d", packageNo))

		e.downloadsMutex.Lock()
		e.Downloads[fileName] = &Download{status: Waiting}
		e.downloadsMutex.Unlock()

		r <- true
	}()

	return r
}

func (e *Engine) handleDccSendPacket(packet irc.Packet) {
	payload, payloadOk := packet.Payload.(irc.PrivMsgDccSendPayload)
	if !payloadOk {
		return
	}

	e.downloadsMutex.Lock()
	request, requestExists := e.Downloads[payload.FileName]
	e.downloadsMutex.Unlock()

	if requestExists || e.UnsafeMode {
		if !requestExists {
			e.downloadsMutex.Lock()
			e.Downloads[payload.FileName] = &Download{status: Waiting}
			request = e.Downloads[payload.FileName]
			e.downloadsMutex.Unlock()
		}

		if request.status != Waiting {
			return
		}

		downloadConn, dialError := e.dialer("tcp", fmt.Sprintf("%s:%d", payload.IP, payload.Port))
		if dialError == nil {
			defer downloadConn.Close()
		}

		writer, writerErr := e.openWriter(e, payload)
		if writerErr == nil {
			defer writer.Close()
		}

		var copyErr error
		if writerErr == nil && dialError == nil {
			_, copyErr = io.CopyN(writer, downloadConn, payload.FileLength)
			if flusher, isFlusher := writer.(interface{ Flush() error }); isFlusher {
				flusher.Flush()
			}
		}

		e.downloadsMutex.Lock()
		if copyErr == nil && writerErr == nil && dialError == nil {
			e.Downloads[payload.FileName].status = Done
		} else {
			e.Downloads[payload.FileName].status = Failed
		}
		e.downloadsMutex.Unlock()
	}
}
