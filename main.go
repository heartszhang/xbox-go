// xbox project main.go
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
)

func main() {
	xtstoken("", "")
}

func new_session() http.Client {
	ckyjar, _ := cookiejar.New(nil)
	return http.Client{Jar: ckyjar}
}
func xtstoken(email, passwd string) (xtoken string, err error) {
	if email == "" {
		email = "gao.qi@bestv.com.cn"
	}
	if passwd == "" {
		passwd = "bestvxbox1"
	}
	session := new_session()
	const oauth2_url = "https://login.live.com/oauth20_authorize.srf"

	para := url.Values{
		"client_id":     []string{"0000000048093EE3"},
		"redirect_uri":  []string{"https://login.live.com/oauth20_desktop.srf"},
		"response_type": []string{"token"},
		"display":       []string{"touch"},
		"scope":         []string{"service::user.auth.xboxlive.com::MBI_SSL"},
		"locale":        []string{"en"},
	}
	ourl := oauth2_url + "?" + para.Encode()
	fmt.Println(ourl)
	resp, err := session.Get(ourl)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.ContentLength, resp.Status, resp.StatusCode)

	txt, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	//urlPost:'https://login.live.com/ppsecure/post.srf?client_id=0000000048093EE3&display=touch&locale=en&redirect_uri=https%3A%2F%2Flogin.live.com%2Foauth20_desktop.srf&response_type=token&scope=service%3A%3Auser.auth.xboxlive.com%3A%3AMBI_SSL&bk=1428934781&uaid=3790edea17da48b9b44c9c33a4cb1cdd'
	urlpost_re := regexp.MustCompile(`urlPost:'([A-Za-z0-9:\?_\-\.&/=%]+)`)
	login_url := urlpost_re.FindStringSubmatch(string(txt))[1]
	fmt.Println(login_url)

	//sFTTag:'<input type="hidden" name="PPFT" id="i0327" value="CnvTxCHxZAfQIkGEKQcUcxmDgDZgpbkRcvQf*qButfGLZj7eQ9hjeqhVA0Pvv6BvyFKjld3b0BPzlqxLMY4q2!qam5d6SpUaW81A00yRFJwT2UMrhYCTPYrwf8KdKRRLlE7PONBkSkR7F4yYMuFnAV5Cz7R6eGd*bgN3cjV8n9OtDtYmDCV*vB*l10HV9010oQ$$"/>'
	fttag_re := regexp.MustCompile(`sFTTag:'.*value="(.*)"/>'`)
	ppft := fttag_re.FindStringSubmatch(string(txt))[1]
	fmt.Println(ppft)

	post_data := url.Values{
		"login":        []string{email},
		"passwd":       []string{passwd},
		"PPFT":         []string{ppft},
		"PPSX":         []string{"Passpor"},
		"SI":           []string{"Sign in"},
		"type":         []string{"11"},
		"NewUser":      []string{"1"},
		"LoginOptions": []string{"1"},
		"i3":           []string{"36728"},
		"m1":           []string{"768"},
		"m2":           []string{"1184"},
		"m3":           []string{"0"},
		"i12":          []string{"1"},
		"i17":          []string{"0"},
		"i18":          []string{"__Login_Host|1"},
	}
	resp, err = session.PostForm(login_url, post_data)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.StatusCode, resp.Status, resp.ContentLength)
	fmt.Println(resp.Header)
	fmt.Println(resp.Header.Get("Location"))
	resp.Body.Close()
	return
}
