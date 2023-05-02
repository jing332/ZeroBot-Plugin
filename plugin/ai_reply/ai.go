package aireply

import (
	"errors"
	"github.com/FloatTech/AnimeAPI/tts"
	"github.com/FloatTech/AnimeAPI/tts/genshin"
	aireply "github.com/FloatTech/ZeroBot-Plugin/plugin/ai_reply/aireply-api"
	ctrl "github.com/FloatTech/zbpctrl"
	zero "github.com/wdvxdr1123/ZeroBot"
)

// 数据结构: [4 bits] [4 bits] [8 bits] [8 bits]
// 			[ttscn模式] [百度模式] [tts模式] [回复模式]

// defaultttsindexkey
// 数据结构: [4 bits] [4 bits] [8 bits]
// 			[ttscn模式] [百度模式] [tts模式]

// [tts模式]: 0~63 genshin 64 baidu 65 ttscn

const (
	lastgsttsindex = 63 + iota
	baiduttsindex
	ttscnttsindex
)

// extrattsname is the tts other than genshin vits
var extrattsname = []string{"百度", "TTSCN"}

var ttscnspeakers = [...]string{
	"晓晓（女 - 年轻人）",
	"云扬（男 - 年轻人）",
	"晓辰（女 - 年轻人 - 抖音热门）",
	"晓涵（女 - 年轻人）",
	"晓墨（女 - 年轻人）",
	"晓秋（女 - 中年人）",
	"晓睿（女 - 老年）",
	"晓双（女 - 儿童）",
	"晓萱（女 - 年轻人）",
	"晓颜（女 - 年轻人）",
	"晓悠（女 - 儿童）",
	"云希（男 - 年轻人 - 抖音热门）",
	"云野（男 - 中年人）",
	"晓梦（女 - 年轻人）",
	"晓伊（女 - 儿童）",
	"晓甄（女 - 年轻人）",
}


var (
	原  = newapikeystore("./data/tts/o.txt")
	ཆཏ = newapikeystore("./data/tts/c.txt")
	百  = newapikeystore("./data/tts/b.txt")
)

type replymode []string

func (r replymode) setReplyMode(ctx *zero.Ctx, name string) error {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	var ok bool
	var index int64
	for i, s := range r {
		if s == name {
			ok = true
			index = int64(i)
			break
		}
	}
	if !ok {
		return errors.New("no such mode")
	}
	m, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
	if !ok {
		return errors.New("no such plugin")
	}
	return m.SetData(gid, (m.GetData(index)&^0xff)|(index&0xff))
}

func (r replymode) getReplyMode(ctx *zero.Ctx) aireply.AIReply {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	m, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
	if ok {
		switch m.GetData(gid) & 0xff {
		case 0:
			return aireply.NewQYK(aireply.QYKURL, aireply.QYKBotName)
		case 1:
			return aireply.NewXiaoAi(aireply.XiaoAiURL, aireply.XiaoAiBotName)
		case 2:
			k := ཆཏ.k
			if k != "" {
				return aireply.NewChatGPT(aireply.ChatGPTURL, k)
			}
			return aireply.NewQYK(aireply.QYKURL, aireply.QYKBotName)
		}
	}
	return aireply.NewQYK(aireply.QYKURL, aireply.QYKBotName)
}

var ttsins = func() map[string]tts.TTS {
	m := make(map[string]tts.TTS, 128)
	for _, mode := range append(genshin.SoundList[:], extrattsname...) {
		m[mode] = nil
	}
	return m
}()

var ttsModes = func() []string {
	s := append(genshin.SoundList[:], make([]string, 64-len(genshin.SoundList))...) // 0-63
	s = append(s, extrattsname...)                                                  // 64 65 ...
	return s
}()

func list(list []string, num int) string {
	s := ""
	for i, value := range list {
		s += value
		if (i+1)%num == 0 {
			s += "\n"
		} else {
			s += " | "
		}
	}
	return s
}
