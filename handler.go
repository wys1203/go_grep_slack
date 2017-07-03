package main

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"fmt"

	"strconv"

	"net/http"

	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
	"github.com/nlopes/slack"
)

// Yesterday is a representation of timestamp range
type Yesterday struct {
	startUnix string
	endUnix   string
	date      string
}

func unixTimestampOfYesterday(t time.Time) Yesterday {
	zone, _ := time.LoadLocation("Asia/Taipei")
	year, month, day := t.AddDate(0, 0, -1).Date()
	start := time.Date(year, month, day, 0, 0, 0, 0, zone)
	end := time.Date(year, month, day, 23, 59, 59, 0, zone)
	return Yesterday{strconv.FormatInt(start.Unix(), 10), strconv.FormatInt(end.Unix(), 10), fmt.Sprintf("%d-%02d-%02d", year, month, day)}
}

func grep(pattern string, text string) (string, error) {
	_, err := regexp.MatchString(pattern, text)
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(pattern)
	return re.FindString(text), err
}

// Handle is the entrypoint of aws lambda function
func Handle(evt json.RawMessage, ctx *runtime.Context) (interface{}, error) {
	api := slack.New("YOUR SLACK TOKEN")
	// If you set debugging, it will log all requests to the console
	// Useful when encountering issues
	// api.SetDebug(true)

	yesterday := unixTimestampOfYesterday(time.Now())
	history, err := api.GetChannelHistory("YOUR SLACK CHANNEL ID", slack.HistoryParameters{Latest: yesterday.endUnix, Oldest: yesterday.startUnix, Count: 100})
	if err != nil {
		panic(err)
	}
	m := map[string]int{}
	for _, message := range history.Messages {
		msg, _ := grep(".*\n", message.Msg.Text)
		msg = strings.Replace(msg, "\n", "", -1)
		m[msg]++
	}

	s := fmt.Sprintf("**%s**<br />", yesterday.date)
	for k, v := range m {
		s += fmt.Sprintf("%d %+3v<br />", v, k)
	}

	var r http.Request
	r.ParseForm()
	r.Form.Add("thread_id", "YOUR QUIP DOC THREAD ID")
	r.Form.Add("location", "2")
	r.Form.Add("format", "markdown")
	r.Form.Add("section_id", "YOUR SECTION ID IN QUIP DOC THREAD")
	r.Form.Add("content", s)
	bodystr := strings.TrimSpace(r.Form.Encode())
	req, err := http.NewRequest("POST", "https://platform.quip.com/1/threads/edit-document", strings.NewReader(bodystr))
	req.Header.Add("Authorization", "Bearer YOUR QUIP AUTH TOKEN")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	var resp *http.Response
	resp, _ = http.DefaultClient.Do(req)
	return resp.Status, nil
}

func main() {}
