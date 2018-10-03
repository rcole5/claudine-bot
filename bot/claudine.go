package bot

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gempir/go-twitch-irc"
	"github.com/jinzhu/gorm"
	"github.com/rcole5/claudine-bot"
	"github.com/rcole5/claudine-bot/models"
	"strings"
	"text/template"
	"time"
)

var (
	Client *twitch.Client
	service claudine_bot.Service
)

func New(s claudine_bot.Service, user string, token string, db *gorm.DB) {
	// Init the service
	service = s

	// Connect to twitch
	Client = twitch.NewClient(user, token)

	// Listen for new messages
	Client.OnNewMessage(handleMessage)

	// TODO: Everything about this is bad. Change it ASAP.
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		currentChannels := make(map[string]struct{})
		for range ticker.C {
			channels, err := s.ListChannel(context.Background())
			if err != nil {
				panic(err)
			}
			for _, channel := range channels {
				_, ok := currentChannels[string(channel)]
				if !ok {
					fmt.Println("Joined:", strings.TrimSpace(string(channel)))
					Client.Join(string(channel))
					currentChannels[string(channel)] = struct{}{}

					var commands []*models.Command
					db.Find(&commands)
					for _, comm := range commands {
						fmt.Println(comm.Channel)
						if comm.Channel != string(channel) {
							continue
						}
						fmt.Println("Added command", comm.Trigger)
						s.NewCommand(context.Background(), string(channel), claudine_bot.Command{
							Trigger: comm.Trigger,
							Action: comm.Action,
						})
					}

				}
			}
		}
	}()

	//for _, channel := range channels {
	//	fmt.Println("Joined:", strings.TrimSpace(string(channel)))
	//	Client.Join(string(channel))
	//}

	// Start the bot
	if err := Client.Connect(); err != nil {
		panic(err)
	}
}

func handleMessage(channel string, user twitch.User, message twitch.Message) {
	fmt.Printf("%s@%s: %s\n", user.DisplayName, channel, message.Text)
	msg := strings.Split(message.Text, " ")
	if msg[0][0] != '!' {
		fmt.Println("Not a command")
		return
	}

	if isMod(user) {
		if msg[0] == "!add" {
			if len(msg) < 3 {
				Client.Say(channel, "Not enough args. Syntax is !add command response.")
				return
			}
			_, err := service.NewCommand(context.Background(), channel, claudine_bot.Command{
				Trigger: msg[1],
				Action: strings.Join(msg[2:], " "),
			})
			if err != nil {
				Client.Say(channel, "This command already exists.")
				return
			}
			Client.Say(channel, "Command added. VoHiYo")
			return
		} else if msg[0] == "!remove" {
			if len(msg) < 2 {
				Client.Say(channel, "Not enough args. Syntax is !remove command.")
				return
			}
			err := service.DeleteCommand(context.Background(), channel, msg[1])
			if err != nil {
				Client.Say(channel, "This commands doesn't exist, baka.")
				return
			}
			Client.Say(channel, "Command deleted.")
			return
		}
	}
	command, err := service.GetCommand(context.Background(), channel, msg[0][1:])
	if err != nil {
		// Command doesn't exist. Should prob add an error code here in case it's a different error.
	}

	if command.Trigger != "" {
		// Parse any variables
		t, err := template.New("Parse Command").Parse(command.Action)
		if err != nil {
			fmt.Println(err)
			Client.Say(channel, "Failed to parse command")
			return
		}

		// Prepare the variables
		vars := Variables{
			User: user.Username,
			UserID: user.UserID,
		}

		buf := new(bytes.Buffer)
		t.Execute(buf, vars)

		Client.Say(channel, buf.String())
	}
}

type Variables struct {
	User string
	UserID int64
}

func isMod(user twitch.User) bool {
	return user.Badges["broadcaster"] == 1 || user.Badges["moderator"] == 1
}
