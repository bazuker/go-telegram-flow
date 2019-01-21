package menu

/*
	Menu flow is a paginated menu of inline buttons for Telegram
	Author: Daniil Furmanov
	License: MIT
*/

import (
	"github.com/tucnak/tr"
	tb "gopkg.in/tucnak/telebot.v2"
	"sync"
	"sync/atomic"
)

/*
	A flow is essentially a high-level representation of a menu
*/
type Menu struct {
	flowId        string
	serial        uint32
	root          *Node
	bot           *tb.Bot
	dialogs       map[string]*Dialog
	defaultLocale string
	engine        *tr.Engine
	mx            sync.RWMutex
}

/*
	A dialog is an abstract piece that holds a menu message sent by the bot
	and a language that the interface is displayed
*/
type Dialog struct {
	Message  *tb.Message
	Language string
}

/*
	Creates a new flow and initializes the specified locale directory
	Warning! When setting a flowId treat it gently, like picking a directory name, same rules applies.
	It will fail without a notice if you put special characters or symbols (except for underscore) in it.
	Suggested names: flow1, flow_1, MyFlow
*/
func NewMenuFlow(id string, bot *tb.Bot, langDir, defaultLocale string) (*Menu, error) {
	engine, err := tr.NewEngine(langDir, defaultLocale, false)
	if err != nil {
		return nil, err
	}
	f := &Menu{
		flowId:        id,
		serial:        0,
		bot:           bot,
		dialogs:       make(map[string]*Dialog),
		defaultLocale: defaultLocale,
		engine:        engine,
		mx:            sync.RWMutex{},
	}
	atomic.StoreUint32(&f.serial, 0)
	f.root = &Node{id: "0", flow: f, mustUpdate: false, markup: make(map[string]*tb.ReplyMarkup)}
	return f, nil
}

/*
	Get flow's unique identificator
*/
func (f *Menu) GetId() string {
	return f.flowId
}

/*
	Count all nodes in the tree
*/
func (f *Menu) CountNodes() int {
	return int(atomic.LoadUint32(&f.serial))
}

/*
	Get attached Telegram bot
*/
func (f *Menu) GetBot() *tb.Bot {
	return f.bot
}

/*
	Get the root node
*/
func (f *Menu) GetRoot() *Node {
	return f.root
}

/*
	Retrieves a dialog by user id
*/
func (f *Menu) GetDialog(id string) (*Dialog, bool) {
	f.mx.RLock()
	d, ok := f.dialogs[id]
	f.mx.RUnlock()
	return d, ok
}

/*
	Sets a dialog by a user id
	Only internal use is intended
*/
func (f *Menu) setDialog(id string, dialog *Dialog) {
	f.mx.Lock()
	f.dialogs[id] = dialog
	f.mx.Unlock()
}

/*
	Helper handler for forward buttons
*/
func (f *Menu) HandleForward(e *Node, c *tb.Callback) bool {
	return true
}

/*
	Helper handler for back buttons
*/
func (f *Menu) HandleBack(e *Node, c *tb.Callback) bool {
	return false
}

/*
	Creates a new node in the flow
*/
func (f *Menu) NewNode(text string, endpoint NodeEndpoint) *Node {
	return newNode(f, text, endpoint, f.root)
}

/*
	Creates a new back button node in the flow
	that automatically takes a user one page back
*/
func (f *Menu) NewBackNode(text string) *Node {
	return newNode(f, text, f.HandleBack, f.root)
}

/*
	Builds the flow for a specified locale
*/
func (f *Menu) Build(lang string) *Menu {
	f.root.build(f.flowId, lang)
	return f
}

/*
	Sends a new instance of a menu to the user with a specified locale
	Tries to delete the old menu before sending a new one
*/
func (f *Menu) Start(to tb.Recipient, text, lang string) error {
	if d, ok := f.GetDialog(to.Recipient()); ok {
		f.bot.Delete(d.Message)
	}
	msg, err := f.bot.Send(to, text, f.root.markup[lang])
	if err != nil {
		return err
	}
	f.setDialog(to.Recipient(), &Dialog{Message: msg, Language: lang})
	return nil
}
