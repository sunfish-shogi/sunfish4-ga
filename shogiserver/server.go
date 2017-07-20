package shogiserver

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/sunfish-shogi/sunfish4-ga/util"
)

var kifuDir = regexp.MustCompile(`^[0-9]+$`)

type ShogiServer struct {
	Cmd *exec.Cmd
}

func (s *ShogiServer) Dir() string {
	return path.Join(util.WorkDir(), "shogi-server")
}

func (s *ShogiServer) Setup() error {
	cmd := exec.Command("git", "clone", "--depth", "1", "--branch", "master", "git://git.pf.osdn.jp/gitroot/s/su/sunfish-shogi/shogi-server.git", "shogi-server")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, "failed to clone shogi-server")
	}

	s.Cmd = exec.Command("ruby", "shogi-server", "test", "4081")
	s.Cmd.Dir = s.Dir()
	err = s.Cmd.Start()
	if err != nil {
		return errors.Wrap(err, "failed to start shogi-server")
	}
	return nil
}

func (s *ShogiServer) Stop() {
	if s.Cmd != nil && s.Cmd.Process != nil {
		if err := s.Cmd.Process.Kill(); err != nil {
			log.Println(err)
		} else if _, err := s.Cmd.Process.Wait(); err != nil {
			log.Println(err)
		}
	}
}

func (s *ShogiServer) MakeRate() (Rate, error) {
	files, err := ioutil.ReadDir(s.Dir())
	if err != nil {
		return Rate{}, err
	}

	cmdParams := make([]string, 0, 8)
	cmdParams = append(cmdParams, path.Join(s.Dir(), "mk_game_results"))
	for i := range files {
		if files[i].IsDir() && kifuDir.MatchString(files[i].Name()) {
			cmdParams = append(cmdParams, files[i].Name())
		}
	}
	cmdParams = append(cmdParams, "|")
	cmdParams = append(cmdParams, "grep", "-v", "abnormal")
	cmdParams = append(cmdParams, "|")
	cmdParams = append(cmdParams, path.Join(s.Dir(), "mk_rate"))

	cmd := exec.Command("sh", "-c", strings.Join(cmdParams, " "))
	cmd.Dir = s.Dir()
	buf, err := cmd.Output()
	if err != nil {
		return Rate{}, err
	}

	return UnmarshalRate(buf)
}
