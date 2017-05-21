package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/marcusolsson/tui-go"
	"github.com/prometheus/log"
	"github.com/ryanuber/columnize"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type brInfo struct {
	Name   string
	Author object.Signature
	Branch plumbing.ReferenceName
}

func (i brInfo) String() string {
	return fmt.Sprintf("%s|%s|o %s", i.Author.When.String()[:19], i.Author.Name, i.Name)
}

func main() {
	path := "."
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	repo, err := git.PlainOpen(path)
	if err != nil {
		panic(err)
	}

	brs, err := repo.Branches()
	if err != nil {
		panic(err)
	}

	infoMap := make(map[string]brInfo)
	var info []brInfo
	brs.ForEach(func(br *plumbing.Reference) error {
		commit, err := repo.CommitObject(br.Hash())
		if err != nil {
			log.Error(err.Error())
			return nil
		}
		name := br.Name()
		brInfo := brInfo{name.Short(), commit.Author, name}
		infoMap[name.Short()] = brInfo
		info = append(info, brInfo)
		return nil
	})

	sort.Slice(info, func(i, j int) bool { return info[i].Author.When.Unix() > info[j].Author.When.Unix() })
	var infostr []string
	for _, inf := range info {
		infostr = append(infostr, inf.String())
	}
	infostr = strings.Split(columnize.SimpleFormat(infostr), "\n")

	// ---

	l := tui.NewList()
	l.SetFocused(true)
	l.AddItems(infostr...)
	l.SetSelected(0)

	status := tui.NewStatusBar("")
	status.SetText("[press enter to switch to selected branch]")
	status.SetPermanentText("[press esc or q to quit]")
	root := tui.NewVBox(
		l,
		tui.NewSpacer(),
		status,
	)

	t := tui.NewTheme()
	t.SetStyle("list.item", tui.Style{Bg: tui.ColorBlack, Fg: tui.ColorWhite})
	t.SetStyle("list.item.selected", tui.Style{Bg: tui.ColorGreen, Fg: tui.ColorWhite})

	ui := tui.New(root)
	ui.SetTheme(t)
	ui.SetKeybinding(tui.KeyEsc, func() { ui.Quit() })
	ui.SetKeybinding('q', func() { ui.Quit() })

	l.OnItemActivated(func(l *tui.List) {
		w, err := repo.Worktree()
		if err != nil {
			status.SetText(err.Error())
		}
		err = w.Checkout(&git.CheckoutOptions{
			Branch: infoMap[info[l.Selected()].Name].Branch,
			Force:  true,
		})
		if err != nil {
			status.SetText(err.Error())
		}
		status.SetText("switched to " + info[l.Selected()].Name)
	})

	if err := ui.Run(); err != nil {
		panic(err)
	}
}
