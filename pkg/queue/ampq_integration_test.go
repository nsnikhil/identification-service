package queue_test

//
//import (
//	"github.com/streadway/amqp"
//	"github.com/stretchr/testify/suite"
//	"identification-service/pkg/config"
//	"identification-service/pkg/queue"
//	"identification-service/pkg/test"
//	"testing"
//	"time"
//)
//
//type ampqTestSuite struct {
//	qu  queue.AMQP
//	cfg config.QueueConfig
//	ch  *amqp.Channel
//	suite.Suite
//}
//
//func (qt *ampqTestSuite) SetupSuite() {
//	cfg := config.NewConfig("../../local.env").QueueConfig()
//
//	qu := queue.NewAMQP(cfg.Address())
//
//	conn, err := amqp.Dial(cfg.Address())
//	qt.Require().NoError(err)
//
//	ch, err := conn.Channel()
//	qt.Require().NoError(err)
//
//	time.Sleep(time.Second)
//
//	qt.qu = qu
//	qt.cfg = cfg
//	qt.ch = ch
//}
//
//func (qt *ampqTestSuite) TearDownSuite() {
//	qt.Require().NoError(qt.qu.Close())
//}
//
//func TestAmpqTestSuite(t *testing.T) {
//	suite.Run(t, new(ampqTestSuite))
//}
//
//func (qt *ampqTestSuite) TestPushSuccess() {
//	queueName := test.RandString(8)
//	data := test.RandBytes(8)
//
//	defer deleteQueue(qt, queueName)
//
//	done := make(chan bool)
//
//	go func() {
//		err := qt.qu.Push(queueName, data)
//		qt.Require().NoError(err)
//		done <- true
//	}()
//
//	select {
//	case <-done:
//	case <-time.After(time.Second * 10):
//		qt.Fail("pushed failed, timed out")
//	}
//}
//
//func (qt *ampqTestSuite) TestUnPushSuccess() {
//	queueName := test.RandString(8)
//	data := test.RandBytes(8)
//
//	defer deleteQueue(qt, queueName)
//
//	done := make(chan bool)
//
//	go func() {
//		err := qt.qu.UnsafePush(queueName, data)
//		qt.Require().NoError(err)
//		done <- true
//	}()
//
//	select {
//	case <-done:
//	case <-time.After(time.Second * 10):
//		qt.Fail("pushed failed, timed out")
//	}
//}
//
//func (qt *ampqTestSuite) TestStreamSuccess() {
//	queueName := test.RandString(8)
//	data := test.RandBytes(8)
//
//	defer deleteQueue(qt, queueName)
//
//	go func() {
//		err := qt.qu.UnsafePush(queueName, data)
//		qt.Require().NoError(err)
//	}()
//
//	value := make(chan []byte)
//
//	go func() {
//		ch, err := qt.qu.Stream(queueName)
//		qt.Require().NoError(err)
//
//		v := <-ch
//		value <- v.Body
//	}()
//
//	select {
//	case v := <-value:
//		qt.Assert().Equal(data, v)
//	case <-time.After(time.Second * 10):
//		qt.Fail("stream failed, timed out")
//	}
//}
//
//func deleteQueue(qt *ampqTestSuite, queueName string) {
//	_, err := qt.ch.QueueDelete(queueName, false, false, true)
//	qt.Require().NoError(err)
//}
