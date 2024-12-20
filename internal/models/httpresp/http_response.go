package httpresp

type Response struct {
	Data        any
	StatusCode  int
	ContentType string
}

type OptionFunc func(*Response)

func WithStatusCode(code int) OptionFunc {
	return func(r *Response) {
		r.StatusCode = code
	}
}

func WithContentType(contentType string) OptionFunc {
	return func(r *Response) {
		r.ContentType = contentType
	}
}
