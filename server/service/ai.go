package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"server/core/db"
	"server/model"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/1szt/go-pkg/env"
)

// ------------------------ MatchGoods AI 商品匹配 ------------------------
// 前端传自然语言点单文字 → 后端取库中商品（id+名称）拼提示词 → 调大模型 →
// 解析返回的 JSON 数组 [{"id":"1","qty":2}] → 校验后回传命中的商品完整信息，
// 前端据此把商品加入购物车。

const matchTextMaxRunes = 200 // 点单文字长度上限（字符）

type MatchGoodsInput struct {
	Text string `json:"text"` // 用户自然语言点单描述，如「两杯柠檬水」
}
type MatchedGoods struct {
	GoodsView
	Qty int `json:"qty"` // 匹配到的购买数量
}
type MatchGoodsOutput struct {
	List []MatchedGoods `json:"list"` // 命中商品（空数组 = 没匹配到）
}

// matchPair 大模型返回的单个匹配项（id 可能是字符串或数字，用 RawMessage 兼容两种）
type matchPair struct {
	ID  json.RawMessage `json:"id"`
	Qty int             `json:"qty"`
}

// fenceRe 剥掉模型可能包裹的 ```json 代码块
var fenceRe = regexp.MustCompile("(?s)```(?:json)?\\s*(.*?)```")

func MatchGoods(i MatchGoodsInput) (o MatchGoodsOutput, err error) {
	o.List = make([]MatchedGoods, 0)

	text := strings.TrimSpace(i.Text)
	if text == "" {
		return o, errors.New("请输入点单内容")
	}
	if utf8.RuneCountInString(text) > matchTextMaxRunes {
		return o, fmt.Errorf("点单内容不能超过 %d 个字", matchTextMaxRunes)
	}
	apiKey := env.Get("LFDD_AI_API_KEY")
	if apiKey == "" {
		return o, errors.New("未配置大模型：请在 data/.env 设置 LFDD_AI_API_KEY")
	}

	// 1. 取上架商品列表作为提示词素材
	var goods []model.Goods
	if err = db.DB.Where("status = ?", 1).Order("sort ASC, id DESC").Find(&goods).Error; err != nil {
		return o, err
	}
	if len(goods) == 0 {
		return o, errors.New("暂无上架商品")
	}
	byID := make(map[int64]model.Goods, len(goods))
	var list strings.Builder
	for _, g := range goods {
		line := fmt.Sprintf("- id=%d 名称=%s", g.ID, g.Name)
		if g.Spec != "" {
			line += " 规格=" + g.Spec
		}
		list.WriteString(line + "\n")
		byID[g.ID] = g
	}

	// 2. 调大模型（OpenAI 兼容 /chat/completions）
	system := "你是点餐匹配助手。根据用户描述从商品列表中挑选商品并给出数量。" +
		"只输出一个 JSON 数组，元素格式 {\"id\":\"商品id\",\"qty\":购买数量}；" +
		"id 必须来自商品列表，qty 为 1-99 的整数；没有匹配的商品时输出 []。" +
		"不要输出任何其他文字、解释或代码块标记。"
	user := fmt.Sprintf("商品列表：\n%s\n用户点单：%s", list.String(), text)

	content, err := chatCompletion(apiKey, system, user)
	if err != nil {
		return o, err
	}

	// 3. 解析并校验匹配结果（只回传库里真实存在的上架商品）
	pairs, err := parseMatchJSON(content)
	if err != nil {
		return o, errors.New("大模型返回格式异常，请重试")
	}

	merged := map[int64]int{} // id → 数量（合并重复项）
	order := []int64{}
	for _, p := range pairs {
		id, e := parseMatchID(p.ID)
		if e != nil || p.Qty <= 0 {
			continue
		}
		g, ok := byID[id]
		if !ok {
			continue // 模型编造的 id 直接丢弃
		}
		if _, seen := merged[id]; !seen {
			order = append(order, id)
		}
		merged[id] += p.Qty
		if g.Stock > 0 && merged[id] > g.Stock {
			merged[id] = g.Stock // 数量封顶到库存
		}
	}

	for _, id := range order {
		g := byID[id]
		o.List = append(o.List, MatchedGoods{
			GoodsView: GoodsView{
				ID: g.ID, MenuID: g.MenuID, Name: g.Name, Price: g.Price,
				Spec: g.Spec, Stock: g.Stock, Image: g.Image, Sort: g.Sort, Status: g.Status,
			},
			Qty: merged[id],
		})
	}
	return o, nil
}

// chatCompletion 调 OpenAI 兼容的 chat/completions，返回首个 choice 的 content
func chatCompletion(apiKey, system, user string) (string, error) {
	baseURL := env.Get("LFDD_AI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}
	model := env.Get("LFDD_AI_MODEL")
	if model == "" {
		model = "deepseek-chat"
	}

	body, _ := json.Marshal(map[string]any{
		"model":       model,
		"temperature": 0,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(baseURL, "/")+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.New("大模型请求失败，请检查网络")
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 上限 1MB
	if err != nil {
		return "", err
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err = json.Unmarshal(data, &out); err != nil {
		return "", errors.New("大模型响应解析失败")
	}
	if out.Error != nil {
		return "", fmt.Errorf("大模型错误: %s", out.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("大模型请求失败（%d）", resp.StatusCode)
	}
	if len(out.Choices) == 0 {
		return "", errors.New("大模型未返回内容")
	}
	return out.Choices[0].Message.Content, nil
}

// parseMatchJSON 解析模型输出为匹配项数组（容忍代码块包裹/前后杂散文字）
func parseMatchJSON(content string) ([]matchPair, error) {
	s := strings.TrimSpace(content)
	if m := fenceRe.FindStringSubmatch(s); m != nil {
		s = strings.TrimSpace(m[1])
	}
	// 截取首个 [ ... ] 区间，规避模型在数组前后加文字
	if start := strings.IndexByte(s, '['); start >= 0 {
		if end := strings.LastIndexByte(s, ']'); end > start {
			s = s[start : end+1]
		}
	}
	var pairs []matchPair
	if err := json.Unmarshal([]byte(s), &pairs); err != nil {
		return nil, err
	}
	return pairs, nil
}

// parseMatchID 兼容模型把 id 输出成字符串 "1" 或数字 1
func parseMatchID(raw json.RawMessage) (int64, error) {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		var id int64
		_, e := fmt.Sscanf(s, "%d", &id)
		return id, e
	}
	var id int64
	err := json.Unmarshal(raw, &id)
	return id, err
}
