package menu

import (
	"fmt"
	"github.com/tucnak/tr"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"strconv"
	"sync/atomic"
)

/*
	Callback function declaration that includes a "node" that a call was made from
*/
type NodeEndpoint func(e *Node, c *tb.Callback) bool

/*
	An element of a menu that holds all required information by a page
	a.k.a a button that holds other buttons for the next page
*/
type Node struct {
	id         string
	flow       *Menu
	path       string
	text       string
	endpoint   NodeEndpoint
	markup     map[string]*tb.ReplyMarkup
	prev       *Node
	nodes      []*Node
	mustUpdate bool
}

/*
	Creates a new node in the flow
*/
func newNode(root *Menu, text string, endpoint NodeEndpoint, prev *Node) *Node {
	id := atomic.AddUint32(&root.serial, 1)
	return &Node{
		id:         strconv.Itoa(int(id)),
		flow:       root,
		text:       text,
		path:       text,
		endpoint:   endpoint,
		prev:       prev,
		markup:     make(map[string]*tb.ReplyMarkup),
		mustUpdate: false,
	}
}

/*
	Get related flow
*/
func (e *Node) GetFlow() *Menu {
	return e.flow
}

/*
	Get node's identificator
*/
func (e *Node) GetId() string {
	return e.id
}

/*
	Get node's default text
*/
func (e *Node) GetText() string {
	return e.text
}

/*
	Get node's locale path
*/
func (e *Node) GetPath() string {
	return e.path
}

/*
	Get node's callback endpoint
*/
func (e *Node) GetEndpoint() NodeEndpoint {
	return e.endpoint
}

/*
	Get previous (parent) node in the tree
*/
func (e *Node) Previous() *Node {
	return e.prev
}

/*
	Get all children nodes
*/
func (e *Node) GetNodes() []*Node {
	return e.nodes
}

/*
	Get a markup in a specified language
	Caution! Menu must be built for the specified language beforehand
*/
func (e *Node) GetMarkup(lang string) *tb.ReplyMarkup {
	return e.markup[lang]
}

/*
	Adds a new node to the current node
	Returns the current node
*/
func (e *Node) Add(text string, endpoint NodeEndpoint) *Node {
	e.AddSub(text, endpoint)
	return e
}

/*
	Adds a new node with many sub-nodes to the current node
	Returns the current node
*/
func (e *Node) AddWith(text string, endpoint NodeEndpoint, elements ...*Node) *Node {
	newElement := e.AddSub(text, endpoint)
	newElement.AddManySub(elements)
	return e
}

/*
	Adds a new sub node
	Returns the new node
*/
func (e *Node) AddSub(text string, endpoint NodeEndpoint) *Node {
	newElement := newNode(e.flow, text, endpoint, e)
	if e.nodes == nil {
		e.nodes = make([]*Node, 1)
		e.nodes[0] = newElement
		return newElement
	}
	e.nodes = append(e.nodes, newElement)
	return newElement
}

/*
	Adds many new sub nodes
	Returns the current node
*/
func (e *Node) AddManySub(elements []*Node) *Node {
	if e.nodes == nil {
		e.nodes = make([]*Node, len(elements))
		for i, el := range elements {
			el.prev = e
			e.nodes[i] = el
		}
		return e
	}
	e.nodes = append(e.nodes, elements...)
	return e
}

/*
	Sets a new caption for the flow
	that will be updated in the next menu iteration
	params are automatically placed in the text if provided
*/
func (e *Node) SetCaption(c *tb.Callback, text string, params ...interface{}) *Node {
	if d, ok := e.flow.GetDialog(c.Sender.Recipient()); ok {
		if len(params) > 0 {
			text = fmt.Sprintf(text, params...)
		}
		if d.Message.Text != text {
			d.Message.Text = text
			e.mustUpdate = true
		}
	}
	return e
}

/*
	Gets a language currently used in a dialog by the user
*/
func (e *Node) GetLanguage(c *tb.Callback) string {
	if d, ok := e.flow.GetDialog(c.Sender.Recipient()); ok {
		return d.Language
	}
	return e.flow.defaultLocale
}

/*
	Sets a language for the user's dialog
*/
func (e *Node) SetLanguage(c *tb.Callback, lang string) *Node {
	if d, ok := e.flow.GetDialog(c.Sender.Recipient()); ok {
		d.Language = lang
		e.mustUpdate = true
		e.next(c)
	}
	return e
}

/*
	Updates the menu
*/
func (e *Node) update(c *tb.Callback, d *Dialog, markup *tb.ReplyMarkup) {
	newMsg, err := e.flow.bot.Edit(d.Message, d.Message.Text, markup)
	if err != nil {
		log.Println("failed to continue", c.Sender.ID, err)
		return
	}
	e.mustUpdate = false
	d.Message = newMsg
}

/*
	Goes back to the previous menu
*/
func (e *Node) back(c *tb.Callback) *Node {
	d, ok := e.flow.GetDialog(c.Sender.Recipient())
	if !ok {
		log.Println(c.Sender.ID, "does not exist")
		return nil
	}
	if e.prev == nil || e.prev.prev == nil {
		if e.mustUpdate {
			e.update(c, d, e.flow.root.markup[d.Language])
			return e
		}
		return nil
	}
	newMsg, err := e.flow.bot.Edit(d.Message, d.Message.Text, e.prev.prev.markup[d.Language])
	if err != nil {
		log.Println("failed to back", c.Sender.ID, err)
		return nil
	}
	d.Message = newMsg
	return e.prev
}

/*
	Continues to the following and/or updates the menu
*/
func (e *Node) next(c *tb.Callback) {
	nodes := len(e.nodes)
	if nodes < 1 && !e.mustUpdate {
		return
	}
	d, ok := e.flow.GetDialog(c.Sender.Recipient())
	if !ok {
		log.Println(c.Sender.ID, "does not exist")
		return
	}
	markup := e.markup
	if nodes < 1 {
		markup = e.prev.markup
	}
	e.update(c, d, markup[d.Language])
}

/*
	Builds the flow and creates markup for a tree of nodes in a specified locale
*/
func (e *Node) build(basePath, lang string) {
	if e.prev != nil {
		e.path = basePath + "/" + e.text
	} else {
		e.path = basePath
	}
	buttons := make([][]tb.InlineButton, len(e.nodes))
	for i, child := range e.nodes {
		child.build(e.path, lang)
		buttons[i] = []tb.InlineButton{
			{
				Unique: "flow" + lang + e.flow.flowId + child.id,
				Text:   tr.Lang(lang).Tr(child.path),
			},
		}
		if child.endpoint != nil {
			e.flow.bot.Handle(&buttons[i][0], child.handle)
		} else {
			e.flow.bot.Handle(&buttons[i][0], child.handleDeadEnd)
		}
	}
	e.markup[lang] = &tb.ReplyMarkup{
		InlineKeyboard: buttons,
	}
}

/*
	Default handler for pagination
*/
func (e *Node) handle(c *tb.Callback) {
	err := e.flow.bot.Respond(c)
	if err != nil {
		log.Println("failed to respond", c.Sender.ID, err)
		return
	}
	if e.endpoint(e, c) {
		e.next(c)
	} else {
		e.back(c)
	}
}

/*
	Handler for menu buttons with no provided endpoint (callback)
*/
func (e *Node) handleDeadEnd(c *tb.Callback) {
	err := e.flow.bot.Respond(c)
	if err != nil {
		log.Println("failed to respond", c.Sender.ID, err)
		return
	}
	e.next(c)
}
