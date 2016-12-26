package sms

import (
	"github.com/kkserver/kk-lib/kk/app"
)

type SMSSendTaskResult struct {
	app.Result
}

type SMSSendTask struct {
	app.Task
	Phone   string                 `json:"phone"`
	Content string                 `json:"content"`
	Options map[string]interface{} `json:"options"`
	Result  SMSSendTaskResult
}

func (T *SMSSendTask) GetResult() interface{} {
	return &T.Result
}
