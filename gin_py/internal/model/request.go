package model

// ChatProcessRequest 聊天处理请求
type ChatProcessRequest struct {
	Prompt        string                     `json:"prompt"`                  // 用户消息
	Options       *ChatProcessRequestOptions `json:"options,omitempty"`       // 可选参数
	SystemMessage string                     `json:"systemMessage,omitempty"` // 系统提示词
}

// ChatProcessRequestOptions 聊天请求选项
type ChatProcessRequestOptions struct {
	ParentMessageId string `json:"parentMessageId,omitempty"` // 上一条消息ID
}

// ChatMessageResponse SSE 流式响应消息
type ChatMessageResponse struct {
	ID              string      `json:"id"`               // 消息ID
	Text            string      `json:"text"`             // 累计文本
	Role            string      `json:"role"`             // 角色
	Delta           string      `json:"delta"`            // 本次增量文本
	Detail          interface{} `json:"detail,omitempty"` // 完整详情
	TokenCount      int         `json:"tokenCount"`       // Token 数量
	ParentMessageId string      `json:"parentMessageId"`  // 父消息ID
}
