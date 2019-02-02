package menu

/*
	Menu flow is a paginated menu of inline buttons for Telegram
	Author: Daniil Furmanov
	License: MIT
*/

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/tucnak/tr"
	tb "gopkg.in/tucnak/telebot.v2"
	"sync"
	"sync/atomic"
)

/*
	A flow is essentially a high-level representation of a menu
*/
type Menu struct {
	id            string
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
	Position *Node
}

/*
	Creates a new flow and initializes the specified locale directory
	Warning! When setting a id treat it gently, like picking a directory name, same rules applies.
	It will fail without a notice if you put special characters or symbols (except for underscore) in it.
	Suggested names: flow1, flow_1, MyFlow
*/
func NewMenuFlow(id string, bot *tb.Bot, engine *tr.Engine) (*Menu, error) {
	f := &Menu{
		id:      id,
		serial:  0,
		bot:     bot,
		dialogs: make(map[string]*Dialog),
		engine:  engine,
		mx:      sync.RWMutex{},
	}
	atomic.StoreUint32(&f.serial, 0)
	f.root = &Node{id: id + "_root", flow: f, mustUpdate: false, markups: make(map[string]*tb.ReplyMarkup)}
	return f, nil
}

/*
	Get menu's unique identificator
*/
func (f *Menu) GetId() string {
	return f.id
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
	Deletes a dialog by a user id
	Only internal use is intended
*/
func (f *Menu) deleteDialog(id string) {
	f.mx.Lock()
	delete(f.dialogs, id)
	f.mx.Unlock()
}

/*
	Sets a new caption for the menu
	The caption will be updated right away
	Params are automatically placed in the text if provided
*/
func (f *Menu) SetCaption(recipient tb.Recipient, text string, params ...interface{}) *Menu {
	if d, ok := f.GetDialog(recipient.Recipient()); ok {
		if len(params) > 0 {
			text = fmt.Sprintf(text, params...)
		}
		if d.Message.Text != text {
			d.Message.Text = text
			d.Position.update(recipient, d, d.Position.GetMarkup(d.Language))
		}
	}
	return f
}

/*
	Helper handler for forward buttons
*/
func (f *Menu) HandleForward(e *Node, c *tb.Callback) int {
	return Forward
}

/*
	Helper handler for back buttons
*/
func (f *Menu) HandleBack(e *Node, c *tb.Callback) int {
	return Back
}

/*
	Creates a new node in the flow
*/
func (f *Menu) NewNode(text string, endpoint Callback) *Node {
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
	f.root.build(f.id, lang)
	return f
}

/*
	Sends a new instance of a menu to a user with a specified locale
	Tries to delete the old menu before sending a new one
*/
func (f *Menu) Start(to tb.Recipient, text, lang string) error {
	if d, ok := f.GetDialog(to.Recipient()); ok {
		f.bot.Delete(d.Message)
	}
	msg, err := f.bot.Send(to, text, f.root.markups[lang], tb.Silent)
	if err != nil {
		return err
	}
	f.setDialog(to.Recipient(), &Dialog{Message: msg, Language: lang, Position: f.root})
	return nil
}

/*
	Sends an instance of a menu to a user starting at a specified node
	Tries to delete the old menu before sending a new one
*/
func (f *Menu) StartAt(to tb.Recipient, text, lang string, at *Node) error {
	d, ok := f.GetDialog(to.Recipient())
	if ok {
		f.bot.Delete(d.Message)
	} else {
		d = &Dialog{}
	}
	msg, err := f.bot.Send(to, text, at.markups[lang], tb.Silent)
	if err != nil {
		return err
	}
	d.Message = msg
	d.Language = lang
	d.Position = at
	f.setDialog(to.Recipient(), d)
	return nil
}

/*
	Takes a user to a specified menu position (page)
*/
func (f *Menu) MoveTo(to tb.Recipient, text, lang string, position *Node) error {
	d, ok := f.GetDialog(to.Recipient())
	if !ok {
		return errors.New("dialog not found")
	}
	msg, err := f.bot.Edit(d.Message, text, position.markups[lang], tb.Silent)
	if err != nil {
		return err
	}
	d.Message = msg
	d.Language = lang
	d.Position = position
	f.setDialog(to.Recipient(), d)
	return nil
}

/*
	Removes the menu from a user and deletes the session
*/
func (f *Menu) Stop(to tb.Recipient, text, lang string) error {
	if d, ok := f.GetDialog(to.Recipient()); ok {
		f.bot.Delete(d.Message)
	}
	f.deleteDialog(to.Recipient())
	return nil
}
