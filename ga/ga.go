package ga

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	server "github.com/sunfish-shogi/sunfish4-ga/shogiserver"
)

type GAManager struct {
	Config Config

	server   server.ShogiServer
	inds     []*individual
	normInds []*individual
	scores   []map[int32]scoreType
}

type scoreType struct {
	win  float64
	loss float64
}

func NewGAManager(config Config) *GAManager {
	validateConfig(config)
	scores := make([]map[int32]scoreType, len(config.Params))
	for i := range scores {
		scores[i] = make(map[int32]scoreType)
		param := config.Params[i]
		for val := param.MinimumValue; val <= param.MaximumValue; val += param.Step {
			scores[i][val] = scoreType{}
		}
	}
	return &GAManager{
		Config: config,
		scores: scores,
	}
}

func (ga *GAManager) Run() error {
	defer ga.Destroy()

	err := ga.Start()
	if err != nil {
		return err
	}

	for gn := 1; ; gn++ {
		ga.PrintGeneration(gn)

		time.Sleep(ga.Config.Duration)

		err = ga.Next()
		if err != nil {
			log.Println(err)
		}
	}
}

func (ga *GAManager) Start() error {
	err := ga.server.Setup()
	if err != nil {
		return err
	}

	values := generateNormalValues(ga.Config)
	ga.normInds = make([]*individual, 0, ga.Config.Concurrency)
	for i := 0; i < ga.Config.Concurrency; i++ {
		ind := newIndividual(fmt.Sprintf("n-%d", i), i, ga.Config, values)
		ga.normInds = append(ga.normInds, ind)
	}

	values = generateRandomValues(ga.Config)
	ga.inds = make([]*individual, 0, ga.Config.Concurrency)
	for i := 0; i < ga.Config.Concurrency; i++ {
		ind := newIndividual(fmt.Sprintf("i-%d", i), i, ga.Config, values)
		ga.inds = append(ga.inds, ind)
	}

	err = startIndividuals(append(ga.inds, ga.normInds...))
	return err
}

func (ga *GAManager) Next() error {
	log.Println("Scores")
	for _, ind := range ga.inds {
		if err := ind.UpdateScore(); err != nil {
			log.Println(err)
		}
		log.Printf("%s %d - %d\n", ind.id, ind.win, ind.loss)
	}
	log.Println()

	// Update Scores
	ga.UpdateScores()

	// Best Values
	log.Println("Best Values")
	bestValues := ga.CalculateBestValues()
	log.Println(stringifyValues(bestValues))
	log.Println()

	// New Generation
	inds := make([]*individual, 0, ga.Config.Concurrency)
	for i := 0; i < ga.Config.Concurrency; i++ {
		ind := newIndividual(fmt.Sprintf("i-%d", i), i, ga.Config, bestValues)
		inds = append(inds, ind)
	}

	// Stop Previous Generation
	stopIndividuals(append(ga.inds, ga.normInds...))

	// Replace to New Generation
	ga.inds = inds

	// Start Next Generation
	startIndividuals(append(ga.inds, ga.normInds...))

	return nil
}

func (ga *GAManager) PrintGeneration(gn int) {
	log.Printf("Generation: %d\n", gn)
	for i := range ga.inds {
		log.Printf("%s %s\n", ga.inds[i].id, stringifyValues(ga.inds[i].values))
	}
	log.Println()
}

func (ga *GAManager) Destroy() {
	for i := range ga.inds {
		ga.inds[i].stopWithClean()
	}
	ga.server.Stop()
}

func (ga *GAManager) UpdateScores() {
	for i := range ga.Config.Params {
		for _, ind := range ga.inds {
			value := ind.values[i]
			score := ga.scores[i][value]

			score.win += float64(ind.win)
			score.loss += float64(ind.loss)

			ga.scores[i][value] = score
		}
	}
}

func (ga *GAManager) CalculateBestValues() []int32 {
	values := make([]int32, len(ga.Config.Params))
	var maxValues string

	for i := range ga.Config.Params {
		maxX := float64(-1.0)
		for value, score := range ga.scores[i] {
			sum := score.win + score.loss
			var x float64
			if sum >= 1 {
				x = score.win / sum
			} else {
				x = rand.Float64() // [0.0,1.0)
			}
			if x > maxX {
				values[i] = value
				maxX = x
			}
		}
		maxValues = maxValues + " " + strconv.FormatFloat(maxX, 'f', 2, 64)

		const epsilon = 0.1
		if rand.Float64() < epsilon {
			vs := make([]int32, len(ga.scores[i]))
			var i int
			for value := range ga.scores[i] {
				vs[i] = value
				i++
			}
			values[i] = vs[rand.Int31n(int32(len(vs)))]
			maxValues = maxValues + "(e)"
		}
	}
	log.Println("ucb1: ", maxValues)

	return values
}
