package dto

// Response is the generic envelope used by endpoints that follow the
// {code, msg, data} convention.
type Response struct {
	Code  int         `json:"code"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data,omitempty"`
	Error []string    `json:"error,omitempty"`
}
