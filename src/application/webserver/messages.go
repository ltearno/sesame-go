package webserver

type PageContext struct {
	Name string
}

type DocumentContext struct {
	Document struct {
		Name string
	}
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
