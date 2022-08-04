package encoder

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/no-mole/neptune/crypto/aes"
)

func New(key []byte, opts ...Option) *Config {
	conf := &Config{
		Gzip: false,
		salt: key,
	}
	for _, opt := range opts {
		opt(conf)
	}
	return conf
}

type Option func(config *Config)

func WithGzip(gzip bool) Option {
	return func(config *Config) {
		config.Gzip = gzip
	}
}

type Config struct {
	Gzip bool `json:"gzip"`
	salt []byte
}

func (conf *Config) Encode(t interface{}) (string, error) {
	if t == nil {
		return "", errors.New("encoder is nil")
	}
	data, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	encrypted := aes.Encrypt(data, conf.salt)
	if len(encrypted) == 0 {
		return "", errors.New("encode fail")
	}
	if conf.Gzip {
		var res bytes.Buffer
		//gz := gzip.NewWriter(&res)
		gz, _ := gzip.NewWriterLevel(&res, gzip.BestCompression)
		_, _ = gz.Write(encrypted)
		_ = gz.Flush()
		_ = gz.Close()
		return base64.RawURLEncoding.EncodeToString(res.Bytes()), nil
	}
	return base64.RawURLEncoding.EncodeToString(encrypted), err
}

func (conf *Config) Decode(str string, dst interface{}) error {
	data, err := base64.RawURLEncoding.DecodeString(str)
	if err != nil {
		return err
	}
	if conf.Gzip {
		rdata := bytes.NewReader(data)
		r, err := gzip.NewReader(rdata)
		if err != nil {
			return err
		}
		data, err = ioutil.ReadAll(r)
		if err != nil {
			return err
		}
	}
	body := aes.Decrypt(data, conf.salt)
	return json.Unmarshal(body, &dst)
}
