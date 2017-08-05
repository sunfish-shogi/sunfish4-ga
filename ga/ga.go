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
	lastID   int
	indMap   map[string]*individual
	normal   *individual
	allInds  []*individual
	currInds []*individual
}

func NewGAManager(config Config) *GAManager {
	validateConfig(config)
	return &GAManager{
		Config: config,
		indMap: make(map[string]*individual),
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

		time.Sleep(time.Hour * 4)

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

	ga.normal = newIndividual(ga.nextID(), ga.Config)
	ga.normal.initParamNormal()

	ga.currInds = make([]*individual, 0, ga.Config.NumberOfIndividual)
	for i := 0; i < ga.Config.NumberOfIndividual; i++ {
		ind := newIndividual(ga.nextID(), ga.Config)
		ind.initParamByRandom()
		ga.currInds = append(ga.currInds, ind)
	}
	ga.updateIndMap()

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

	ga.normal.score = 0
	for _, ind := range ga.indMap {
		ind.score = 0
	}
	for pi := range rate.Players {
		for _, player := range rate.Players[pi] {
			if ind, ok := ga.indMap[player.Name]; ok {
				ind.score = player.Rate
			} else if ga.normal.id == player.Name {
				ga.normal.score = player.Rate
			}
		}
	}
	for _, ind := range ga.indMap {
		ind.score -= ga.normal.score
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
	for i := 0; i < 8; i++ {
		ind := newIndividual(ga.nextID(), ga.Config)
		ind.initParamByRandom()
		log.Printf("random: => %s", ind.id)
		inds = append(inds, ind)
	}

	for {
		if len(inds) >= ga.Config.NumberOfIndividual {
			break
		}

		i1 := ga.selectIndividual("")
		i2 := ga.selectIndividual(i1.id)

		ind := ga.crossover(i1, i2)
		log.Printf("crossover: %s x %s => %s", i1.id, i2.id, ind.id)

		if rand.Intn(8) < 1 /* 1/8 */ {
			ga.mutate(ind)
			log.Printf("mutate: %s", ind.id)
		}

		inds = append(inds, ind)
	}
	log.Println()

	// Stop Previous Generation
	stopIndividuals(ga.currInds)

	// Replace to New Generation
	ga.currInds = inds
	ga.updateIndMap()

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

func (ga *GAManager) crossover(i1, i2 *individual) *individual {
	ind := newIndividual(ga.nextID(), ga.Config)
	values := make([]int32, len(ga.Config.Params))
	for i := range values {
		if rand.Intn(2) == 0 {
			values[i] = i1.values[i]
		} else {
			values[i] = i2.values[i]
		}
	}
	ind.initParam(values)
	return ind
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

func (ga *GAManager) nextID() string {
	ga.lastID++
	return strconv.Itoa(ga.lastID)
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

func (ga *GAManager) updateIndMap() {
	for i := range ga.currInds {
		ga.indMap[ga.currInds[i].id] = ga.currInds[i]
	}

	ga.allInds = make([]*individual, 0, len(ga.indMap))
	for _, ind := range ga.indMap {
		ga.allInds = append(ga.allInds, ind)
	}
}
