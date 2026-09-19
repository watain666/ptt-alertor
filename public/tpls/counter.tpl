{{define "counter"}}
{{if .Count}}<div class="counter-line page-width" id="counter-board" data-ws-host="{{.WSHost}}"><span class="green-dot" aria-hidden="true"></span> 已替大家送出 <strong id="counter">{{range .Count}}{{.}}{{end}}</strong> 則關心的消息。</div>{{end}}
{{end}}
