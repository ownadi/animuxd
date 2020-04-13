package irc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRplWelcomePacket(t *testing.T) {
	res := Parse(":irc.infernet.org 001 foobar :Welcome to the Rizon Internet Relay Chat Network gcrrvjzfGr")

	assert.Equal(t, RplWelcome, res.Type)
	assert.Equal(t, "foobar", res.Payload)
}

func TestRplWhoisChannelsPacket(t *testing.T) {
	res := Parse(":magnet.rizon.net 319 foo Ginpachi-Sensei :%#HorribleSubs %#NIBL %#news")

	assert.Equal(t, RplWhoisChannels, res.Type)

	payload, ok := res.Payload.(RplWhoisChannelsPayload)
	assert.True(t, ok)

	assert.Equal(t, "Ginpachi-Sensei", payload.nick)
	assert.Equal(t, []string{"HorribleSubs", "NIBL", "news"}, payload.channels)
}

func TestRplEndOfNames(t *testing.T) {
	res := Parse(":irc.rizon.club 366 gharibol #footest :End of /NAMES list.")

	assert.Equal(t, RplEndOfNames, res.Type)
	assert.Equal(t, "footest", res.Payload)
}

func TestErrNicknameInUse(t *testing.T) {
	res := Parse(":magnet.rizon.net 433 * gourangaharibol :Nickname is already in use.")

	assert.Equal(t, ErrNicknameInUse, res.Type)
	assert.Equal(t, "gourangaharibol", res.Payload)
}

func TestUnknownOnRandomInput(t *testing.T) {
	res := Parse("FOO BAR BAZ")

	assert.Equal(t, Unknown, res.Type)
}

func TestOnlySubsetOfIrc(t *testing.T) {
	res := Parse(":solenoid.rizon.net 002 a1bcwy :Your host is solenoid.rizon.net, running version plexus-4(hybrid-8.1.20)")

	assert.Equal(t, Unknown, res.Type)
}

func TestPrivMsgDccSend(t *testing.T) {
	res := Parse(":Gintoki!~Gin@oshiete.ginpachi.sensei PRIVMSG ownadi :\x01DCC SEND Gin.txt 2130706433 39095 339260|")

	assert.Equal(t, PrivMsgDccSend, res.Type)

	payload, ok := res.Payload.(PrivMsgDccSendPayload)
	assert.True(t, ok)

	assert.Equal(t, "Gin.txt", payload.fileName)
	assert.Equal(t, uint64(339260), payload.fileLength)
	assert.Equal(t, uint64(39095), payload.port)
	assert.Equal(t, "127.0.0.1", payload.ip.String())
}

func TestPrivMsgDccSendWithSpaces(t *testing.T) {
	res := Parse(":[C-W]Archive!~sakura@distro.cartoon-world.org PRIVMSG av1vfca :\x01DCC SEND \"Great Teacher Onizuka - 25 [x264-AC3-DVD][Sakura][C-W][B9F96CF8].mkv\" 2130706433 48467 541715509|")

	assert.Equal(t, PrivMsgDccSend, res.Type)

	payload, ok := res.Payload.(PrivMsgDccSendPayload)
	assert.True(t, ok)

	assert.Equal(t, "Great Teacher Onizuka - 25 [x264-AC3-DVD][Sakura][C-W][B9F96CF8].mkv", payload.fileName)
	assert.Equal(t, uint64(541715509), payload.fileLength)
	assert.Equal(t, uint64(48467), payload.port)
	assert.Equal(t, "127.0.0.1", payload.ip.String())
}

func TestPrivMsgDccSendBroken(t *testing.T) {
	res := Parse(":[C-W]Archive!~sakura@distro.cartoon-world.org PRIVMSG av1vfca :\x01DCC SEND \"Great Teacher Onizuka - 25 [x264-AC3-DVD][Sakura][C-W][B9F96CF8].mkv\" 213070foo bar baz|")

	assert.Equal(t, Unknown, res.Type)
}

func TestRandomPrivMsg(t *testing.T) {
	res := Parse(":[C-W]Archive!~sakura@distro.cartoon-world.org PRIVMSG av1vfca :Hello!")

	assert.Equal(t, Unknown, res.Type)
}
