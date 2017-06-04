package gitbr

import (
	"fmt"
	"sort"
	"strings"

	"github.com/marcusolsson/tui-go"
	"github.com/prometheus/log"
	"github.com/ryanuber/columnize"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// UIRunner wraps the function to Run an UI
type UIRunner interface {
	Run() error
}

// Open returns an UIRunner from a git repository filesystem path.
func Open(path string) (UIRunner, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}

	brs, err := extract(repo)
	if err != nil {
		return nil, err
	}

	return newTuiUIRunner(repo, brs), nil
}

type branch struct {
	Name   string
	Author object.Signature
	Branch plumbing.ReferenceName
}

func (i branch) String() string {
	return fmt.Sprintf("%s|%s|o %s", i.Author.When.String()[:19], i.Author.Name, i.Name)
}

type branches map[string]*branch

func (brs branches) sort() []*branch {
	var list []*branch
	for _, br := range brs {
		list = append(list, br)
	}
	// sort by date desc
	sort.Slice(list, func(i, j int) bool { return list[i].Author.When.Unix() > list[j].Author.When.Unix() })
	return list
}

func (brs branches) displayData() []string {
	var data []string
	for _, br := range brs.sort() {
		data = append(data, br.String())
	}
	data = strings.Split(columnize.SimpleFormat(data), "\n")
	return data
}

func extract(repo *git.Repository) (branches, error) {
	brs, err := repo.Branches()
	if err != nil {
		return nil, err
	}

	brsByName := make(branches)
	brs.ForEach(func(br *plumbing.Reference) error {
		commit, err := repo.CommitObject(br.Hash())
		if err != nil {
			log.Error(err.Error())
			return nil
		}
		name := br.Name()
		branch := &branch{name.Short(), commit.Author, name}
		brsByName[name.Short()] = branch
		return nil
	})

	return brsByName, nil
}

func newTuiUIRunner(repo *git.Repository, brs branches) tui.UI {
	l := tui.NewList()
	l.SetFocused(true)
	list := brs.displayData()
	l.AddItems(list...)
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
		br := brs.sort()[l.Selected()]
		err = w.Checkout(&git.CheckoutOptions{
			Branch: brs[br.Name].Branch,
			Force:  true,
		})
		if err != nil {
			status.SetText(err.Error())
		}
		status.SetText("switched to " + br.Name)
	})

	return ui
}
