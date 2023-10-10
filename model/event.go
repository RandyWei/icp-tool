package model

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type EventName string

const (
	Event_PRRSER EventName = "parser"
	Event_SAVE   EventName = "save"
)

type EventStatus string

const (
	Event_PARSER_LOADING EventStatus = "loading"
	Event_PARSER_RESULT  EventStatus = "result"
	Event_PARSER_ERROR   EventStatus = "error"
	Event_SAVE_SUCCESS   EventStatus = "success"
	Event_SAVE_FAILED    EventStatus = "failed"
)

type EventData struct {
	Status EventStatus `json:"status"`
	Data   interface{} `json:"data"`
}

// 事件
type Event struct {
	Ctx  context.Context
	Name EventName
	Data EventData
}

func (event *Event) Send() {
	runtime.EventsEmit(event.Ctx, string(event.Name), event.Data)
}
