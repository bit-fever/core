//=============================================================================
/*
Copyright © 2023 Andrea Carboni andrea.carboni71@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
//=============================================================================

package req

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"github.com/bit-fever/core"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

//=============================================================================

var clientMap = map[string] *http.Client {}

//=============================================================================
//===
//=== Public methods
//===
//=============================================================================

func AddClient(id string, caCert string, clientCert string, clientKey string) {
	clientMap[id] = createClient(caCert, clientCert, clientKey)
}

//=============================================================================

func GetClient(id string) *http.Client {
	return clientMap[id]
}

//=============================================================================

func DoGet(client *http.Client, url string, output any, token string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+ token)
	}

	res, err := client.Do(req)
	return buildResponse(res, err, &output)
}

//=============================================================================

func DoPost(client *http.Client, url string, params any, output any, token string) error {
	body, err := json.Marshal(&params)
	if err != nil {
		slog.Error("Error marshalling POST parameter", "error", err.Error())
		return err
	}

	reader := bytes.NewReader(body)

	req, err := http.NewRequest("POST", url, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "Application/json")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+ token)
	}

	res, err := client.Do(req)
	return buildResponse(res, err, &output)
}

//=============================================================================

func DoPut(client *http.Client, url string, params any, output any, token string) error {
	body, err := json.Marshal(&params)
	if err != nil {
		slog.Error("Error marshalling PUT parameter", "error", err.Error())
		return err
	}

	reader := bytes.NewReader(body)

	req, err := http.NewRequest("PUT", url, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "Application/json")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+ token)
	}

	res, err := client.Do(req)
	return buildResponse(res, err, &output)
}

//=============================================================================

func DoDelete(client *http.Client, url string, output any, token string) error {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		slog.Error("Error creating a DELETE request", "error", err.Error())
		return err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+ token)
	}

	res, err := client.Do(req)
	return buildResponse(res, err, &output)
}

//=============================================================================
//===
//=== Private methods
//===
//=============================================================================

func createClient(caCert string, clientCert string, clientKey string) *http.Client {
	cert, err := os.ReadFile("config/"+ caCert)
	core.ExitIfError(err)

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(cert)

	certificate, err := tls.LoadX509KeyPair("config/"+ clientCert, "config/"+ clientKey)
	core.ExitIfError(err)

	return &http.Client{
		Timeout: time.Minute * 3,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{certificate},
			},
		},
	}
}

//=============================================================================

func buildResponse(res *http.Response, err error, output any) error {
	if err != nil {
		slog.Error("Error sending request", "error", err.Error())
		return err
	}

	if res.StatusCode == 401 {
		slog.Error("Error from the server", "error", res.Status)
		return errors.New("Authorization failed")
	}

	if res.StatusCode == 404 {
		slog.Error("Error from the server", "error", res.Status)
		return errors.New("Not found")
	}

	if res.StatusCode >= 400 {
		slog.Error("Error from the server", "error", res.Status)
		return errors.New("Client error: "+ res.Status)
	}

	//--- Read the response body
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("Error reading response", "error", err.Error())
		return err
	}

	err = json.Unmarshal(body, &output)
	if err != nil {
		slog.Error("Bad JSON response from server", "error", err.Error())
	}

	return nil
}

//=============================================================================
