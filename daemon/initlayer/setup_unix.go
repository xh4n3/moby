// +build linux freebsd

package initlayer

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/docker/docker/pkg/idtools"
)

// Setup populates a directory with mountpoints suitable
// for bind-mounting things into the container.
//
// This extra layer is used by all containers as the top-most ro layer. It protects
// the container from unwanted side-effects on the rw layer.
// layer 的默认配置 setup
func Setup(initLayer string, rootUID, rootGID int) error {
	for pth, typ := range map[string]string{
		"/dev/pts":         "dir",
		"/dev/shm":         "dir",
		"/proc":            "dir",
		"/sys":             "dir",
		"/.dockerenv":      "file",
		"/etc/resolv.conf": "file",
		"/etc/hosts":       "file",
		"/etc/hostname":    "file",
		"/dev/console":     "file",
		"/etc/mtab":        "/proc/mounts",
	} {
		parts := strings.Split(pth, "/")
		prev := "/"
		for _, p := range parts[1:] {
			prev = filepath.Join(prev, p)
			// 类似于删除
			// https://linux.die.net/man/2/unlink
			// initLayer 应该对应于宿主机上某个路径，这里要删除 initLayer/dev/pts 等等
			syscall.Unlink(filepath.Join(initLayer, prev))
		}

		if _, err := os.Stat(filepath.Join(initLayer, pth)); err != nil {
			// 看是否删除成功，如果成功，即进入该 if body
			if os.IsNotExist(err) {
				// 以 root 用户创建所有目录
				if err := idtools.MkdirAllNewAs(filepath.Join(initLayer, filepath.Dir(pth)), 0755, rootUID, rootGID); err != nil {
					return err
				}
				switch typ {
				case "dir":
					if err := idtools.MkdirAllNewAs(filepath.Join(initLayer, pth), 0755, rootUID, rootGID); err != nil {
						return err
					}
				case "file":
					f, err := os.OpenFile(filepath.Join(initLayer, pth), os.O_CREATE, 0755)
					if err != nil {
						return err
					}
					f.Chown(rootUID, rootGID)
					f.Close()
				default:
					if err := os.Symlink(typ, filepath.Join(initLayer, pth)); err != nil {
						return err
					}
				}
			} else {
				return err
			}
		}
	}

	// Layer is ready to use, if it wasn't before.
	return nil
}
