package menu

import (
	"github.com/tucnak/tr"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"strconv"
	"sync/atomic"
)

/*
	An element of a menu that holds all required information by a page
	a.k.a a button that holds other buttons for the next page
 */
type Node struct {
	Id         string
	Flow       *Flow
	Caption    *string
	Path       string
	Text       string
	Endpoint   FlowCallback
	Markup     map[string]*tb.ReplyMarkup
	Prev       *Node
	Nodes      []*Node
	MustUpdate bool
}

/*
	Creates a new node in the flow
 */
func newNode(root *Flow, text string, endpoint FlowCallback, prev *Node) *Node {
	id := atomic.AddUint32(&root.Serial, 1)
	return &Node{
		Id: strconv.Itoa(int(id)),
		Flow: root,
		Text: text,
		Path: text,
		Endpoint: endpoint,
		Prev: prev,
		Markup: make(map[string]*tb.ReplyMarkup),
		MustUpdate: false,
	}
}

/*
	Adds a new node to the current node
	Returns the current node
 */
func (e *Node) Add(text string, endpoint FlowCallback) *Node {
	e.AddSub(text, endpoint)
	return e
}

/*
	Adds a new node with many sub-nodes to the current node
	Returns the current node
 */
func (e *Node) AddWith(text string, endpoint FlowCallback, elements ...*Node) *Node {
	newElement := e.AddSub(text, endpoint)
	newElement.AddManySub(elements)
	return e
}

/*
	Adds a new sub node
	Returns the new node
 */
func (e *Node) AddSub(text string, endpoint FlowCallback) *Node {
	newElement := newNode(e.Flow, text, endpoint, e)
	if e.Nodes == nil {
		e.Nodes = make([]*Node, 1)
		e.Nodes[0] = newElement
		return newElement
	}
	e.Nodes = append(e.Nodes, newElement)
	return newElement
}

/*
	Adds many new sub nodes
	Returns the current node
 */
func (e *Node) AddManySub(elements []*Node) *Node {
	if e.Nodes == nil {
		e.Nodes = make([]*Node, len(elements))
		for i, el := range elements {
			el.Prev = e
			e.Nodes[i] = el
		}
		return e
	}
	e.Nodes = append(e.Nodes, elements...)
	return e
}

/*
	Sets a new caption for the flow
	that will be updated after .next was called
 */
func (e *Node) SetCaption(c *tb.Callback, text string) *Node {
	if d, ok := e.Flow.GetDialog(c.Sender.ID); ok {
		if d.Message.Text != text {
			d.Message.Text = text
			e.MustUpdate = true
		}
	}
	return e
}

/*
	Gets a language currently used in a dialog by the user
*/
func (e *Node) GetLanguage(c *tb.Callback) string {
	if d, ok := e.Flow.GetDialog(c.Sender.ID); ok {
		return d.Language
	}
	return e.Flow.defaultLocale
}

/*
	Sets a language for the user's dialog
*/
func (e *Node) SetLanguage(c *tb.Callback, lang string) *Node {
	if d, ok := e.Flow.GetDialog(c.Sender.ID); ok {
		d.Language = lang
		e.MustUpdate = true
		e.next(c)
	}
	return e
}

/*
	Goes back to the previous menu
 */
func (e *Node) back(c *tb.Callback) *Node {
	if e.Prev == nil || e.Prev.Prev == nil {
		return nil
	}
	d, ok := e.Flow.GetDialog(c.Sender.ID)
	if !ok {
		log.Println(c.Sender.ID, "does not exist")
		return nil
	}
	newMsg, err := e.Flow.Bot.Edit(d.Message, d.Message.Text, e.Prev.Prev.Markup[d.Language])
	if err != nil {
		log.Println("failed to back", c.Sender.ID, err)
		return nil
	}
	d.Message = newMsg
	return e.Prev
}

/*
	Continues to the following and/or updates the menu
 */
func (e *Node) next(c *tb.Callback) {
	nodes := len(e.Nodes)
	if nodes < 1 && !e.MustUpdate {
		return
	}
	d, ok := e.Flow.GetDialog(c.Sender.ID)
	if !ok {
		log.Println(c.Sender.ID, "does not exist")
		return
	}
	markup := e.Markup
	if nodes < 1 {
		markup = e.Prev.Markup
	}
	newMsg, err := e.Flow.Bot.Edit(d.Message, d.Message.Text, markup[d.Language])
	if err != nil {
		log.Println("failed to continue", c.Sender.ID, err)
		return
	}
	e.MustUpdate = false
	d.Message = newMsg
}

/*
	Builds the flow and creates markup for a tree of nodes in a specified locale
 */
func (e *Node) build(basePath, lang string) {
	if e.Prev != nil {
		e.Path = basePath + "/" + e.Text
	} else {
		e.Path = basePath
	}
	buttons := make([][]tb.InlineButton, len(e.Nodes))
	for i, child := range e.Nodes {
		child.build(e.Path, lang)
		buttons[i] = []tb.InlineButton{
			{
				Unique: "flow" + lang + e.Flow.FlowId + child.Id,
				Text:   tr.Lang(lang).Tr(child.Path),
			},
		}
		if child.Endpoint != nil {
			e.Flow.Bot.Handle(&buttons[i][0], child.handle)
		} else {
			e.Flow.Bot.Handle(&buttons[i][0], child.handleDeadEnd)
		}
	}
	e.Markup[lang] = &tb.ReplyMarkup{
		InlineKeyboard: buttons,
	}
}

/*
	Default handler for pagination
 */
func (e *Node) handle(c *tb.Callback) {
	log.Println("wtf")
	err := e.Flow.Bot.Respond(c)
	if err != nil {
		log.Println("failed to respond", c.Sender.ID, err)
		return
	}
	if e.Endpoint(e, c) {
		e.next(c)
	} else {
		e.back(c)
	}
}

/*
	Handler for menu buttons with no provided endpoint (callback)
*/
func (e *Node) handleDeadEnd(c *tb.Callback) {
	err := e.Flow.Bot.Respond(c)
	if err != nil {
		log.Println("failed to respond", c.Sender.ID, err)
		return
	}
	e.next(c)
}
