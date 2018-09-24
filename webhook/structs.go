package webhook

type RegistyHook struct {
	Events []Event `json:"events"`
}

type Actor struct {
	Name string `json:"name"`
}

type Request struct {
	Addr      string `json:"addr"`
	Host      string `json:"host"`
	ID        string `json:"id"`
	Method    string `json:"method"`
	Useragent string `json:"useragent"`
}

type Source struct {
	Addr       string `json:"addr"`
	InstanceID string `json:"instanceID"`
}

type Target struct {
	Digest     string `json:"digest"`
	Length     int    `json:"length"`
	MediaType  string `json:"mediaType"`
	Repository string `json:"repository"`
	Size       int    `json:"size"`
	Tag        string `json:"tag"`
	URL        string `json:"url"`
}

type Event struct {
	Action    string  `json:"action"`
	Actor     Actor   `json:"actor"`
	ID        string  `json:"id"`
	Request   Request `json:"request"`
	Source    Source  `json:"source"`
	Target    Target  `json:"target"`
	Timestamp string  `json:"timestamp"`
}
