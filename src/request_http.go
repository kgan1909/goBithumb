package src

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type httpRequester struct {
	connectKey string
	secretKey  string

	basicUrl string

	publicClient  *http.Client
	privateClient *http.Client
}

func newHttpRequester(connectKey string, secretKey string) *httpRequester {

	httpRequester := httpRequester{}

	httpRequester.connectKey = connectKey
	httpRequester.secretKey = secretKey
	httpRequester.basicUrl = "https://api.bithumb.com"

	Timelog(utf8.ValidString(connectKey))
	Timelog(utf8.ValidString(secretKey))

	httpRequester.publicClient = &http.Client{}
	httpRequester.privateClient = &http.Client{}

	return &httpRequester
}

func (h *httpRequester) requestPublic(order publicOrder, data string) []byte {

	Timelog(h.basicUrl + string(order) + "/" + data)

	request, err := http.NewRequest("GET", h.basicUrl+string(order)+"/"+data, nil)
	if err != nil {
		panic("Failed to create Request")
	}

	response, err := h.publicClient.Do(request)
	if err != nil {
		panic("Failed to receive Data, check server status")
	}

	byteResponse, err := ioutil.ReadAll(response.Body)
	if err != nil {
		Timelog("Failed to receive Data")
		panic(err)
	}
	return byteResponse
}

func (h *httpRequester) requestPrivate(order privateOrder, data interface{}, values url.Values) []byte {

	Timelog("connectKey : ", h.connectKey)
	Timelog("secretKey : ", h.secretKey)
	Timelog(values.Encode())

	test := TEST{}

	Decocer := json.NewEncoder(&test)
	_ = Decocer.Encode(data)
	rawDataByte := test.Container
	Timelog("RAW : ", string(rawDataByte))

	//replaced := bytes.ReplaceAll(rawDataByte, []byte("\""), []byte("'"))
	//dataByte := bytes.NewBuffer(rawDataByte)

	dataByte := url.Values{}
	dataByte.Set("currency", "btc")
	dataByte.Set("endpoint", "/info/balance")

	nonce := time.Now().UnixNano() / int64(time.Millisecond)
	strNonce := strconv.FormatInt(nonce, 10)

	//request, err := http.NewRequest("POST", h.basicUrl+string(order), dataByte)
	request, err := http.NewRequest("POST", h.basicUrl+string(order), strings.NewReader(dataByte.Encode()))
	//request, err := http.NewRequest("POST", "http://0.0.0.0:9876", strings.NewReader(dataByte.Encode()))
	if err != nil {
		panic("Failed to create Request")
	}
	encrypted := h.encryptData(order, strNonce, []byte(values.Encode()))
	Timelog("Encrypted : ", encrypted)

	request.Header.Add("User-Agent", "tester")
	request.Header.Add("Api-Key", h.connectKey)
	//request.Header.Add("Secret-Key", h.secretKey)
	request.Header.Add("Api-Sign", encrypted)
	request.Header.Add("Api-Nonce", strNonce)
	request.Header.Add("Connection", "keep-alive")
	request.Header.Add("Accept-Encoding", "gzip, deflate")
	request.Header.Add("Accept", "*/*")
	request.Header.Add("Content-type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(dataByte.Encode())))

	response, err := h.privateClient.Do(request)
	if err != nil {
		panic("Failed to receive Data, check server stauts")
	}

	byteResponse, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic("Failed to receive Data")
	}
	return byteResponse
}

func (h *httpRequester) encryptData(order privateOrder, nonce string, data []byte) string {
	//reqRawString := string(order) + string(0) + string(data) + string(0) + nonce
	reqRawString := string(order) + string(data) + nonce

	hmacParsed := hmac.New(sha512.New, []byte(h.secretKey))
	hmacParsed.Write([]byte(reqRawString))

	hexData := hex.EncodeToString(hmacParsed.Sum(nil))
	byteHexData := []byte(hexData)
	hmacParsed.Reset()

	result := base64.StdEncoding.EncodeToString(byteHexData)
	return result
}

func Test2() {

	//original := NewBithumbRequester("57bc35837f7f00c6f64a25d25ef69f6f", "7539214f665dfbd945dac04010b16eea")
	//val := original.GetCandleStick(BTC, KRW, H24)
	//Timelog(len(val.Data))

}
