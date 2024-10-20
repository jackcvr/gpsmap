package main

import (
	"github.com/jackcvr/gpsmap/orm"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"gorm.io/gorm"
	"log"
)

type TGBot struct {
	*telego.Bot
	db *gorm.DB
}

func (bot *TGBot) SendMessage(params *telego.SendMessageParams) error {
	_, err := bot.Bot.SendMessage(params)
	return err
}

func (bot *TGBot) Broadcast(msg string, ids ...uint) error {
	for _, id := range ids {
		if err := bot.SendMessage(tu.Message(tu.ID(int64(id)), msg)); err != nil {
			return err
		}
	}
	return nil
}

func (bot *TGBot) NotifyUsers(msg string) error {
	var ids []uint
	if err := bot.db.Model(&orm.TGChat{}).Pluck("ID", &ids).Error; err != nil {
		return err
	}
	return bot.Broadcast(msg, ids...)
}

func StartTGBot(config TGBotConfig, db *gorm.DB, debug bool) *TGBot {
	var opts []telego.BotOption
	if debug {
		opts = append(opts, telego.WithDefaultDebugLogger())
	}
	telegoBot, err := telego.NewBot(config.Token, opts...)
	if err != nil {
		panic(err)
	}
	bot := &TGBot{Bot: telegoBot, db: db}
	sendMessage := func(params *telego.SendMessageParams) {
		if err = bot.SendMessage(params); err != nil {
			errLog.Print(err)
		}
	}

	var botUser *telego.User
	if botUser, err = bot.GetMe(); err != nil {
		panic(err)
	} else {
		log.Printf("Bot info: %+v", botUser)
	}

	var updates <-chan telego.Update
	updates, err = bot.UpdatesViaLongPolling(nil)
	if err != nil {
		panic(err)
	}
	var bh *th.BotHandler
	bh, err = th.NewBotHandler(bot.Bot, updates)
	if err != nil {
		panic(err)
	}

	bh.Handle(func(_ *telego.Bot, update telego.Update) {
		msg := tu.Message(
			tu.ID(update.Message.Chat.ID),
			"receiver added",
		)
		chat := orm.TGChat{ID: uint(update.Message.Chat.ID)}
		if err = db.Create(&chat).Error; err != nil {
			errLog.Print(err)
			msg = tu.Messagef(
				tu.ID(update.Message.Chat.ID),
				"error: %s",
				err,
			)
		}
		sendMessage(msg)
	}, th.CommandEqual("start"))

	bh.Handle(func(_ *telego.Bot, update telego.Update) {
		msg := tu.Message(
			tu.ID(update.Message.Chat.ID),
			"receiver removed",
		)
		chat := orm.TGChat{ID: uint(update.Message.Chat.ID)}
		if err = db.Delete(&chat).Error; err != nil {
			errLog.Print(err)
			msg = tu.Messagef(
				tu.ID(update.Message.Chat.ID),
				"error: %s",
				err,
			)
		}
		sendMessage(msg)
	}, th.CommandEqual("stop"))

	bh.Handle(func(_ *telego.Bot, update telego.Update) {
		sendMessage(tu.Message(
			tu.ID(update.Message.Chat.ID),
			"Unknown command",
		))
	}, th.AnyCommand())

	go func() {
		defer bh.Stop()
		defer bot.StopLongPolling()
		bh.Start()
	}()

	return bot
}
