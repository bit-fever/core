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
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//=============================================================================

const MaxQueryLimit = 5000

//=============================================================================
//===
//=== Parameter retrieval
//===
//=============================================================================

func GetPagingParams(c *gin.Context) (offset int, limit int, errV error) {

	offset, err1 := GetParamAsInt(c, "offset", 0)

	if err1 != nil || offset < 0 {
		return 0, 0, NewBadRequestError("Invalid 'offset' param: %v", offset)
	}

	//--- Extract limit

	limit, err2 := GetParamAsInt(c, "limit", MaxQueryLimit)

	if err2 != nil || limit < 1 || limit > MaxQueryLimit {
			return 0, 0, NewBadRequestError("Invalid 'limit' param: %v", limit)
		}

	return offset, limit, nil
}

//=============================================================================

func GetParamAsBool(c *gin.Context, name string, defValue bool) (bool, error) {
	params := c.Request.URL.Query()

	if ! params.Has(name) {
		return defValue, nil
	}

	value := params.Get(name)

	res, err := strconv.ParseBool(value)

	if err == nil {
		return res, nil
	}

	return false, NewBadRequestError("Parameter '%v' has not a boolean value: %v", name, value)
}

//=============================================================================

func GetParamAsInt(c *gin.Context, name string, defValue int) (int, error) {
	params := c.Request.URL.Query()

	if ! params.Has(name) {
		return defValue, nil
	}

	value := params.Get(name)

	res, err := strconv.ParseInt(value, 10, 32)

	if err == nil {
		return int(res), nil
	}

	return 0, NewBadRequestError("Parameter '%v' has not an integer value: %v", name, value)
}

//=============================================================================

func GetParamAsString(c *gin.Context, name string, defValue string) string {
	params := c.Request.URL.Query()

	if ! params.Has(name) {
		return defValue
	}

	value := params.Get(name)

	if value == "" {
		return defValue
	}

	return value
}

//=============================================================================

func BindParamsFromQuery(c *gin.Context, obj any) (err error) {
	if err := c.ShouldBindQuery(obj); err != nil {
		message := parseError(err)
		return NewBadRequestError(message, nil)
	}

	return nil
}

//=============================================================================

func BindParamsFromBody(c *gin.Context, obj any) (err error) {
	if err := c.ShouldBind(obj); err != nil {
		message := parseError(err)
		return NewBadRequestError(message, nil)
	}

	return nil
}

//=============================================================================

func GetIdFromUrl(c *gin.Context) (uint, error) {
	sId := c.Param("id")
	iId, err := strconv.ParseInt(sId, 10, 64)

	if err != nil || iId<0 {
		return 0, NewBadRequestError("Invalid ID in url: %v", sId)
	}

	return uint(iId), nil
}

//=============================================================================

type listResponse struct {
	Offset   int  `json:"offset"`
	Limit    int  `json:"limit"`
	Overflow bool `json:"overflow"`
	Result   any  `json:"result"`
}

//-----------------------------------------------------------------------------

func ReturnList(c *gin.Context, result any, offset int, limit int, size int) error {
	c.JSON(http.StatusOK, &listResponse{
		Offset:   offset,
		Limit:    limit,
		Overflow: size == MaxQueryLimit,
		Result:   result,
	})

	return nil
}

//=============================================================================
//===
//=== Private methods
//===
//=============================================================================

func parseError(err error) string {
	switch typedError := any(err).(type) {
	case validator.ValidationErrors:
		for _, e := range typedError {
			return parseFieldError(e)
		}

	case *json.UnmarshalTypeError:
		return parseMarshallingError(*typedError)

	case *strconv.NumError:
		return parseConvertError(*typedError)
	}

	return err.Error()
}

//=============================================================================

func parseFieldError(e validator.FieldError) string {
	field := strings.ToLower(e.Field())
	fieldPrefix := fmt.Sprintf("The field %s", field)
	tag := strings.Split(e.Tag(), "|")[0]

	switch tag {
	case "required":
		return fmt.Sprintf("Missing the '%s' parameter", field)

	case "required_without":
		return fmt.Sprintf("%s is required if %s is not supplied", fieldPrefix, e.Param())

	case "lt", "ltfield":
		param := e.Param()
		if param == "" {
			param = time.Now().Format(time.RFC3339)
		}
		return fmt.Sprintf("%s must be less than %s", fieldPrefix, param)

	case "gt", "gtfield":
		param := e.Param()
		if param == "" {
			param = time.Now().Format(time.RFC3339)
		}
		return fmt.Sprintf("%s must be greater than %s", fieldPrefix, param)

	default:
		return fmt.Errorf("%v", e).Error()
	}
}

//=============================================================================

func parseMarshallingError(e json.UnmarshalTypeError) string {
	return fmt.Sprintf("Invalid type: '%s' must be a %s", strings.ToLower(e.Field), e.Type.String())
}

//=============================================================================

func parseConvertError(e strconv.NumError) string {
	return fmt.Sprintf("Parameter must be an integer: %s", e.Num)
}

//=============================================================================
