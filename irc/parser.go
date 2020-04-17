package irc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

type PacketType int

const (
	Ping PacketType = iota
	RplWelcome
	RplWhoisChannels
	RplEndOfNames
	ErrNicknameInUse
	PrivMsgDccSend
	Unknown
)

type Packet struct {
	Type    PacketType
	Payload interface{}
}

type RplWhoisChannelsPayload struct {
	nick     string
	channels []string
}

type PrivMsgDccSendPayload struct {
	fileName   string
	fileLength uint64
	ip         net.IP
	port       uint64
}

const (
	ping             = "PING"
	privmsg          = "PRIVMSG"
	rplWelcome       = "001"
	rplWhoisChannels = "319"
	rplEndOfNames    = "366"
	errNicknameInUse = "433"
	dccSendMsgStart  = "\x01DCC SEND "
)

var pattern = regexp.MustCompile(
	fmt.Sprintf(
		"^:\\S* (%s|%s|%s|%s|%s) (\\S*) :?(.*)$",
		privmsg, rplWhoisChannels, rplWelcome, rplEndOfNames, errNicknameInUse,
	),
)
var pingPattern = regexp.MustCompile(
	fmt.Sprintf("^(%s) :(\\S*)$", ping),
)
var dccSendMsgPattern = regexp.MustCompile(
	"^\x01?DCC SEND \"?([^\"]*)\"? ([0-9]*) ([0-9]*) ([0-9]*)",
)

func Parse(line string) Packet {
	captures := pingPattern.FindAllStringSubmatch(line, -1)

	if captures == nil {
		captures = pattern.FindAllStringSubmatch(line, -1)
	}

	if captures == nil {
		return Packet{Type: Unknown}
	}
	parts := captures[0]

	if parts[1] == ping {
		return Packet{Type: Ping, Payload: parts[2]}
	}

	if parts[1] == rplWelcome {
		return Packet{Type: RplWelcome, Payload: parts[2]}
	}

	if parts[1] == rplWhoisChannels {
		return Packet{Type: RplWhoisChannels, Payload: parseRplWhoisChannelsPayload(parts[3])}
	}

	if parts[1] == rplEndOfNames {
		return Packet{Type: RplEndOfNames, Payload: parseRplEndOfNamesChannel(parts[3])}
	}

	if parts[1] == errNicknameInUse {
		return Packet{Type: ErrNicknameInUse, Payload: strings.SplitN(parts[3], " ", 2)[0]}
	}

	if parts[1] == privmsg {
		if strings.HasPrefix(parts[3], dccSendMsgStart) {
			payload, err := parseDccSendMsgPayload(parts[3])
			if err == nil {
				return Packet{Type: PrivMsgDccSend, Payload: payload}
			}
		}
	}

	return Packet{Type: Unknown}
}

func parseRplWhoisChannelsPayload(data string) RplWhoisChannelsPayload {
	parts := strings.Split(data, " ")
	channelTrashesPattern := regexp.MustCompile("^:?%#")

	channels := make([]string, len(parts)-1)
	for i, dirtyChannel := range parts[1:] {
		channels[i] = channelTrashesPattern.ReplaceAllString(dirtyChannel, "")
	}

	return RplWhoisChannelsPayload{nick: parts[0], channels: channels}
}

func parseRplEndOfNamesChannel(data string) string {
	return strings.SplitN(data, " ", 2)[0][1:]
}

func parseDccSendMsgPayload(data string) (PrivMsgDccSendPayload, error) {
	msgCaptures := dccSendMsgPattern.FindAllStringSubmatch(data, -1)

	if len(msgCaptures) >= 1 {
		msgParts := msgCaptures[0]
		if msgCaptures != nil && len(msgParts) >= 5 {
			port, parsePortErr := strconv.ParseUint(msgParts[3], 10, 64)
			fileLength, parseFileLengthErr := strconv.ParseUint(msgParts[4], 10, 64)
			ipU64, parseIPErr := strconv.ParseUint(msgParts[2], 10, 64)

			ip := make(net.IP, 4)
			binary.BigEndian.PutUint32(ip, uint32(ipU64))

			if parsePortErr != nil || parseFileLengthErr != nil || parseIPErr != nil {
				return PrivMsgDccSendPayload{}, errors.New("Could not parse number")
			}

			return PrivMsgDccSendPayload{fileName: msgParts[1], ip: ip, port: port, fileLength: fileLength}, nil
		}
	}

	return PrivMsgDccSendPayload{}, errors.New("Wrong format")
}
