package list

import (
	"github.com/pkg/errors"
	"github.com/tucnak/tr"
	tb "gopkg.in/tucnak/telebot.v2"
	"sync"
)

/*
	Callback that triggers when a user sends an item from the list
	if the function returns true - the user is being evicted from the active sessions
*/
type Callback func(list *List, path string, c *tb.Message) bool

var (
	ErrInvalidTextPath = errors.New("invalid paths")
	ErrInvalidLanguage = errors.New("locale does not exist")
)

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
	sessions map[string]string
	paths    []string
	callback Callback
	mx       sync.RWMutex
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
		sessions: make(map[string]string), // user id -> language
		paths:    textPaths,
		callback: callback,
		mx:       sync.RWMutex{},
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
	Get list's unique identificator
*/
func (l *List) GetId() string {
	return l.id
}

/*
	Get attached Telegram bot
*/
func (l *List) GetBot() *tb.Bot {
	return l.bot
}

/*
	Starts a list flow for the user
*/
func (l *List) Start(to tb.Recipient, textPath, language string) error {
	if _, ok := l.engine.Langs[language]; !ok {
		return ErrInvalidLanguage
	}
	l.setSession(to, language)
	_, err := l.bot.Send(to, l.engine.Lang(language).Tr(textPath), l.GetMarkup(language))
	return err
}

/*
	Retrieves a session language by recipient
*/
func (l *List) GetSession(of tb.Recipient) (string, bool) {
	l.mx.RLock()
	d, ok := l.sessions[of.Recipient()]
	l.mx.RUnlock()
	return d, ok
}

/*
	Sets a session language for a recipient
	Only internal use is intended
*/
func (l *List) setSession(of tb.Recipient, language string) {
	l.mx.Lock()
	l.sessions[of.Recipient()] = language
	l.mx.Unlock()
}

/*
	Sets a session language for a recipient
	Only internal use is intended
*/
func (l *List) deleteSession(of tb.Recipient) {
	l.mx.Lock()
	delete(l.sessions, of.Recipient())
	l.mx.Unlock()
}

/*
	A default handler that aggregates all the incoming responses for the list
*/
func (l *List) handler(m *tb.Message) {
	if lang, ok := l.GetSession(m.Sender); ok {
		if link, ok := l.links[lang][m.Text]; ok {
			if l.callback(l, l.paths[link], m) {
				// delete the session if it was marked as completed
				l.deleteSession(m.Sender)
			}
		}
	}
}
