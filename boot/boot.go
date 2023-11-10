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

package boot

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/bit-fever/core"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"io"
	"log/slog"
	"net/http"
	"os"
)

//=============================================================================
//===
//=== Public functions
//===
//=============================================================================

func InitApplication(component string, config any) *slog.Logger {
	readConfig(component, config)

	logFile := "config/"+ component +".log"
	return initLogger(component, logFile, config.(*core.Config))
}

//=============================================================================

func RunHttpServer(router *gin.Engine, config any) {

	slog.Info("Starting HTTPS server...")
	rootCAs, err := x509.SystemCertPool()
	core.ExitIfError(err)

	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	caCert, err := os.ReadFile("config/ca.crt")
	core.ExitIfError(err)

	if ok := rootCAs.AppendCertsFromPEM(caCert); !ok {
		core.ExitWithMessage("Failed to append CA cert to local certificate pool")
	}

	tlsConfig := &tls.Config{
		ClientCAs:  rootCAs,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	server := &http.Server{
		Addr:      config.(*core.Config).Application.BindAddress,
		TLSConfig: tlsConfig,
		Handler:   router,
	}

	slog.Info("Running")
	err = server.ListenAndServeTLS("config/server.crt", "config/server.key")
	core.ExitIfError(err)
}

//=============================================================================
//===
//=== Private functions
//===
//=============================================================================

func readConfig(component string, config any) {
	viper.SetConfigName(component)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/bit-fever/")
	viper.AddConfigPath("$HOME/.bit-fever/"+component)
	viper.AddConfigPath("config")

	err := viper.ReadInConfig()
	core.ExitIfError(err)

	err = viper.Unmarshal(config)
	core.ExitIfError(err)
}

//=============================================================================

func initLogger(component string, logFile string, config *core.Config) *slog.Logger {

	//--- Create log file

	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	core.ExitIfError(err)

	var wrt io.Writer = f

	if ! config.Application.Production {
		wrt = io.MultiWriter(os.Stdout, f)
	}

	//--- create logger

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	if !config.Application.Debug {
		opts = nil
	}

	logger := slog.New(slog.NewJSONHandler(wrt, opts)).With(
		slog.String("component", component),
		slog.Int   ("pid",       os.Getpid()),
	)

	slog.SetDefault(logger)
	gin.DefaultWriter = wrt

	return logger
}

//=============================================================================
