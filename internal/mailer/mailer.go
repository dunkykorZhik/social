package mailer

import "embed"

const (
	FromName            = "GopherSocial"
	maxRetires          = 3
	UserWelcomeTemplate = "user_activation.tmpl"
)

//go:embed "templates"
var FS embed.FS

type Client interface {
	Send(templateFile, username, email string, data any) (int, error)
}
