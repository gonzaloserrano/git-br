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
	"gopkg.in/src-d/go-git.v4/utils/merkletrie"
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

	return newTuiUI(repo, brs), nil
}

type branch struct {
	Name   string
	Author object.Signature
	Branch plumbing.ReferenceName
	Tree   *object.Tree
}

func (b branch) String() string {
	author := b.Author.Name
	if len(author) > 16 {
		author = author[0:15] + "..."
	}
	name := b.Name
	if len(name) > 32 {
		name = name[0:31] + "..."
	}
	return fmt.Sprintf("%s|%s|o %s", b.Author.When.String()[2:19], author, name)
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
	err = brs.ForEach(func(br *plumbing.Reference) error {
		commit, err := repo.CommitObject(br.Hash())
		if err != nil {
			log.Error(err.Error())
			return nil
		}
		tree, err := commit.Tree()
		if err != nil {
			log.Error(err.Error())
			return nil
		}
		name := br.Name()
		branch := &branch{name.Short(), commit.Author, name, tree}
		brsByName[name.Short()] = branch
		return nil
	})

	if err != nil {
		return nil, err
	}

	return brsByName, nil
}

func newTuiUI(repo *git.Repository, brs branches) tui.UI {
	t := tui.NewTable(0, 0)

	t.SetColumnStretch(0, 1)
	t.SetColumnStretch(1, 1)
	t.SetColumnStretch(2, 2)

	sortedBrs := brs.sort()

	for _, br := range sortedBrs {
		author := br.Author.Name
		if len(author) > 16 {
			author = author[0:15] + "..."
		}
		name := br.Name
		if len(name) > 26 {
			name = name[0:25] + "..."
		}
		t.AppendRow(
			tui.NewLabel(br.Author.When.String()[2:19]),
			tui.NewLabel(author),
			tui.NewLabel(name),
		)
	}

	diffView := tui.NewLabel("")

	status := tui.NewStatusBar("")
	status.SetText("[press enter to switch to selected branch]")
	status.SetPermanentText("[press esc or q to quit]")
	top := tui.NewHBox(t, diffView)
	top.SetSizePolicy(tui.Preferred, tui.Preferred)
	root := tui.NewVBox(
		top,
		tui.NewSpacer(),
		status,
	)
	root.SetSizePolicy(tui.Preferred, tui.Preferred)

	th := tui.NewTheme()
	th.SetStyle("table.cell.selected", tui.Style{Bg: tui.ColorGreen, Fg: tui.ColorWhite})

	ui := tui.New(root)
	ui.SetTheme(th)
	ui.SetKeybinding(tui.KeyEsc, func() { ui.Quit() })
	ui.SetKeybinding('q', func() { ui.Quit() })
	t.OnItemActivated(func(t *tui.Table) {
		w, err := repo.Worktree()
		if err != nil {
			status.SetText(err.Error())
			return
		}
		br := sortedBrs[t.Selected()]
		err = w.Checkout(&git.CheckoutOptions{
			Branch: br.Branch,
			Force:  true,
		})

		if err != nil {
			status.SetText(err.Error())
			return
		}
		status.SetText("switched to " + br.Name)
	})
	t.OnSelectionChanged(func(t *tui.Table) {
		br := sortedBrs[t.Selected()]
		fromBrName := "master"
		if br.Name == fromBrName {
			diffView.SetText("")
			return
		}
		fromBr, ok := brs[fromBrName]
		if !ok {
			status.SetText(fmt.Sprintf("no origin %s branch", fromBrName))
			return
		}
		changes, err := object.DiffTree(fromBr.Tree, br.Tree)
		if err != nil {
			status.SetText(err.Error())
			return
		}
		var changesMsg string
		if len(changes) == 0 {
			changesMsg = fmt.Sprintf("no changes between %s and %s", fromBrName, br.Name)
		} else {
			changesMsg = changesToString(fromBrName, changes)
		}
		diffView.SetText(changesMsg)
	})
	t.Select(0)

	return ui
}

func changesToString(fromBrName string, changes object.Changes) string {
	changesMsg := fmt.Sprintf("changes between %s and selected:\n\n", fromBrName)
	if len(changes) > 30 {
		changesMsg += "\ttoo many changes to display :-("
		return changesMsg
	}
	for _, c := range changes {
		action, err := c.Action()
		if err != nil {
			// @TODO log
			continue
		}
		var actionStr, fileName string
		switch action {
		case merkletrie.Insert:
			actionStr = "A"
			fileName = c.To.Name
		case merkletrie.Delete:
			actionStr = "D"
			fileName = c.From.Name
		case merkletrie.Modify:
			actionStr = "M"
			fileName = c.From.Name
		}
		changesMsg += fmt.Sprintf("\t%s: %s\n", actionStr, fileName)
	}

	return changesMsg
}
