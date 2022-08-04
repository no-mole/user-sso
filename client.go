package user_sso

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/no-mole/neptune/json"
	"golang.org/x/oauth2"
)

type Client struct {
	userInfoEndpoint string
	logoutEndpoint   string
	accessEndpoint   string
	oauth2           *oauth2.Config
}

type Config struct {
	Endpoint     string   `json:"endpoint" form:"endpoint" yaml:"endpoint" toml:"endpoint"`
	ClientId     string   `json:"client_id" form:"client_id" yaml:"client_id" toml:"client_id"`
	ClientSecret string   `json:"client_secret" form:"client_secret" yaml:"client_secret" toml:"client_secret"`
	RedirectUrl  string   `json:"redirect_url" form:"redirect_url" yaml:"redirect_url" toml:"redirect_url"`
	Scopes       []string `json:"scopes" form:"scopes" yaml:"scopes" toml:"scopes"`
}

func NewClient(conf *Config) *Client {
	endpoint := strings.TrimRight(conf.Endpoint, "/?")
	cli := &Client{
		oauth2: &oauth2.Config{
			ClientID:     conf.ClientId,
			ClientSecret: conf.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:   fmt.Sprintf("%s/oauth/authorize", endpoint),
				TokenURL:  fmt.Sprintf("%s/oauth/token", endpoint),
				AuthStyle: oauth2.AuthStyleInParams,
			},
			RedirectURL: conf.RedirectUrl,
			Scopes:      conf.Scopes,
		},
		userInfoEndpoint: fmt.Sprintf("%s/oauth/user", endpoint),
		logoutEndpoint:   fmt.Sprintf("%s/user/logout", endpoint),
		accessEndpoint:   fmt.Sprintf("%s/user/access", endpoint),
	}
	return cli
}

type UserInfo struct {
	Name     string            `json:"name,omitempty"`   //姓名
	Email    string            `json:"email,omitempty"`  //邮箱
	Avatar   string            `json:"avatar,omitempty"` //base64 头像
	Token    *oauth2.Token     `json:"encoder"`
	Metadata map[string]string `json:"md,omitempty"`
}

// Get get metadata
func (u *UserInfo) Get(key string) string {
	if u.Metadata == nil {
		u.Metadata = map[string]string{}
	}
	return u.Metadata[key]
}

// Set set metadata
func (u *UserInfo) Set(key, val string) {
	if u.Metadata == nil {
		u.Metadata = map[string]string{}
	}
	u.Metadata[key] = val
}

var FetchUserInfoErr = errors.New("fetch user info error")

// PasswordCredentials use username & password login
func (c *Client) PasswordCredentials(ctx context.Context, username, password string) (*UserInfo, error) {
	oauthToken, err := c.oauth2.PasswordCredentialsToken(ctx, username, password)
	if err != nil {
		return nil, err
	}
	return c.fetchUser(ctx, oauthToken)
}

// AuthUrl use auth code mode
func (c *Client) AuthUrl(state string, opts ...oauth2.AuthCodeOption) string {
	return c.oauth2.AuthCodeURL(state, opts...)
}

// LogoutUrl linkage user service logout
func (c *Client) LogoutUrl(returnUrl string) string {
	if returnUrl == "" {
		return c.logoutEndpoint
	}
	return fmt.Sprintf("%s?return_url=%s", c.logoutEndpoint, url.QueryEscape(returnUrl))
}

// Exchange fetch user info
func (c *Client) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*UserInfo, error) {
	oauthToken, err := c.oauth2.Exchange(ctx, code, opts...)
	if err != nil {
		return nil, err
	}
	return c.fetchUser(ctx, oauthToken)
}

// RefreshToken refresh token
func (c *Client) RefreshToken(ctx context.Context, userInfo *UserInfo) error {
	newToken, err := c.oauth2.TokenSource(ctx, userInfo.Token).Token()
	if err != nil {
		return err
	}
	userInfo.Token = newToken
	return nil
}

// ExpirationSoon  judge how long there is more to ask for token
func (c *Client) ExpirationSoon(_ context.Context, userInfo *UserInfo, failureInterval float64) bool {
	return time.Now().Sub(userInfo.Token.Expiry).Seconds() < failureInterval
}

func (c *Client) fetchUser(_ context.Context, token *oauth2.Token) (*UserInfo, error) {
	req, _ := http.NewRequest("GET", c.userInfoEndpoint, nil)
	req.Header.Add("Authorization", token.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, FetchUserInfoErr
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var info *UserInfo
	err = json.Unmarshal(body, &info)
	if err != nil {
		return nil, FetchUserInfoErr
	}
	info.Token = token
	return info, nil
}

// AccessItems 批量鉴权
type AccessItems struct {
	Items []*AccessItem `json:"items"`
}

type AccessItem struct {
	Resource string `protobuf:"bytes,1,opt,name=resource,proto3" json:"resource,omitempty"`
	Action   string `protobuf:"bytes,2,opt,name=action,proto3" json:"action,omitempty"`
	Ok       bool   `protobuf:"varint,3,opt,name=ok,proto3" json:"ok,omitempty"`
}

type AccessResponse struct {
	Code int64         `json:"code"`
	Msg  string        `json:"msg"`
	Data []*AccessItem `json:"data"`
}

var ErrorNilAccessItems = errors.New("no access items")
var ErrorRequestFailed = errors.New("access request failed")

// Authentication batch authentication
func (c *Client) Authentication(_ context.Context, accessToken string, accessItems []*AccessItem) ([]*AccessItem, error) {
	if accessItems == nil || len(accessItems) == 0 {
		return nil, ErrorNilAccessItems
	}
	data, err := json.Marshal(&AccessItems{Items: accessItems})
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(data)
	req, _ := http.NewRequest("POST", c.accessEndpoint, reader)
	req.Header.Add("Authorization", accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, ErrorRequestFailed
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	accessResp := &AccessResponse{}
	err = json.Unmarshal(body, accessResp)
	if err != nil {
		return nil, err
	}
	if accessResp.Code != 200 {
		return nil, fmt.Errorf("access failed:status=%d,message=%s", accessResp.Code, accessResp.Msg)
	}
	return accessResp.Data, err
}
