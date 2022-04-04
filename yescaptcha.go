package yescaptcha

import (
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/ddliu/go-httpclient"
	"log"
	"net/http"
	"time"
)

type Client struct {
	clientKey   string
	siteKey     string
	siteUrl     string
	captchaType string
	taskId      string
	softId      int64
	balance     int64
	timeout     int
	interval    int
	debug       bool
}

// NewClient 创建对象
func NewClient(clientKey, siteKey, siteUrl, captchaType string, timeout, interval int) *Client {
	return &Client{
		clientKey:   clientKey,
		siteKey:     siteKey,
		siteUrl:     siteUrl,
		captchaType: captchaType,
		softId:      0,
		timeout:     timeout,
		interval:    interval,
	}
}

// GetSoftId 获取软件Id
func (c *Client) GetSoftId() (int64, *CaptchaError) {
	data := map[string]interface{}{
		"clientKey": c.clientKey,
	}
	resp, err := c.post("getSoftID", data)
	if err == nil {
		errorId, _ := jsonparser.GetInt(resp, "errorId")
		if errorId == 0 {
			softId, _ := jsonparser.GetInt(resp, "softID")
			c.softId = softId
			return softId, nil
		} else {
			return 0, resToErr(resp)
		}
	}
	return 0, err
}

// Solve 创建任务获取验证码
func (c *Client) Solve() (string, *CaptchaError) {
	if c.softId == 0 {
		_, err := c.GetSoftId()
		if err != nil {
			return "", err
		}
	}
	log.Println(c.softId)
	_, err := c.CreateTask()
	if err != nil {
		return "", err
	}
	log.Println(c.taskId)
	captcha, err := c.WaitForResult()

	if err != nil {
		return "", err
	}

	return captcha, nil
}

// GetBalance 获取余额
func (c *Client) GetBalance() (int64, *CaptchaError) {
	data := map[string]interface{}{
		"clientKey": c.clientKey,
	}
	resp, err := c.post("getBalance", data)
	if err == nil {
		errorId, _ := jsonparser.GetInt(resp, "errorId")
		if errorId == 0 {
			balance, _ := jsonparser.GetInt(resp, "balance")
			c.balance = balance
			return balance, nil
		} else {
			return 0, resToErr(resp)
		}
	}
	return 0, err
}

// CreateTask 创建任务
func (c *Client) CreateTask() (string, *CaptchaError) {
	data := map[string]interface{}{
		"clientKey": c.clientKey,
		"task": map[string]interface{}{
			"websiteURL": c.siteUrl,
			"websiteKey": c.siteKey,
			"type":       c.captchaType,
		},
		"softId": c.softId,
	}
	resp, err := c.post("createTask", data)
	if err == nil {
		errorId, _ := jsonparser.GetInt(resp, "errorId")
		if errorId == 0 {
			taskId, _ := jsonparser.GetString(resp, "taskId")
			c.taskId = taskId
			return taskId, nil
		} else {
			return "", resToErr(resp)
		}
	}
	return "", err
}

func (c *Client) WaitForResult() (string, *CaptchaError) {

	start := time.Now()
	now := start
	for now.Sub(start) < (time.Duration(c.timeout) * time.Second) {

		time.Sleep(time.Duration(c.interval) * time.Second)

		captcha, err := c.GetTaskResult()

		if err == nil {
			return captcha, nil
		}

		now = time.Now()
	}

	return "", NewCaptchaError("ERROR_WAIT_CAPTCHA_TIME_OUT", "等待验证码超时")
}

// GetTaskResult 获取任务结果
func (c *Client) GetTaskResult() (string, *CaptchaError) {
	data := map[string]interface{}{
		"clientKey": c.clientKey,
		"taskId":    c.taskId,
	}
	resp, err := c.post("getTaskResult", data)
	if err == nil {
		errorId, _ := jsonparser.GetInt(resp, "errorId")
		if errorId == 0 {
			status, _ := jsonparser.GetString(resp, "status")
			if status == "ready" {
				captcha, _ := jsonparser.GetString(resp, "solution", "gRecaptchaResponse")
				return captcha, nil
			}
			if status == "processing" {
				return "", NewCaptchaError("ERROR_PROCESSING", "正在识别中")
			}
		}
		return "", resToErr(resp)
	}
	return "", err
}

func (c *Client) post(action string, data map[string]interface{}) ([]byte, *CaptchaError) {
	url := fmt.Sprintf("https://hk.yescaptcha.com/%s", action)
	res, _ := httpclient.WithHeader("ContentType", "application/x-www-form-urlencoded").
		PostJson(url, data)

	if res == nil || res.Response == nil {
		return nil, NewCaptchaError("ERROR_POST_NOT_RESPONSE", fmt.Sprintf("%s 请求无返回", url))
	}
	if res.StatusCode != http.StatusOK {
		return nil, NewCaptchaError("ERROR_POST_STATUS_CODE", fmt.Sprintf("%s 返回状态码: %d", url, res.StatusCode))
	}

	resp, _ := res.ReadAll()

	return resp, nil
}

func resToErr(resp []byte) *CaptchaError {
	errorCode, _ := jsonparser.GetString(resp, "errorCode")
	errorDescription, _ := jsonparser.GetString(resp, "errorDescription")
	return NewCaptchaError(errorCode, errorDescription)
}
