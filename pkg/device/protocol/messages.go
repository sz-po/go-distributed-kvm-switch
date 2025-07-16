package protocol

type CallId string
type ServiceName string
type MethodName string

type MessageEnvelope struct {
	Call   *MethodCallMessage   `json:"call,omitempty"`
	Result *MethodResultMessage `json:"result,omitempty"`
}

type MethodCallMessage struct {
	CallId      CallId      `json:"callId"`
	ServiceName ServiceName `json:"serviceName"`
	MethodName  MethodName  `json:"methodName"`
	Payload     any         `json:"payload"`
}

type MethodResultMessage struct {
	CallId  CallId  `json:"callId"`
	Payload any     `json:"payload"`
	Error   *string `json:"error,omitempty"`
}
