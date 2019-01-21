package framework

import (
	"errors"
	"github.com/tucnak/tr"
	"go-telegram-flow/chain"
	"go-telegram-flow/menu"
	tb "gopkg.in/tucnak/telebot.v2"
	"sync"
	"time"
)

var ErrChainDoesNotExist = errors.New("chain does not exist")

type UnknownChainMessage = func(c *chain.Chain, m *tb.Message)

type BotFramework struct {
	bot            *tb.Bot
	textEngine     *tr.Engine
	name           string
	path           string
	defaultLocale  string
	chains         map[string]*chain.Chain
	menus          map[string]*menu.Menu
	sessions       map[string]string
	mx             sync.RWMutex
	defaultHandler UnknownChainMessage
}

func NewFramework(name, token, path, defaultLocale string) (*BotFramework, error) {
	b, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, err
	}
	engine, err := tr.NewEngine(path, defaultLocale, false)
	if err != nil {
		return nil, err
	}
	fw := &BotFramework{
		bot:           b,
		textEngine:    engine,
		name:          name,
		path:          path,
		defaultLocale: defaultLocale,
		chains:        make(map[string]*chain.Chain),
		menus:         make(map[string]*menu.Menu),
		sessions:      make(map[string]string),
		mx:            sync.RWMutex{},
	}
	fw.bot.Handle(tb.OnText, fw.process)
	fw.bot.Handle(tb.OnPhoto, fw.process)
	fw.bot.Handle(tb.OnLocation, fw.process)
	fw.bot.Handle(tb.OnContact, fw.process)
	fw.bot.Handle(tb.OnVideo, fw.process)
	fw.bot.Handle(tb.OnVideoNote, fw.process)
	fw.bot.Handle(tb.OnVoice, fw.process)
	fw.bot.Handle(tb.OnDocument, fw.process)
	fw.bot.Handle(tb.OnSticker, fw.process)
	fw.bot.Handle(tb.OnAudio, fw.process)
	return fw, nil
}

func (fw *BotFramework) process(m *tb.Message) {
	fw.mx.RLock()
	chainId, ok := fw.sessions[m.Sender.Recipient()]
	if !ok {
		fw.mx.RUnlock()
		return
	}
	c, _ := fw.chains[chainId]
	fw.mx.RUnlock()
	if !c.Process(m) {
		if fw.defaultHandler != nil {
			fw.defaultHandler(c, m)
		}
	}
}

func (fw *BotFramework) GetBot() *tb.Bot {
	return fw.bot
}

func (fw *BotFramework) SetDefaultUnknownMessageHandler(handler UnknownChainMessage) *BotFramework {
	fw.defaultHandler = handler
	return fw
}

func (fw *BotFramework) AddChain(c *chain.Chain) *BotFramework {
	fw.mx.Lock()
	fw.chains[c.GetId()] = c
	fw.mx.Unlock()
	return fw
}

func (fw *BotFramework) AddMenu(m *menu.Menu) *BotFramework {
	fw.mx.Lock()
	fw.menus[m.GetId()] = m
	fw.mx.Unlock()
	return fw
}

func (fw *BotFramework) RunChain(to tb.Recipient, chainId, textPath string, options ...interface{}) error {
	fw.mx.RLock()
	c, ok := fw.chains[chainId]
	fw.mx.RUnlock()
	if !ok {
		return ErrChainDoesNotExist
	}
	return c.Start(to, fw.textEngine.Tr(textPath), options)
}

func (fw *BotFramework) RunMenu(to *tb.User, chainId, textPath string) error {
	fw.mx.RLock()
	m, ok := fw.menus[chainId]
	fw.mx.RUnlock()
	if !ok {
		return ErrChainDoesNotExist
	}
	lang := fw.defaultLocale
	if _, ok := fw.textEngine.Langs[to.LanguageCode]; ok {
		lang = to.LanguageCode
	}
	return m.Start(to, fw.textEngine.Lang(lang).Tr(textPath), lang)
}

func (fw *BotFramework) RunMenuInLang(to tb.Recipient, chainId, lang, textPath string) error {
	fw.mx.RLock()
	m, ok := fw.menus[chainId]
	fw.mx.RUnlock()
	if !ok {
		return ErrChainDoesNotExist
	}
	return m.Start(to, fw.textEngine.Lang(lang).Tr(textPath), lang)
}
