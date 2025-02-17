package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ruskotwo/ready-checker/internal/config"
	"github.com/ruskotwo/ready-checker/internal/domain/pending"
	"github.com/ruskotwo/ready-checker/internal/utils"
	"log/slog"
	"strconv"
	"strings"
	"time"
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

		textRunes := []rune(msg.Text)
		statuses := make(pending.Statuses)
		for _, entity := range msg.Entities {
			if entity.Type == "mention" {
				username := string(textRunes[entity.Offset+1 : entity.Offset+entity.Length])

				statuses[username] = pending.Wait
			}
		}

		b.logger.Info(
			fmt.Sprintf(
				"Handle /check from %d@%s mention %s",
				msg.Chat.ID,
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

		timer, duration := b.pending.Start(strconv.Itoa(int(msg.Chat.ID)), statuses, utils.ExtractDuration(msg.Text))

		listMsg := tgbotapi.NewMessage(msg.Chat.ID, b.makeTextForList(statuses, duration))
		listMsg.ReplyToMessageID = msg.MessageID
		listMsg.ReplyMarkup = keyboard

		sentMsg, err := bot.Send(listMsg)
		if err != nil {
			b.logger.Error(err.Error())
		}

		go b.waitTimer(timer, bot, &sentMsg)
	}
}

func (b *Bot) handleCallback(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	b.logger.Info(fmt.Sprintf("Got %s from %d@%s", query.Data, query.Message.Chat.ID, query.From.UserName))

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

	result, statuses := b.pending.HandleStatus(
		strconv.Itoa(int(query.Message.Chat.ID)),
		query.From.UserName,
		status,
	)

	b.update(
		bot,
		query.Message,
		result,
		statuses,
		true,
	)
}

func (b *Bot) waitTimer(timer chan bool, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	for signal := range timer {
		if !signal {
			return
		}

		b.logger.Info(fmt.Sprintf("End timer for %d", msg.Chat.ID))

		result, statuses := b.pending.GetStatusesWithResult(strconv.Itoa(int(msg.Chat.ID)))

		b.update(
			bot,
			msg,
			result,
			statuses,
			false,
		)

	}
}

func (b *Bot) update(
	bot *tgbotapi.BotAPI,
	msg *tgbotapi.Message,
	result pending.Status,
	statuses pending.Statuses,
	keepWait bool,
) {
	listMsg := tgbotapi.NewEditMessageText(
		msg.Chat.ID,
		msg.MessageID,
		b.makeTextForList(statuses, utils.ExtractDuration(msg.Text)),
	)
	if keepWait && result == pending.Wait {
		listMsg.ReplyMarkup = msg.ReplyMarkup
	}
	if _, err := bot.Send(listMsg); err != nil {
		b.logger.Error(err.Error())
	}

	if keepWait && result == pending.Wait {
		b.logger.Info(fmt.Sprintf("Keep wait chat %d", msg.Chat.ID))
		return
	}

	resultMsg := tgbotapi.NewMessage(msg.Chat.ID, b.makeTextForResult(result, statuses))
	resultMsg.ReplyToMessageID = msg.MessageID

	if _, err := bot.Send(resultMsg); err != nil {
		b.logger.Error(err.Error())
	}
}

func (b *Bot) makeTextForList(pendingUsers pending.Statuses, duration time.Duration) string {
	text := fmt.Sprintf("Проверка готовности (%s):\n\n", utils.FormatDurationToString(duration))

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
	case pending.Wait:
		text += "⬜️ Время истекло.\n"
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
