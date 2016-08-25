package dist

import (
	"bytes"
	"fmt"

	"gopkg.in/mailgun/mailgun-go.v1"
)

type notifyEmailContext struct {
	Release Release
	Date    string
}

// NotifyAll sends email notification to all subscribers
func NotifyAll(release Release) error {
	ctx := notifyEmailContext{
		Release: release,
		Date:    release.Date.Time().Format("20060102150405"),
	}

	subIds := []string{}
	subIDMap := map[string]*Sub{}
	subs, err := ListSubs()
	if err != nil {
		return fmt.Errorf("notify: list subs: %s", err.Error())
	}
	for i := range subs {
		s := &subs[i]
		subIds = append(subIds, s.ID)
		subIDMap[s.ID] = s
	}

	links, err := createLinks(subIds, release.ID)
	if err != nil {
		return fmt.Errorf("notify: create links: %s", err.Error())
	}

	m, err := getNotifyMessage(ctx)
	if err != nil {
		return fmt.Errorf("notify: create message: %s", err.Error())
	}
	for _, link := range links {
		m.AddRecipientAndVariables(subIDMap[link.SubID].Email, map[string]interface{}{
			"Link": makeLink(link.ID),
		})
	}

	_, _, err = mg.Send(m)
	if err != nil {
		return fmt.Errorf("notify: send: %s", err.Error())
	}

	return nil
}

func getNotifyMessage(ctx notifyEmailContext) (*mailgun.Message, error) {
	buf := bytes.NewBuffer(nil)
	var subject, content string
	err := notifyEmailSubjectTemplate.Execute(buf, ctx)
	if err != nil {
		return nil, err
	}
	subject = string(buf.Bytes())
	buf.Reset()
	err = notifyEmailContentTemplate.Execute(buf, ctx)
	if err != nil {
		return nil, err
	}
	content = string(buf.Bytes())
	return mailgun.NewMessage("DreamHacks <notify@dreamdota.com>", subject, content), nil
}

// NotifySubscriber sends email to a subscriber
func NotifySubscriber(sub Sub, release Release) error {
	return nil
}
