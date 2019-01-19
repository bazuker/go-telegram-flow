package chain

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

/*
	Node is an element in a double-linked list
*/
type Node struct {
	id       string
	flow     *Flow
	endpoint FlowCallback
	prev     *Node
	next     *Node
	event    string
}

/*
	Creates a following element in the list
*/
func (e *Node) Then(id string, endpoint FlowCallback, expectedEvent string) *Node {
	newNode := &Node{
		id:       id,
		flow:     e.flow,
		endpoint: endpoint,
		prev:     e,
		next:     nil,
		event:    expectedEvent,
	}
	e.next = newNode
	return newNode
}

/*
	Get related flow
*/
func (e *Node) GetFlow() *Flow {
	return e.flow
}

/*
	Get node's identificator
*/
func (e *Node) GetId() string {
	return e.id
}

/*
	Get node's callback endpoint
*/
func (e *Node) GetEndpoint() FlowCallback {
	return e.endpoint
}

/*
	Get the previous node in the list
*/
func (e *Node) Previous() *Node {
	return e.prev
}

/*
	Get the next node in the list
*/
func (e *Node) Next() *Node {
	return e.next
}

/*
	Tries to find a node with ID down the list
*/
func (e *Node) SearchDown(nodeId string) (*Node, bool) {
	temp := e
	for {
		temp = temp.next
		if temp == nil {
			break
		}
		if temp.id == nodeId {
			return temp, true
		}
	}
	return nil, false
}

/*
	Tries to find a node with ID up the list
*/
func (e *Node) SearchUp(nodeId string) (*Node, bool) {
	temp := e
	for {
		temp = temp.prev
		if temp == nil {
			break
		}
		if temp.id == nodeId {
			return temp, true
		}
	}
	return nil, false
}

/*
	Checks if the message type is matching the node type
*/
func (e *Node) CheckEvent(m *tb.Message) bool {
	switch e.event {
	case tb.OnText:
		if len(m.Text) < 1 {
			return false
		}
	case tb.OnPhoto:
		if m.Photo == nil {
			return false
		}
	case tb.OnLocation:
		if m.Location == nil {
			return false
		}
	case tb.OnContact:
		if m.Contact == nil {
			return false
		}
	case tb.OnAudio:
		if m.Audio == nil {
			return false
		}
	case tb.OnVideoNote:
		if m.VideoNote == nil {
			return false
		}
	case tb.OnVideo:
		if m.Video == nil {
			return false
		}
	case tb.OnVoice:
		if m.Voice == nil {
			return false
		}
	case tb.OnDocument:
		if m.Document == nil {
			return false
		}
	case tb.OnSticker:
		if m.Sticker == nil {
			return false
		}
	}
	return true
}
