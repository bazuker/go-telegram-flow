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
type Flow struct {
	flowId        string
	serial        uint32
	root          *Node
	bot           *tb.Bot
	dialogs       map[string]*Dialog
	defaultLocale string
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
func NewFlow(flowId string, bot *tb.Bot, langDir, defaultLocale string) (*Flow, error) {
	err := tr.Init(langDir, defaultLocale)
	if err != nil {
		return nil, err
	}
	f := &Flow{
		flowId:        flowId,
		serial:        0,
		bot:           bot,
		dialogs:       make(map[string]*Dialog),
		defaultLocale: defaultLocale,
		mx:            sync.RWMutex{},
	}
	atomic.StoreUint32(&f.serial, 0)
	f.root = &Node{id: "0", flow: f, mustUpdate: false, markup: make(map[string]*tb.ReplyMarkup)}
	return f, nil
}

/*
	Get flow's unique identificator
*/
func (f *Flow) GetFlowId() string {
	return f.flowId
}

/*
	Count all nodes in the tree
*/
func (f *Flow) CountNodes() int {
	return int(atomic.LoadUint32(&f.serial))
}

/*
	Get attached Telegram bot
*/
func (f *Flow) GetBot() *tb.Bot {
	return f.bot
}

/*
	Get the root node
*/
func (f *Flow) GetRoot() *Node {
	return f.root
}

/*
	Retrieves a dialog by user id
*/
func (f *Flow) GetDialog(id string) (*Dialog, bool) {
	f.mx.RLock()
	d, ok := f.dialogs[id]
	f.mx.RUnlock()
	return d, ok
}

/*
	Sets a dialog by a user id
	Only internal use is intended
*/
func (f *Flow) setDialog(id string, dialog *Dialog) {
	f.mx.Lock()
	f.dialogs[id] = dialog
	f.mx.Unlock()
}

/*
	Helper handler for forward buttons
*/
func (f *Flow) HandleForward(e *Node, c *tb.Callback) bool {
	return true
}

/*
	Helper handler for back buttons
*/
func (f *Flow) HandleBack(e *Node, c *tb.Callback) bool {
	return false
}

/*
	Creates a new node in the flow
*/
func (f *Flow) NewNode(text string, endpoint NodeEndpoint) *Node {
	return newNode(f, text, endpoint, f.root)
}

/*
	Creates a new back button node in the flow
	that automatically takes a user one page back
*/
func (f *Flow) NewBackNode(text string) *Node {
	return newNode(f, text, f.HandleBack, f.root)
}

/*
	Builds the flow for a specified locale
*/
func (f *Flow) Build(lang string) *Flow {
	f.root.build(f.flowId, lang)
	return f
}

/*
	Sends a new instance of a menu to the user with a specified locale
	Tries to delete the old menu before sending a new one
*/
func (f *Flow) Display(to tb.Recipient, text, lang string) error {
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
