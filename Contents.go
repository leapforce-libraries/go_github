package go_github

import (
	"fmt"
	errortools "github.com/leapforce-libraries/go_errortools"
	go_http "github.com/leapforce-libraries/go_http"
	"net/http"
)

type Contents struct {
	Name        string  `json:"name"`
	Path        string  `json:"path"`
	Sha         string  `json:"sha"`
	Size        int     `json:"size"`
	Url         string  `json:"url"`
	HtmlUrl     string  `json:"html_url"`
	GitUrl      string  `json:"git_url"`
	DownloadUrl *string `json:"download_url"`
	Type        string  `json:"type"`
	Links       struct {
		Self string `json:"self"`
		Git  string `json:"git"`
		Html string `json:"html"`
	} `json:"_links"`
}

type GetContentsConfig struct {
	Owner string
	Repo  string
	Path  string
}

func (service *Service) GetContents(cfg *GetContentsConfig) (*[]Contents, *errortools.Error) {
	var contents []Contents

	requestConfig := go_http.RequestConfig{
		Method:        http.MethodGet,
		Url:           service.url(fmt.Sprintf("repos/%s/%s/contents/%s", cfg.Owner, cfg.Repo, cfg.Path)),
		ResponseModel: &contents,
	}
	_, _, e := service.httpRequest(&requestConfig)
	if e != nil {
		return nil, e
	}

	return &contents, nil
}
