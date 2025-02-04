package db_lib

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/pkg/task_logger"
	"github.com/ansible-semaphore/semaphore/util"
	"os"
	"os/exec"
	"strings"
	"time"
)

type BashApp struct {
	Logger     task_logger.Logger
	Template   db.Template
	Repository db.Repository
	reader     bashReader
}

type bashReader struct {
	input  *string
	logger task_logger.Logger
}

func (r *bashReader) Read(p []byte) (n int, err error) {

	r.logger.SetStatus(task_logger.TaskWaitingConfirmation)

	for {
		time.Sleep(time.Second * 3)
		if r.input != nil {
			break
		}
	}

	copy(p, *r.input+"\n")
	r.logger.SetStatus(task_logger.TaskRunningStatus)
	return len(*r.input) + 1, nil
}

func (t *BashApp) makeCmd(command string, args []string, environmentVars *[]string) *exec.Cmd {
	cmd := exec.Command(command, args...) //nolint: gas
	cmd.Dir = t.GetFullPath()

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", util.Config.TmpPath))
	cmd.Env = append(cmd.Env, fmt.Sprintf("PWD=%s", cmd.Dir))

	if environmentVars != nil {
		cmd.Env = append(cmd.Env, args...)
	}

	if environmentVars != nil {
		cmd.Env = append(cmd.Env, *environmentVars...)
	}

	// Remove sensitive env variables from cmd process
	for _, env := range getSensitiveEnvs() {
		cmd.Env = append(cmd.Env, env+"=")
	}

	return cmd
}

func (t *BashApp) runCmd(command string, args []string) error {
	cmd := t.makeCmd(command, args, nil)
	t.Logger.LogCmd(cmd)
	return cmd.Run()
}

func (t *BashApp) GetFullPath() (path string) {
	path = t.Repository.GetFullPath(t.Template.ID)
	return
}

func (t *BashApp) SetLogger(logger task_logger.Logger) task_logger.Logger {
	t.Logger = logger
	t.Logger.AddStatusListener(func(status task_logger.TaskStatus) {

	})
	t.reader.logger = logger
	return logger
}

func (t *BashApp) InstallRequirements() error {
	return nil
}

func (t *BashApp) Run(args []string, environmentVars *[]string, inputs map[string]string, cb func(*os.Process)) error {
	cmd := t.makeCmd("bash", args, environmentVars)
	t.Logger.LogCmd(cmd)
	//cmd.Stdin = &t.reader
	cmd.Stdin = strings.NewReader("")
	err := cmd.Start()
	if err != nil {
		return err
	}
	cb(cmd.Process)
	return cmd.Wait()
}
