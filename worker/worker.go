package worker

import (
	"bytes"
	"github.com/ddalogin/bicycle-ci/models"
	"os/exec"
	"strconv"
	"strings"
)

// Выполнить этап билда
func RunStep(project models.Project, cmd *exec.Cmd, result *models.Step) {
	var stdout, stderr bytes.Buffer
	var env []string
	env = append(env, "ID="+strconv.Itoa(int(project.Id)))
	env = append(env, "NAME="+project.Name)
	env = append(env, "DEPLOY_DIR="+strings.TrimSpace(*project.DeployDir))
	env = append(env, "ARTIFACT_DIR="+strings.TrimSpace(*project.ArtifactDir))
	env = append(env, "SSHKEY="+*project.DeployPrivate)

	cmd.Env = env
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		result.Error = err.Error()
		result.Status = models.STEP_STATUS_FAILED
	}
	cmd.Wait()
	result.StdOut = string(stdout.Bytes())
	result.StdErr = string(stderr.Bytes())

	if result.Status == models.STEP_STATUS_RUNING {
		result.Status = models.STEP_STATUS_SUCCESS
	}

	return
}
