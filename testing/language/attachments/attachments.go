package attachments

// Attachment represents a file or data attached to a test
type Attachment struct {
	Name      string
	MediaType string
	Data      []byte
}

// Provider is an interface for accessing attachments from the test world
type Provider interface {
	GetAttachments() []Attachment
	ClearAttachments()
}
