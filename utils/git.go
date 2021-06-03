package utils

import (
	"os"
	"strings"

	Log "github.com/wellmoon/go/logger"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

var GIT_ROOT = "https://codeup.aliyun.com/5f8cff393035265285849090/leridge/"

func Clone(project string, clonePath string, branch string, gitUsername string, gitPassword string) string {
	if len(branch) == 0 {
		branch = "master"
	}
	allGitPath := GIT_ROOT + project + ".git"
	Log.Debug("开始从git下载[%v]", project)
	r, err := git.PlainClone(clonePath+project, false, &git.CloneOptions{
		URL:  allGitPath,
		Auth: &http.BasicAuth{Username: gitUsername, Password: gitPassword},
		// RemoteName:        "",
		ReferenceName:     plumbing.NewBranchReferenceName(branch),
		SingleBranch:      true,
		NoCheckout:        false,
		Depth:             1,
		RecurseSubmodules: 0,
		Progress:          os.Stdout,
		Tags:              0,
		InsecureSkipTLS:   false,
		CABundle:          []byte{},
	})

	if err != nil {
		if strings.Contains(err.Error(), "couldn't find remote ref") {
			// 如果分支不存在，更新master分支
			return "branch not exist"
		}
		Log.Error("error:%v", err)
	}
	Log.Debug("下载[%v]完成", project)
	ref, err := r.Head()
	if err != nil {
		Log.Error("Head error, %v", err)
		return ""
	}
	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		Log.Error("CommitObject error, %v", err)
		return ""
	}
	Log.Trace("[%v]最后一次提交为：\n[%v]", project, commit)
	return "success"
}

func Pull(project string, repository string, gitUsername string, gitPassword string) bool {
	//allGitPath := GIT_ROOT + project + ".git"
	// We instantiate a new repository targeting the given path (the .git folder)
	r, err := git.PlainOpen(repository)
	if err != nil {
		Log.Error("PlainOpen error, %v", err)
		return false
	}

	// Get the working directory for the repository
	w, err := r.Worktree()
	if err != nil {
		Log.Error("Worktree error, %v", err)
		return false
	}

	// Pull the latest changes from the origin remote and merge into the current branch
	Log.Debug("准备从git更新[%v]", project)
	err = w.Pull(&git.PullOptions{
		RemoteName: "origin",
		Depth:      1,
		Auth:       &http.BasicAuth{Username: gitUsername, Password: gitPassword},
	})
	if err != nil {
		if !strings.Contains(err.Error(), "already up-to-date") {
			Log.Error("Pull error, %v", err)
			return false
		} else {
			Log.Trace("[%v]已经是最新代码，无需更新(pull)", project)
		}
	} else {
		Log.Debug("[%v]更新(pull)完毕", project)
	}

	// Print the latest commit that was just pulled
	ref, err := r.Head()
	if err != nil {
		Log.Error("Head error, %v", err)
		return false
	}
	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		Log.Error("CommitObject error, %v", err)
		return false
	}

	Log.Trace("[%v]最后一次提交为：\n[%v]", project, commit)
	return true
}
