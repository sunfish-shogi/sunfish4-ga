package ga

import "log"

type Param struct {
	Name         string
	Normal       int32
	MinimumValue int32
	MaximumValue int32
}

type Config struct {
	Params             []Param
	NumberOfIndividual int
}

func validateConfig(config Config) {
	if config.NumberOfIndividual == 0 {
		log.Fatal("NumberOfIndividual must not be zero")
	}

	if config.NumberOfIndividual%2 != 0 {
		log.Fatal("NumberOfIndividual must be even number")
	}
}
