package requests

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"net/url"
	"sams_helper/conf"
	"time"
)

type Request struct {
	Headers *http.Header `json:"headers"`
	Client  *http.Client `json:"client"`
}

func (request *Request) InitRequest(setting conf.Setting) error {
	u, _ := url.Parse(setting.ProxySet.ProxyUrl)
	t := &http.Transport{
		MaxIdleConns:    10,
		MaxConnsPerHost: 10,
		IdleConnTimeout: time.Duration(10) * time.Second,
		Proxy:           http.ProxyURL(u),
	}

	if setting.ProxySet.IsEnabled {
		request.Client = &http.Client{
			Timeout:   60 * time.Second,
			Transport: t,
		}
	} else {
		request.Client = &http.Client{
			Timeout: 60 * time.Second,
		}
	}

	request.Headers = &http.Header{
		"Host":            []string{"api-sams.walmartmobile.cn"},
		"Content-Type":    []string{"application/json"},
		"Accept":          []string{"*/*"},
		"Auth-Token":      []string{setting.AuthToken},
		"Accept-Language": []string{"zh-Hans-CN;q=1, en-CN;q=0.9, ga-IE;q=0.8"},
	}

	switch setting.DeviceType {
	case 2:
		request.Headers.Set("device-type", "mini_program")
		request.Headers.Set("user-agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 11_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E217 MicroMessenger/6.8.0(0x16080000) NetType/WIFI Language/en Branch/Br_trunk MiniProgramEnv/Mac")
	default: // 默认 ios
		request.Headers.Set("device-type", "ios")
		request.Headers.Set("user-agent", "SamClub/5.0.47 (iPhone; iOS 15.4.1; Scale/3.00)SamClub/5.0.47 (iPhone; iOS 15.4.1; Scale/3.00)")
	}

	if len(setting.DeviceId) > 30 {
		request.Headers.Set("device-id", setting.DeviceId)
	}

	return nil
}

func (request *Request) GET(url string) (error, gjson.Result) {
	req, _ := http.NewRequest("GET", url, nil)
	return request.do(req)
}

func (request *Request) POST(url string, data []byte) (error, gjson.Result) {
	req, _ := http.NewRequest("POST", url, bytes.NewReader(data))
	return request.do(req)
}

func (request *Request) do(req *http.Request) (error, gjson.Result) {
	req.Header = *request.Headers

	resp, err := request.Client.Do(req)
	if err != nil {
		return err, gjson.Result{}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err, gjson.Result{}
	}
	_ = resp.Body.Close()
	if resp.StatusCode == 200 {
		result := gjson.Parse(string(body))
		switch result.Get("code").Str {
		case "Success":
			return nil, result
		case "AUTH_FAIL":
			return conf.AuthFailErr, gjson.Result{}
		case "LIMITED":
			return conf.LimitedErr, gjson.Result{}
		case "CART_GOOD_CHANGE":
			return conf.CartGoodChangeErr, gjson.Result{}
		case "CLOSE_ORDER_TIME_EXCEPTION":
			return conf.CloseOrderTimeExceptionErr, gjson.Result{}
		case "DECREASE_CAPACITY_COUNT_ERROR":
			return conf.DecreaseCapacityCountErr, gjson.Result{}
		case "GOODS_EXCEED_LIMIT":
			return conf.GoodsExceedLimitErr, gjson.Result{}
		case "OUT_OF_STOCK":
			return conf.OutOfSellErr, gjson.Result{}
		case "NOT_DELIVERY_CAPACITY_ERROR":
			return conf.NotDeliverCapCityErr, gjson.Result{}
		case "STORE_HAS_CLOSED":
			return conf.StoreHasClosedErr, gjson.Result{}
		case "NO_MATCH_DELIVERY_MODE":
			return conf.NoMatchDeliverMode, gjson.Result{}
		case "FAIL":
			return conf.FAILErr, gjson.Result{}
		case "CART_GOODS_SPU_EXCEED_LIMIT":
			return conf.CartFullErr, gjson.Result{}
		case "GET_DELIVERY_INFO_ERROR":
			return conf.GetDeliveryErr, gjson.Result{}
		case "NotCheckShopPendingErr":
			return conf.NotCheckShopPendingErr, gjson.Result{}
		case "REQUEST_ERROR":
			fmt.Printf("[!] 请求错误 %s\n", result.Get("msg").Str)
			return conf.RequestErr, gjson.Result{}
		case "CLOUD_GOODS_OVER_WEIGHT":
			return conf.GoodsOverWeightErr, gjson.Result{}
		default:
			return errors.New(fmt.Sprintf("code: %s %s", result.Get("code"), result.Get("msg").Str)), gjson.Result{}
		}
	} else {
		return errors.New(fmt.Sprintf("[%v] %s", resp.StatusCode, body)), gjson.Result{}
	}
}
