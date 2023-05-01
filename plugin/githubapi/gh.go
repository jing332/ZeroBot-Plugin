// Package github GitHub 仓库搜索
package github

import (
	"fmt"
	"github.com/FloatTech/ZeroBot-Plugin/plugin/githubapi/ghapi"
	"github.com/FloatTech/floatbox/file"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	json "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var debug = false

var (
	api      *ghapi.GithubAPI
	cfg      config
	cfgFile  string
	cacheDir string
)

func init() { // 插件主体
	engine := control.Register("github-api", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "Github相关接口",
		Help: "- gh set [token / repo] (超管私发)\n" +
			"- >gh action",
		PrivateDataFolder: "github-api",
	})
	cfgFile = engine.DataFolder() + "/config.json"
	cacheDir = engine.DataFolder() + "/cache"
	if file.IsExist(cacheDir) {
		err := os.MkdirAll(cacheDir, 0777)
		if err != nil {
			log.Errorf("[github-api] 创建缓存目录失败: %s", err)
			return
		}
	}
	reload()

	engine.OnPrefix("gh set token", zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		getConfigVar(ctx, "token", func(value string) bool {
			cfg.Token = value
			return true
		})
	})

	engine.OnPrefix("gh set repo", zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		getConfigVar(ctx, "repo", func(value string) bool {
			ss := strings.Split(value, "/")
			if len(ss) != 2 {
				ctx.SendChain(message.Text("格式错误，正确格式为：[owner]/[repo], \n例如: jing332/tts-server-android"))
				return false
			}
			cfg.Repo = value
			return true
		})
	})

	engine.OnRegex(`gh\s+action\s*(.*)$`, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		_ = os.RemoveAll(cacheDir)
		inputRepo := ctx.State["regex_matched"].([]string)[1]
		var repo = ""
		if strings.TrimSpace(inputRepo) == "" {
			repo = cfg.Repo
		} else {
			repo = inputRepo
		}

		ctx.SendChain(message.Text(fmt.Sprintf("Github Actions: \n%s", repo)))
		runs, err := api.GetWorkflowAllRuns(repo)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("加载失败: %s", err)))
			return
		}

		var msg = "请选择Action序号:\n"
		for i, run := range runs.WorkflowRuns {
			msg += fmt.Sprintf("%d - [%s]%s\n\n", i+1, run.Name, run.DisplayTitle)
		}
		ctx.SendChain(ctxext.FakeSenderForwardNode(ctx, message.Text(msg)))

		waitNumMsg(ctx, func(num int) bool {
			if num <= 0 || int(num) > len(runs.WorkflowRuns) {
				ctx.SendChain(message.Text("请输入正确的序号!"))
				return false
			}

			run := runs.WorkflowRuns[num-1]
			ctx.Send(fmt.Sprintf("获取Action产物: [%d]%s", run.RunNumber, run.DisplayTitle))
			arts, err := api.GetArtifacts(repo, run.Id)
			if err != nil {
				ctx.SendChain(message.Text(fmt.Sprintf("获取Action产物失败: %s", err)))
				return false
			}

			msg = ""
			for i, artifact := range arts.Artifacts {
				msg += fmt.Sprintf("%d - (%dMiB)%s\n", i+1, artifact.SizeInBytes/1024/1024, artifact.Name)
			}
			ctx.SendChain(ctxext.FakeSenderForwardNode(ctx, message.Text(msg)))
			waitNumMsg(ctx, func(num int) bool {
				if num <= 0 || num > len(arts.Artifacts) {
					ctx.SendChain(message.Text("请输入正确的序号!"))
					return false
				}

				art := arts.Artifacts[num-1]
				ctx.SendChain(message.Text(fmt.Sprintf("下载产物: [%d]%s", num, art.Name)))
				resp, err := api.DownloadArtifactFromUrl(art.ArchiveDownloadUrl)
				if err != nil {
					ctx.SendChain(message.Text(fmt.Sprintf("下载失败: %s", err)))
					return false
				}
				defer resp.Body.Close()
				bytes, err := io.ReadAll(resp.Body)
				if err != nil {
					ctx.SendChain(message.Text(fmt.Sprintf("下载失败: %s", err)))
					return true
				}

				absPath := filepath.Join(cacheDir, art.Name)
				if debug {
					absPath = "/data/data/com.termux/files/home/qq/download/" + art.Name
				}

				if debug {
					// 上传到SFTP 仅在调式时使用
					info, err := upload(bytes, "qq/download", art.Name)
					if err != nil {
						ctx.SendChain(message.Text(fmt.Sprintf("上传SFTP失败: %s", err)))
						return true
					}
					ctx.SendChain(message.Text(fmt.Sprintf("上传SFTP完成: %d", info.Size())))
				} else {
					_ = os.MkdirAll(absPath, 0666)
					err = os.WriteFile(absPath, bytes, 0666)
					if err != nil {
						ctx.SendChain(message.Text(fmt.Sprintf("下载失败: %s", err)))
						return true
					}
				}

				ctx.SendChain(message.Text(fmt.Sprintf("下载完成，上传到QQ: [%d]%s", num, art.Name)))
				res := ctx.UploadThisGroupFile(absPath, art.Name, "")
				if res.Msg != "" {
					ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("上传失败: %s", res.Msg)))
				}

				return true
			})

			return true
		})

	})
}

// 阻塞监听下一个序号消息
// onNum 返回true则结束监听
func waitNumMsg(ctx *zero.Ctx, onNum func(num int) bool) {
	next := zero.NewFutureEvent("message", 999, false, ctx.CheckSession(), zero.RegexRule(`^\d+$`))
	recv, cancel := next.Repeat()
	defer cancel()
	for {
		select {
		case <-time.After(time.Second * 30):
			ctx.SendChain(message.Text("指令过期"))
		case c := <-recv:
			msg := c.Event.Message.ExtractPlainText()
			num, err := strconv.Atoi(msg)
			if err != nil {
				ctx.SendChain(message.Text("请输入数字"))
			}

			if onNum(num) { // true
				return
			} else {
				continue
			}
		}
	}
}

func getConfigVar(ctx *zero.Ctx, varName string, onSetVar func(value string) bool) {
	ss := strings.Split(strings.TrimSpace(ctx.MessageString()), "gh set "+varName)
	if len(ss) > 0 {
		if onSetVar(strings.TrimSpace(ss[1])) {
			err := saveConfig(cfgFile)
			if err != nil {
				log.Warnln("GithubAPI-保存配置失败", err)
				return
			}
			reload()
			ctx.SendChain(message.Text("设置成功"))
		}
	} else {
		ctx.SendChain(message.Text("请在命令后面加上Token值"))
	}
}

// 保存用户配置
func saveConfig(cfgFile string) error {
	if reader, err := os.Create(cfgFile); err == nil {
		err = json.NewEncoder(reader).Encode(&cfg)
		if err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

func reload() {
	c := readConfigFile()
	if c == nil {
		log.Println("cfg == nil")
		cfg = config{}
	} else {
		cfg = *c
	}

	api = ghapi.NewGithubAPI(cfg.Token)
}

func readConfigFile() (cfg *config) {
	if reader, err := os.Open(cfgFile); err == nil {
		cfg = &config{}
		err = json.NewDecoder(reader).Decode(cfg)
		if err != nil {
			return nil
		}
	}

	return cfg
}

type config struct {
	// Token 为 GitHub PAT
	Token string `json:"token"`
	// Repo 为 GitHub 仓库名
	Repo string `json:"repo"`

	// Owner 为 GitHub 仓库所有者
	//Owner string `json:"owner"`
	//WorkflowName 为 GitHub Action 工作流文件名
	//WorkflowName string `json:"workflowName"`
}
