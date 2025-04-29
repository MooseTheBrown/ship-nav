package ship

type IPCRequest struct {
	Type string `json:"type"`
	Cmd  string `json:"cmd"`
	Data string `json:"data"`
}

type IPCCommandResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type IPCQueryResponse struct {
	Speed    string `json:"speed"`
	Steering string `json:"steering"`
}
