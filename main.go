package main

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/sunfish-shogi/sunfish4-ga/ga"
)

func main() {
	f, err := os.OpenFile("ga.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.SetOutput(io.MultiWriter(os.Stdout, f))

	params := []ga.Param{
		{
			Name:         "EXT_SINGULAR",
			MinimumValue: 0,
			MaximumValue: 4,
			Step:         1,
		},
		{
			Name:         "EXT_DEPTH_CHECK",
			MinimumValue: 0,
			MaximumValue: 4,
			Step:         1,
		},
		{
			Name:         "EXT_DEPTH_ONE_REPLY",
			MinimumValue: 0,
			MaximumValue: 4,
			Step:         1,
		},
		{
			Name:         "EXT_DEPTH_RECAP",
			MinimumValue: 0,
			MaximumValue: 4,
			Step:         1,
		},
		{
			Name:         "NULL_DEPTH_RATE",
			MinimumValue: 4,
			MaximumValue: 16,
			Step:         1,
		},
		{
			Name:         "NULL_DEPTH_REDUCE",
			MinimumValue: 0,
			MaximumValue: 20,
			Step:         1,
		},
		{
			Name:         "NULL_DEPTH_VRATE",
			MinimumValue: 10,
			MaximumValue: 800,
			Step:         10,
		},
		{
			Name:         "REDUCTION_RATE1",
			MinimumValue: 5,
			MaximumValue: 30,
			Step:         5,
		},
		{
			Name:         "REDUCTION_RATE2",
			MinimumValue: 5,
			MaximumValue: 30,
			Step:         5,
		},
		{
			Name:         "RAZOR_MARGIN1",
			MinimumValue: 10,
			MaximumValue: 800,
			Step:         10,
		},
		{
			Name:         "RAZOR_MARGIN2",
			MinimumValue: 10,
			MaximumValue: 800,
			Step:         10,
		},
		{
			Name:         "RAZOR_MARGIN3",
			MinimumValue: 10,
			MaximumValue: 800,
			Step:         10,
		},
		{
			Name:         "RAZOR_MARGIN4",
			MinimumValue: 10,
			MaximumValue: 800,
			Step:         10,
		},
		{
			Name:         "FUT_PRUN_MAX_DEPTH",
			MinimumValue: 4,
			MaximumValue: 64,
			Step:         4,
		},
		{
			Name:         "FUT_PRUN_MARGIN_RATE",
			MinimumValue: 10,
			MaximumValue: 200,
			Step:         10,
		},
		{
			Name:         "FUT_PRUN_MARGIN",
			MinimumValue: 50,
			MaximumValue: 800,
			Step:         10,
		},
		{
			Name:         "PROBCUT_MARGIN",
			MinimumValue: 50,
			MaximumValue: 500,
			Step:         10,
		},
		{
			Name:         "PROBCUT_REDUCTION",
			MinimumValue: 1,
			MaximumValue: 10,
			Step:         1,
		},
		{
			Name:         "ASP_MIN_DEPTH",
			MinimumValue: 2,
			MaximumValue: 10,
			Step:         1,
		},
		{
			Name:         "ASP_1ST_DELTA",
			MinimumValue: 32,
			MaximumValue: 256,
			Step:         8,
		},
		{
			Name:         "ASP_DELTA_RATE",
			MinimumValue: 25,
			MaximumValue: 200,
			Step:         5,
		},
		{
			Name:         "SINGULAR_DEPTH",
			MinimumValue: 4,
			MaximumValue: 12,
			Step:         1,
		},
		{
			Name:         "SINGULAR_MARGIN",
			MinimumValue: 1,
			MaximumValue: 32,
			Step:         1,
		},
	}
	config := ga.Config{
		Params:             params,
		NumberOfIndividual: 33,
		Duration:           time.Minute * 10,
	}

	m := ga.NewGAManager(config)
	err = m.Run()
	if err != nil {
		log.Fatal(err)
	}
}
