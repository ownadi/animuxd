package xdcc

import (
	"animuxd/irc"
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
type WriteOpener func(engine *Engine, payload irc.PrivMsgDccSendPayload) (io.Writer, io.Closer, error)

// Download describes current status and other metadata.
type Download struct {
	Status       DownloadStatus
	CurrentSpeed uint64
	AvgSpeed     uint64
	Downloaded   uint64
	Size         int64
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
	downloadsMutex *sync.Mutex
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
		e.Downloads[fileName] = &Download{Status: Waiting}
		e.downloadsMutex.Unlock()

		r <- true
	}()

	return r
}

// DownloadsJSON writes JSON representation of downloads to given writer.
func (e *Engine) DownloadsJSON(writer io.Writer) error {
	e.downloadsMutex.Lock()
	defer e.downloadsMutex.Unlock()

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

	e.downloadsMutex.Lock()
	request, requestExists := e.Downloads[payload.FileName]
	e.downloadsMutex.Unlock()

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
			done := e.spawnSpeedOMeter(wc, payload)
			_, copyErr = io.CopyN(writer, downloadReader, payload.FileLength)
			done <- true
			if flusher, isFlusher := writer.(interface{ Flush() error }); isFlusher {
				flusher.Flush()
			}
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
	done := make(chan bool)

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
