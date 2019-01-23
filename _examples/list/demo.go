package list

import (
	"github.com/tucnak/tr"
	"go-telegram-flow/list"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"time"
)

func Run(token string) {
	b, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		panic(err)
	}

	if err := tr.Init("_examples/list/lang", "en"); err != nil {
		panic(err)
	}

	list, err := list.NewListFlow("list1", tr.DefaultEngine, b, handle, "blue", "green", "red")
	if err != nil {
		panic(err)
	}

	list.Build("en")

	b.Handle("/start", func(m *tb.Message) {
		if err := list.Start(m.Sender, "hello", "en"); err != nil {
			log.Println(err)
		}
	})

	log.Println("starting...", b.Me.Username)

	b.Start()
}

func handle(list *list.List, path string, m *tb.Message) bool {
	log.Println(path)
	list.GetBot().Send(m.Sender, "You picked "+path)
	return false // do not quit
}
