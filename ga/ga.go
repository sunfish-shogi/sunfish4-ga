package ga

import (
	"log"
	"math"
	"math/rand"
	"time"

	server "github.com/sunfish-shogi/sunfish4-ga/shogiserver"
)

type GAManager struct {
	Config Config

	server   server.ShogiServer
	indMap   map[string]*individual
	allInds  []*individual
	currInds []*individual
}

func NewGAManager(config Config) *GAManager {
	validateConfig(config)
	return &GAManager{
		Config:  config,
		indMap:  make(map[string]*individual),
		allInds: make([]*individual, 0, 1024),
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

	ga.currInds = make([]*individual, 0, ga.Config.NumberOfIndividual)
	for len(ga.currInds) < ga.Config.NumberOfIndividual {
		ind, ok := ga.newIndividual(generateRandomValues(ga.Config), "")
		if ok {
			ga.currInds = append(ga.currInds, ind)
		}
	}

	err = startIndividuals(ga.currInds)
	return err
}

func (ga *GAManager) Next() error {
	rate, err := ga.server.MakeRate()
	if err != nil {
		return err
	}

	for _, ind := range ga.indMap {
		ind.score = 0
	}
	for pi := range rate.Players {
		for _, player := range rate.Players[pi] {
			if ind, ok := ga.indMap[player.Name]; ok {
				ind.score = player.Rate
				ind.win = player.Win
				ind.loss = player.Loss
			}
		}
	}

	// Print Scores
	log.Println("Score")
	for i := range ga.currInds {
		log.Printf("%s %0.3f", ga.currInds[i].id, ga.currInds[i].score)
	}
	log.Println()

	// Best Values
	bestValues := ga.CalculateBestValues()
	log.Println("Best Values")
	log.Println(stringifyValues(bestValues))
	log.Println()

	// New Generation
	inds := make([]*individual, 0, ga.Config.NumberOfIndividual)

	// Indivisuals
	for i := 0; i < ga.Config.NumberOfIndividual; {
		values := make([]int32, len(ga.Config.Params))
		randomValues := generateRandomValues(ga.Config)
		randomIdx := rand.Intn(len(ga.Config.Params))
		for i := range ga.Config.Params {
			if bestValues[i] == nilValue || i == randomIdx {
				values[i] = randomValues[i]
			} else {
				values[i] = bestValues[i]
			}
		}
		ind, ok := ga.newIndividual(values, "")
		if ok {
			log.Printf("random: => %s", ind.id)
			inds = append(inds, ind)
			i++
		}
	}

	// Stop Previous Generation
	stopIndividuals(ga.currInds)

	// Replace to New Generation
	ga.currInds = inds

	// Start Next Generation
	startIndividuals(ga.currInds)

	return nil
}

func (ga *GAManager) PrintGeneration(gn int) {
	log.Printf("Generation: %d\n", gn)
	for i := range ga.currInds {
		log.Printf("%s %s\n", ga.currInds[i].id, stringifyValues(ga.currInds[i].values))
	}
	log.Println()
}

func (ga *GAManager) Destroy() {
	for i := range ga.currInds {
		ga.currInds[i].stopWithClean()
	}
	ga.server.Stop()
}

func (ga *GAManager) newIndividual(values []int32, customID string) (*individual, bool) {
	ind := newIndividual(ga.Config, values, customID)

	if _, exists := ga.indMap[ind.id]; exists {
		return nil, false
	}

	ga.indMap[ind.id] = ind
	if customID == "" {
		ga.allInds = append(ga.allInds, ind)
	}
	return ind, true
}

func (ga *GAManager) CalculateBestValues() []int32 {
	values := make([]int32, len(ga.Config.Params))

	for vi := range ga.Config.Params {
		values[vi] = nilValue

		scores := make(map[int32]struct {
			win  float64
			loss float64
		})
		var total float64

		for ii := range ga.allInds {
			ind := ga.allInds[ii]
			value := ind.values[vi]
			score := scores[value]

			score.win += ind.win
			score.loss += ind.loss
			total += ind.win + ind.loss

			scores[value] = score
		}

		var maxX float64
		for value, score := range scores {
			const c = 1.0
			sum := score.win + score.loss
			if sum >= 1 {
				x := score.win/sum + c*math.Sqrt(2*math.Log(total)/sum)
				if x > maxX {
					values[vi] = value
					maxX = x
				}
			}
		}
	}

	return values
}
