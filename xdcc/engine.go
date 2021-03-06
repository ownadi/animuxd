package xdcc

import (
	"animuxd/irc"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

const timeoutMsec = 2000

type DownloadStatus int

const (
	Waiting DownloadStatus = iota
	Downloading
	Done
	Failed
)

// Dialer is a function that connects somewhere and returns IO.
type Dialer func(engine *Engine, payload irc.PrivMsgDccSendPayload) (io.ReadCloser, error)

// WriteOpener is a function that prepares and returns IO
// that requested files will be written to.
// Returns both writer and closer for convenient usage of bufio.
type WriteOpener func(engine *Engine, payload irc.PrivMsgDccSendPayload) (io.Writer, io.Closer, error)

// Download describes current status and other metadata.
type Download struct {
	Status       DownloadStatus
	CurrentSpeed uint64
	AvgSpeed     uint64
	Downloaded   uint64
	Size         int64
	BotNick      string
	PackageNo    int
}

// DownloadJSON extends Download with some JSON-useful fields.
type DownloadJSON struct {
	FileName string
	*Download
}

// An Engine represents that part of the app which is responsible
// for handling XDCC download method. Works on top of irc.Engine.
type Engine struct {
	ircEngine      irc.IRCEngine
	dialer         Dialer
	openWriter     WriteOpener
	UnsafeMode     bool
	Downloads      map[string]*Download
	downloadsMutex *sync.RWMutex
	ctx            context.Context
	cancelFunc     context.CancelFunc
}

type XDCCEngine interface {
	RequestFile(botNick string, packageNo int, fileName string) <-chan bool
	DownloadsJSON(writer io.Writer) error
}

// Start initializes an engine.
func (e *Engine) Start(ircEngine irc.IRCEngine, dialer Dialer, writeOpener WriteOpener, unsafe bool) {
	e.ircEngine = ircEngine
	e.dialer = dialer
	e.openWriter = writeOpener
	e.UnsafeMode = unsafe
	e.Downloads = map[string]*Download{}
	e.downloadsMutex = &sync.RWMutex{}
	e.ctx, e.cancelFunc = context.WithCancel(ircEngine.Context())

	go e.handleIrcPackets()
}

func (e *Engine) Stop() {
	e.cancelFunc()
}

func (e *Engine) Context() context.Context {
	return e.ctx
}

// Restart starts the Engine on top of a new IRCEngine and resumes uncompleted downloads.
// TODO: Try to RESUME instead of downloading from zero.
func (e *Engine) Restart(ircEngine irc.IRCEngine) {
	e.ircEngine = ircEngine
	e.ctx, e.cancelFunc = context.WithCancel(ircEngine.Context())

	e.downloadsMutex.Lock()
	requestPromises := make([]<-chan bool, 0, len(e.Downloads))
	for fileName, download := range e.Downloads {
		if download.Status != Done {
			e.Downloads[fileName].Status = Waiting
			requestPromise := e.RequestFile(download.BotNick, download.PackageNo, fileName)
			requestPromises = append(requestPromises, requestPromise)
		}
	}
	e.downloadsMutex.Unlock()

	go e.handleIrcPackets()

	for _, promise := range requestPromises {
		<-promise
	}
}

func (e *Engine) handleIrcPackets() {
	defer e.cancelFunc()

	packets := e.ircEngine.IRCPacketsChann()

	for {
		select {
		case <-e.ctx.Done():
			return
		case packet := <-packets:
			if packet.Type == irc.PrivMsgDccSend {
				go func(dccSendPacket irc.Packet) {
					e.handleDccSendPacket(dccSendPacket)
				}(packet)
			}
		}
	}
}

// joinBotChannels joins all channels that bot under given nick
// is present on. Returns promise channel.
func (e *Engine) joinBotChannels(botNick string) chan bool {
	r := make(chan bool, 1)

	go func() {
		defer close(r)

		channelsContext, cancelChannelsContext := context.WithTimeout(e.ctx, timeoutMsec*time.Millisecond)
		channelsPromise := e.ircEngine.ChannelsOfUser(channelsContext, botNick)
		channels := <-channelsPromise

		if channelsContext.Err() != nil {
			r <- false
			cancelChannelsContext()
			return
		}
		cancelChannelsContext()

		joinPromises := make([]<-chan bool, 0, len(channels))
		joinCtx, cancelJoinCtx := context.WithTimeout(e.ctx, timeoutMsec*time.Millisecond)
		for _, channelName := range channels {
			joinPromises = append(joinPromises, e.ircEngine.Join(joinCtx, channelName))
		}
		for _, joinPromise := range joinPromises {
			<-joinPromise
		}

		r <- true
		cancelJoinCtx()
	}()

	return r
}

// RequestFile sends and memoizes download request.
func (e *Engine) RequestFile(botNick string, packageNo int, fileName string) <-chan bool {
	r := make(chan bool, 1)

	go func() {
		defer close(r)

		joinPromise := e.joinBotChannels(botNick)
		<-joinPromise

		e.ircEngine.SendMessage(botNick, fmt.Sprintf("XDCC SEND %d", packageNo))

		e.downloadsMutex.Lock()
		e.Downloads[fileName] = &Download{Status: Waiting, BotNick: botNick, PackageNo: packageNo}
		e.downloadsMutex.Unlock()

		r <- true
	}()

	return r
}

// DownloadsJSON writes JSON representation of downloads to given writer.
func (e *Engine) DownloadsJSON(writer io.Writer) error {
	e.downloadsMutex.RLock()
	defer e.downloadsMutex.RUnlock()

	jsonArray := make([]DownloadJSON, 0, len(e.Downloads))
	for fileName, download := range e.Downloads {
		jsonArray = append(jsonArray, DownloadJSON{
			FileName: fileName,
			Download: download,
		})
	}

	return json.NewEncoder(writer).Encode(jsonArray)
}

func (e *Engine) handleDccSendPacket(packet irc.Packet) {
	payload, payloadOk := packet.Payload.(irc.PrivMsgDccSendPayload)
	if !payloadOk {
		return
	}

	e.downloadsMutex.RLock()
	request, requestExists := e.Downloads[payload.FileName]
	e.downloadsMutex.RUnlock()

	if requestExists || e.UnsafeMode {
		if !requestExists {
			e.downloadsMutex.Lock()
			e.Downloads[payload.FileName] = &Download{Status: Waiting}
			request = e.Downloads[payload.FileName]
			e.downloadsMutex.Unlock()
		}

		if request.Status != Waiting {
			return
		}

		downloadConn, dialError := e.dialer(e, payload)
		if dialError == nil {
			defer downloadConn.Close()
		}

		writer, closer, writerErr := e.openWriter(e, payload)
		if writerErr == nil {
			defer closer.Close()
		}

		var copyErr error
		if writerErr == nil && dialError == nil {
			e.downloadsMutex.Lock()
			e.Downloads[payload.FileName].Size = payload.FileLength
			e.Downloads[payload.FileName].Status = Downloading
			e.downloadsMutex.Unlock()

			wc := &WriteCounter{}
			downloadReader := io.TeeReader(downloadConn, wc)
			endSpeedOMeter := e.spawnSpeedOMeter(wc, payload)
			done := make(chan bool, 1)
			defer close(done)

			// Cancel download when context gets canceled
			go func() {
				select {
				case <-e.ctx.Done():
					closer.Close()
				case <-done:
				}
			}()

			_, copyErr = io.CopyN(writer, downloadReader, payload.FileLength)

			if e.ctx.Err() == nil {
				endSpeedOMeter <- true
			}
			if flusher, isFlusher := writer.(interface{ Flush() error }); isFlusher {
				flusher.Flush()
			}
			done <- true
		}

		e.downloadsMutex.Lock()
		if copyErr == nil && writerErr == nil && dialError == nil {
			e.Downloads[payload.FileName].Status = Done
		} else {
			e.Downloads[payload.FileName].Status = Failed
		}
		e.downloadsMutex.Unlock()
	}
}

func (e *Engine) spawnSpeedOMeter(wc *WriteCounter, payload irc.PrivMsgDccSendPayload) chan<- bool {
	done := make(chan bool, 1)

	startTime := time.Now()
	lastTime := time.Now()

	lastDownloadedBytes := float64(0)

	go func() {
		defer close(done)

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		lastIteration := false
		for {
			select {
			case <-e.ctx.Done():
				lastIteration = true
			case <-done:
				lastIteration = true
			case <-ticker.C:
				lastIteration = false
			}

			downloadedBytes := atomic.LoadUint64(&wc.Total)

			currentTime := time.Now()
			passedTime := currentTime.UnixNano() - lastTime.UnixNano()
			passedAllTime := currentTime.UnixNano() - startTime.UnixNano()
			passedSeconds := float64(passedTime) / float64(time.Second)
			passedAllSeconds := float64(passedAllTime) / float64(time.Second)

			currentSpeed := (float64(downloadedBytes) - lastDownloadedBytes) / passedSeconds
			avgSpeed := float64(downloadedBytes) / passedAllSeconds

			lastTime = currentTime
			lastDownloadedBytes = float64(downloadedBytes)

			e.downloadsMutex.Lock()
			if lastIteration {
				e.Downloads[payload.FileName].CurrentSpeed = 0
			} else {
				e.Downloads[payload.FileName].CurrentSpeed = uint64(currentSpeed)
			}
			e.Downloads[payload.FileName].AvgSpeed = uint64(avgSpeed)
			e.Downloads[payload.FileName].Downloaded = downloadedBytes
			e.downloadsMutex.Unlock()

			if lastIteration {
				return
			}
		}
	}()

	return done
}
