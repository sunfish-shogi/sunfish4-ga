package ga

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	server "github.com/sunfish-shogi/sunfish4-ga/shogiserver"
)

type GAManager struct {
	Config Config

	server server.ShogiServer
	lastID int
	inds   []*individual
}

func NewGAManager(config Config) *GAManager {
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

	ga.inds = make([]*individual, 0, ga.Config.NumberOfIndividual)
	ga.lastID = 0
	for i := 0; i < ga.Config.NumberOfIndividual; i++ {
		ind := newIndividual(ga.nextID(), ga.Config)
		if i == 0 {
			ind.initParamForFirstElite()
		} else {
			ind.initParamByRandom()
		}
		ga.inds = append(ga.inds, ind)
	}

	errs := ga.startIndividuals()
	if len(errs) != 0 {
		return errs[0]
	}
	return nil
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
	for i := 0; i < 2; i++ {
		ind := ga.copyElite(i)
		log.Printf("elite: %s => %s", ga.inds[i].id, ind.id)
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

		if rand.Intn(100) < 1 /* 1/100 */ {
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

	ga.startIndividuals()

	return nil
}

func (ga *GAManager) startIndividuals() []error {
	var errs []error

	// Setup
	var wg sync.WaitGroup
	for _, ind := range ga.inds {
		wg.Add(1)
		go func(ind *individual) {
			defer wg.Done()
			err := ind.setup()
			if err != nil {
				err = errors.Wrap(err, fmt.Sprintf("failed to setup sunfish %s", ind.id))
				errs = append(errs, err)
				log.Println(err)
			}
		}(ind)
	}
	wg.Wait()

	// Start
	for _, ind := range ga.inds {
		err := ind.start()
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to start sunfish %s", ind.id))
			errs = append(errs, err)
			log.Println(err)
		}
	}

	return errs
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
				weight[i] = weight[i-1]*4/5 + 1
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
	i := rand.Intn(len(ind.values))
	min := ga.Config.Params[i].MinimumValue
	max := ga.Config.Params[i].MaximumValue
	t := min + rand.Int31n(max-min+1)
	ind.values[i] = (ind.values[i] + t) / 2
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
	for i := range ga.inds {
		ga.inds[i].stop()
	}
	ga.server.Stop()
}
