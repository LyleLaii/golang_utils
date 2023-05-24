package requestes

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

type RequestsConfig struct {
	VerifyTLS  bool
	CaCertPath string
	MutualTLS  bool
	CaKeyPath  string
	CaCrtPath  string
}

type RequestsClient struct {
	client *http.Client
}

type ExtraConfig func(r *http.Request)

type File struct {
	FileName    string
	FileContent []byte
}
type FormDataBody struct {
	Value map[string]string
	File  map[string]File
}

type PostData func(url string) (*http.Request, error)

type ResponseData struct {
	Status     string
	StatusCode int
	Header     map[string][]string
	Data       []byte
	//data string
}

func (rsp ResponseData) Text() string {
	return string(rsp.Data[:])
}

func (rsp ResponseData) BindJSON(obj interface{}) error {
	if err := json.Unmarshal(rsp.Data, obj); err != nil {
		return err
	}
	return nil
}

func New(c RequestsConfig) (*RequestsClient, error) {

	pool := x509.NewCertPool()
	if c.CaCertPath != "" {
		caCrt, err := ioutil.ReadFile(c.CaCertPath)
		if err != nil {
			return &RequestsClient{}, errors.Wrap(err, "read ca cert file failed")
		}
		pool.AppendCertsFromPEM(caCrt)
	}

	Certificates := make([]tls.Certificate, 0)
	if c.MutualTLS {
		if c.CaKeyPath == "" || c.CaCrtPath == "" {
			return &RequestsClient{}, errors.New("config error: if use mutual TLS, Must specify both of caKeyPath and caCrtPath")
		}
		cliCrt, err := tls.LoadX509KeyPair(c.CaCrtPath, c.CaKeyPath)
		if err != nil {
			return &RequestsClient{}, errors.Wrap(err, "load x509 key pair failed")
		}
		Certificates = []tls.Certificate{cliCrt}
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !c.VerifyTLS,
			RootCAs:            pool,
			Certificates:       Certificates,
		},
	}

	return &RequestsClient{&http.Client{Transport: tr}}, nil
}

func AddHeader(data map[string]string) ExtraConfig {
	add := func(req *http.Request) {
		for k, v := range data {
			req.Header.Set(k, v)
		}
	}
	return add
}

//func AddQueryParam(data map[string]string) ExtraConfig {
//	add := func(req *http.Request) {
//		q := req.URL.Query()
//		for k, v := range data {
//			q.Set(k, v)
//		}
//		req.URL.RawQuery = q.Encode()
//	}
//	return add
//}

// TODO: Ugly, need format trans, only support simple map/struct, optimize it
func AddQueryParam(data interface{}) ExtraConfig {
	dataType := reflect.TypeOf(data).Kind().String()
	if dataType == "map" {
		add := func(req *http.Request) {
			q := req.URL.Query()
			for k, v := range data.(map[string]string) {
				q.Set(k, v)
			}
			req.URL.RawQuery = q.Encode()
		}
		return add
	} else if dataType == "struct" {
		add := func(req *http.Request) {
			q := req.URL.Query()
			rType := reflect.TypeOf(data)
			rVal := reflect.ValueOf(data)
			for k := 0; k < rVal.NumField(); k++ {
				q.Set(rType.Field(k).Name, rVal.Field(k).String())
			}
			req.URL.RawQuery = q.Encode()
		}
		return add
	} else {
		fmt.Println("AddQueryParam now can only support a map or struct")
		return func(req *http.Request) {}
	}
}

func FormData(data map[string]string) PostData {
	add := func(url string) (*http.Request, error) {

		list := make([]string, 0)
		for k, v := range data {
			list = append(list, fmt.Sprintf("%s=%s", k, fmt.Sprint(v)))
		}
		data := strings.Join(list, "&")

		req, err := http.NewRequest("POST", url, strings.NewReader(data))
		if err != nil {
			return &http.Request{}, errors.Wrap(err, "Gen FormData failed")
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		return req, nil

	}

	return add
}

func JsonData(data interface{}) PostData {
	add := func(url string) (*http.Request, error) {
		b, err := json.Marshal(data)
		if err != nil {
			return &http.Request{}, errors.Wrap(err, "json format error")
		}
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
		if err != nil {
			return &http.Request{}, errors.Wrap(err, "crete request failed")
		}
		req.Header.Add("Content-Type", "application/json;charset=utf-8")

		return req, nil
	}

	return add
}

func MultiFormData(data FormDataBody) PostData {
	add := func(url string) (*http.Request, error) {
		buf := new(bytes.Buffer)
		bw := multipart.NewWriter(buf)
		for k, v := range data.value {
			d, _ := bw.CreateFormField(k)
			d.Write([]byte(v))
		}
		for k, v := range data.file {
			f, _ := bw.CreateFormFile(k, v.FileName)
			reader := bytes.NewReader(v.FileContent)
			io.Copy(f, reader)
		}
		bw.Close()
		req, err := http.NewRequest("POST", url, buf)
		if err != nil {
			return &http.Request{}, errors.Wrap(err, "Gen PostFormData failed")
		}
		req.Header.Add("Content-Type", bw.FormDataContentType())
		return req, nil
	}
	return add
}

func generateRepData(resp *http.Response) (ResponseData, error) {
	status := resp.Status
	statusCode := resp.StatusCode
	headers := make(map[string][]string)

	for k, v := range resp.Header {
		headers[k] = v
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ResponseData{Status: status, StatusCode: statusCode, Header: headers, Data: nil}, errors.Wrap(err, "read response data failed")
	}

	return ResponseData{Status: status, StatusCode: statusCode, Header: headers, Data: body}, nil
}

func (r *RequestsClient) Get(url string, extraConfigs ...ExtraConfig) (ResponseData, error) {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return ResponseData{}, errors.Wrap(err, "get method create request failed")
	}

	for _, extraConfig := range extraConfigs {
		extraConfig(req)
	}

	resp, err := r.client.Do(req)

	if err != nil {
		return ResponseData{}, errors.Wrap(err, "http do get failed")
	}

	return generateRepData(resp)
}

func (r *RequestsClient) Post(url string, postData PostData, extraConfigs ...ExtraConfig) (ResponseData, error) {
	req, err := postData(url)
	if err != nil {
		return ResponseData{}, errors.Wrap(err, "post method create request failed")
	}
	for _, extraConfig := range extraConfigs {
		extraConfig(req)
	}

	resp, err := r.client.Do(req)

	if err != nil {
		return ResponseData{}, errors.Wrap(err, "http do post failed")
	}

	return generateRepData(resp)
}

func (r *RequestsClient) PostFormData(url string, MultiFormData PostData, extraConfigs ...ExtraConfig) (ResponseData, error) {
	req, err := MultiFormData(url)
	if err != nil {
		return ResponseData{}, errors.Wrap(err, "post method create request failed")
	}
	for _, extraConfig := range extraConfigs {
		extraConfig(req)
	}

	resp, err := r.client.Do(req)

	if err != nil {
		return ResponseData{}, errors.Wrap(err, "http do post failed")
	}

	return generateRepData(resp)
}
