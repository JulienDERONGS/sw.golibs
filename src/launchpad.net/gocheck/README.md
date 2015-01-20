Fork of `gocheck` from launchpad

~~~
$ mkdir repo.git
$ cd repo.git
$ git init
$ bzr fast-export --plain ../repo.bzr | git fast-import
$ git reset --hard
~~~
