package ga

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path"

	"github.com/pkg/errors"
	"github.com/sunfish-shogi/sunfish4-ga/util"
)

type individual struct {
	id     string
	values []int32
	score  float64
	cmd    *exec.Cmd
	config Config
}

func newIndividual(id string, config Config) *individual {
	ind := &individual{
		id:     id,
		values: make([]int32, len(config.Params)),
		config: config,
	}
	return ind
}

func (ind *individual) initParamForFirstElite() {
	for i := range ind.config.Params {
		ind.values[i] = ind.config.Params[i].FirstEliteValue
	}
}

func (ind *individual) initParamByRandom() {
	for i := range ind.config.Params {
		min := ind.config.Params[i].MinimumValue
		max := ind.config.Params[i].MaximumValue
		ind.values[i] = min + rand.Int31n(max-min+1)
	}
}

func (ind *individual) initParam(values []int32) {
	copy(ind.values, values)
}

func (ind *individual) setup() error {
	err := util.Command("git", "clone", "--depth", "1", "--branch", "expt-ga", "https://github.com/sunfish-shogi/sunfish4.git", ind.id).Run()
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

func (ind *individual) stop() {
	if ind.cmd != nil && ind.cmd.Process != nil {
		if err := ind.cmd.Process.Kill(); err != nil {
			log.Println(err)
		} else if _, err := ind.cmd.Process.Wait(); err != nil {
			log.Println(err)
		}
	}
	os.RemoveAll(ind.Dir())
}

func (ind *individual) Dir() string {
	return path.Join(util.WorkDir(), ind.id)
}
