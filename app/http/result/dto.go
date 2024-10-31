package result

type Result struct {
	Code    int         `json:"code"`
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 方法，封装成功响应的结构体
func Success(data interface{}) Result {
	return Result{
		Code:    1,
		Success: true,
		Message: "Success",
		Data:    data,
	}
}

// SuccessWithMessage 方法，封装成功响应的结构体
func SuccessWithMessage(message string) Result {
	return Result{
		Code:    1,
		Success: true,
		Message: message,
	}
}

// SuccessWithOKMessage 方法，封装成功响应的结构体， message 操作成功
func SuccessWithOKMessage() Result {
	return Result{
		Code:    1,
		Success: true,
		Message: "操作成功",
	}
}

// Fail 方法，封装失败响应的结构体
func Fail(message string) Result {
	return Result{
		Code:    0,
		Success: false,
		Message: message,
		Data:    nil,
	}
}
