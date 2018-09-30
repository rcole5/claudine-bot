package twitch

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// MessageType different message types possible to receive via IRC
type MessageType int

const (
	// WHISPER private messages
	WHISPER MessageType = 0
	// PRIVMSG standard chat message
	PRIVMSG MessageType = 1
	// CLEARCHAT timeout messages
	CLEARCHAT MessageType = 2
	// ROOMSTATE changes like sub mode
	ROOMSTATE MessageType = 3
	// USERNOTICE messages like subs, resubs, raids, etc
	USERNOTICE MessageType = 4
	// USERSTATE messages
	USERSTATE MessageType = 5
)

type message struct {
	Type        MessageType
	Time        time.Time
	Channel     string
	UserID      int64
	Username    string
	DisplayName string
	UserType    string
	Color       string
	Action      bool
	Badges      map[string]int
	Emotes      []*Emote
	Tags        map[string]string
	Text        string
	Raw         string
}

// Emote twitch emotes
type Emote struct {
	Name  string
	ID    string
	Count int
}

func parseMessage(line string) *message {
	if !strings.HasPrefix(line, "@") {
		return &message{
			Text: line,
			Raw:  line,
		}
	}
	spl := strings.SplitN(line, " :", 3)
	if len(spl) < 3 {
		return parseOtherMessage(line)
	}
	action := false
	tags, middle, text := spl[0], spl[1], spl[2]
	if strings.HasPrefix(text, "\u0001ACTION ") && strings.HasSuffix(text, "\u0001") {
		action = true
		text = text[8 : len(text)-1]
	}
	msg := &message{
		Time:   time.Now(),
		Text:   text,
		Tags:   map[string]string{},
		Action: action,
	}
	msg.Username, msg.Type, msg.Channel = parseMiddle(middle)
	parseTags(msg, tags[1:])
	msg.UserID, _ = strconv.ParseInt(msg.Tags["user-id"], 10, 64)
	if msg.Type == CLEARCHAT {
		targetUser := msg.Text
		msg.Username = targetUser

		msg.Text = fmt.Sprintf("%s was timed out for %s: %s", targetUser, msg.Tags["ban-duration"], msg.Tags["ban-reason"])
	}
	msg.Raw = line
	return msg
}

func parseOtherMessage(line string) *message {
	msg := &message{}
	split := strings.Split(line, " ")

	switch split[2] {
	case "ROOMSTATE":
		msg.Type = ROOMSTATE
	case "USERNOTICE":
		msg.Type = USERNOTICE
	case "CLEARCHAT":
		msg.Type = CLEARCHAT
	case "USERSTATE":
		msg.Type = USERSTATE
	}
	msg.Tags = make(map[string]string)

	// Parse out channel if it exists in this line
	if len(split) >= 4 && len(split[3]) > 1 && split[3][0] == '#' {
		// Remove # from channel
		msg.Channel = split[3][1:]
	}

	tagsString := strings.Fields(strings.TrimPrefix(split[0], "@"))
	tags := strings.Split(tagsString[0], ";")
	for _, tag := range tags {
		tagSplit := strings.Split(tag, "=")

		value := ""
		if len(tagSplit) > 1 {
			value = tagSplit[1]
		}

		msg.Tags[tagSplit[0]] = value
	}

	if msg.Type == CLEARCHAT {
		msg.Text = "Chat has been cleared by a moderator"
	}
	return msg
}

func parseMiddle(middle string) (string, MessageType, string) {
	var username string
	var msgType MessageType
	var channel string

	for i, c := range middle {
		if c == '!' {
			username = middle[:i]
			middle = middle[i:]
		}
	}
	start := -1
	for i, c := range middle {
		if c == ' ' {
			if start == -1 {
				start = i + 1
			} else {
				typ := middle[start:i]
				switch typ {
				case "PRIVMSG":
					msgType = PRIVMSG
				case "WHISPER":
					msgType = WHISPER
				case "CLEARCHAT":
					msgType = CLEARCHAT
				}
				middle = middle[i:]
			}
		}
	}
	for i, c := range middle {
		if c == '#' {
			channel = middle[i+1:]
		}
	}

	return username, msgType, channel
}

func parseTags(msg *message, tagsRaw string) {
	tags := strings.Split(tagsRaw, ";")
	for _, tag := range tags {
		spl := strings.SplitN(tag, "=", 2)
		value := strings.Replace(spl[1], "\\:", ";", -1)
		value = strings.Replace(value, "\\s", " ", -1)
		value = strings.Replace(value, "\\\\", "\\", -1)
		switch spl[0] {
		case "badges":
			msg.Badges = parseBadges(value)
		case "color":
			msg.Color = value
		case "display-name":
			msg.DisplayName = value
		case "emotes":
			msg.Emotes = parseTwitchEmotes(value, msg.Text)
		case "user-type":
			msg.UserType = value
		default:
			msg.Tags[spl[0]] = value
		}
	}
}

func parseBadges(badges string) map[string]int {
	m := map[string]int{}
	spl := strings.Split(badges, ",")
	for _, badge := range spl {
		s := strings.SplitN(badge, "/", 2)
		if len(s) < 2 {
			continue
		}
		n, _ := strconv.Atoi(s[1])
		m[s[0]] = n
	}
	return m
}

func parseTwitchEmotes(emoteTag, text string) []*Emote {
	emotes := []*Emote{}

	if emoteTag == "" {
		return emotes
	}

	runes := []rune(text)

	emoteSlice := strings.Split(emoteTag, "/")
	for i := range emoteSlice {
		spl := strings.Split(emoteSlice[i], ":")
		pos := strings.Split(spl[1], ",")
		sp := strings.Split(pos[0], "-")
		start, _ := strconv.Atoi(sp[0])
		end, _ := strconv.Atoi(sp[1])
		id := spl[0]
		e := &Emote{
			ID:    id,
			Count: strings.Count(emoteSlice[i], "-"),
			Name:  string(runes[start : end+1]),
		}

		emotes = append(emotes, e)
	}
	return emotes
}
