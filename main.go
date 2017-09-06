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
			Name:         "EXT_DEPTH_CHECK",
			Normal:       3,
			MinimumValue: 0,
			MaximumValue: 8,
		},
		{
			Name:         "EXT_DEPTH_ONE_REPLY",
			Normal:       2,
			MinimumValue: 0,
			MaximumValue: 8,
		},
		{
			Name:         "EXT_DEPTH_RECAP",
			Normal:       1,
			MinimumValue: 0,
			MaximumValue: 8,
		},
		{
			Name:         "NULL_DEPTH_RATE",
			Normal:       11,
			MinimumValue: 4,
			MaximumValue: 16,
		},
		{
			Name:         "NULL_DEPTH_REDUCE",
			Normal:       12,
			MinimumValue: 0,
			MaximumValue: 20,
		},
		{
			Name:         "NULL_DEPTH_VRATE",
			Normal:       150,
			MinimumValue: 10,
			MaximumValue: 800,
		},
		{
			Name:         "REDUCTION_RATE1",
			Normal:       10,
			MinimumValue: 5,
			MaximumValue: 30,
		},
		{
			Name:         "REDUCTION_RATE2",
			Normal:       10,
			MinimumValue: 5,
			MaximumValue: 30,
		},
		{
			Name:         "RAZOR_MARGIN1",
			Normal:       300,
			MinimumValue: 10,
			MaximumValue: 800,
		},
		{
			Name:         "RAZOR_MARGIN2",
			Normal:       400,
			MinimumValue: 10,
			MaximumValue: 800,
		},
		{
			Name:         "RAZOR_MARGIN3",
			Normal:       400,
			MinimumValue: 10,
			MaximumValue: 800,
		},
		{
			Name:         "RAZOR_MARGIN4",
			Normal:       450,
			MinimumValue: 10,
			MaximumValue: 800,
		},
		{
			Name:         "FUT_PRUN_MAX_DEPTH",
			Normal:       28,
			MinimumValue: 4,
			MaximumValue: 64,
		},
		{
			Name:         "FUT_PRUN_MARGIN_RATE",
			Normal:       75,
			MinimumValue: 10,
			MaximumValue: 200,
		},
		{
			Name:         "FUT_PRUN_MARGIN",
			Normal:       500,
			MinimumValue: 50,
			MaximumValue: 800,
		},
		{
			Name:         "PROBCUT_MARGIN",
			Normal:       200,
			MinimumValue: 50,
			MaximumValue: 500,
		},
		{
			Name:         "PROBCUT_REDUCTION",
			Normal:       4,
			MinimumValue: 1,
			MaximumValue: 10,
		},
		{
			Name:         "ASP_MIN_DEPTH",
			Normal:       6,
			MinimumValue: 2,
			MaximumValue: 10,
		},
		{
			Name:         "ASP_1ST_DELTA",
			Normal:       128,
			MinimumValue: 32,
			MaximumValue: 256,
		},
		{
			Name:         "ASP_DELTA_RATE",
			Normal:       50,
			MinimumValue: 25,
			MaximumValue: 200,
		},
		{
			Name:         "SINGULAR_DEPTH",
			Normal:       8,
			MinimumValue: 4,
			MaximumValue: 12,
		},
		{
			Name:         "SINGULAR_MARGIN",
			Normal:       3,
			MinimumValue: 1,
			MaximumValue: 32,
		},
	}
	config := ga.Config{
		Params:             params,
		NumberOfIndividual: 32,
		Duration:           time.Hour * 4,
	}

	m := ga.NewGAManager(config)
	err = m.Run()
	if err != nil {
		log.Fatal(err)
	}
}
