# Contributing to MASA golibs

## Setup

MASA developers use what is commonly referenced to as a "[triangular workflow](https://github.blog/2015-07-29-git-2-5-including-multiple-worktrees-and-triangular-workflows/#improved-support-for-triangular-workflows)".

This is a widely used model where every contributing developer is expected to possess its own [fork](https://help.github.com/en/github/getting-started-with-github/fork-a-repo) of the [corporate repository](https://github.com/masagroup/sw.golibs) and where all his work should take place.

You will also need to make sure the [swbuilder](https://github.com/swbuilder) bot has a READ access on your fork, in order to build and test your code on our CI server. This can be done in the _access_ section of your fork's settings, on GitHub. If you are a direct member of the [MASA group organization](https://github.com/masagroup) no action is required.

## GitHub workflow

1. Create a dedicated branch on your fork

    - If you plan to work on an issue or a story that has been logged into JIRA, the name of the branch should be prefixed by its ID

    - Most of the time this branch will be originated from `masa/master`

    - Example:
        > `git checkout -b SWNG-64-xxxx`

2. Push your work to this new branch

    - Commits created on this branch should reference any relevant JIRA page

    - Commit messages should be prefixed by the name of the packages they impacts (example "[extractor,ts]")

    - Commit messages should be postfixed by the time it took to complete the commit in effort points spent (example [EPS:2])

    - If this is a long-lived branch (e.g. a long-haul refactoring), regularly rebase it onto its origin branch

    - Example:
        > `git commit -m "[extractor,ts] SWNG-64: fixes out of bounds error when accessing steamed buffer"`

3. When your work is ready to be submitted, open a Pull Request (PR) requesting a merge of this branch into the `master` branch of the corporate repository

    - Make sure your branch is up to date with the target branch

    - If relevant, the title of your PR should reference the issue or story it is addressing

    - This PR should contain only the minimum code necessary to tackle the problem at hand (typically, no merge commits)

    - Please provide a succinct description of what the PR does. Do not hesitate the emphasize anything your reviewers should be aware of

    - If you think you can use some input, feel free to open a [Draft PR](https://github.blog/2019-02-14-introducing-draft-pull-requests/). This will allow others to review your submission but will block merging until you decide otherwise

    - Keep an eye on your PR afterwards and rebase your branch regularly if needed. Don't expect anyone to review a PR with conflicts

    - Example: https://github.com/masagroup/sw.webclient/pull/16

## Coding style

We are using gofmt by default to format all go code.

## Doubts?

Just ask :wink:
