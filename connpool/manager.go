package connpool

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/miekg/dns"
	mydns "github.com/shell909090/goproxy/dns"
)

const (
	str_sess = `
<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd">
<html>
  <head>
    <title>session list</title>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
    <meta name="author" content="Shell.Xu">
  </head>
  <body>
    <table>
      <tr>
        <td width="50%"><a href="cutoff">cutoff</a></td>
      </tr>
      <tr>
	<td>
	  <form action="lookup">
	    <input type="text" name="host" value="www.google.com">
	    <input type="submit" value="lookup">
	  </form>
	</td>
      </tr>
    </table>
    <table>
      <tr>
	<th>Sess</th><th>Id</th><th>State</th><th width="50%">Target</th>
      </tr>
      {{if .GetSize}}
      {{range $tun := .GetTunnels}}
      <tr>
	<td>{{$tun.LocalAddr}}</td>
	<td>{{$tun.GetSize}}</td>
	<td>{{$tun.Uptime}}</td>
	<td>{{$tun.RemoteAddr}}</td>
      </tr>
      {{range $conn := $tun.GetConnections}}
      <tr>
	{{with $conn}}
	<td></td>
	<td>{{$conn.GetStreamId}}</td>
	<td>{{$conn.GetStatusString}}</td>
	<td>{{$conn.GetTarget}}</td>
	{{else}}
	<td></td>
	<td>half closed</td>
	{{end}}
      </tr>
      {{end}}
      {{end}}
      {{else}}
      <tr><td>no session</td></tr>
      {{end}}
    </table>
  </body>
</html>`
	str_addrs = `
<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd">
<html>
  <head>
    <title>address list</title>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
    <meta name="author" content="Shell.Xu">
  </head>
  <body>
    <pre>
{{.}}
    </pre>
  </body>
</html>`
)

var (
	tmpl_sess *template.Template
	tmpl_addr *template.Template
)

func init() {
	var err error
	tmpl_sess, err = template.New("main").Parse(str_sess)
	if err != nil {
		panic(err)
	}

	tmpl_addr, err = template.New("address").Parse(str_addrs)
	if err != nil {
		panic(err)
	}
}

func (pool *Pool) HandlerMain(w http.ResponseWriter, req *http.Request) {
	err := tmpl_sess.Execute(w, pool)
	if err != nil {
		logger.Error(err.Error())
	}
	return
}

func HandlerLookup(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	hosts, ok := q["host"]
	if !ok {
		w.WriteHeader(400)
		w.Write([]byte("no domain"))
		return
	}

	rslt := ""
	if exhg, ok := mydns.DefaultResolver.(mydns.Exchanger); ok {
		quiz := new(dns.Msg)
		quiz.SetQuestion(dns.Fqdn(hosts[0]), dns.TypeA)
		quiz.RecursionDesired = true

		resp, err := exhg.Exchange(quiz)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "error %s", err)
			return
		}

		rslt = resp.String()
	} else {
		addrs, err := mydns.DefaultResolver.LookupIP(hosts[0])
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "error %s", err)
			return
		}
		for _, addr := range addrs {
			rslt += addr.String() + "\r\n"
		}
	}

	err := tmpl_addr.Execute(w, rslt)
	if err != nil {
		logger.Error(err.Error())
	}
	return
}

func (pool *Pool) HandlerCutoff(w http.ResponseWriter, req *http.Request) {
	pool.CutAll()
	return
}
