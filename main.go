package main

import (
	"io"
	"log"
	"os"

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
			Name:            "EXT_DEPTH_CHECK",
			FirstEliteValue: 3,
			MinimumValue:    0,
			MaximumValue:    8,
		},
		{
			Name:            "EXT_DEPTH_ONE_REPLY",
			FirstEliteValue: 2,
			MinimumValue:    0,
			MaximumValue:    8,
		},
		{
			Name:            "EXT_DEPTH_RECAP",
			FirstEliteValue: 1,
			MinimumValue:    0,
			MaximumValue:    8,
		},
		{
			Name:            "NULL_DEPTH_RATE",
			FirstEliteValue: 11,
			MinimumValue:    4,
			MaximumValue:    16,
		},
		{
			Name:            "NULL_DEPTH_REDUCE",
			FirstEliteValue: 12,
			MinimumValue:    0,
			MaximumValue:    20,
		},
		{
			Name:            "NULL_DEPTH_VRATE",
			FirstEliteValue: 150,
			MinimumValue:    10,
			MaximumValue:    800,
		},
		{
			Name:            "REDUCTION_RATE1",
			FirstEliteValue: 10,
			MinimumValue:    5,
			MaximumValue:    30,
		},
		{
			Name:            "REDUCTION_RATE2",
			FirstEliteValue: 10,
			MinimumValue:    5,
			MaximumValue:    30,
		},
		{
			Name:            "RAZOR_MARGIN1",
			FirstEliteValue: 300,
			MinimumValue:    10,
			MaximumValue:    800,
		},
		{
			Name:            "RAZOR_MARGIN2",
			FirstEliteValue: 400,
			MinimumValue:    10,
			MaximumValue:    800,
		},
		{
			Name:            "RAZOR_MARGIN3",
			FirstEliteValue: 400,
			MinimumValue:    10,
			MaximumValue:    800,
		},
		{
			Name:            "RAZOR_MARGIN4",
			FirstEliteValue: 450,
			MinimumValue:    10,
			MaximumValue:    800,
		},
		{
			Name:            "FUT_PRUN_MAX_DEPTH",
			FirstEliteValue: 28,
			MinimumValue:    4,
			MaximumValue:    64,
		},
		{
			Name:            "FUT_PRUN_MARGIN_RATE",
			FirstEliteValue: 75,
			MinimumValue:    10,
			MaximumValue:    200,
		},
		{
			Name:            "FUT_PRUN_MARGIN",
			FirstEliteValue: 500,
			MinimumValue:    50,
			MaximumValue:    800,
		},
		{
			Name:            "PROBCUT_MARGIN",
			FirstEliteValue: 200,
			MinimumValue:    50,
			MaximumValue:    500,
		},
	}
	config := ga.Config{
		Params:             params,
		NumberOfIndividual: 7,
	}

	m := ga.NewGAManager(config)
	err = m.Run()
	if err != nil {
		log.Fatal(err)
	}
}
