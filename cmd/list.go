package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

type Options struct {
	days    int
	reverse bool
}

var opt = &Options{}

type Files []os.FileInfo

func (f Files) Len() int {
	return len(f)
}

func (f Files) Less(i, j int) bool {
	return f[i].Sys().(*syscall.Stat_t).Ctim.Nano() <
		f[j].Sys().(*syscall.Stat_t).Ctim.Nano()
}

func (f Files) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

// ゴミ箱の中のファイル一覧を表示
func list(path string) (files []string, err error) {
	files = make([]string, 0, len(files))

	fileInfo, err := ioutil.ReadDir(path)
	if err != nil {
		return
	}

	if opt.reverse {
		sort.Sort(sort.Reverse(Files(fileInfo)))
	} else {
		sort.Sort(Files(fileInfo))
	}

	const executable os.FileMode = 0111
	const green = "\x1b[32m\x1b[1m%s"
	const blue = "\x1b[34m\x1b[1m%s"
	const cyan = "\x1b[36m\x1b[1m%s"
	const white = "\x1b[37m\x1b[0m%s"

	now := time.Now()
	daysAgo := now.AddDate(0, 0, -opt.days)

	for _, info := range fileInfo {
		internalStat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			err = fmt.Errorf("fileInfo.Sys(): cast error")
			return
		}
		if (internalStat.Ctim.Nano() - daysAgo.UnixNano()) < 0 {
			continue
		}

		if info.IsDir() {
			files = append(files, fmt.Sprintf(blue, info.Name()))
		} else if info.Mode()&os.ModeSymlink != 0 {
			files = append(files, fmt.Sprintf(cyan, info.Name()))
		} else if info.Mode()&executable != 0 {
			files = append(files, fmt.Sprintf(green, info.Name()))
		} else {
			files = append(files, fmt.Sprintf(white, info.Name()))
		}
	}

	return
}

func createListCmd(trashPath string) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "The list of the trash",
		Run: func(cmd *cobra.Command, args []string) {
			files, err := list(trashPath)
			if err != nil {
				log.Fatalln(err)
			}
			for _, file := range files {
				fmt.Println(file)
			}
		},
	}

	cmd.Flags().IntVarP(&opt.days, "days", "d", 31, "How many days ago")
	cmd.Flags().BoolVarP(&opt.reverse, "reverse", "r", false, "display in reverse order")

	return cmd
}
