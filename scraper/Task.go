package scraper

/* task definition */
type Task struct {
	URL string
	Depth int
	Callback func(Result)
}

type Result struct {
	Title string
	Description string 
	Links []string
	Metadata map[string]string 
	StatusCode int
	RawHTML []byte
	Error error
}