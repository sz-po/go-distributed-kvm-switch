package process

type ExecutionSpecification struct {
	ExecutablePath       string            `json:"executablePath"`
	Arguments            []string          `json:"arguments"`
	WorkingDir           string            `json:"workingDir"`
	EnvironmentVariables map[string]string `json:"environmentVariables"`
}

type Specification struct {
	Execution ExecutionSpecification `json:"execution"`
}

type Status struct {
	IsRunning bool   `json:"isRunning"`
	ProcessID int    `json:"processId"`
	ExitCode  int    `json:"exitCode"`
	Error     string `json:"error"`
}
