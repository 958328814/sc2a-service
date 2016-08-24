package dist

import (
	"bytes"
	"fmt"

	"gopkg.in/mailgun/mailgun-go.v1"
)

type notifyEmailContext struct {
	Release Release
	Date    string
	Link    string
}

// NotifyAll sends email notification to all subscribers
func NotifyAll(release Release) error {
	ctx := notifyEmailContext{
		Release: release,
		Date:    release.Date.Format("20060102150405"),
		Link:    "",
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

	for _, link := range links {
		ctx.Link = makeLink(link.ID)
		err := sendNotifyEmail(ctx, subIDMap[link.SubID].Email)
		if err != nil {
			return fmt.Errorf("notify: send to %s: %s", link.SubID, err.Error())
		}
	}

	return nil
}

func sendNotifyEmail(ctx notifyEmailContext, email string) error {
	buf := bytes.NewBuffer(nil)
	var subject, content string
	err := notifyEmailSubjectTemplate.Execute(buf, ctx)
	if err != nil {
		return err
	}
	subject = string(buf.Bytes())
	buf.Reset()
	err = notifyEmailContentTemplate.Execute(buf, ctx)
	if err != nil {
		return err
	}
	content = string(buf.Bytes())
	m := mailgun.NewMessage("notify@dreamdota.com", subject, content, email)
	_, _, err = mg.Send(m)
	return err
}

// NotifySubscriber sends email to a subscriber
func NotifySubscriber(sub Sub, release Release) error {
	return nil
}
