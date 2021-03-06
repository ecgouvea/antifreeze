package apifakes

import (
	"github.com/cloudfoundry/loggregator_consumer"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
)

type OldFakeLoggregatorConsumer struct {
	RecentCalledWith struct {
		AppGuid   string
		AuthToken string
	}

	RecentReturns struct {
		Messages  []*logmessage.LogMessage
		Err       []error
		callIndex int
	}

	TailFunc func(appGuid, token string) (<-chan *logmessage.LogMessage, error)

	IsClosed bool

	OnConnectCallback func()

	closeChan chan bool
}

func NewFakeLoggregatorConsumer() *OldFakeLoggregatorConsumer {
	return &OldFakeLoggregatorConsumer{
		closeChan: make(chan bool, 1),
	}
}

func (c *OldFakeLoggregatorConsumer) Recent(appGuid string, authToken string) ([]*logmessage.LogMessage, error) {
	c.RecentCalledWith.AppGuid = appGuid
	c.RecentCalledWith.AuthToken = authToken

	var err error
	if c.RecentReturns.callIndex < len(c.RecentReturns.Err) {
		err = c.RecentReturns.Err[c.RecentReturns.callIndex]
		c.RecentReturns.callIndex++
	}

	return c.RecentReturns.Messages, err
}

func (c *OldFakeLoggregatorConsumer) Close() error {
	c.IsClosed = true
	c.closeChan <- true
	return nil
}

func (c *OldFakeLoggregatorConsumer) SetOnConnectCallback(cb func()) {
	c.OnConnectCallback = cb
}

func (c *OldFakeLoggregatorConsumer) Tail(appGuid string, authToken string) (<-chan *logmessage.LogMessage, error) {
	return c.TailFunc(appGuid, authToken)
}

func (c *OldFakeLoggregatorConsumer) WaitForClose() {
	<-c.closeChan
}

func (c *OldFakeLoggregatorConsumer) SetDebugPrinter(debugPrinter loggregator_consumer.DebugPrinter) {
	<-c.closeChan
}
