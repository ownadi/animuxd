package irc

import (
	"bufio"
	"fmt"
	"net"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var userPattern = regexp.MustCompile("USER (\\S*) \\* \\* (\\S*)")
var nickPattern = regexp.MustCompile("NICK (\\S*)")

func TestNick(t *testing.T) {
	engine := &Engine{nick: "foo"}

	assert.Equal(t, "foo", engine.Nick())
}

func TestReqister(t *testing.T) {
	client, server := net.Pipe()
	reader := bufio.NewReader(client)

	engine := &Engine{}
	engine.Start(server)
	registerPromise := engine.Register(999999)

	userRequest, _ := reader.ReadString('\r')
	userRequestParts := userPattern.FindAllStringSubmatch(userRequest, -1)[0]
	assert.Len(t, userRequestParts, 1+2)

	nickRequest, _ := reader.ReadString('\r')
	nickRequestParts := nickPattern.FindAllStringSubmatch(nickRequest, -1)[0]
	assert.Len(t, nickRequestParts, 1+1)

	client.Write([]byte(fmt.Sprintf(":irc.infernet.org 001 %s :Welcome to the Rizon Internet Relay Chat Network gcrrvjzfGr\r\n", nickRequestParts[1])))

	assert.True(t, <-registerPromise)
}

func TestReqisterNickInUse(t *testing.T) {
	client, server := net.Pipe()
	reader := bufio.NewReader(client)

	engine := &Engine{}
	engine.Start(server)
	registerPromise := engine.Register(999999)

	userRequest, _ := reader.ReadString('\r')
	userRequestParts := userPattern.FindAllStringSubmatch(userRequest, -1)[0]
	assert.Len(t, userRequestParts, 1+2)

	nickRequest, _ := reader.ReadString('\r')
	nickRequestParts := nickPattern.FindAllStringSubmatch(nickRequest, -1)[0]
	assert.Len(t, nickRequestParts, 1+1)
	firstNick := nickRequestParts[1]

	client.Write([]byte(fmt.Sprintf(":magnet.rizon.net 433 * %s :Nickname is already in use.\r\n", firstNick)))

	userRequest, _ = reader.ReadString('\r')
	userRequestParts = userPattern.FindAllStringSubmatch(userRequest, -1)[0]
	assert.Len(t, userRequestParts, 1+2)

	nickRequest, _ = reader.ReadString('\r')
	nickRequestParts = nickPattern.FindAllStringSubmatch(nickRequest, -1)[0]
	assert.Len(t, nickRequestParts, 1+1)
	secondNick := nickRequestParts[1]

	client.Write([]byte(fmt.Sprintf(":irc.infernet.org 001 %s :Welcome to the Rizon Internet Relay Chat Network gcrrvjzfGr\r\n", secondNick)))

	assert.True(t, <-registerPromise)

	assert.NotEqual(t, firstNick, secondNick)
	assert.NotEqual(t, firstNick, engine.Nick())
	assert.Equal(t, secondNick, engine.Nick())
}

func TestReqisterTimeouts(t *testing.T) {
	client, server := net.Pipe()
	reader := bufio.NewReader(client)

	engine := &Engine{}
	engine.Start(server)
	registerPromise := engine.Register(50)

	userRequest, _ := reader.ReadString('\r')
	userRequestParts := userPattern.FindAllStringSubmatch(userRequest, -1)[0]
	assert.Len(t, userRequestParts, 1+2)

	nickRequest, _ := reader.ReadString('\r')
	nickRequestParts := nickPattern.FindAllStringSubmatch(nickRequest, -1)[0]
	assert.Len(t, nickRequestParts, 1+1)
	firstNick := nickRequestParts[1]

	time.Sleep(time.Duration(50) * time.Millisecond)

	userRequest, _ = reader.ReadString('\r')
	userRequestParts = userPattern.FindAllStringSubmatch(userRequest, -1)[0]
	assert.Len(t, userRequestParts, 1+2)

	nickRequest, _ = reader.ReadString('\r')
	nickRequestParts = nickPattern.FindAllStringSubmatch(nickRequest, -1)[0]
	assert.Len(t, nickRequestParts, 1+1)
	secondNick := nickRequestParts[1]

	client.Write([]byte(fmt.Sprintf(":irc.infernet.org 001 %s :Welcome to the Rizon Internet Relay Chat Network gcrrvjzfGr\r\n", secondNick)))

	assert.True(t, <-registerPromise)

	assert.NotEqual(t, firstNick, engine.Nick())
	assert.Equal(t, secondNick, engine.Nick())
}

func TestPongs(t *testing.T) {
	client, server := net.Pipe()
	scanner := bufio.NewScanner(client)

	engine := &Engine{}
	engine.Start(server)

	client.Write([]byte("PING :foo\r\n"))
	scanner.Scan()
	response := scanner.Text()

	assert.Equal(t, "PONG :foo", response)
}

func TestJoin(t *testing.T) {
	client, server := net.Pipe()
	scanner := bufio.NewScanner(client)

	engine := &Engine{}
	engine.Start(server)

	promise := engine.Join("foo", 999999)

	scanner.Scan()
	assert.Equal(t, "JOIN #foo", scanner.Text())

	client.Write([]byte(":irc.rizon.club 366 gharibol #foo :End of /NAMES list.\r\n"))

	assert.True(t, <-promise)
}

func TestJoinTimeouts(t *testing.T) {
	client, server := net.Pipe()
	scanner := bufio.NewScanner(client)

	engine := &Engine{}
	engine.Start(server)

	promise := engine.Join("foo", 50)

	scanner.Scan()
	assert.Equal(t, "JOIN #foo", scanner.Text())

	assert.True(t, <-promise)
}

func TestChannelsOfUser(t *testing.T) {
	client, server := net.Pipe()
	scanner := bufio.NewScanner(client)

	engine := &Engine{}
	engine.Start(server)

	promise := engine.ChannelsOfUser("JohnDoe", 999999)

	scanner.Scan()
	assert.Equal(t, "WHOIS JohnDoe", scanner.Text())

	client.Write([]byte(":magnet.rizon.net 319 foo JohnDoe :%#HorribleSubs %#NIBL %#news\r\n"))

	assert.Equal(t, []string{"HorribleSubs", "NIBL", "news"}, <-promise)
}

func TestChannelsOfUserTimeouts(t *testing.T) {
	client, server := net.Pipe()
	scanner := bufio.NewScanner(client)

	engine := &Engine{}
	engine.Start(server)

	promise := engine.ChannelsOfUser("JohnDoe", 50)

	scanner.Scan()
	assert.Equal(t, "WHOIS JohnDoe", scanner.Text())

	assert.Equal(t, []string{}, <-promise)
}
