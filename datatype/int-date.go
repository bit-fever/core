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

package datatype

import (
	"fmt"
	"time"
)

//=============================================================================

type IntDate int

//=============================================================================

func (dt IntDate) Year() int {
	return  int(dt / 10000)
}

//=============================================================================

func (dt IntDate) Month() int {
	return int((dt / 100) % 100)
}

//=============================================================================

func (dt IntDate) Day() int {
	return int(dt % 100)
}

//=============================================================================

func (dt IntDate) String() string {
	return fmt.Sprintf("%4d-%2d-%2d", dt.Year(), dt.Month(), dt.Day())
}

//=============================================================================

func (dt IntDate) IsNil() bool {
	return dt == 0
}

//=============================================================================

func (dt IntDate) IsValid() bool {
	d := dt.Day()
	m := dt.Month()

	if m<1 || m>12 {
		return false
	}

	if m==4 || m==6 || m==9 || m==11 {
		return d>=1 && d<=30
	}

	if m==2 {
		return d>=1 && d<=29
	}

	return d>=1 && d<=31
}

//=============================================================================

func (dt IntDate) ToDateTime(endDay bool, loc *time.Location) time.Time {
	hh := 0
	mm := 0
	ss := 0

	if endDay {
		hh = 23
		mm = 59
		ss = 59
	}

	return time.Date(dt.Year(), time.Month(dt.Month()), dt.Day(), hh, mm, ss, 0, loc)
}

//=============================================================================
