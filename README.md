# git-br

A simple interactive cli tool to handle your local git branches.

I wrote it because normally I work with tons of local branches and I needed a better way to handle them than `git branch`.

## install

Download, compile and install with `go get -u github.com/gonzaloserrano/git-br/cmd/git-br`.

If you don't have `$GOPATH/bin` in your `$PATH`, you can for e.g `$ cp $GOPATH/bin/git-br /usr/local/bin`.

## use

Type `git br` in your repo or provide a path as a first argument.

## todo

- [ ] use shift-enter to switch branch and quit
- [ ] highlight master/develop branches
- [ ] display file information, e.g diff with master
- [ ] remove columnize dep
- [ ] add live mode: use a goroutine to refresh branches
- [ ] add delete feature
- [ ] add delete all branches that are already merged feature

## license

This program is released under the Apache 2.0 license, see LICENSE file.
