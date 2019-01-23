package chain

import (
	"go-telegram-flow/chain"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"strings"
	"time"
)

var flow *chain.Chain

var markup *tb.ReplyMarkup

func Run(token string) {
	b, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		panic(err)
	}

	flow, err = chain.NewChainFlow("flow1", b)
	if err != nil {
		panic(err)
	}

	flow.SetDefaultHandler(defaultResponse).GetRoot().
		Then("name", stageName, tb.OnText).
		Then("phone", stagePhone, tb.OnContact).
		Then("share_location", stageShareLocation, tb.OnText).
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

	log.Println("starting...", b.Me.Username)

	b.Start()
}

func defaultResponse(e *chain.Node, m *tb.Message) *chain.Node {
	e.GetFlow().GetBot().Send(m.Sender, "Sorry, I dont' get you")
	return e // stays on the same stage
}

func stageName(e *chain.Node, c *tb.Message) *chain.Node {
	if len(c.Text) < 2 {
		e.GetFlow().GetBot().Send(c.Sender, "Doesn't look like a name to me.. try again", markup)
		return e // stays on the same stage
	}
	log.Println(c.Sender.Recipient(), "goes through", e.GetId())
	e.GetFlow().GetBot().Send(c.Sender, "Good one! What's your phone?", markup)
	return e.Next() // continue
}

func stagePhone(e *chain.Node, c *tb.Message) *chain.Node {
	log.Println(c.Sender.Recipient(), "goes through", e.GetId())
	e.GetFlow().GetBot().Send(c.Sender, "Perfect. Would you mind sharing your location? Yes/no", markup)
	return e.Next() // continue
}

func stageShareLocation(e *chain.Node, c *tb.Message) *chain.Node {
	log.Println(c.Sender.Recipient(), "goes through", e.GetId())
	text := strings.ToLower(c.Text)
	if text == "yes" {
		e.GetFlow().GetBot().Send(c.Sender, "Great! Just send me your location now")
		node, _ := e.SearchDown("location")
		return node
	} else if text == "no" {
		e.GetFlow().GetBot().Send(c.Sender, "Okay. You are all set now!")
		return nil
	}
	e.GetFlow().GetBot().Send(c.Sender, "I am a bot! Please, answer yes or no")
	return e
}

func stageLocation(e *chain.Node, c *tb.Message) *chain.Node {
	log.Println(c.Sender.Recipient(), "goes through", e.GetId())
	e.GetFlow().GetBot().Send(c.Sender, "You are all set now!")
	return nil // only return nil when it's over
}
