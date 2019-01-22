package list

import (
	"github.com/pkg/errors"
	"github.com/tucnak/tr"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Callback func(list *List, path string, c *tb.Message)

var ErrInvalidTextPath = errors.New("invalid paths")

/*
	List is a set of prepared replies in a selected language
	that is able to perform a callback when a user selects an answer from the list
*/
type List struct {
	id       string
	engine   *tr.Engine
	bot      *tb.Bot
	markups  map[string]*tb.ReplyMarkup
	links    map[string]map[string]int
	paths    []string
	callback Callback
}

/*
	Creates a new list
*/
func NewListFlow(id string, textEngine *tr.Engine, bot *tb.Bot, callback Callback, textPaths ...string) (*List, error) {
	if textPaths == nil || len(textPaths) < 1 {
		return nil, ErrInvalidTextPath
	}
	return &List{
		id:       id,
		engine:   textEngine,
		bot:      bot,
		markups:  make(map[string]*tb.ReplyMarkup),
		links:    make(map[string]map[string]int),
		paths:    textPaths,
		callback: callback,
	}, nil
}

/*
	Builds markups for a specific locale
*/
func (l *List) Build(lang string) *List {
	buttons := make([][]tb.ReplyButton, len(l.paths))
	l.links[lang] = make(map[string]int)
	for i, p := range l.paths {
		text := l.engine.Lang(lang).Tr(p)
		btn := []tb.ReplyButton{
			{
				Text: text,
			},
		}
		l.links[lang][text] = i
		l.bot.Handle(&btn[0], l.handler)
		buttons[i] = btn
	}
	l.markups[lang] = &tb.ReplyMarkup{
		ReplyKeyboard: buttons,
		ForceReply:    true,
	}
	return l
}

/*
	Gets a built markup in a specified language
*/
func (l *List) GetMarkup(lang string) *tb.ReplyMarkup {
	return l.markups[lang]
}

/*
	A default handler that aggregates all the incoming responses for the list
*/
func (l *List) handler(m *tb.Message) {
	lang := l.engine.DefaultLocale.Name
	if userLang, ok := l.engine.Langs[m.Sender.LanguageCode]; ok {
		lang = userLang.Name
	}
	if link, ok := l.links[lang][m.Text]; ok {
		l.callback(l, l.paths[link], m)
	}
}
