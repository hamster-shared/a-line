package service

import (
	"errors"
	"fmt"
	"github.com/hamster-shared/a-line/pkg/application"
	"github.com/hamster-shared/a-line/pkg/consts"
	db2 "github.com/hamster-shared/a-line/pkg/db"
	"github.com/hamster-shared/a-line/pkg/parameter"
	"github.com/hamster-shared/a-line/pkg/utils"
	"gorm.io/gorm"
	"log"
	"time"
)

type ILoginService interface {
}

type LoginService struct {
	db            *gorm.DB
	githubService *GithubService
}

func NewLoginService() *LoginService {
	return &LoginService{
		db:            application.GetBean[*gorm.DB]("db"),
		githubService: application.GetBean[*GithubService]("githubService"),
	}
}

func (l *LoginService) LoginWithGithub(data parameter.LoginParam) (db2.User, error) {
	data.ClientSecret = consts.ClientSecrets
	var userData db2.User
	var token parameter.Token
	url := "https://github.com/login/oauth/access_token"
	res, err := utils.NewHttp().NewRequest().SetQueryParams(map[string]string{
		"client_id":     data.ClientId,
		"client_secret": data.ClientSecret,
		"code":          data.Code,
	}).SetResult(&token).SetHeader("Accept", "application/json").Post(url)
	if res.StatusCode() != 200 {
		return userData, err
	}
	if err != nil {
		return userData, err
	}
	userInfo, err := l.githubService.GetUserInfo(token.AccessToken)
	if err != nil {
		return userData, err
	}
	userData.Id = uint(*userInfo.ID)
	userData.Username = *userInfo.Login
	userData.AvatarUrl = *userInfo.AvatarURL
	userData.HtmlUrl = *userInfo.HTMLURL
	userData.CreateTime = time.Now()
	userData.Token = token.AccessToken
	l.db.Save(&userData)
	accessToken := utils.AesEncrypt(token.AccessToken, consts.SecretKey)
	userData.Token = accessToken
	return userData, nil
}

func (l *LoginService) GithubRepoAuth(authData parameter.AuthParam) (string, error) {
	authData.ClientSecret = consts.ClientSecrets
	var userData db2.User
	var token parameter.Token
	res := l.db.Model(db2.User{}).Where("id = ?", authData.UserId).First(&userData)
	if res.Error != nil {
		log.Println("login user not fond ", res.Error)
		return "", res.Error
	}

	url := "https://github.com/login/oauth/access_token"
	response, err := utils.NewHttp().NewRequest().SetQueryParams(map[string]string{
		"client_id":     authData.ClientId,
		"client_secret": authData.ClientSecret,
		"code":          authData.Code,
	}).SetResult(&token).SetHeader("Accept", "application/json").Post(url)
	if response.StatusCode() != 200 {
		log.Println(string(response.Body()))
		return "", errors.New(fmt.Sprintf("auth failed:%s", string(response.Body())))
	}
	if err != nil {
		log.Println("repo auth failed ", err)
		return "", err
	}
	userData.Token = token.AccessToken
	l.db.Save(&userData)
	accessToken := utils.AesEncrypt(token.AccessToken, consts.SecretKey)
	return accessToken, nil
}