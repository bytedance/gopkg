# How to Contribute
## Code of Conduct
Adapted from the [Contributor Covenant](https://www.contributor-covenant.org/)

## Your First Pull Request
We use github for our codebase. You can start by reading [How To Pull Request](https://docs.github.com/en/github/collaborating-with-issues-and-pull-requests/about-pull-requests).

## Semantic Versioning
We use [gomod](https://golang.org/ref/mod) as our dependencies manager, also follow the [semantic versioning](https://semver.org/). For better experience, when we make breaking changes, we also introduce deprecation warnings in a minor version so that our users learn about the upcoming changes and migrate their code in advance.
## Branch Organization
We use [git-flow](https://nvie.com/posts/a-successful-git-branching-model/) as our branch organization, as known as [FDD](https://en.wikipedia.org/wiki/Feature-driven_development)

## Bugs
### 1. How to Find Known Issues
We are using [Github Issues](https://github.com/bytedance/gopkg/issues) for our public bugs. We keep a close eye on this and try to make it clear when we have an internal fix in progress. Before filing a new task, try to make sure your problem doesn’t already exist.

### 2. Reporting New Issues
Providing a reduced test code is a recommended way for reporting issues. Then can placed in:
. Just in issues
. [Golang Playground](https://play.golang.org/)

### 3. Security Bugs
Please do not report the safe disclosure of bugs to public issues. Contact us by [Support Email](mailto:gopkg@bytedance.com)

## How to Get in Touch
- [Email](mailto:gopkg@bytedance.com)

## Submit a Pull Request
Before you submit your Pull Request (PR) consider the following guidelines:
1. Search [GitHub](https://github.com/bytedance/gopkg/pulls) for an open or closed PR that relates to your submission. You don't want to duplicate existing efforts.
2. 
3. Be sure that an issue describes the problem you're fixing, or documents the design for the feature you'd like to add. Discussing the design upfront helps to ensure that we're ready to accept your work.
4. [Fork](https://docs.github.com/en/github/getting-started-with-github/fork-a-repo) the angular/angular repo.
5. In your forked repository, make your changes in a new git branch:
    ```
    git checkout -b my-fix-branch develop
    ```
6. Create your patch, including appropriate test cases.
7. Follow our [Style Guides](#style).
8. Commit your changes using a descriptive commit message that follows [AngularJS Git Commit Message Conventions](https://docs.google.com/document/d/1QrDFcIiPjSLDn3EL15IJygNPiHORgU1_OOAqWjiDU5Y/edit).
   Adherence to these conventions is necessary because release notes are automatically generated from these messages.
9. Push your branch to GitHub:
    ```
    git push origin my-fix-branch
    ```
10. In GitHub, send a pull request to `gopkg:develop`

## Contribution Prerequisites
- Our code has been fully test with [Go](https://golang.org/) v1.15.0+, so you have installed at v1.15.0+.
- You need lint tools check before submit your pull request. [gofmt](https://golang.org/pkg/cmd/gofmt/) and [golangci-lint](https://github.com/golangci/golangci-lint)
- You are familiar with [Github](https://github.com) 
- Maybe you need familiar with [Actions](https://github.com/features/actions)(our default workflow tool).

## <a name="style"></a> Code Style Guides
Also see [Pingcap General advice](https://pingcap.github.io/style-guide/general.html).

Good resources:
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
