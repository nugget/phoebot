package chat

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/Tnze/go-mc/data"
	pk "github.com/Tnze/go-mc/net/packet"
)

//Message is a message sent by other
type Message jsonChat

type jsonChat struct {
	Text string `json:"text"`

	Bold          bool   `json:"bold,omitempty"`          //粗体
	Italic        bool   `json:"Italic,omitempty"`        //斜体
	UnderLined    bool   `json:"underlined,omitempty"`    //下划线
	StrikeThrough bool   `json:"strikethrough,omitempty"` //删除线
	Obfuscated    bool   `json:"obfuscated,omitempty"`    //随机
	Color         string `json:"color,omitempty"`

	Translate string            `json:"translate,omitempty"`
	With      []json.RawMessage `json:"with,omitempty"` // How can go handle an JSON array with Object and String?
	Extra     []jsonChat        `json:"extra,omitempty"`
}

//UnmarshalJSON decode json to Message
func (m *Message) UnmarshalJSON(jsonMsg []byte) (err error) {
	if jsonMsg[0] == '"' {
		err = json.Unmarshal(jsonMsg, &m.Text) //Unmarshal as jsonString
	} else {
		err = json.Unmarshal(jsonMsg, (*jsonChat)(m)) //Unmarshal as jsonChat
	}
	return
}

//Decode a ChatMsg packet
func (m *Message) Decode(r pk.DecodeReader) error {
	var Len pk.VarInt
	if err := Len.Decode(r); err != nil {
		return err
	}

	return json.NewDecoder(io.LimitReader(r, int64(Len))).Decode(m)
}

var colors = map[string]int{
	"black":        30,
	"dark_blue":    34,
	"dark_green":   32,
	"dark_aqua":    36,
	"dark_red":     31,
	"dark_purple":  35,
	"gold":         33,
	"gray":         37,
	"dark_gray":    90,
	"blue":         94,
	"green":        92,
	"aqua":         96,
	"red":          91,
	"light_purple": 95,
	"yellow":       93,
	"white":        97,
}

// String return the message with escape sequence for ansi color.
// On windows, you may want print this string using
// github.com/mattn/go-colorable.
func (m Message) String() string {
	var msg, format strings.Builder
	if m.Bold {
		format.WriteString("1;")
	}
	if m.Italic {
		format.WriteString("3;")
	}
	if m.UnderLined {
		format.WriteString("4;")
	}
	if m.StrikeThrough {
		format.WriteString("9;")
	}
	if m.Color != "" {
		fmt.Fprintf(&format, "%d;", colors[m.Color])
	}
	if format.Len() > 0 {
		msg.WriteString("\033[" + format.String()[:format.Len()-1] + "m")
	}
	msg.WriteString(m.Text)

	//handle translate
	if m.Translate != "" {
		args := make([]interface{}, len(m.With))
		for i, v := range m.With {
			var arg Message
			arg.UnmarshalJSON(v) //ignore error
			args[i] = arg
		}

		fmt.Fprintf(&msg, data.EnUs[m.Translate], args...)
	}

	if m.Extra != nil {
		for i := range m.Extra {
			msg.WriteString(Message(m.Extra[i]).String())
		}
	}

	if format.Len() > 0 {
		msg.WriteString("\033[0m")
	}
	return msg.String()
}
