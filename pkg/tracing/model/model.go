package model

type TracingType string

const (
	ConstType         TracingType = "const"
	ProbabilisticType TracingType = "probabilistic"
	RatelimitingType  TracingType = "ratelimiting"
	RemoteType        TracingType = "remote"

	ConstSendAllTracers = 1
	ConstSendNoTracers  = 0
	// ratelimitingParam = 10.
	// probabilisticParam = 0.1.
)
