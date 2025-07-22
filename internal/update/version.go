package update

import "fmt"

var GithubRepo string = "xeyossr/anitr-cli"
var githubApi string = fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", GithubRepo)
var repoLink string = fmt.Sprintf("https://github.com/%s", GithubRepo)
var CurrentVersion string = "v4.1.1"
