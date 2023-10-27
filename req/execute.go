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
	"io"
	"log"
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
	client, err := createClient(caCert, clientCert, clientKey)

	if err != nil {
		log.Fatalf("Cannot create http client for '"+ id +"' : "+ err.Error())
	}

	clientMap[id] = client
}

//=============================================================================

func GetClient(id string) *http.Client {
	return clientMap[id]
}

//=============================================================================

func DoGet(client *http.Client, url string, output any) error {
	res, err := client.Get(url)
	return buildResponse(res, err, &output)
}

//=============================================================================

func DoPut(client *http.Client, url string, params any, output any) error {
	body, err := json.Marshal(&params)
	if err != nil {
		log.Printf("Error marshalling put parameter: %v", err)
		return err
	}

	reader := bytes.NewReader(body)
	res, err := client.Post(url, "Application/json", reader)
	return buildResponse(res, err, &output)
}

//=============================================================================
//===
//=== Private methods
//===
//=============================================================================

func createClient(caCert string, clientCert string, clientKey string) (*http.Client, error) {
	cert, err := os.ReadFile("config/"+ caCert)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(cert)

	certificate, err := tls.LoadX509KeyPair("config/"+ clientCert, "config/"+ clientKey)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Timeout: time.Minute * 3,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{certificate},
			},
		},
	}, nil
}

//=============================================================================

func buildResponse(res *http.Response, err error, output any) error {
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return err
	}

	if res.StatusCode >= 400 {
		log.Printf("Error from the server: %v", res.Status)
		return err
	}

	//--- Read the response body
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return err
	}

	err = json.Unmarshal(body, &output)
	if err != nil {
		log.Printf("Bad JSON response from server:\n%v", err)
	}

	return nil
}

//=============================================================================
