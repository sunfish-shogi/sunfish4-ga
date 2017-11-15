package ga

import (
	"log"
	"math/rand"
	"time"
)

type Param struct {
	Name         string
	Normal       int32
	MinimumValue int32
	MaximumValue int32
	Step         int32
}

type Config struct {
	Params      []Param
	Concurrency int
	Duration    time.Duration
}

func validateConfig(config Config) {
	if config.Concurrency == 0 {
		log.Fatal("NumberOfIndividual must not be zero")
	}
}

func generateNormalValues(config Config) []int32 {
	values := make([]int32, len(config.Params))
	for i := range config.Params {
		values[i] = config.Params[i].Normal
	}
	return values
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
