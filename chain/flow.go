package chain

/*
	Chain flow is list of event listeners organized by type for Telegram
	Author: Daniil Furmanov
	License: MIT
*/

import (
	"github.com/pkg/errors"
	tb "gopkg.in/tucnak/telebot.v2"
	"sync"
)

/*
	A flow is chain or double-linked list of events organized by type
*/
type Chain struct {
	flowId         string
	root           *Node
	bot            *tb.Bot
	defaultLocale  string
	positions      map[string]*Node
	defaultHandler NodeEndpoint
	mx             sync.RWMutex
}

var ErrChainIsEmpty = errors.New("chain has zero handlers")

/*
	Creates a new chain flow
*/
func NewChainFlow(flowId string, bot *tb.Bot) (*Chain, error) {
	f := &Chain{
		bot:            bot,
		positions:      make(map[string]*Node),
		defaultHandler: nil,
		mx:             sync.RWMutex{},
	}
	f.root = &Node{id: flowId, flow: f, endpoint: nil, prev: nil, next: nil}
	return f, nil
}

/*
	Get flow's unique identificator
*/
func (f *Chain) GetFlowId() string {
	return f.flowId
}

/*
	Get attached Telegram bot
*/
func (f *Chain) GetBot() *tb.Bot {
	return f.bot
}

/*
	Get the root node
*/
func (f *Chain) GetRoot() *Node {
	return f.root
}

/*
	Gets the user position in the flow
*/
func (f *Chain) GetPosition(of tb.Recipient) (*Node, bool) {
	f.mx.RLock()
	node, ok := f.positions[of.Recipient()]
	f.mx.RUnlock()
	return node, ok
}

/*
	Sets the user current position in the flow
*/
func (f *Chain) SetPosition(of tb.Recipient, node *Node) {
	f.mx.Lock()
	f.positions[of.Recipient()] = node
	f.mx.Unlock()
}

/*
	Deletes the user current position in the flow
*/
func (f *Chain) DeletePosition(of tb.Recipient) {
	f.mx.Lock()
	delete(f.positions, of.Recipient())
	f.mx.Unlock()
}

/*
	Search for a node with ID
*/
func (f *Chain) Search(nodeId string) (*Node, bool) {
	return f.root.SearchDown(nodeId)
}

/*
	Get the root node
*/
func (f *Chain) DefaultHandler(endpoint NodeEndpoint) *Chain {
	f.defaultHandler = endpoint
	return f
}

func (f *Chain) Start(to tb.Recipient, text string, options ...interface{}) (err error) {
	if f.root.next == nil {
		return ErrChainIsEmpty
	}
	if len(options) > 0 {
		// a workaround for nil options
		// otherwise the message will not be sent
		_, err = f.GetBot().Send(to, text, options)
	} else {
		_, err = f.GetBot().Send(to, text)
	}
	if err == nil {
		f.mx.Lock()
		f.positions[to.Recipient()] = f.root.next
		f.mx.Unlock()
	}
	return
}

/*
	Process with the next flow iteration
	Returns true only if the iteration was successful
*/
func (f *Chain) Process(m *tb.Message) bool {
	if m == nil {
		return false
	}
	sender := m.Sender
	node, ok := f.GetPosition(sender)
	if !ok {
		// the flow hasn't started for the user
		return false
	}
	if node == nil {
		f.DeletePosition(sender)
		return false
	}
	if !node.CheckEvent(m) || node.endpoint == nil {
		// input is invalid for the particular node
		if f.defaultHandler != nil {
			next := f.defaultHandler(node, m)
			if next != node {
				f.SetPosition(sender, next)
			}
			return true
		}
		return false
	}
	next := node.endpoint(node, m)
	if next != node {
		f.SetPosition(sender, next)
	}
	return true
}
