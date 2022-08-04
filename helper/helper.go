package helper

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	userSso "github.com/no-mole/user-sso"
)

type Helper struct {
	HeaderKey  string `json:"header_key"`
	CookieKey  string `json:"cookie_key"`
	ContextKey string `json:"context_key"`
}

func New(headerKey, cookieKey, contextKey string) *Helper {
	return &Helper{
		HeaderKey:  headerKey,
		CookieKey:  cookieKey,
		ContextKey: contextKey,
	}
}

func (h *Helper) FromGin(ctx *gin.Context) string {
	str := h.FromHeader(ctx)
	if str == "" {
		return h.FromCookie(ctx)
	}
	return str
}

func (h *Helper) FromHeader(ctx *gin.Context) string {
	return ctx.GetHeader(h.HeaderKey)
}

func (h *Helper) WithHeader(ctx *gin.Context, value string) {
	ctx.Set(h.HeaderKey, value)
}

func (h *Helper) FromCookie(ctx *gin.Context) string {
	str, _ := ctx.Cookie(h.CookieKey)
	return str
}

func (h *Helper) WithCookie(ctx *gin.Context, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	ctx.SetCookie(h.CookieKey, value, maxAge, path, domain, secure, httpOnly)
}

var ErrorNotLogin = errors.New("not login")

func (h *Helper) FromContext(ctx context.Context) (*userSso.UserInfo, error) {
	val := ctx.Value(h.ContextKey)
	if u, ok := val.(*userSso.UserInfo); ok {
		return u, nil
	}
	return nil, ErrorNotLogin
}

func (h *Helper) WithContext(ctx context.Context, u *userSso.UserInfo) context.Context {
	return context.WithValue(ctx, h.ContextKey, u)
}

func (h *Helper) WithGin(ctx *gin.Context, u *userSso.UserInfo) {
	ctx.Set(h.ContextKey, u)
}
