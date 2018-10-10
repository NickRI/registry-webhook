package webhook

//RegistyHook top of incomming data
type RegistyHook struct {
	Events []Event `json:"events"`
}

//Actor holds actors name
type Actor struct {
	Name string `json:"name"`
}

//Request contains the data about request
type Request struct {
	Addr      string `json:"addr"`
	Host      string `json:"host"`
	ID        string `json:"id"`
	Method    string `json:"method"`
	Useragent string `json:"useragent"`
}

//Source holds data about sours
type Source struct {
	Addr       string `json:"addr"`
	InstanceID string `json:"instanceID"`
}

//Target holds target data
type Target struct {
	Digest     string `json:"digest"`
	Length     int    `json:"length"`
	MediaType  string `json:"mediaType"`
	Repository string `json:"repository"`
	Size       int    `json:"size"`
	Tag        string `json:"tag"`
	URL        string `json:"url"`
}

//Event represents event data
type Event struct {
	Action    string  `json:"action"`
	Actor     Actor   `json:"actor"`
	ID        string  `json:"id"`
	Request   Request `json:"request"`
	Source    Source  `json:"source"`
	Target    Target  `json:"target"`
	Timestamp string  `json:"timestamp"`
}
