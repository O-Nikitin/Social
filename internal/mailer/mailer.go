package mailer

import "embed"

const (
	FromName            = "Social"
	maxSendingRetries   = 3 //Total time should not be longer than req ctx timeout!!!
	UserWelcomeTemplate = "user_invitation.tmpl"
)

//go:embed "template"
var FS embed.FS

//go:generate mockgen -source=./mailer.go -destination=../../cmd/api/mock/mailer/Mock_Mailer.go -package=mock_mailer Client

type Client interface {
	Send(templateFile, username, email string,
		data any, isSandbox bool) (int, error)
}
