package helm

import (
	"context"
	"os"
	"os/exec"
	"time"

	"github.com/pkg/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type Helm struct {
	namespace string
	WorkDir   string
}

const (
	ReleaseName   = "app-release"
	FlagNamespace = "--namespace"
)

var log = logf.Log.WithName("helm_controller")

func NewHelm(namespace string) *Helm {
	return &Helm{
		namespace: namespace,
		WorkDir:   os.Getenv("DEPLOYMENT_DIR") + "/app-deployment-generated",
	}
}

func (h *Helm) execCommand(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "helm", args...)
	cmd.Dir = h.WorkDir

	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return "", errors.New("command timed out")
	}
	if err != nil {
		return "", errors.Wrapf(err, "command failed, output: %v", string(out))
	}

	log.Info("cmd output", "output", string(out))
	log.Info("command successfully executed")

	return string(out), nil
}

func (h *Helm) getRelease() (string, error) {
	out, err := h.execCommand("list", "-q", FlagNamespace, h.namespace)

	return string(out), err
}

func (h *Helm) install() error {
	_, err := h.execCommand("install", ReleaseName, FlagNamespace, h.namespace, ".")

	return err
}

func (h *Helm) upgrade() error {
	_, err := h.execCommand("upgrade", ReleaseName, FlagNamespace, h.namespace, ".")

	return err
}

func (h *Helm) Deploy() error {
	if release, err := h.getRelease(); err == nil {
		if release == "" {
			return h.install()
		} else {
			return h.upgrade()
		}
	} else {
		return err
	}
}

func (h *Helm) Undeploy() error {
	_, err := h.execCommand("uninstall", ReleaseName, FlagNamespace, h.namespace)

	return err
}
