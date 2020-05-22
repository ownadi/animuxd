package irc

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"time"
)

const nickLength = 7

type onPacketCallback func(Packet)

type IRCEngine interface {
	IRCPacketsChann() chan Packet
	Join(channelName string, timeout int64) <-chan bool
	ChannelsOfUser(nick string, timeout int64) chan []string
	SendMessage(nick string, body string)
}

// An Engine represents that part of the app which is responsible
// for handling low-level IRC protocol related stuff.
type Engine struct {
	nick                    string
	ircStream               io.ReadWriteCloser
	ircPacketsChan          chan Packet
	onRplWelcome            map[string]onPacketCallback
	onErrNicknameInUse      map[string]onPacketCallback
	onRplEndOfNames         map[string]onPacketCallback
	onRplWhoisChannels      map[string]onPacketCallback
	onRplWelcomeMutex       *sync.RWMutex
	onErrNicknameInUseMutex *sync.RWMutex
	onRplEndOfNamesMutex    *sync.RWMutex
	onRplWhoisChannelsMutex *sync.RWMutex
}

// Nick returns current registered nick.
func (e *Engine) Nick() string {
	return e.nick
}

// IRCPacketsChann returns channel of packets.
func (e *Engine) IRCPacketsChann() chan Packet {
	return e.ircPacketsChan
}

// Start initializes the engine.
func (e *Engine) Start(ircStream io.ReadWriteCloser) {
	e.ircStream = ircStream
	e.nick = ""
	e.onErrNicknameInUse = map[string]onPacketCallback{}
	e.onRplWelcome = map[string]onPacketCallback{}
	e.onRplEndOfNames = map[string]onPacketCallback{}
	e.onRplWhoisChannels = map[string]onPacketCallback{}
	e.onErrNicknameInUseMutex = &sync.RWMutex{}
	e.onRplWelcomeMutex = &sync.RWMutex{}
	e.onRplEndOfNamesMutex = &sync.RWMutex{}
	e.onRplWhoisChannelsMutex = &sync.RWMutex{}

	ircScanner := bufio.NewScanner(e.ircStream)
	r := make(chan Packet, runtime.NumCPU())
	e.ircPacketsChan = r

	go func() {
		for ircScanner.Scan() {
			ircLine := ircScanner.Text()

			go func(line string) {
				packet := Parse(line)

				if packet.Type == RplWelcome {
					e.onRplWelcomeMutex.RLock()
					for _, callback := range e.onRplWelcome {
						callback(packet)
					}
					e.onRplWelcomeMutex.RUnlock()
				}

				if packet.Type == ErrNicknameInUse {
					e.onErrNicknameInUseMutex.RLock()
					for _, callback := range e.onErrNicknameInUse {
						callback(packet)
					}
					e.onErrNicknameInUseMutex.RUnlock()
				}

				if packet.Type == RplEndOfNames {
					e.onRplEndOfNamesMutex.RLock()
					for _, callback := range e.onRplEndOfNames {
						callback(packet)
					}
					e.onRplEndOfNamesMutex.RUnlock()
				}

				if packet.Type == RplWhoisChannels {
					e.onRplWhoisChannelsMutex.RLock()
					for _, callback := range e.onRplWhoisChannels {
						callback(packet)
					}
					e.onRplWhoisChannelsMutex.RUnlock()
				}

				if packet.Type == Ping {
					e.send(fmt.Sprintf("PONG :%s", packet.Payload))
				}

				r <- packet
			}(ircLine)
		}
	}()
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

			e.onRplWelcomeMutex.Lock()
			e.onRplWelcome[currentNick] = welcomeCallback
			e.onRplWelcomeMutex.Unlock()
			e.onErrNicknameInUseMutex.Lock()
			e.onErrNicknameInUse[currentNick] = nickTakenCallback
			e.onErrNicknameInUseMutex.Unlock()

			e.send(fmt.Sprintf("USER %s * * %s", currentNick, currentNick))
			e.send(fmt.Sprintf("NICK %s", currentNick))

			select {
			case <-time.After(time.Duration(timeout) * time.Millisecond):
				registrationSuccess = false
			case success := <-successChann:
				registrationSuccess = success
			}

			e.onRplWelcomeMutex.Lock()
			delete(e.onRplWelcome, currentNick)
			e.onRplWelcomeMutex.Unlock()
			e.onErrNicknameInUseMutex.Lock()
			delete(e.onErrNicknameInUse, currentNick)
			e.onErrNicknameInUseMutex.Unlock()
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
		e.onRplEndOfNamesMutex.Lock()
		e.onRplEndOfNames[channelWithoutHash] = callback
		e.onRplEndOfNamesMutex.Unlock()

		e.send(fmt.Sprintf("JOIN %s", channelWithHash))

		select {
		case <-time.After(time.Duration(timeout) * time.Millisecond):
			r <- true
		case <-callbackSuccessChann:
			r <- true
		}

		e.onRplEndOfNamesMutex.Lock()
		delete(e.onRplEndOfNames, channelWithoutHash)
		e.onRplEndOfNamesMutex.Unlock()
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
		e.onRplWhoisChannelsMutex.Lock()
		e.onRplWhoisChannels[nick] = callback
		e.onRplWhoisChannelsMutex.Unlock()

		e.send(fmt.Sprintf("WHOIS %s", nick))

		select {
		case <-time.After(time.Duration(timeout) * time.Millisecond):
			r <- make([]string, 0)
		case channels := <-callbackChann:
			r <- channels
		}

		e.onRplWhoisChannelsMutex.Lock()
		delete(e.onRplWhoisChannels, nick)
		e.onRplWhoisChannelsMutex.Unlock()
	}()

	return r
}

// SendMessage sends a message to user under given nick.
func (e *Engine) SendMessage(nick string, body string) {
	e.send(fmt.Sprintf("PRIVMSG %s :%s", nick, body))
}

func (e *Engine) send(data string) {
	fmt.Fprintf(e.ircStream, "%s\r\n", data)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randNick() string {
	b := make([]rune, nickLength)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
