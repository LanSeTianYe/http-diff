package task

type Payload struct {
	Params  string `json:"params"`
	Headers string `json:"headers"`
	Body    string `json:"body"`
}
