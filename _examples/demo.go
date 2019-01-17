package _examples

import (
	"go-telegram-flow/menu"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"strconv"
	"time"
)

var total = 0

func Run() {
	b, err := tb.NewBot(tb.Settings{
		Token:  "YOUR_BOT_TOKEN_HERE",
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		panic(err)
	}

	const defaultLocale = "en"

	flow, err := menu.NewFlow("flow1", b, "_examples/lang", defaultLocale)
	if err != nil {
		panic(err)
	}
	/*

				An example of a menu

	---------------------------------------------
		greetings -> say hello
	---------------------------------------------
						pizza -> margarita ...
								 <-back
		orders  -> 		sushi -> nigiri ...
								 <-back
					 	<-back
	---------------------------------------------
		invoice -> show the total
	---------------------------------------------
		language -> switch the language
	---------------------------------------------

	*/
	flow.Root.
		Add("greetings", userPressGreeting).
		AddWith("order", userPress,
			flow.New("pizza", userPress).
				Add("margarita", userOrderPizza).
				Add("pepperoni", userOrderPizza).Add("back", userPressBack),
			flow.New("sushi", userPress).
				Add("temaki", userOrderSushi).
				Add("nigiri", userOrderSushi).Add("sasazushi", userOrderSushi).
				Add("back", userPressBack),
			flow.New("back", userPressBack),
		).
		Add("invoice", userPressInvoice).
		Add("language", userPressLanguage).Flow.Build("en").Build("ru")

	b.Handle("/start", func(m *tb.Message) {
		err = flow.Display(m.Sender,  "Hello there", defaultLocale)
		if err != nil {
			log.Println("failed to display the menu", err)
		}
	})

	log.Println("starting...")

	b.Start()
}

func userPressGreeting(e *menu.Node, c *tb.Callback) bool {
	log.Println(c.Sender.Recipient(), "press", e.Text)
	if _, err := e.Flow.Bot.Send(c.Sender, "Hi there"); err != nil {
		log.Println(err)
	}
	return true // continue
}

func userOrderSushi(e *menu.Node, c *tb.Callback) bool {
	log.Println(c.Sender.Recipient(), "press", e.Text)
	e.SetCaption(c, "Added " + e.Text + " to your order")
	total += 5
	return true
}

func userOrderPizza(e *menu.Node, c *tb.Callback) bool {
	log.Println(c.Sender.Recipient(), "press", e.Text)
	e.SetCaption(c, "Added " + e.Text + " to your order")
	total += 10
	return true
}

func userPressInvoice(e *menu.Node, c *tb.Callback) bool {
	log.Println(c.Sender.Recipient(), "press", e.Text)
	e.SetCaption(c, "Your total is $" + strconv.Itoa(total))
	return true
}

func userPressLanguage(e *menu.Node, c *tb.Callback) bool {
	log.Println(c.Sender.Recipient(), "press", e.Text)
	if e.GetLanguage(c) == "en" {
		e.SetLanguage(c, "ru")
	} else {
		e.SetLanguage(c, "en")
	}
	return true // continue
}

func userPress(e *menu.Node, c *tb.Callback) bool {
	log.Println(c.Sender.Recipient(), "press", e.Text)
	return true
}

func userPressBack(e *menu.Node, c *tb.Callback) bool {
	log.Println(c.Sender.Recipient(), "press", e.Text)
	return false // go back
}