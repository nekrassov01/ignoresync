package env

import (
	"time"

	_ "embed"
)

// stackMaxWaitDur is the maximum wait time for stack creation.
const stackMaxWaitDur time.Duration = time.Minute * 15

//go:embed template.yaml
var template string
