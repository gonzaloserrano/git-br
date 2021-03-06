package git

import (
	"io"
	"path/filepath"
	"strings"

	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/filemode"
	"gopkg.in/src-d/go-git.v4/plumbing/format/index"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage"
	"gopkg.in/src-d/go-git.v4/utils/ioutil"

	"gopkg.in/src-d/go-billy.v2"
)

// Commit stores the current contents of the index in a new commit along with
// a log message from the user describing the changes.
func (w *Worktree) Commit(msg string, opts *CommitOptions) (plumbing.Hash, error) {
	if err := opts.Validate(w.r); err != nil {
		return plumbing.ZeroHash, err
	}

	if opts.All {
		if err := w.autoAddModifiedAndDeleted(); err != nil {
			return plumbing.ZeroHash, err
		}
	}

	idx, err := w.r.Storer.Index()
	if err != nil {
		return plumbing.ZeroHash, err
	}

	h := &commitIndexHelper{
		fs: w.fs,
		s:  w.r.Storer,
	}

	tree, err := h.buildTreeAndBlobObjects(idx)
	if err != nil {
		return plumbing.ZeroHash, err
	}

	commit, err := w.buildCommitObject(msg, opts, tree)
	if err != nil {
		return plumbing.ZeroHash, err
	}

	return commit, w.updateHEAD(commit)
}

func (w *Worktree) autoAddModifiedAndDeleted() error {
	s, err := w.Status()
	if err != nil {
		return err
	}

	for path, fs := range s {
		if fs.Worktree != Modified && fs.Worktree != Deleted {
			continue
		}

		if _, err := w.Add(path); err != nil {
			return err
		}

	}

	return nil
}

func (w *Worktree) updateHEAD(commit plumbing.Hash) error {
	head, err := w.r.Storer.Reference(plumbing.HEAD)
	if err != nil {
		return err
	}

	name := plumbing.HEAD
	if head.Type() != plumbing.HashReference {
		name = head.Target()
	}

	ref := plumbing.NewHashReference(name, commit)
	return w.r.Storer.SetReference(ref)
}

func (w *Worktree) buildCommitObject(msg string, opts *CommitOptions, tree plumbing.Hash) (plumbing.Hash, error) {
	commit := &object.Commit{
		Author:       *opts.Author,
		Committer:    *opts.Committer,
		Message:      msg,
		TreeHash:     tree,
		ParentHashes: opts.Parents,
	}

	obj := w.r.Storer.NewEncodedObject()
	if err := commit.Encode(obj); err != nil {
		return plumbing.ZeroHash, err
	}
	return w.r.Storer.SetEncodedObject(obj)
}

// commitIndexHelper converts a given index.Index file into multiple git objects
// reading the blobs from the given filesystem and creating the trees from the
// index structure. The created objects are pushed to a given Storer.
type commitIndexHelper struct {
	fs billy.Filesystem
	s  storage.Storer

	trees   map[string]*object.Tree
	entries map[string]*object.TreeEntry
}

// buildTreesAndBlobs builds the objects and push its to the storer, the hash
// of the root tree is returned.
func (h *commitIndexHelper) buildTreeAndBlobObjects(idx *index.Index) (plumbing.Hash, error) {
	const rootNode = ""
	h.trees = map[string]*object.Tree{rootNode: {}}
	h.entries = map[string]*object.TreeEntry{}

	for _, e := range idx.Entries {
		if err := h.commitIndexEntry(e); err != nil {
			return plumbing.ZeroHash, err
		}
	}

	return h.copyTreeToStorageRecursive(rootNode, h.trees[rootNode])
}

func (h *commitIndexHelper) commitIndexEntry(e *index.Entry) error {
	parts := strings.Split(e.Name, string(filepath.Separator))

	var path string
	for _, part := range parts {
		parent := path
		path = filepath.Join(path, part)

		if !h.buildTree(e, parent, path) {
			continue
		}

		if err := h.copyIndexEntryToStorage(e); err != nil {
			return err
		}
	}

	return nil
}

func (h *commitIndexHelper) buildTree(e *index.Entry, parent, path string) bool {
	if _, ok := h.trees[path]; ok {
		return false
	}

	if _, ok := h.entries[path]; ok {
		return false
	}

	te := object.TreeEntry{Name: filepath.Base(path)}

	if path == e.Name {
		te.Mode = e.Mode
		te.Hash = e.Hash
	} else {
		te.Mode = filemode.Dir
		h.trees[path] = &object.Tree{}
	}

	h.trees[parent].Entries = append(h.trees[parent].Entries, te)
	return true
}

func (h *commitIndexHelper) copyIndexEntryToStorage(e *index.Entry) error {
	_, err := h.s.EncodedObject(plumbing.BlobObject, e.Hash)
	if err == nil {
		return nil
	}

	if err != plumbing.ErrObjectNotFound {
		return err
	}

	return h.doCopyIndexEntryToStorage(e)
}

func (h *commitIndexHelper) doCopyIndexEntryToStorage(e *index.Entry) (err error) {
	fi, err := h.fs.Stat(e.Name)
	if err != nil {
		return err
	}

	obj := h.s.NewEncodedObject()
	obj.SetType(plumbing.BlobObject)
	obj.SetSize(fi.Size())

	reader, err := h.fs.Open(e.Name)
	if err != nil {
		return err
	}

	defer ioutil.CheckClose(reader, &err)

	writer, err := obj.Writer()
	if err != nil {
		return err
	}

	defer ioutil.CheckClose(writer, &err)

	if _, err := io.Copy(writer, reader); err != nil {
		return err
	}

	_, err = h.s.SetEncodedObject(obj)
	return err
}

func (h *commitIndexHelper) copyTreeToStorageRecursive(parent string, t *object.Tree) (plumbing.Hash, error) {
	for i, e := range t.Entries {
		if e.Mode != filemode.Dir && !e.Hash.IsZero() {
			continue
		}

		path := filepath.Join(parent, e.Name)

		var err error
		e.Hash, err = h.copyTreeToStorageRecursive(path, h.trees[path])
		if err != nil {
			return plumbing.ZeroHash, err
		}

		t.Entries[i] = e
	}

	o := h.s.NewEncodedObject()
	if err := t.Encode(o); err != nil {
		return plumbing.ZeroHash, err
	}

	return h.s.SetEncodedObject(o)
}
