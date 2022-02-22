package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

func Send(bot *tb.Bot, to tb.Recipient, msg string) {
	_, err := bot.Send(to, msg, tb.ModeMarkdown)
	if err != nil {
		log.Fatalf("message send error: %v", err)
		return
	}
}

func SendError(bot *tb.Bot, to tb.Recipient, msg string) {
	_, err := bot.Send(to, msg, tb.ModeMarkdown)
	if err != nil {
		log.Fatalf("message send error: %v", err)
		return
	}
}

func SendAdmin(bot *tb.Bot, to []User, msg string) {
	SendMany(bot, to, fmt.Sprintf("*[Admin]* %s", msg))
}

func SendKeyboardList(bot *tb.Bot, to tb.Recipient, msg string, list []string) {
	var buttons []tb.ReplyButton
	for _, item := range list {
		buttons = append(buttons, tb.ReplyButton{Text: item})
	}

	var replyKeys [][]tb.ReplyButton
	for _, b := range buttons {
		replyKeys = append(replyKeys, []tb.ReplyButton{b})
	}

	_, err := bot.Send(to, msg, &tb.ReplyMarkup{
		ReplyKeyboard:   replyKeys,
		OneTimeKeyboard: true,
	})
	if err != nil {
		log.Fatalf("message send error: %v", err)
		return
	}
}

func SendMany(bot *tb.Bot, to []User, msg string) {
	for _, user := range to {
		_, err := bot.Send(user, msg, tb.ModeMarkdown)
		if err != nil {
			log.Fatalf("message send error: %v", err)
			return
		}
	}
}

func DisplayName(u *tb.User) string {
	if u.FirstName != "" && u.LastName != "" {
		return EscapeMarkdown(fmt.Sprintf("%s %s", u.FirstName, u.LastName))
	}

	return EscapeMarkdown(u.FirstName)
}

func EscapeMarkdown(s string) string {
	s = strings.Replace(s, "[", "\\[", -1)
	s = strings.Replace(s, "]", "\\]", -1)
	s = strings.Replace(s, "_", "\\_", -1)
	return s
}

func FormatDate(t time.Time) string {
	if t.IsZero() {
		return "Unknown"
	}
	return t.Format("02.01.2006")
}

func FormatDateTime(t time.Time) string {
	if t.IsZero() {
		return "Unknown"
	}
	return t.Format("02.01.2006 15:04:05")
}

func GetRootFolderFromPath(path string) string {
	return strings.Title(filepath.Base(filepath.Dir(path)))
}

func GetUserName(m *tb.Message) string {
	var username string
	if len(m.Sender.Username) > 0 {
		username = m.Sender.Username
	} else {
		username = fmt.Sprintf("%s %s", m.Sender.FirstName, m.Sender.LastName)
	}
	return strings.TrimSpace(strings.ToLower(username))
}
