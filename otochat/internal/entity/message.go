package entity

type Type int

const (
	Broadcast Type = iota
	OneToOne
	OnlineUsers
)

type MessageResponse struct {
	Code int         `json:"code"`
	Type Type        `json:"type"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type MessageRequest struct {
	Type Type        `json:"type"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func NewMessageResponse(code int, t Type, msg string, data interface{}) MessageResponse {
	return MessageResponse{
		Code: code,
		Type: t,
		Msg:  msg,
		Data: data,
	}
}
