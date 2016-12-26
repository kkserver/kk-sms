package sms

import (
	"github.com/kkserver/kk-lib/kk/app"
	"github.com/kkserver/kk-lib/kk/app/remote"
)

type SMSApp struct {
	app.App
	Remote *remote.Service
	Ali    *SMSAliService
	Send   *SMSSendTask
}
