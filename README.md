# git-br

A very simple interactive cli tool to handle your local git branches. Currently in alpha stage.

![output](https://user-images.githubusercontent.com/349328/26901406-116d2b6e-4bd6-11e7-8116-7bfa211cd25e.gif)

I wrote it because normally I work with tons of local branches and I needed a better way to handle them than `git branch`.

I know [tig](https://github.com/jonas/tig) has a branches (refs) view, but I wanted to try something to show the changes easier.

## install

Download, compile and install with `go get -u github.com/gonzaloserrano/git-br/cmd/git-br`.

If you don't have `$GOPATH/bin` in your `$PATH`, you can for e.g `$ cp $GOPATH/bin/git-br /usr/local/bin`.

## use

Type `git br` in your repo or provide a path as a first argument.

## todo

- [ ] use colored labels for distinguish diff added, modified, deleted
- [ ] if no path is passed, and we are not in the .git dir, look it up in the parent tree
- [ ] display repo path
- [ ] highlight master/develop branches
- [ ] use enter to switch and quite, shift-enter to just switch
- [ ] add delete feature
- [ ] improve performance when moving between lines, add delay + cancelation
- [ ] show origin branches like tig
- [ ] add delete all branches that are already merged feature
- [ ] add live mode: use a goroutine to refresh branches
- [ ] remove columnize dep

## license

This program is released under the Apache 2.0 license, see LICENSE file.
