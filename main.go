package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

// 構造体の定義
type Page struct {
	Title string
	Body []byte
}

// ページに保存するメソッド
func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

/**
ページを読み込む関数
正常値の場合はポインタ，異常値の場合はerrorを返す
 */
func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	// エラーハンドリング
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

// テンプレートファイルをパースするのは一度だけ
var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

/** テンプレートハンドラー
tmpl: 各htmlファイル
 */
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page)  {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// パスのバリデーション
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

// HandlerFunc（関数http.HandleFuncに渡すのに適している）の関数を返すラッパー関数
// fn: save, edit, viewハンドラーの一つ
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	// クロージャ．引数にresponse, request, title
	return func(w http.ResponseWriter, r *http.Request) {
		// リクエストパスからページのタイトルを抽出
		m := validPath.FindStringSubmatch(r.URL.Path)
		// エラーハンドリング
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

/**
閲覧ページのハンドラー
http.HandleFuncのハンドラー
http.ResponseWriter: HTTPサーバからのレスポンスをアセンブルし，クライアントにデータを送信
http.Request: クライアントのリクエストのデータ構造
 */
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

// 編集ページのハンドラー
func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

// 保存ページのハンドラー
func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func main() {
	// handlerを使ってリクエストを処理できるようhttpパッケージに指示するメソッド
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	// エラーの場合のみ返すメソッドをlog.Fatalでラップ
	log.Fatal(http.ListenAndServe(":8080", nil))
}