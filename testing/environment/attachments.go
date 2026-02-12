package environment

// Attachment represents a file or data attached to a test
type Attachment struct {
	Name      string
	MediaType string
	Data      []byte
}

// AttachmentProvider is an interface for accessing attachments from the test world
type AttachmentProvider interface {
	GetAttachments() []Attachment
	ClearAttachments()
}
