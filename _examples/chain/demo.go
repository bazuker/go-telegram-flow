package chain

import (
	"go-telegram-flow/chain"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"time"
)

var flow *chain.Flow

var markup *tb.ReplyMarkup

func Run(token string) {
	var err error

	b, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		panic(err)
	}

	flow, err = chain.NewFlow("flow1", b)
	if err != nil {
		panic(err)
	}

	flow.DefaultHandler(defaultResponse).GetRoot().
		Then("name", stageName, tb.OnText).
		Then("phone", stagePhone, tb.OnContact).
		Then("location", stageLocation, tb.OnLocation)

	btnSharePhone := tb.ReplyButton{
		Contact: true,
		Text:    "Share phone number",
	}
	btnShareLocation := tb.ReplyButton{
		Location: true,
		Text:     "Share location",
	}
	replySharePhone := [][]tb.ReplyButton{
		{
			btnSharePhone,
		},
		{
			btnShareLocation,
		},
	}
	markup = &tb.ReplyMarkup{
		ReplyKeyboard: replySharePhone,
		ForceReply:    true,
	}

	b.Handle("/start", func(m *tb.Message) {
		log.Println("starting the flow for", m.Sender.Recipient())
		if err := flow.Start(m.Sender, "Hi, what's your name?"); err != nil {
			log.Println("failed to start the conversation", err)
		}
	})

	b.Handle(tb.OnText, func(m *tb.Message) {
		flow.Process(m)
	})

	b.Handle(tb.OnContact, func(m *tb.Message) {
		flow.Process(m)
	})

	b.Handle(tb.OnLocation, func(m *tb.Message) {
		flow.Process(m)
	})

	log.Println("starting...")

	b.Start()
}

func defaultResponse(e *chain.Node, m *tb.Message) bool {
	e.GetFlow().GetBot().Send(m.Sender, "Sorry, I dont' get you")
	return true // does not matter
}

func stageName(e *chain.Node, c *tb.Message) bool {
	if len(c.Text) < 2 {
		e.GetFlow().GetBot().Send(c.Sender, "Doesn't look like a name to me.. try again", markup)
		return false
	}
	log.Println(c.Sender.Recipient(), "goes through", e.GetId())
	e.GetFlow().GetBot().Send(c.Sender, "Good one! What's your phone?", markup)
	return true // continue
}

func stagePhone(e *chain.Node, c *tb.Message) bool {
	log.Println(c.Sender.Recipient(), "goes through", e.GetId())
	e.GetFlow().GetBot().Send(c.Sender, "Perfect. Just send me your location and we are done", markup)
	return true // continue
}

func stageLocation(e *chain.Node, c *tb.Message) bool {
	log.Println(c.Sender.Recipient(), "goes through", e.GetId())
	e.GetFlow().GetBot().Send(c.Sender, "You are all set now!")
	return true // continue
}
