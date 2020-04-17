package irc

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"runtime"
	"strings"
	"time"
)

const nickLength = 7

type onPacketCallback func(Packet)

// An Engine represents that part of the app which is responsible
// for handling low-level IRC protocol related stuff.
type Engine struct {
	nick               string
	ircStream          io.ReadWriteCloser
	onRplWelcome       map[string]onPacketCallback
	onErrNicknameInUse map[string]onPacketCallback
	onRplEndOfNames    map[string]onPacketCallback
	onRplWhoisChannels map[string]onPacketCallback
}

// Nick returns current registered nick.
func (e *Engine) Nick() string {
	return e.nick
}

// Start initializes the engine.
// Returns channel of parsed IRC packets.
func (e *Engine) Start(ircStream io.ReadWriteCloser) <-chan Packet {
	e.ircStream = ircStream
	e.nick = ""
	e.onErrNicknameInUse = map[string]onPacketCallback{}
	e.onRplWelcome = map[string]onPacketCallback{}
	e.onRplEndOfNames = map[string]onPacketCallback{}
	e.onRplWhoisChannels = map[string]onPacketCallback{}

	ircScanner := bufio.NewScanner(e.ircStream)
	r := make(chan Packet, runtime.NumCPU())

	go func() {
		for ircScanner.Scan() {
			ircLine := ircScanner.Text()

			go func() {
				packet := Parse(ircLine)

				if packet.Type == RplWelcome {
					for _, callback := range e.onRplWelcome {
						callback(packet)
					}
				}

				if packet.Type == ErrNicknameInUse {
					for _, callback := range e.onErrNicknameInUse {
						callback(packet)
					}
				}

				if packet.Type == RplEndOfNames {
					for _, callback := range e.onRplEndOfNames {
						callback(packet)
					}
				}

				if packet.Type == RplWhoisChannels {
					for _, callback := range e.onRplWhoisChannels {
						callback(packet)
					}
				}

				if packet.Type == Ping {
					e.send(fmt.Sprintf("PONG :%s", packet.Payload))
				}

				r <- packet
			}()
		}
	}()

	return r
}

// Register tries to register IRC nick.
// In most cases should be called right after Start.
// On success sends true on the returned channel.
func (e *Engine) Register(timeout int64) <-chan bool {
	r := make(chan bool)

	go func() {
		defer close(r)

		registrationSuccess := false
		for !registrationSuccess {
			successChann := make(chan bool)
			defer close(successChann)

			currentNick := randNick()

			welcomeCallback := func(packet Packet) {
				if packet.Payload == currentNick {
					e.nick = currentNick
					successChann <- true
				}
			}

			nickTakenCallback := func(packet Packet) {
				if packet.Payload == currentNick {
					successChann <- false
				}
			}

			e.onRplWelcome[currentNick] = welcomeCallback
			e.onErrNicknameInUse[currentNick] = nickTakenCallback

			e.send(fmt.Sprintf("USER %s * * %s", currentNick, currentNick))
			e.send(fmt.Sprintf("NICK %s", currentNick))

			select {
			case <-time.After(time.Duration(timeout) * time.Millisecond):
				registrationSuccess = false
			case success := <-successChann:
				registrationSuccess = success
			}

			delete(e.onRplWelcome, currentNick)
			delete(e.onErrNicknameInUse, currentNick)
		}

		r <- true
	}()

	return r
}

// Join tries to join IRC channel.
// On success sends true on the returned channel.
func (e *Engine) Join(channelName string, timeout int64) <-chan bool {
	r := make(chan bool)

	go func() {
		defer close(r)

		callbackSuccessChann := make(chan bool)
		defer close(callbackSuccessChann)

		channelWithHash := channelName
		if !strings.HasPrefix(channelName, "#") {
			channelWithHash = fmt.Sprintf("#%s", channelName)
		}
		channelWithoutHash := channelWithHash[1:]

		callback := func(packet Packet) {
			if packet.Payload == channelWithoutHash {
				callbackSuccessChann <- true
			}
		}
		e.onRplEndOfNames[channelWithoutHash] = callback

		e.send(fmt.Sprintf("JOIN %s", channelWithHash))

		select {
		case <-time.After(time.Duration(timeout) * time.Millisecond):
			r <- true
		case <-callbackSuccessChann:
			r <- true
		}

		delete(e.onRplEndOfNames, channelWithoutHash)
	}()

	return r
}

// ChannelsOfUser tries to obtain channels of user under given nick.
// Sends the result on the returned channel.
func (e *Engine) ChannelsOfUser(nick string, timeout int64) chan []string {
	r := make(chan []string)

	go func() {
		defer close(r)

		callbackChann := make(chan []string)
		defer close(callbackChann)

		callback := func(packet Packet) {
			payload, payloadOk := packet.Payload.(RplWhoisChannelsPayload)

			if !payloadOk {
				return
			}

			if payload.nick == nick {
				callbackChann <- payload.channels
			}
		}
		e.onRplWhoisChannels[nick] = callback

		e.send(fmt.Sprintf("WHOIS %s", nick))

		select {
		case <-time.After(time.Duration(timeout) * time.Millisecond):
			r <- make([]string, 0)
		case channels := <-callbackChann:
			r <- channels
		}

		delete(e.onRplWhoisChannels, nick)
	}()

	return r
}

func (e *Engine) send(data string) {
	e.ircStream.Write([]byte(fmt.Sprintf("%s\r\n", data)))
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randNick() string {
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, nickLength)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
