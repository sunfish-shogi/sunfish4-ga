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

	server server.ShogiServer
	inds   []*individual
	scores []map[int32]scoreType
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

	usedIDs := make(map[string]struct{})
	ga.inds = make([]*individual, 0, ga.Config.NumberOfIndividual)
	for len(ga.inds) < ga.Config.NumberOfIndividual {
		ind := newIndividual(ga.Config, generateRandomValues(ga.Config))
		if _, exists := usedIDs[ind.id]; exists {
			continue
		}
		ga.inds = append(ga.inds, ind)
	}

	err = startIndividuals(ga.inds)
	return err
}

func (ga *GAManager) Next() error {
	rate, err := ga.server.MakeRate()
	if err != nil {
		return err
	}

	indMap := make(map[string]*individual)
	for _, ind := range ga.inds {
		indMap[ind.id] = ind
	}
	for pi := range rate.Players {
		for _, player := range rate.Players[pi] {
			if ind, ok := indMap[player.Name]; ok {
				ind.score = player.Rate
				ind.win = player.Win
				ind.loss = player.Loss
			}
		}
	}

	// Print Scores
	log.Println("Score")
	for i := range ga.inds {
		log.Printf("%s %0.3f", ga.inds[i].id, ga.inds[i].score)
	}
	log.Println()

	// Update Scores
	ga.UpdateScores()

	// Best Values
	bestValues := ga.CalculateBestValues()
	log.Println("Best Values")
	log.Println(stringifyValues(bestValues))
	log.Println()

	// New Generation
	inds := make([]*individual, 0, ga.Config.NumberOfIndividual)

	// Indivisuals
	usedIDs := make(map[string]struct{})
	for i := 0; i < ga.Config.NumberOfIndividual; i++ {
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
		ind := newIndividual(ga.Config, values)
		if _, exists := usedIDs[ind.id]; exists {
			continue
		}
		inds = append(inds, ind)
	}

	// Stop Previous Generation
	stopIndividuals(ga.inds)

	// Replace to New Generation
	ga.inds = inds

	// Start Next Generation
	startIndividuals(ga.inds)

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

			score.win += ind.win
			score.loss += ind.loss

			ga.scores[i][value] = score
		}
	}
}

func (ga *GAManager) CalculateBestValues() []int32 {
	values := make([]int32, len(ga.Config.Params))

	for i := range ga.Config.Params {
		values[i] = nilValue

		var total float64
		for value := range ga.scores[i] {
			total += ga.scores[i][value].win
			total += ga.scores[i][value].loss
		}

		var maxX float64
		for value, score := range ga.scores[i] {
			const c = 1.0
			sum := score.win + score.loss
			if sum >= 1 {
				x := score.win/sum + c*math.Sqrt(2*math.Log(total)/sum)
				if x > maxX {
					values[i] = value
					maxX = x
				}
			}
		}
	}

	return values
}
