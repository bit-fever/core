//=============================================================================
/*
Copyright © 2025 Andrea Carboni andrea.carboni71@gmail.com

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

package msg

import "time"

//=============================================================================
//===
//=== Events
//===
//=============================================================================

type EventLevel int8

const (
	EventLevelInfo    = 0
	EventLevelWarning = 1
	EventLevelError   = 2
)

//=============================================================================

type Event struct {
	Username   string
	Level      EventLevel
	EventDate  time.Time
	Code       string
	Title      string
	Message    string
	Parameters map[string]any
}

//=============================================================================

func SendEventByCode(username string, code string, params map[string]any) error {
	e := Event{
		Username  : username,
		EventDate : time.Now(),
		Code      : code,
		Parameters: params,
	}

	return SendMessage(ExEvent, SourceEvent, TypeCreate, e)
}

//=============================================================================

func SendEvent(username string, level EventLevel, title, message string, params map[string]any) error {
	e := Event{
		Username  : username,
		Level     : level,
		EventDate : time.Now(),
		Title     : title,
		Message   : message,
		Parameters: params,
	}

	return SendMessage(ExEvent, SourceEvent, TypeCreate, e)
}

//=============================================================================
