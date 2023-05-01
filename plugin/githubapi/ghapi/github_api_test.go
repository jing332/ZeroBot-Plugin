package ghapi

import (
	"io"
	"strings"
	"testing"
)

var token = "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"
var api = NewGithubAPI(token, "jing332", "tts-server-android")

var runId = 4836488819

func TestGithubAPI_GetArtifactInfo(t *testing.T) {
	url, err := api.GetArtifactFile(671692321, ArtifactFormatZip)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(url)
}

func TestGithubAPI_GetArtifacts(t *testing.T) {
	ret, err := api.GetArtifacts(runId)
	if err != nil {
		t.Fatal(err)
	}

	for i, artifact := range ret.Artifacts {
		t.Logf("%d - %d %s %dMB = %s", i, artifact.Id, artifact.Name, artifact.SizeInBytes/1024/1024, artifact.ArchiveDownloadUrl)
		if strings.HasPrefix(artifact.Name, "TTS-Server_") {
			resp, err := api.DownloadArtifactFromUrl(artifact.ArchiveDownloadUrl)
			if err != nil {
				t.Fatal(err)
			}
			bytes, err := io.ReadAll(resp.Body)
			if err != nil {
				resp.Body.Close()
				t.Fatal(err)
			}
			t.Log(len(bytes))

			resp.Body.Close()
		}
	}
}

func TestGithubAPI_WorkflowRuns(t *testing.T) {
	result, err := api.GetWorkflowRuns("test.yml")
	if err != nil {
		t.Fatal(err)
	}

	for i, run := range result.WorkflowRuns {
		t.Logf("%d - %s|%s: %s", i, run.Status, run.Conclusion, run.DisplayTitle)
	}
}
