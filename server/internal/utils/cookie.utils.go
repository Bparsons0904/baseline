package utils

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type Cookie struct {
	Name    string
	Value   string
	Expires time.Time
}

func ApplyCookie(c *fiber.Ctx, cookie Cookie) {
	c.Cookie(&fiber.Cookie{
		Name:     cookie.Name,
		Value:    cookie.Value,
		Expires:  cookie.Expires,
		HTTPOnly: true,
	})
}

func ExpireCookie(c *fiber.Ctx, key string) {
	ApplyCookie(c, Cookie{
		Name:    key,
		Value:   "",
		Expires: time.Now().Add(1 * time.Second),
	})
}
