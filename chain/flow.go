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

type FlowCallback func(e *Node, c *tb.Message) bool

/*
	A flow is chain or double-linked list of events organized by type
*/
type Flow struct {
	flowId         string
	root           *Node
	bot            *tb.Bot
	defaultLocale  string
	status         map[string]*Node
	defaultHandler FlowCallback
	mx             sync.RWMutex
}

var ErrChainIsEmpty = errors.New("chain has zero handlers")

/*
	Creates a new chain flow
*/
func NewFlow(flowId string, bot *tb.Bot) (*Flow, error) {
	f := &Flow{
		bot:            bot,
		status:         make(map[string]*Node),
		defaultHandler: nil,
		mx:             sync.RWMutex{},
	}
	f.root = &Node{id: flowId, flow: f, endpoint: nil, prev: nil, next: nil}
	return f, nil
}

/*
	Get flow's unique identificator
*/
func (f *Flow) GetFlowId() string {
	return f.flowId
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
	Get the root node
*/
func (f *Flow) DefaultHandler(endpoint FlowCallback) *Flow {
	f.defaultHandler = endpoint
	return f
}

func (f *Flow) Start(to tb.Recipient, text string, options ...interface{}) (err error) {
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
		f.status[to.Recipient()] = f.root.next
		f.mx.Unlock()
	}
	return
}

/*
	Process with the next flow iteration
	Returns true only if a user can be taken to the next node
*/
func (f *Flow) Process(m *tb.Message) bool {
	if m == nil {
		return false
	}

	recipient := m.Sender.Recipient()

	f.mx.RLock()
	node, ok := f.status[recipient]
	f.mx.RUnlock()
	if !ok {
		// the flow hasn't started for the user
		return false
	}
	if !node.CheckEvent(m) || node.endpoint == nil {
		// input is invalid for the particular node
		if f.defaultHandler != nil {
			f.defaultHandler(node, m)
		}
		return false
	}
	ok = node.endpoint(node, m)
	if ok {
		if node.next == nil {
			delete(f.status, recipient)
			return true
		}
		f.mx.Lock()
		f.status[recipient] = node.next
		f.mx.Unlock()
		return true
	}
	return false
}
