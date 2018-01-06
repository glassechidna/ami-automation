package shared

type OutputFormat struct {
	Outputs map[string][]*string
	AmiId string
	AmiIds map[string]string
	WaitCommand string `json:",omitempty"`
}
