package ga

import (
	"fmt"
	"log"
	"math/big"
	"os"
	"os/exec"
	"path"
	"sync"

	"github.com/pkg/errors"
	"github.com/sunfish-shogi/sunfish4-ga/util"
	"github.com/tv42/base58"
	"golang.org/x/sync/errgroup"
)

type individual struct {
	id         string
	values     []int32
	score      float64
	scoreLower float64
	scoreUpper float64
	cmd        *exec.Cmd
	config     Config
}

func newIndividual(config Config, values []int32, customID string) *individual {
	ind := &individual{
		values: values,
		config: config,
	}
	if customID != "" {
		ind.id = customID
	} else {
		ind.setUniqueID()
	}
	return ind
}

func (ind *individual) setUniqueID() {
	x := big.NewInt(0)
	for i := range ind.config.Params {
		value := ind.values[i]
		param := ind.config.Params[i]

		x.Mul(x, big.NewInt(int64(param.MaximumValue-param.MinimumValue)))
		x.Add(x, big.NewInt(int64(value-param.MinimumValue)))
	}
	ind.id = string(base58.EncodeBig(nil, x))
}

func (ind *individual) setup() error {
	err := util.Command("git", "clone", "--depth", "1", "--branch", "master", "https://github.com/sunfish-shogi/sunfish4.git", ind.id).Run()
	if err != nil {
		return errors.Wrap(err, "failed to clone sunfish4")
	}

	err = ind.writeParamHpp()
	if err != nil {
		return err
	}

	cmd := util.Command("make", "csa")
	cmd.Dir = path.Join(ind.Dir())
	cmd.Run()
	if err != nil {
		return errors.Wrap(err, "failed to make csa")
	}

	err = util.Symlink(path.Join(util.WorkDir(), "eval.bin"), path.Join(ind.Dir(), "eval.bin"))
	if err != nil {
		return errors.Wrap(err, "failed to create symbolic link for eval.bin")
	}

	err = util.Symlink(path.Join(util.WorkDir(), "book.bin"), path.Join(ind.Dir(), "book.bin"))
	if err != nil {
		return errors.Wrap(err, "failed to create symbolic link for book.bin")
	}

	err = ind.writeCsaIni()
	if err != nil {
		return err
	}

	return nil
}

func (ind *individual) start() error {
	if ind.cmd != nil {
		return fmt.Errorf("sunfish_csa already started")
	}

	ind.cmd = util.Command(path.Join(ind.Dir(), "sunfish_csa"), "-s")
	ind.cmd.Dir = path.Join(ind.Dir())
	err := ind.cmd.Start()
	if err != nil {
		return errors.Wrap(err, "failed to start sunfish_csa")
	}
	return nil
}

func (ind *individual) writeParamHpp() error {
	f, err := os.OpenFile(path.Join(ind.Dir(), "src/search/Param.hpp"), os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	for i := range ind.config.Params {
		fmt.Fprintf(f, "#define %s %d\n", ind.config.Params[i].Name, ind.values[i])
	}

	return f.Close()
}

func (ind *individual) writeCsaIni() error {
	f, err := os.OpenFile(path.Join(ind.Dir(), "config/csa.ini"), os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	f.WriteString("[Server]\n")
	f.WriteString("Host      = localhost\n")
	f.WriteString("Port      = 4081\n")
	f.WriteString("Pass      = test-600-10,SunTest\n")
	f.WriteString("Floodgate = 1\n")
	f.WriteString("User      = " + ind.id + "\n")
	f.WriteString("\n")
	f.WriteString("[Search]\n")
	f.WriteString("Depth    = 48\n")
	f.WriteString("Limit    = 1\n")
	f.WriteString("Repeat   = 1000000\n")
	f.WriteString("Worker   = 1\n")
	f.WriteString("Ponder   = 0\n")
	f.WriteString("UseBook  = 1\n")
	f.WriteString("HashMem  = 128\n")
	f.WriteString("MarginMs = 500\n")
	f.WriteString("\n")
	f.WriteString("[KeepAlive]\n")
	f.WriteString("KeepAlive = 1\n")
	f.WriteString("KeepIdle  = 10\n")
	f.WriteString("KeepIntvl = 5\n")
	f.WriteString("KeepCnt   = 10\n")
	f.WriteString("\n")
	f.WriteString("[File]\n")
	f.WriteString("KifuDir   = out/csa_kifu\n")

	return f.Close()
}

func (ind *individual) stopWithClean() {
	ind.kill()
	ind.clean()
}

func (ind *individual) kill() {
	if ind.cmd != nil && ind.cmd.Process != nil {
		if err := ind.cmd.Process.Kill(); err != nil {
			log.Println(err)
		} else if _, err := ind.cmd.Process.Wait(); err != nil {
			log.Println(err)
		}
		ind.cmd = nil
	}
}

func (ind *individual) clean() {
	os.RemoveAll(ind.Dir())
}

func (ind *individual) Dir() string {
	return path.Join(util.WorkDir(), ind.id)
}

func startIndividuals(inds []*individual) error {
	// Setup
	var eg errgroup.Group
	for _, _ind := range inds {
		ind := _ind
		eg.Go(func() error {
			err := ind.setup()
			if err != nil {
				err = errors.Wrap(err, fmt.Sprintf("failed to setup sunfish %s", ind.id))
				log.Println(err)
				return err
			}
			return nil
		})
	}
	rerr := eg.Wait()

	// Start
	for _, ind := range inds {
		err := ind.start()
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to start sunfish %s", ind.id))
			log.Println(err)
			if rerr == nil {
				rerr = err
			}
		}
	}

	return rerr
}

func stopIndividuals(inds []*individual) {
	// Kill
	wg := &sync.WaitGroup{}
	for _, ind := range inds {
		wg.Add(1)
		func(ind *individual) {
			defer wg.Done()
			ind.kill()
		}(ind)
	}
	wg.Wait()

	// Clean
	for _, ind := range inds {
		ind.clean()
	}
}
