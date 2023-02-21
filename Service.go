package go_github

import (
	"fmt"
	"github.com/leapforce-libraries/go_oauth2/tokensource"
	"net/http"
	"time"

	errortools "github.com/leapforce-libraries/go_errortools"
	go_http "github.com/leapforce-libraries/go_http"
	oauth2 "github.com/leapforce-libraries/go_oauth2"
	go_token "github.com/leapforce-libraries/go_oauth2/token"
)

const (
	apiName            string = "Github"
	apiUrl             string = "https://api.github.com"
	authUrl            string = "https://github.com/login/oauth/authorize"
	tokenUrl           string = "https://github.com/login/oauth/access_token"
	tokenHttpMethod    string = http.MethodPost
	defaultRedirectUrl string = "http://localhost:8080/oauth/redirect"
)

type authorizationMode string

const (
	authorizationModeOAuth2      authorizationMode = "oauth2"
	authorizationModeAccessToken authorizationMode = "accesstoken"
)

type ServiceWithOAuth2Config struct {
	ClientId      string
	ClientSecret  string
	TokenSource   tokensource.TokenSource
	RedirectUrl   *string
	RefreshMargin *time.Duration
}

type Service struct {
	authorizationMode authorizationMode
	clientId          string
	accessToken       *string
	httpService       *go_http.Service
	oAuth2Service     *oauth2.Service
	errorResponse     *ErrorResponse
}

func NewServiceWithOAuth2(serviceConfig *ServiceWithOAuth2Config) (*Service, *errortools.Error) {
	if serviceConfig == nil {
		return nil, errortools.ErrorMessage("ServiceConfig must not be a nil pointer")
	}

	if serviceConfig.ClientId == "" {
		return nil, errortools.ErrorMessage("ClientId not provided")
	}

	redirectUrl := defaultRedirectUrl
	if serviceConfig.RedirectUrl != nil {
		redirectUrl = *serviceConfig.RedirectUrl
	}

	oAuth2ServiceConfig := oauth2.ServiceConfig{
		ClientId:        serviceConfig.ClientId,
		ClientSecret:    serviceConfig.ClientSecret,
		RedirectUrl:     redirectUrl,
		AuthUrl:         authUrl,
		TokenUrl:        tokenUrl,
		RefreshMargin:   serviceConfig.RefreshMargin,
		TokenHttpMethod: tokenHttpMethod,
		TokenSource:     serviceConfig.TokenSource,
	}
	oauth2Service, e := oauth2.NewService(&oAuth2ServiceConfig)
	if e != nil {
		return nil, e
	}

	return &Service{
		authorizationMode: authorizationModeOAuth2,
		clientId:          serviceConfig.ClientId,
		oAuth2Service:     oauth2Service,
	}, nil
}

type ServiceWithAccessTokenConfig struct {
	AccessToken string
}

func NewServiceWithAccessToken(cfg *ServiceWithAccessTokenConfig) (*Service, *errortools.Error) {
	if cfg == nil {
		return nil, errortools.ErrorMessage("ServiceConfig must not be a nil pointer")
	}

	if cfg.AccessToken == "" {
		return nil, errortools.ErrorMessage("AccessToken not provided")
	}

	httpService, e := go_http.NewService(&go_http.ServiceConfig{})
	if e != nil {
		return nil, e
	}

	return &Service{
		authorizationMode: authorizationModeAccessToken,
		accessToken:       &cfg.AccessToken,
		httpService:       httpService,
	}, nil
}

func (service *Service) url(path string) string {
	return fmt.Sprintf("%s/%s", apiUrl, path)
}

func (service *Service) httpRequest(requestConfig *go_http.RequestConfig) (*http.Request, *http.Response, *errortools.Error) {
	var request *http.Request
	var response *http.Response
	var e *errortools.Error

	// add error model
	service.errorResponse = &ErrorResponse{}
	requestConfig.ErrorModel = service.errorResponse

	header := http.Header{}
	header.Set("Accept", "application/vnd.github+json")

	if service.authorizationMode == authorizationModeOAuth2 {
		requestConfig.NonDefaultHeaders = &header
		request, response, e = service.oAuth2Service.HttpRequest(requestConfig)
	} else if service.authorizationMode == authorizationModeAccessToken {
		// add accesstoken to header
		header.Set("Authorization", fmt.Sprintf("Bearer %s", *service.accessToken))
		requestConfig.NonDefaultHeaders = &header

		request, response, e = service.httpService.HttpRequest(requestConfig)
	}

	if e != nil {
		if service.errorResponse.Message != "" {
			e.SetMessage(service.errorResponse.Message)
		}
	}

	if e != nil {
		return request, response, e
	}

	return request, response, nil
}

func (service *Service) AuthorizeUrl(scope string, state *string) string {
	return service.oAuth2Service.AuthorizeUrl(scope, nil, nil, state)
}

func (service *Service) ValidateToken() (*go_token.Token, *errortools.Error) {
	return service.oAuth2Service.ValidateToken()
}

func (service *Service) GetTokenFromCode(r *http.Request) *errortools.Error {
	return service.oAuth2Service.GetTokenFromCode(r, nil)
}

func (service Service) ApiName() string {
	return apiName
}

func (service Service) ApiKey() string {
	return service.clientId
}

func (service Service) ApiCallCount() int64 {
	return service.oAuth2Service.ApiCallCount()
}

func (service Service) ApiReset() {
	service.oAuth2Service.ApiReset()
}
