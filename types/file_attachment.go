package types

// A file attachment is represented as a JSON object.
// The file attachment may point to a document previously uploaded by the
// uploads/add API call, or by any external resource.
//
// TODO see https://developer.todoist.com/sync/v9/#file-attachments, it looks
// like there are some extra properties for images and audio files
type FileAttachment struct {
	// The name of the file
	FileName string `json:"file_name"`
	// The size of the file in bytes
	FileSize int `json:"file_size"`
	// MIME type
	FileType string `json:"file_type"`
	// The URL where the file is located. Todoist doesn't cache the remote content
	// on their servers, nor does it stream or expose files directly from
	// third-party resources. Avoid providing links to non HTTPS resources, as
	// exposing them in todoist may issue a browser warning
	FileUrl string `json:"file_url"`
	// Upload completion state (e.g. pending, completed)
	UploadState string `json:"upload_state"`
}
