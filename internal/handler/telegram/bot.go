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
	config  *config.AppConfig
	logger  *slog.Logger
	pending *pending.Pending
}

func NewBot(
	config *config.AppConfig,
	logger *slog.Logger,
	pending *pending.Pending,
) *Bot {
	return &Bot{
		config,
		logger,
		pending,
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

		statuses := make(pending.Statuses)
		for _, entity := range msg.Entities {
			if entity.Type == "mention" {
				statuses[msg.Text[entity.Offset+1:entity.Offset+entity.Length]] = pending.Wait
			}
		}

		b.logger.Info(
			fmt.Sprintf(
				"Handle /check from %s mention %s",
				msg.From.UserName,
				fmt.Sprint(statuses),
			),
		)

		if len(statuses) == 0 {
			b.sendErrorNeedMentions(bot, msg)
			return
		}

		statuses[msg.From.UserName] = pending.Wait

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Готов", readyButton),
				tgbotapi.NewInlineKeyboardButtonData("Не готов", canselButton),
			),
		)

		listMsg := tgbotapi.NewMessage(msg.Chat.ID, b.makeTextForList(statuses))
		listMsg.ReplyToMessageID = msg.MessageID
		listMsg.ReplyMarkup = keyboard

		if err := b.pending.Start(strconv.Itoa(int(msg.Chat.ID)), statuses); err != nil {
			b.logger.Error(err.Error())
			return
		}

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

	var status pending.Status
	switch query.Data {
	case readyButton:
		status = pending.Ready
	case canselButton:
		status = pending.Cancel
	}

	result, statuses, err := b.pending.Update(
		strconv.Itoa(int(query.Message.Chat.ID)),
		query.From.UserName,
		status,
	)
	if err != nil {
		return
	}

	listMsg := tgbotapi.NewEditMessageText(
		query.Message.Chat.ID,
		query.Message.MessageID,
		b.makeTextForList(statuses),
	)
	if result == pending.Wait {
		listMsg.ReplyMarkup = query.Message.ReplyMarkup
	}
	if _, err := bot.Send(listMsg); err != nil {
		b.logger.Error(err.Error())
	}

	if result == pending.Wait {
		return
	}

	resultMsg := tgbotapi.NewMessage(query.Message.Chat.ID, b.makeTextForResult(result, statuses))
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

func (b *Bot) makeTextForResult(result pending.Status, statuses pending.Statuses) (text string) {
	switch result {
	case pending.Undefined:
		text += "🟨 Не все подтвердили готовность\n"
	case pending.Ready:
		text += "🟩 Все подтвердили готовность!\n"
	case pending.Cancel:
		text += "🟥 Никто не подтвердил готовность.\n"
	}

	for user := range statuses {
		text += fmt.Sprintf("@%s ", user)
	}

	return
}

func (b *Bot) sendErrorNeedMentions(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	errorMsg := tgbotapi.NewMessage(msg.Chat.ID, "Необходимо упомянуть пользователей")
	errorMsg.ReplyToMessageID = msg.MessageID
	if _, err := bot.Send(errorMsg); err != nil {
		b.logger.Error(err.Error())
	}
}
