package encoder

import (
	"testing"
)

type userInfo struct {
	Name     string            `json:"name,omitempty"`   //姓名
	Email    string            `json:"email,omitempty"`  //邮箱
	Avatar   string            `json:"avatar,omitempty"` //base64 头像
	Metadata map[string]string `json:"md,omitempty"`
}

func TestEncoder(t *testing.T) {
	conf := New([]byte("biomind1024"))

	data := &userInfo{
		Name:   "aaa",
		Email:  "aaa@bbb.com",
		Avatar: "xxxxx",
		Metadata: map[string]string{
			"a": "b",
			"c": "d",
		},
	}
	encoded, err := conf.Encode(data)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	if encoded != "HWgow1W6JzHEc9dPJmQfCGuVENAq2KWg6ftO3NBaDtHuK75rwmR89IoC83MwM1PvAUuWdH_jGlli9nXwJq6xvlTgjlseXUiJDFre0_dCXLQ" {
		t.Fatalf("%s", encoded)
	}

	var decoded *userInfo

	err = conf.Decode(encoded, &decoded)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	if decoded.Name != data.Name || decoded.Email != data.Email || decoded.Metadata["a"] != data.Metadata["a"] {
		t.Fatalf("%+v", decoded)
	}
}

func TestEncoderGzip(t *testing.T) {
	conf := New([]byte("biomind1024"), WithGzip(true))

	data := &userInfo{
		Name:   "aaa",
		Email:  "aaa@bbb.com",
		Avatar: "xxxxx",
		Metadata: map[string]string{
			"a": "b",
			"c": "d",
		},
	}
	encoded, err := conf.Encode(data)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	if encoded != "H4sIAAAAAAAC_wBQAK__HWgow1W6JzHEc9dPJmQfCGuVENAq2KWg6ftO3NBaDtHuK75rwmR89IoC83MwM1PvAUuWdH_jGlli9nXwJq6xvlTgjlseXUiJDFre0_dCXLQAAAD__wEAAP__0H_1JVAAAAA" {
		t.Fatalf("%s", encoded)
	}

	var decoded *userInfo

	err = conf.Decode(encoded, &decoded)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	if decoded.Name != data.Name || decoded.Email != data.Email || decoded.Metadata["a"] != data.Metadata["a"] {
		t.Fatalf("%+v", decoded)
	}
}
