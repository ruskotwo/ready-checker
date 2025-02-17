package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ruskotwo/ready-checker/internal/config"
	"github.com/ruskotwo/ready-checker/internal/domain/pending"
	"log/slog"
	"strconv"
	"strings"
)

const (
	readyButton  = "ready"
	canselButton = "cansel"
)

type Bot struct {
	config         *config.AppConfig
	logger         *slog.Logger
	pendingStorage *pending.Storage
}

func NewBot(
	config *config.AppConfig,
	logger *slog.Logger,
	pendingStorage *pending.Storage,
) *Bot {
	return &Bot{
		config,
		logger,
		pendingStorage,
	}
}

func (b *Bot) Start() {
	bot, err := tgbotapi.NewBotAPI(b.config.TelegramBotToken)
	if err != nil {
		panic(err)
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		switch true {
		case update.Message != nil:
			b.handleMessage(bot, update.Message)
		case update.CallbackQuery != nil:
			b.handleCallback(bot, update.CallbackQuery)
		}
	}
}

func (b *Bot) handleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	if strings.HasPrefix(msg.Text, "/check") {
		if len(msg.Entities) == 0 {
			b.sendErrorNeedMentions(bot, msg)
		}

		pendingUsers := make(pending.Statuses)
		for _, entity := range msg.Entities {
			if entity.Type == "mention" {
				pendingUsers[msg.Text[entity.Offset+1:entity.Offset+entity.Length]] = pending.Wait
			}
		}

		b.logger.Info(
			fmt.Sprintf(
				"Handle /check from %s mention %s",
				msg.From.UserName,
				fmt.Sprint(pendingUsers),
			),
		)

		if len(pendingUsers) == 0 {
			b.sendErrorNeedMentions(bot, msg)
			return
		}

		pendingUsers[msg.From.UserName] = pending.Wait

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Готов", readyButton),
				tgbotapi.NewInlineKeyboardButtonData("Не готов", canselButton),
			),
		)

		listMsg := tgbotapi.NewMessage(msg.Chat.ID, b.makeTextForList(pendingUsers))
		listMsg.ReplyToMessageID = msg.MessageID
		listMsg.ReplyMarkup = keyboard

		b.pendingStorage.Clean(strconv.Itoa(int(msg.Chat.ID)))
		b.pendingStorage.SetMany(strconv.Itoa(int(msg.Chat.ID)), pendingUsers)

		if _, err := bot.Send(listMsg); err != nil {
			b.logger.Error(err.Error())
		}
	}
}

func (b *Bot) handleCallback(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	b.logger.Info(fmt.Sprintf("Got %s from %s", query.Data, query.From.UserName))

	if query.Data != readyButton && query.Data != canselButton {
		return
	}

	pendingUsers := b.pendingStorage.Get(strconv.Itoa(int(query.Message.Chat.ID)))

	switch query.Data {
	case readyButton:
		pendingUsers[query.From.UserName] = pending.Ready
	case canselButton:
		pendingUsers[query.From.UserName] = pending.Cancel
	}

	nobodyWait := true
	for _, status := range pendingUsers {
		if status == pending.Wait {
			nobodyWait = false
			break
		}
	}

	listMsg := tgbotapi.NewEditMessageText(
		query.Message.Chat.ID,
		query.Message.MessageID,
		b.makeTextForList(pendingUsers),
	)
	if !nobodyWait {
		listMsg.ReplyMarkup = query.Message.ReplyMarkup
	}
	if _, err := bot.Send(listMsg); err != nil {
		b.logger.Error(err.Error())
	}

	b.pendingStorage.Set(strconv.Itoa(int(query.Message.Chat.ID)), query.From.UserName, pendingUsers[query.From.UserName])

	if !nobodyWait {
		return
	}

	resultMsg := tgbotapi.NewMessage(query.Message.Chat.ID, b.makeTextForResult(pendingUsers))
	resultMsg.ReplyToMessageID = query.Message.MessageID

	if _, err := bot.Send(resultMsg); err != nil {
		b.logger.Error(err.Error())
	}
}

func (b *Bot) makeTextForList(pendingUsers pending.Statuses) string {
	text := "Проверка готовности:\n\n"

	for user, status := range pendingUsers {
		switch status {
		case pending.Wait:
			text += "⬜️"
		case pending.Ready:
			text += "🟩"
		case pending.Cancel:
			text += "🟥"
		}

		text += fmt.Sprintf(" @%s\n", user)
	}

	return text
}

func (b *Bot) makeTextForResult(pendingUsers pending.Statuses) string {
	text := "От всех пользователей получен ответ:\n"

	for user := range pendingUsers {
		text += fmt.Sprintf("@%s ", user)
	}

	return text
}

func (b *Bot) sendErrorNeedMentions(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	errorMsg := tgbotapi.NewMessage(msg.Chat.ID, "Необходимо упомянуть пользователей")
	errorMsg.ReplyToMessageID = msg.MessageID
	if _, err := bot.Send(errorMsg); err != nil {
		b.logger.Error(err.Error())
	}
}
