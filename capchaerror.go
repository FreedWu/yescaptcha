package yescaptcha

type CaptchaError struct {
	code string `json:"code"`
	msg  string `json:"msg"`
}

func (e *CaptchaError) Error() string {
	return e.msg
}

func (e *CaptchaError) Code() string {
	return e.msg
}

func NewCaptchaError(code, msg string) *CaptchaError {
	return &CaptchaError{
		code: code,
		msg:  msg,
	}
}
