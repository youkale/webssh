package tui

import "net/url"

type PtyMessageType string

const PtyMessageTypeRequest PtyMessageType = "requests"
const PtyMessageTypeInfo PtyMessageType = "info"

type RequestMessage struct {
	Method string
	Path   url.URL
}

type PtyMessage struct {
	Type    PtyMessageType
	Payload interface{}
}
