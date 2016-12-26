package sms

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/kkserver/kk-lib/kk/app"
	"github.com/kkserver/kk-lib/kk/json"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type SMSAliService struct {
	app.Service
	Init *app.InitTask
	Send *SMSSendTask

	BaseURL         string
	AccessKeyId     string
	AccessKeySecret string
	Sign            string

	client *http.Client
}

func (S *SMSAliService) Handle(a app.IApp, task app.ITask) error {
	return app.ServiceReflectHandle(a, task, S)
}

func encodeURL(u string) string {
	s := url.QueryEscape(u)
	s = strings.Replace(s, "+", "%20", 0)
	s = strings.Replace(s, "*", "%2A", 0)
	s = strings.Replace(s, "%7E", "~", 0)
	return s
}

func (S *SMSAliService) HandleInitTask(a *SMSApp, task *app.InitTask) error {

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(pemCerts)
	S.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: pool},
		},
	}

	return nil
}

func (S *SMSAliService) HandleSMSSendTask(a *SMSApp, task *SMSSendTask) error {

	var data map[string]string = map[string]string{}

	data["Format"] = "JSON"
	data["Version"] = "2016-09-27"
	data["AccessKeyId"] = S.AccessKeyId
	data["SignatureMethod"] = "HMAC-SHA1"
	data["Timestamp"] = time.Now().UTC().Format(time.RFC3339)
	data["SignatureVersion"] = "1.0"
	data["SignatureNonce"] = fmt.Sprintf("%d%d", time.Now().UnixNano(), rand.Int63())
	data["Action"] = "SingleSendSms"
	data["SignName"] = S.Sign
	data["TemplateCode"] = task.Content
	data["RecNum"] = task.Phone
	b, _ := json.Encode(task.Options)
	data["ParamString"] = string(b)

	var keys []string = []string{}

	for key, _ := range data {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	sb := bytes.NewBuffer(nil)

	var idx = 0

	for _, key := range keys {
		if idx != 0 {
			sb.WriteString("&")
		}
		sb.WriteString(encodeURL(key))
		sb.WriteString("=")
		sb.WriteString(encodeURL(data[key]))
		idx = idx + 1
	}

	sign := fmt.Sprintf("POST&%s&%s", encodeURL("/"), encodeURL(sb.String()))

	m := hmac.New(sha1.New, []byte(fmt.Sprintf("%s&", S.AccessKeySecret)))
	m.Write([]byte(sign))

	sign = base64.StdEncoding.EncodeToString(m.Sum(nil))

	sb = bytes.NewBuffer(nil)

	sb.WriteString("Signature")
	sb.WriteString("=")
	sb.WriteString(encodeURL(sign))

	for _, key := range keys {
		sb.WriteString("&")
		sb.WriteString(key)
		sb.WriteString("=")
		sb.WriteString(encodeURL(data[key]))
	}

	log.Println(data)

	resp, err := S.client.Post(S.BaseURL, "application/x-www-form-urlencoded", sb)

	if err != nil {
		task.Result.Errno = ERROR_SMS
		task.Result.Errmsg = err.Error()
	} else if resp.StatusCode == 200 {
		var body = make([]byte, resp.ContentLength)
		_, _ = resp.Body.Read(body)
		defer resp.Body.Close()
		log.Println(string(body))
	} else {
		var body = make([]byte, resp.ContentLength)
		_, _ = resp.Body.Read(body)
		defer resp.Body.Close()
		log.Println(string(body))
		task.Result.Errno = ERROR_SMS
		task.Result.Errmsg = fmt.Sprintf("[%d] %s", resp.StatusCode, resp.Status)
	}

	return nil
}
