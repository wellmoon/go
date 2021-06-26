package socket

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"

	Log "github.com/wellmoon/go/logger"
)

type Message struct {
	EventName  string `json:"eventName"`
	MsgType    string `json:"msgType"`
	CmdIdx     string `json:"cmdIdx"`
	Code       int    `json:"code"`
	Content    string `json:"content"`
	TargetId   string `json:"targetId"`
	contentMap map[string]string
	Wait       *sync.WaitGroup
}

type MsgHandler func(message *Message, conn interface{})

var sendMsgMap map[string]*Message = make(map[string]*Message)
var recvMsgMap map[string]*Message = make(map[string]*Message)

func NewMessage(str string) (Message, error) {
	message := Message{}
	err := json.Unmarshal([]byte(str), &message)
	if err != nil {
		Log.Error("消息格式不正确：%v\n", str)
	}
	return message, err
}

func (message *Message) String() string {
	b, err := json.Marshal(message)
	if err != nil {
		return ""
	}
	return string(b)
}

func (message *Message) NewResponse(code int) *Message {
	if !strings.Contains(message.MsgType, "Request") {
		return nil
	}
	response := Message{}
	response.CmdIdx = message.CmdIdx
	if strings.HasPrefix(message.MsgType, "Client") {
		response.MsgType = "ServerResponse"
	} else {
		response.MsgType = "ClientResponse"
	}
	response.EventName = message.EventName
	response.Code = code
	return &response
}

func (message *Message) GetContent() *map[string]string {
	if len(message.contentMap) == 0 {
		message.contentMap = make(map[string]string)
		json.Unmarshal([]byte(message.Content), &message.contentMap)
	}
	return &message.contentMap
}

func (message *Message) SendMsg(conn interface{}) *Message {
	messageByte, _ := json.Marshal(message)
	messageStr := string(messageByte)
	messageStr = messageStr + "\r\n"
	Log.Trace("发送消息:%v\n", messageStr)

	v := reflect.ValueOf(conn)

	msg := reflect.ValueOf([]byte(messageStr))
	in := []reflect.Value{msg}
	var ret []reflect.Value

	if reflect.TypeOf(conn).String() == "*net.Conn" {
		// 如果反射对象是接口，需要调用实际的对象的方法，避免出现错误：panic: reflect: call of reflect.Value.Call on zero Value
		ret = v.Elem().MethodByName("Write").Call(in)
	} else {
		ret = v.MethodByName("Write").Call(in)
	}

	// 如果err不是nil
	if _, ok := ret[1].Interface().(error); ok {
		Log.Error("发送消息失败：%v\n", message.Content)
	}

	if strings.Contains(message.MsgType, "Request") {
		var wg sync.WaitGroup
		wg.Add(1)
		message.Wait = &wg
		Log.Trace("把消息放入发送队列:%v\n", message.CmdIdx)
		sendMsgMap[message.CmdIdx] = message
		wg.Wait()
		response := recvMsgMap[message.CmdIdx]
		Log.Trace("获取到[%v]返回消息:%v\n", message.CmdIdx, response.String())
		return response
	}
	return nil
}

func OnMessage(conn interface{}, msgChan chan string, handler MsgHandler) {
	var line string
	for {
		msg := <-msgChan
		if len(msg) == 0 {
			Log.Debug("获取管道消息失败，退出")
			break
		}
		Log.Trace("从管道中获取到消息：%v\n", msg)
		line = line + msg
		if strings.Contains(msg, "\r\n") {
			arr := strings.Split(line, "\r\n")
			line = arr[1]
			msg = arr[0]
			message := new(Message)
			err := json.Unmarshal([]byte(msg), message)
			if err != nil {
				Log.Error("消息解析失败，消息内容:%v\n", msg)
			} else {
				cmdIdx := message.CmdIdx
				msgType := message.MsgType
				if strings.Contains(msgType, "Response") {
					recvMsgMap[cmdIdx] = message
					sendMessage := sendMsgMap[cmdIdx]
					sendMessage.Wait.Done()
					Log.Trace("删除发送队列消息:%v\n", cmdIdx)
					delete(sendMsgMap, cmdIdx)

				} else {
					// 如果是request，执行相应的处理器
					go handler(message, conn)
				}
			}
		}
	}
}
