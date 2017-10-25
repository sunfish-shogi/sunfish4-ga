package ga

import (
	"log"
	"math/rand"
	"time"
)

type Param struct {
	Name         string
	MinimumValue int32
	MaximumValue int32
	Step         int32
}

type Config struct {
	Params             []Param
	NumberOfIndividual int
	Duration           time.Duration
}

func validateConfig(config Config) {
	if config.NumberOfIndividual == 0 {
		log.Fatal("NumberOfIndividual must not be zero")
	}

	if config.NumberOfIndividual%2 == 0 {
		log.Fatal("NumberOfIndividual must not be even number")
	}
}

func generateRandomValues(config Config) []int32 {
	values := make([]int32, len(config.Params))
	for i := range config.Params {
		min := config.Params[i].MinimumValue
		max := config.Params[i].MaximumValue
		step := config.Params[i].Step
		values[i] = min + rand.Int31n((max-min+1)/step)*step
	}
	return values
}
