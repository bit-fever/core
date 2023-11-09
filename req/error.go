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
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"os"
)

//=============================================================================

func NewRequestError(message string, params ...any) error {
	msg := fmt.Sprintf(message, params)
	err := AppError{
		RequestError: errors.New(msg),
	}

	return err
}

//=============================================================================

func NewAccessError(message string, params ...any) error {
	msg := fmt.Sprintf(message, params)
	err := AppError{
		AccessError: errors.New(msg),
	}

	return err
}

//=============================================================================

func NewServerError(message string, params ...any) error {
	msg := fmt.Sprintf(message, params)
	err := AppError{
		ServerError: errors.New(msg),
	}

	return err
}

//=============================================================================

func NewServerErrorByError(err error) error {
	if err == nil {
		return nil
	}

	return AppError{
		ServerError: err,
	}
}

//=============================================================================

func ReturnError(c *gin.Context, err error) {
	if err != nil {
		var ae AppError
		if errors.As(err, &ae) {
			if ae.RequestError != nil {
				writeError(c, http.StatusBadRequest, err.Error(), nil)
			} else if ae.AccessError != nil {
				writeError(c, http.StatusForbidden, err.Error(), nil)
			} else if ae.ServerError != nil {
				writeError(c, http.StatusInternalServerError, err.Error(), nil)
			} else {
				writeError(c, http.StatusInternalServerError, "Bad AppError object", nil)
			}
		} else {
			writeError(c, http.StatusInternalServerError, "Found non AppError object : "+ err.Error(), nil)
		}
	}
}

//=============================================================================

func ReturnUnauthorizedError(c *gin.Context, message string) {
	writeError(c, http.StatusUnauthorized, message, nil)
}

//=============================================================================

func ReturnForbiddenError(c *gin.Context, message string) {
	writeError(c, http.StatusForbidden, message, nil)
}

//=============================================================================
//===
//=== Private methods
//===
//=============================================================================

type errorResponse struct {
	Code    int    `json:"code"`
	Error   string `json:"error"`
	Details any    `json:"details,omitempty"`
}

//-----------------------------------------------------------------------------

func writeError(c *gin.Context, errorCode int, errorMessage string, details any) {

	slog.Error(errorMessage,
		"client", c.ClientIP(),
		"pid",  os.Getpid(),
		"code", errorCode,
		"data", details)

	c.JSON(errorCode, &errorResponse{
		Code:    errorCode,
		Error:   errorMessage,
		Details: details,
	})
}

//=============================================================================
