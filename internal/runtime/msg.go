package runtime

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Msg methods don't return errors because they can be used not only at startup.
// If runtime functions need to validate message at startup, they must do it by themselves.
type Msg interface {
	fmt.Stringer
	Type() MsgType
	Bool() bool
	Int() int64
	Float() float64
	Str() string
	List() []Msg
	Map() map[string]Msg
}

type MsgType uint8

const (
	UnknownMsgType MsgType = 0
	BoolMsgType    MsgType = 1
	IntMsgType     MsgType = 2
	FloatMsgType   MsgType = 3
	StrMsgType     MsgType = 4
	ListMsgType    MsgType = 5
	MapMsgType     MsgType = 6
)

// Empty

type emptyMsg struct{}

func (emptyMsg) String() string      { return "<empty>" }
func (emptyMsg) Type() MsgType       { return UnknownMsgType }
func (emptyMsg) Bool() bool          { return false }
func (emptyMsg) Int() int64          { return 0 }
func (emptyMsg) Float() float64      { return 0 }
func (emptyMsg) Str() string         { return "" }
func (emptyMsg) List() []Msg         { return nil }
func (emptyMsg) Map() map[string]Msg { return nil }

// Int

type IntMsg struct {
	emptyMsg
	v int64
}

func (msg IntMsg) Type() MsgType                { return IntMsgType }
func (msg IntMsg) Int() int64                   { return msg.v }
func (msg IntMsg) String() string               { return strconv.Itoa(int(msg.v)) }
func (msg IntMsg) MarshalJSON() ([]byte, error) { return []byte(msg.String()), nil }

func NewIntMsg(n int64) IntMsg {
	return IntMsg{
		emptyMsg: emptyMsg{},
		v:        n,
	}
}

// Float

type FloatMsg struct {
	emptyMsg
	v float64
}

func (msg FloatMsg) Type() MsgType                { return FloatMsgType }
func (msg FloatMsg) Float() float64               { return msg.v }
func (msg FloatMsg) String() string               { return fmt.Sprint(msg.v) }
func (msg FloatMsg) MarshalJSON() ([]byte, error) { return []byte(msg.String()), nil }

func NewFloatMsg(n float64) FloatMsg {
	return FloatMsg{
		emptyMsg: emptyMsg{},
		v:        n,
	}
}

// Str

type StrMsg struct {
	emptyMsg
	v string
}

func (msg StrMsg) Type() MsgType  { return StrMsgType }
func (msg StrMsg) Str() string    { return msg.v }
func (msg StrMsg) String() string { return msg.v }
func (msg StrMsg) MarshalJSON() ([]byte, error) {
	// Escape special characters in the string before marshaling
	jsonString, err := json.Marshal(msg.String())
	if err != nil {
		return nil, err
	}
	return jsonString, nil
}

func NewStrMsg(s string) StrMsg {
	return StrMsg{
		emptyMsg: emptyMsg{},
		v:        s,
	}
}

// Bool

type BoolMsg struct {
	emptyMsg
	v bool
}

func (msg BoolMsg) Type() MsgType                { return BoolMsgType }
func (msg BoolMsg) Bool() bool                   { return msg.v }
func (msg BoolMsg) String() string               { return strconv.FormatBool(msg.v) }
func (msg BoolMsg) MarshalJSON() ([]byte, error) { return []byte(msg.String()), nil }

func NewBoolMsg(b bool) BoolMsg {
	return BoolMsg{
		emptyMsg: emptyMsg{},
		v:        b,
	}
}

// List
type ListMsg struct {
	emptyMsg
	v []Msg
}

func (msg ListMsg) Type() MsgType { return ListMsgType }
func (msg ListMsg) List() []Msg   { return msg.v }
func (msg ListMsg) String() string {
	bb, err := msg.MarshalJSON()
	if err != nil {
		panic(err)
	}
	return string(bb)
}
func (msg ListMsg) MarshalJSON() ([]byte, error) { return json.Marshal(msg.v) }

func NewListMsg(v ...Msg) ListMsg {
	return ListMsg{
		emptyMsg: emptyMsg{},
		v:        v,
	}
}

// Map
type MapMsg struct {
	emptyMsg
	v map[string]Msg
}

func (msg MapMsg) Type() MsgType       { return MapMsgType }
func (msg MapMsg) Map() map[string]Msg { return msg.v }
func (msg MapMsg) MarshalJSON() ([]byte, error) {
	jsonData, err := json.Marshal(msg.v)
	if err != nil {
		panic(err)
	}

	jsonString := string(jsonData)
	jsonString = strings.ReplaceAll(jsonString, ":", ": ")
	jsonString = strings.ReplaceAll(jsonString, ",", ", ")

	return []byte(jsonString), nil
}
func (msg MapMsg) String() string {
	b, err := msg.MarshalJSON()
	if err != nil {
		panic(err)
	}
	return string(b)
}

func NewMapMsg(m map[string]Msg) MapMsg {
	return MapMsg{
		emptyMsg: emptyMsg{},
		v:        m,
	}
}
