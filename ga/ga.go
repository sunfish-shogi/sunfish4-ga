package ga

import (
	"log"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	server "github.com/sunfish-shogi/sunfish4-ga/shogiserver"
)

type GAManager struct {
	Config Config

	server   server.ShogiServer
	indMap   map[string]*individual
	normal   *individual
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

	var ok bool
	ga.normal, ok = ga.newIndividual(generateNormalValues(ga.Config), "normal")
	if !ok {
		log.Fatal("fatal error: failed to generate normal player")
	}

	ga.currInds = make([]*individual, 0, ga.Config.NumberOfIndividual)
	for len(ga.currInds) < ga.Config.NumberOfIndividual {
		ind, ok := ga.newIndividual(generateRandomValues(ga.Config), "")
		if ok {
			ga.currInds = append(ga.currInds, ind)
		}
	}

	err = startIndividuals(append(ga.currInds, ga.normal))
	return err
}

type indsDescScoreOrder []*individual

func (inds indsDescScoreOrder) Len() int           { return len(inds) }
func (inds indsDescScoreOrder) Swap(i, j int)      { inds[i], inds[j] = inds[j], inds[i] }
func (inds indsDescScoreOrder) Less(i, j int) bool { return inds[i].score > inds[j].score }

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
			}
		}
	}
	for _, ind := range ga.indMap {
		if ind.id != ga.normal.id {
			ind.score -= ga.normal.score
		}
	}

	// Sort by Score
	sort.Stable(indsDescScoreOrder(ga.allInds))
	sort.Stable(indsDescScoreOrder(ga.currInds))

	// Print Scores
	log.Println("Score")
	for i := range ga.currInds {
		log.Printf("%s %0.3f", ga.currInds[i].id, ga.currInds[i].score)
	}
	log.Println()

	// New Generation
	inds := make([]*individual, 0, ga.Config.NumberOfIndividual)

	// Elitism
	log.Printf("elite: %s", ga.allInds[0].id)
	inds = append(inds, ga.allInds[0])

	// Random
	numberOfRandomPlayers := ga.Config.NumberOfIndividual / 4
	for i := 0; i < numberOfRandomPlayers; {
		ind, ok := ga.newIndividual(generateRandomValues(ga.Config), "")
		if ok {
			log.Printf("random: => %s", ind.id)
			inds = append(inds, ind)
			i++
		}
	}

	for {
		if len(inds) >= ga.Config.NumberOfIndividual {
			break
		}

		i1 := ga.selectIndividual("")
		i2 := ga.selectIndividual(i1.id)

		ind, ok := ga.crossover(i1, i2)
		if ok {
			log.Printf("crossover: %s x %s => %s", i1.id, i2.id, ind.id)

			if rand.Intn(8) < 1 /* 1/8 */ {
				ga.mutate(ind)
				log.Printf("mutate: %s", ind.id)
			}

			inds = append(inds, ind)
			continue
		}

		ind, ok = ga.newIndividual(generateRandomValues(ga.Config), "")
		if ok {
			log.Printf("random: => %s", ind.id)
			inds = append(inds, ind)
		}
	}
	log.Println()

	// Stop Previous Generation
	stopIndividuals(ga.currInds)

	// Replace to New Generation
	ga.currInds = inds

	// Start Next Generation
	startIndividuals(ga.currInds)

	return nil
}

func (ga *GAManager) selectIndividual(excludeID string) *individual {
	weight := make([]int, len(ga.currInds))
	var sum int
	for i := range weight {
		if ga.currInds[i].id != excludeID {
			if i == 0 {
				weight[0] = 1024
			} else {
				weight[i] = weight[i-1]*9/10 + 1
			}
			sum += weight[i]
		}
	}

	r := rand.Intn(sum)

	for i := range weight {
		r -= weight[i]
		if r <= 0 {
			return ga.currInds[i]
		}
	}
	return ga.currInds[0]
}

func (ga *GAManager) crossover(i1, i2 *individual) (*individual, bool) {
	values := make([]int32, len(ga.Config.Params))
	for i := range values {
		if rand.Intn(2) == 0 {
			values[i] = i1.values[i]
		} else {
			values[i] = i2.values[i]
		}
	}
	return ga.newIndividual(values, "")
}

func (ga *GAManager) mutate(ind *individual) {
	n := rand.Intn(2) + 1
	for ; n > 0; n-- {
		i := rand.Intn(len(ind.values))
		min := ga.Config.Params[i].MinimumValue
		max := ga.Config.Params[i].MaximumValue
		t := min + rand.Int31n(max-min+1)
		ind.values[i] = t
	}
}

func (ga *GAManager) PrintGeneration(gn int) {
	log.Printf("Generation: %d\n", gn)
	for i := range ga.currInds {
		ss := make([]string, len(ga.currInds[i].values))
		for vi, v := range ga.currInds[i].values {
			ss[vi] = strconv.Itoa(int(v))
		}
		log.Printf("%s [%s]\n", ga.currInds[i].id, strings.Join(ss, ","))
	}
	log.Println()
}

func (ga *GAManager) Destroy() {
	ga.normal.stopWithClean()
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
