package service

import (
	"context"
	"github.com/google/go-github/v48/github"
	"github.com/hamster-shared/a-line/pkg/utils"
	"log"
)

type IGithubService interface {
	CheckName(owner, projectName string) (bool, error)
	CreateRepo(templateOwner, templateRepo, repoName, repoOwner string) error
}

type GithubService struct {
	ctx context.Context
}

func NewGithubService() *GithubService {
	return &GithubService{
		ctx: context.Background(),
	}
}

func (g *GithubService) CheckName(token, owner, projectName string) bool {
	client := utils.NewGithubClient(g.ctx, token)
	_, res, _ := client.Repositories.Get(g.ctx, owner, projectName)
	if res.StatusCode == 404 {
		return true
	}
	return false
}

func (g *GithubService) CreateRepo(token, templateOwner, templateRepo, repoName, repoOwner string) (*github.Repository, *github.Response, error) {
	client := utils.NewGithubClient(g.ctx, token)
	var data github.TemplateRepoRequest
	data.Name = &repoName
	data.Owner = &repoOwner
	repo, res, err := client.Repositories.CreateFromTemplate(g.ctx, templateOwner, templateRepo, &data)
	if err != nil {
		log.Println("create github repository failed ", err.Error())
		return nil, res, err
	}
	return repo, res, nil
}

func (g *GithubService) GetUserInfo(token string) (*github.User, error) {
	client := utils.NewGithubClient(g.ctx, token)
	user, _, err := client.Users.Get(g.ctx, "")
	if err != nil {
		log.Println("get github user info failed ", err.Error())
		return nil, err
	}
	return user, nil
}