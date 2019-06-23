package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-ini/ini"
	"github.com/gocolly/colly"
)

type ConfigList struct {
	LoginUrl      string
	LoginAddress  string
	LoginPassword string
	TrainingId    int
}

var Config ConfigList
var Args []string

func init() {
	// 設定ファイル読み込み
	conf, err := ini.Load("config.ini")
	if err != nil {
		log.Fatal("設定ファイルが見つかりません。")
	}
	Config = ConfigList{
		LoginUrl:      conf.Section("login_info").Key("login_url").MustString(""),
		LoginAddress:  conf.Section("login_info").Key("login_address").MustString(""),
		LoginPassword: conf.Section("login_info").Key("login_password").MustString(""),
		TrainingId:    conf.Section("summarize_score").Key("target_training_id").MustInt(),
	}

	// 実行時引数設定
	flag.Parse()
	Args = flag.Args()

	var args0 string
	if len(Args) > 0 {
		args0 = Args[0]
	} else {
		args0 = "無し"
	}

	// 設定情報を取得
	fmt.Printf("ログインページ：%s\nメールアドレス：%s\n研修ID：%d\n実行時引数：%s\n\n",
		Config.LoginUrl, Config.LoginAddress, Config.TrainingId, args0)
}

func main() {

	urlBase, _ := url.Parse(Config.LoginUrl)

	c := colly.NewCollector()
	// c.OnRequest(func(r *colly.Request) {
	// 	log.Println("リクエストURL: ", r.URL.String())
	// })
	c2 := c.Clone()
	c3 := c.Clone()
	c4 := c.Clone()

	// ログイン時に使用するトークン
	var token string

	// ログイン画面を取得したときにトークンを取得する
	c.OnHTML("input[name='_token']", func(e *colly.HTMLElement) {
		token = e.Attr("value")
		//log.Println("_token: ", token)
	})

	// ログイン画面にアクセス
	c.Visit(Config.LoginUrl)

	var trainingUrl string
	// 研修一覧画面取得時の処理を定義
	c2.OnHTML("table tr", func(e *colly.HTMLElement) {
		e.DOM.FilterFunction(func(_ int, el *goquery.Selection) bool {
			trainingId, _ := strconv.Atoi(el.Find("td:first-child").Text())
			return trainingId == Config.TrainingId
		}).Each(func(_ int, el *goquery.Selection) {
			trainingUrl, _ = el.Find("a").Attr("href")
			ref, _ := url.Parse(trainingUrl)
			trainingUrl = urlBase.ResolveReference(ref).String()
		})
	})
	// ログイン情報送信
	c2.Post(Config.LoginUrl, map[string]string{"_token": token, "email": Config.LoginAddress, "password": Config.LoginPassword})

	urlList := []string{}
	particiapntNames := []string{}
	c3.OnHTML(`table`, func(e *colly.HTMLElement) {
		if e.Index == 1 {
			e.ForEach("tr td:nth-child(2) a", func(_ int, el *colly.HTMLElement) {
				participantPage, _ := el.DOM.Attr("href")
				particiapntNames = append(particiapntNames, el.DOM.Text())
				//log.Println("participantPage=" + participantPage)
				ref2, _ := url.Parse(participantPage)
				url := urlBase.ResolveReference(ref2).String()
				//log.Println(url)
				urlList = append(urlList, url)
			})
		}
	})

	c3.Visit(trainingUrl)

	scoreList := []string{}
	for i, u := range urlList {
		ret := findScore(i, c4, u, particiapntNames[i])
		scoreList = append(scoreList, ret)
	}

	// ファイルへの出力
	file, err := os.OpenFile("score_list.csv", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal("得点情報をファイルに書き込むのに失敗しました。")
	}
	defer file.Close()
	err = file.Truncate(0)
	if err != nil {
		log.Fatal("得点情報をファイルに書き込むのに失敗しました。")
	}
	writer := csv.NewWriter(file)
	for _, s := range scoreList {
		writer.Write([]string{s})
	}
	writer.Flush()
}

func findScore(i int, c *colly.Collector, url string, particiapntName string) string {

	// 戻り値
	var ret string

	c5 := c.Clone()
	c5.OnHTML("table", func(e *colly.HTMLElement) {

		if e.Index == 3 {
			tr := e.DOM.Find("tr")
			var examName string
			var score string
			if len(Args) > 0 {
				index, _ := strconv.Atoi(Args[0])
				examName = tr.Eq(index).Find("td").Eq(1).Text()
				score = tr.Eq(index).Find("td").Last().Text()
			} else {
				examName = tr.Last().Find("td").Eq(1).Text()
				score = tr.Last().Find("td").Last().Text()
			}

			examName = strings.TrimSpace(examName)
			particiapntName = strings.TrimSpace(particiapntName)
			score = strings.TrimSpace(score)
			if i == 0 {
				fmt.Printf("収集対象： %s\n", examName)
			}
			fmt.Printf("%s %s\n", particiapntName, score)

			if score == "未登録" {
				ret = "-"
			} else {
				ret = score
			}
		}
	})
	c5.Visit(url)
	return ret
}
