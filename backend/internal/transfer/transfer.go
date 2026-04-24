package transfer

type TransferMode string

const (
	Local     TransferMode = "local"
	WebDAV    TransferMode = "dav"
	WebDAVSec TransferMode = "davs"
)

type TransferDir struct {
	TransferMode TransferMode
	Host         *string
	Path         string
}
