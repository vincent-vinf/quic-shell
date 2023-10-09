package types

import "encoding/json"

const (
	AppProtocol = "quic-echo-example"
)

type MessageType string

const (
	ReportID MessageType = "ReportID"
)

type Message struct {
	Type MessageType
	ID   string
}

func UnpackMsg(data []byte) (*Message, error) {
	m := &Message{}
	err := json.Unmarshal(data, m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func PackMsg(m *Message) ([]byte, error) {
	return json.Marshal(m)
}
