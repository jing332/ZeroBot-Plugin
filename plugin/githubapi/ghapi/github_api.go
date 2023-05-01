package ghapi

import (
	"fmt"
	json "github.com/json-iterator/go"
	"net/http"
)

type GithubAPI struct {
	Client *http.Client
	//Owner   string
	//Repo    string
	Token   string
	Version string
}

const (
	ghApiVersion = "2022-11-28"
)

func NewGithubAPI(token string) *GithubAPI {
	return &GithubAPI{
		Client: newClient(),
		//Repo:    repo,
		Token:   token,
		Version: ghApiVersion,
	}
}

func newClient() *http.Client {
	// 根据环境变量获取代理信息
	transport := &http.Transport{Proxy: http.ProxyFromEnvironment}
	return &http.Client{Transport: transport}
}

func (g *GithubAPI) httpGetResponse(url string) (*http.Response, error) {
	if g.Client == nil {
		g.Client = newClient()
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+g.Token)
	req.Header.Set("X-GitHub-Api-Version", g.Version)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (g *GithubAPI) baseApiUrl(repo string) string {
	return fmt.Sprintf("https://api.github.com/repos/%s", repo)
}

var (
	ArtifactFormatZip = "zip"
)

func (g *GithubAPI) GetArtifactFile(repo string, artifactId int, format string) (resp *http.Response, err error) {
	if format == "" {
		format = ArtifactFormatZip
	}
	return g.httpGetResponse(g.baseApiUrl(repo) + fmt.Sprintf("/actions/artifacts/%d/%s", artifactId, format))
}

func (g *GithubAPI) DownloadArtifactFromUrl(url string) (resp *http.Response, err error) {
	return g.httpGetResponse(url)
}

func (g *GithubAPI) GetArtifacts(repo string, runId int64) (result *ArtifactsResult, err error) {
	resp, err := g.httpGetResponse(g.baseApiUrl(repo) + fmt.Sprintf("/actions/runs/%d/artifacts", runId))
	if err != nil {
		return nil, err
	}

	result = &ArtifactsResult{}
	err = httpHandleResponse(resp, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func httpHandleResponse(resp *http.Response, obj interface{}) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败: %s", resp.Status)
	}

	err := json.NewDecoder(resp.Body).Decode(obj)
	if err != nil {
		return fmt.Errorf("解析JSON失败: %w", err)
	}
	return nil
}

// GetWorkflowRuns 获取工作流结果列表
func (g *GithubAPI) GetWorkflowRuns(repo, flowName string) (result *WorkFlowRunsResult, err error) {
	url := fmt.Sprintf(g.baseApiUrl(repo) + "/actions/workflows/" + flowName + "/runs")
	resp, err := g.httpGetResponse(url)

	result = &WorkFlowRunsResult{}
	err = httpHandleResponse(resp, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetWorkflowAllRuns 获取所有工作流结果
func (g *GithubAPI) GetWorkflowAllRuns(repo string) (result *WorkFlowRunsResult, err error) {
	url := fmt.Sprintf(g.baseApiUrl(repo) + "/actions/runs")
	resp, err := g.httpGetResponse(url)

	result = &WorkFlowRunsResult{}
	err = httpHandleResponse(resp, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
