package dos

type Function struct {
	FunctionName string `json:"function_name"`
	IsValid      bool   `json:"is_valid"`
	Reason       string `json:"reason"`
	Suggestion   string `json:"suggestion"`
}

type LLMResponseList []*Function
