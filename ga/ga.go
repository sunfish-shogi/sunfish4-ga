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

	server server.ShogiServer
	lastID int
	normal *individual
	inds   []*individual
}

func NewGAManager(config Config) *GAManager {
	validateConfig(config)
	return &GAManager{
		Config: config,
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

	ga.inds = make([]*individual, 0, ga.Config.NumberOfIndividual)
	for i := 0; i < ga.Config.NumberOfIndividual; i++ {
		ind := newIndividual(ga.nextID(), ga.Config)
		ind.initParamByRandom()
		ga.inds = append(ga.inds, ind)
	}

	err = startIndividuals(append(ga.inds, ga.normal))
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

	for pi := range rate.Players {
		for _, player := range rate.Players[pi] {
			if player.Win+player.Loss < 100 {
				continue
			}
			for i := range ga.inds {
				if ga.inds[i].id == player.Name {
					ga.inds[i].score = player.Rate
					break
				}
			}
		}
	}

	// Sort by Score
	if ga.inds[0].score != 0 {
		sort.Stable(indsDescScoreOrder(ga.inds))
	} else {
		sort.Stable(indsDescScoreOrder(ga.inds[1:]))
	}

	// Print Scores
	log.Println("Score")
	for i := range ga.inds {
		log.Printf("%s %0.3f", ga.inds[i].id, ga.inds[i].score)
	}
	log.Println()

	// New Generation
	inds := make([]*individual, 0, ga.Config.NumberOfIndividual)

	// Elitism
	ind := ga.copyElite(0)
	log.Printf("elite: %s => %s", ga.inds[0].id, ind.id)
	inds = append(inds, ind)

	// Random
	ind = newIndividual(ga.nextID(), ga.Config)
	ind.initParamByRandom()
	log.Printf("random: => %s", ind.id)
	inds = append(inds, ind)

	for {
		if len(inds) >= ga.Config.NumberOfIndividual {
			break
		}

		i1 := ga.selectIndividual("")
		i2 := ga.selectIndividual(i1.id)

		ind := ga.crossover(i1, i2)
		log.Printf("crossover: %s x %s => %s", i1.id, i2.id, ind.id)

		if rand.Intn(10) < 1 /* 1/10 */ {
			ga.mutate(ind)
			log.Printf("mutate: %s", ind.id)
		}

		inds = append(inds, ind)
	}
	log.Println()

	// Stop Previous Generation
	for i := range ga.inds {
		ga.inds[i].stop()
	}

	// Replace to New Generation
	ga.inds = inds

	startIndividuals(ga.inds)

	return nil
}

func (ga *GAManager) copyElite(idx int) *individual {
	ind := newIndividual(ga.nextID(), ga.Config)
	ind.initParam(ga.inds[idx].values)
	return ind
}

func (ga *GAManager) selectIndividual(excludeID string) *individual {
	weight := make([]int, len(ga.inds))
	var sum int
	for i := range weight {
		if ga.inds[i].id != excludeID {
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
			return ga.inds[i]
		}
	}
	return ga.inds[0]
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
		ind.values[i] = (ind.values[i] + t) / 2
	}
}

func (ga *GAManager) nextID() string {
	ga.lastID++
	return strconv.Itoa(ga.lastID)
}

func (ga *GAManager) PrintGeneration(gn int) {
	log.Printf("Generation: %d\n", gn)
	for i := range ga.inds {
		ss := make([]string, len(ga.inds[i].values))
		for vi, v := range ga.inds[i].values {
			ss[vi] = strconv.Itoa(int(v))
		}
		log.Printf("%s [%s]\n", ga.inds[i].id, strings.Join(ss, ","))
	}
	log.Println()
}

func (ga *GAManager) Destroy() {
	ga.normal.stop()
	for i := range ga.inds {
		ga.inds[i].stop()
	}
	ga.server.Stop()
}
