/*
 * Copyright 2023 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pkg

import (
	"context"
	"errors"
	"fmt"
	"github.com/slack-go/slack"
	"log"
	"strings"
	"time"
)

// NewSlackWriter creates a writer, that sends the accumulated writes on close to the provided url (like a slack webhook message)
func NewSlackWriter(url string, debug bool) *SlackWriter {
	return &SlackWriter{webhook: url, debug: debug}
}

type SlackWriter struct {
	buf     strings.Builder
	webhook string
	debug   bool
}

func (this *SlackWriter) Write(p []byte) (n int, err error) {
	return this.buf.Write(p)
}

// Close sends buffered messages and resets the buffer
// writer may be reused after close
func (this *SlackWriter) Close() error {
	defer this.buf.Reset()
	log.Println("send output to url")
	return SendSlackNotification(this.webhook, this.buf.String())
}

func SendSlackNotification(webhook string, message string) error {
	if webhook == "" {
		return nil
	}
	msg := fmt.Sprintf("Mopher Notification:\n%v\n", message)
	timeout, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := slack.PostWebhookContext(timeout, webhook, &slack.WebhookMessage{
		Text: msg,
	})
	if err != nil {
		err = errors.New(strings.ReplaceAll(err.Error(), webhook, "***"))
		log.Println("ERROR: unable to send slack message:", err)
		return err
	}
	return nil
}
