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
	Callback function description that includes a "node" that a call was made from
 */
type FlowCallback func (e *Node, c *tb.Callback) bool

/*
	A flow is essentially a high-level representation of a menu
 */
type Flow struct {
	FlowId  string
	Serial  uint32
	Root    *Node
	Bot     *tb.Bot
	dialogs map[int]*Dialog
	defaultLocale string
	mx sync.RWMutex
}

/*
	A dialog is an abstract piece that holds a menu message sent by the bot
	and a language that the interface is displayed
 */
type Dialog struct {
	Message *tb.Message
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
	f :=  &Flow{
		FlowId:  flowId,
		Serial:  0,
		Bot:     bot,
		dialogs: make(map[int]*Dialog),
		defaultLocale: defaultLocale,
		mx: sync.RWMutex{},
	}
	atomic.StoreUint32(&f.Serial, 0)
	f.Root = &Node{Id: "0", Flow: f, MustUpdate: false, Markup: make(map[string]*tb.ReplyMarkup)}
	return f, nil
}

/*
	Retrieves a dialog by user id
 */
func (f *Flow) GetDialog(id int) (*Dialog, bool) {
	f.mx.RLock()
	d, ok := f.dialogs[id]
	f.mx.RUnlock()
	return d, ok
}

/*
	Sets a dialog by a user id
	Only internal use is intended
 */
func (f *Flow) setDialog(id int, dialog *Dialog) {
	f.mx.Lock()
	f.dialogs[id] = dialog
	f.mx.Unlock()
}

/*
	Creates a new node in the flow
 */
func (f *Flow) New(text string, endpoint FlowCallback) *Node {
	return newNode(f, text, endpoint, f.Root)
}

/*
	Builds the flow for a specified locale
 */
func (f *Flow) Build(lang string) *Flow {
	f.Root.build(f.FlowId, lang)
	return f
}

/*
	Sends a new instance of a menu to the user with a specified locale
 */
func (f *Flow) Display(to *tb.User, text, lang string) error {
	msg, err := f.Bot.Send(to, text, f.Root.Markup[lang])
	if err != nil {
		return err
	}
	f.setDialog(to.ID, &Dialog{Message: msg, Language: lang})
	return nil
}
