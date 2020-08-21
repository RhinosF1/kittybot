package kitty

import (
	"bufio"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Reply sends a message to where the message came from (user or channel)
func (bot *Bot) Reply(m *Message, text string) {
	var target string
	if strings.Contains(m.To, "#") {
		target = m.To
	} else {
		target = m.From
	}
	bot.Msg(target, text)
}

// Msg sends a message to 'who' (user or channel)
func (bot *Bot) Msg(who, text string) {
	const command = "PRIVMSG"
	for _, line := range bot.splitText(text, command, who) {
		bot.Send(command + " " + who + " :" + line)
	}
}

// MsgMaxSize returns maximum number of bytes that fit into one message.
// Useful, for example, if you want to generate a wall of emojis that fit into one message,
// or you want to cap some output to one message
func (bot *Bot) MsgMaxSize(who string) int {
	const command = "PRIVMSG"
	maxSize := bot.maxMsgSize(command, who)
	return maxSize
}

// Notice sends a NOTICE message to 'who' (user or channel)
func (bot *Bot) Notice(who, text string) {
	const command = "NOTICE"
	for _, line := range bot.splitText(text, command, who) {
		bot.Send(command + " " + who + " :" + line)
	}
}

// NoticeMaxSize returns maximum number of bytes that fit into one message.
// Useful, for example, if you want to generate a wall of emojis that fit into one message,
// or you want to cap some output to one message
func (bot *Bot) NoticeMaxSize(who string) int {
	const command = "NOTICE"
	maxSize := bot.maxMsgSize(command, who)
	return maxSize
}

// Action sends an action to 'who' (user or channel)
func (bot *Bot) Action(who, text string) {
	msg := fmt.Sprintf("\u0001ACTION %s\u0001", text)
	bot.Msg(who, msg)
}

// Topic sets the channel 'c' topic (requires bot has proper permissions)
func (bot *Bot) Topic(c, topic string) {
	str := fmt.Sprintf("TOPIC %s :%s", c, topic)
	bot.Send(str)
}

// Send any command to the server
func (bot *Bot) Send(command string) {
	bot.outgoing <- command
}

// ChMode is used to change users modes in a channel
// operator = "+o" deop = "-o"
// ban = "+b"
func (bot *Bot) ChMode(user, channel, mode string) {
	bot.Send("MODE " + channel + " " + mode + " " + user)
}

// Join a channel
func (bot *Bot) Join(ch string) {
	bot.Send("JOIN " + ch)
}

// Part a channel
func (bot *Bot) Part(ch, msg string) {
	bot.Send("PART " + ch + " " + msg)
}

func (bot *Bot) isClosing() bool {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	return bot.closing
}

// SetNick sets the bots nick on the irc server.
// This does not alter *Bot.Nick, so be vary of that
func (bot *Bot) SetNick(nick string) {
	bot.Send(fmt.Sprintf("NICK %s", nick))
}

func (bot *Bot) maxMsgSize(command, who string) int {
	// Maximum message size that fits into 512 bytes.
	// Carriage return and linefeed are not counted here as they
	// are added by handleOutgoingMessages()
	maxSize := 510 - len(fmt.Sprintf(":%s %s %s :", bot.Prefix(), command, who))
	if _, present := bot.CapStatus(CapIdentifyMSG); present {
		maxSize--
	}
	// https://ircv3.net/specs/extensions/multiline
	if bot.MsgSafetyBuffer {
		maxSize -= 10
	}
	return maxSize
}

// Splits a given string into a string slice, in chunks ending
// either with \n, or with \r\n, or splitting text to maximally allowed size.
func (bot *Bot) splitText(text, command, who string) []string {
	var ret []string

	maxSize := bot.maxMsgSize(command, who)

	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		for len(line) > maxSize {
			totalSize := 0
			runeSize := 0
			// utf-8 aware splitting
			for _, v := range line {
				runeSize = utf8.RuneLen(v)
				if totalSize+runeSize > maxSize {
					ret = append(ret, line[:totalSize])
					line = line[totalSize:]
					totalSize = runeSize
					continue
				}
				totalSize += runeSize
			}

		}
		ret = append(ret, line)
	}
	return ret
}
