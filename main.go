// xbox project main.go
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
)

func main() {
	xtstoken("gao.qi@bestv.com.cn", "bestvxbox1")
}

// with cookie support
func new_session() http.Client {
	ckyjar, _ := cookiejar.New(nil)
	return http.Client{Jar: ckyjar}
}

//CheckRedirect func(req *Request, via []*Request) error
// disable redirect for access login-post-url, because we need the first step's location
//func disable_redirect(req *http.Request, via []*http.Request) error {
//	return errors.New("no-redirect")
//}

func xtstoken(email, passwd string) (xtoken string, err error) {
	session := new_session()
	const oauth2_url = "https://login.live.com/oauth20_authorize.srf"

	ourl := oauth2_url + "?" + url.Values{
		"client_id":     []string{"0000000048093EE3"}, // client-id's magic number
		"redirect_uri":  []string{"https://login.live.com/oauth20_desktop.srf"},
		"response_type": []string{"token"},
		"display":       []string{"touch"},
		"scope":         []string{"service::user.auth.xboxlive.com::MBI_SSL"},
		"locale":        []string{"en"},
	}.Encode()
	//	log.Println(ourl)
	resp, err := session.Get(ourl)
	if err != nil {
		log.Fatal(err)
	}
	//	fmt.Println(resp.ContentLength, resp.Status, resp.StatusCode)

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
	/*
		login_url, ppft, err := urlpost_ppft(&session)
		if err != nil {
			log.Fatal(err)
		}
	*/

	var disable_redirect = errors.New("disable-redirect-this-session")
	session.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return disable_redirect
	}
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
	if e, ok := err.(*url.Error); err != nil && (!ok || e.Err != disable_redirect) {
		log.Fatal(err)
	}
	//
	//	fmt.Println(resp.StatusCode, resp.Status, resp.ContentLength)
	//	fmt.Println(resp.Header)

	var access_token string
	if u, err := resp.Location(); err == nil {
		if x, err := url.ParseQuery(u.Fragment); err == nil {
			access_token = x["access_token"][0]
		}
	}
	resp.Body.Close()
	if access_token == "" {
		fmt.Println("no access-token fetched")
		return
	}
	session.CheckRedirect = nil // enable redirect again
	//	acctoken, err := authen_access_token(email, passwd, ppft, login_url)

	dss, _ := json.Marshal(map[string]interface{}{
		"RelyingParty": "http://auth.xboxlive.com",
		"TokenType":    "JWT",
		"Properties": map[string]string{
			"AuthMethod": "RPS",
			"SiteName":   "user.auth.xboxlive.com",
			"RpsTicket":  access_token,
		},
	})

	const authen_url = "https://user.auth.xboxlive.com/user/authenticate"
	resp, err = session.Post(authen_url, "application/json", bytes.NewReader(dss))
	if err != nil {
		log.Fatal(err)
	}
	var result struct {
		Token         string
		DisplayClaims struct {
			Xui []struct {
				Uhs string `json:"uhs,omitempty"` // used in authenticate
				Xid string `json:"xid,omitempty"` // used in xsts-token
			} `json:"xui"`
		}
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	uhs, user_token := result.DisplayClaims.Xui[0].Uhs, result.Token

	dss, _ = json.Marshal(map[string]interface{}{
		"RelyingParty": "https://xaaa.bbtv.cn",
		"TokenType":    "JWT",
		"Properties": map[string]interface{}{
			"UserTokens": []string{user_token},
			"SandboxId":  "RETAIL",
		},
	})

	const auth_url = "https://xsts.auth.xboxlive.com/xsts/authorize"
	resp, err = session.Post(auth_url, "application/json", bytes.NewReader(dss))
	if err != nil {
		log.Fatal(err)
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}

	xsts_token := "XBL3.0 x=" + uhs + ";" + result.Token
	xid := result.DisplayClaims.Xui[0].Xid

	log.Println(xsts_token)
	log.Println(xid)
	return
}
