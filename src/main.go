package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

type Config struct {
	Email struct {
		IMAP     string `json:"imap"`
		SMTP     string `json:"smtp"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"email"`
	Target string `json:"target"`
}

var (
	log *slog.Logger
	cfg *Config
)

func init() {
	log = slog.New(NewSimpleHandler(os.Stdout, slog.LevelDebug))
	slog.SetDefault(log)

	log.Info("Loading config...")
	cfg = new(Config)
	e := cfg.Load()
	if e != nil {
		log.Error("Load failed, %s", e)
		os.Exit(1)
	}
}

func (c *Config) Load() error {
	f, e := os.Open("config.json")
	if e != nil {
		return e
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	return decoder.Decode(c)
}

// 连接到服务器并登录
func connect() *client.Client {
	log.Info("Connecting to server...")
	// 连接到服务器
	c, e := client.DialTLS(cfg.Email.IMAP, nil)
	if e != nil {
		log.Error("Dial failed", "error", e)
		return nil
	}
	log.Info("Connected")

	// 登录
	if e := c.Login(cfg.Email.Username, cfg.Email.Password); e != nil {
		log.Error("Login failed", "error", e)
		return nil
	}
	log.Info("Logged in")

	return c
}

// 列出所有文件夹
func boxes(c *client.Client) []*imap.MailboxInfo {
	boxes := make(chan *imap.MailboxInfo, 10)
	if e := c.List("", "*", boxes); e != nil {
		log.Error("List boxes failed", "error", e)
		return make([]*imap.MailboxInfo, 0)
	}
	log.Info("Mailboxes:")
	bs := make([]*imap.MailboxInfo, 0)
	for box := range boxes {
		log.Info("  * " + box.Name)
		bs = append(bs, box)
	}
	return bs
}

// 获取邮件
func fetch(c *client.Client, boxName string) {
	log.Info("Fetch", "mailbox", boxName)
	// 选择收件箱
	mbox, e := c.Select(boxName, true)
	if e != nil {
		log.Error("Select box failed", "error", e)
		return
	}
	from := 1
	to := mbox.Messages
	log.Debug("Has messages", "count", to)

	dir := filepath.Join(cfg.Target, boxName)
	names, e := ListNames(dir)
	if e != nil {
		log.Error("List names failed", "error", e)
		return
	}
	if len(names) > 0 {
		from = NameToInt(names[len(names)-1])
	}

	for i := uint32(from); i <= to; i++ {
		log.Debug("Deal with message", "index", i)

		seqset := new(imap.SeqSet)
		seqset.AddNum(i)

		// 获取邮件的全文
		section := &imap.BodySectionName{}
		items := []imap.FetchItem{imap.FetchInternalDate, imap.FetchRFC822Size, imap.FetchEnvelope, section.FetchItem()}
		messages := make(chan *imap.Message, 1)
		e = c.Fetch(seqset, items, messages)
		if e != nil {
			log.Error("Fetch message failed", "error", e)
			continue
		}

		msg := <-messages

		subject := msg.Envelope.Subject
		datetime := msg.InternalDate
		log.Debug("  ", "Subject", subject)
		log.Debug("  ", "Datetime", datetime)
		log.Debug("  ", "Size", msg.Size)

		r := msg.GetBody(section)
		if r == nil {
			log.Error("  Server didn't returned message body")
			continue
		}
		log.Debug("  Body", "length", r.Len())

		out, e := os.OpenFile(fmt.Sprintf("%s/%d.eml", dir, msg.SeqNum), os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
		if e != nil {
			log.Error("  Open file error", "error", e)
			continue
		}
		io.Copy(out, r)
		out.Close()
	}
	log.Info("Done!")
}

func main() {
	c := connect()
	if c == nil {
		return
	}
	defer c.Logout()

	boxes := boxes(c)
	for _, box := range boxes {
		fetch(c, box.Name)
	}
}
