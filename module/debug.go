package module

import (
	"fmt"
	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
	"math/rand"
	"time"
)

func (ctx *ComposeContext) isDebugEnabled() bool {
	for key, values := range ctx.httpRequest.URL.Query() {
		if key == "debug" {
			for _, value := range values {
				if value == "true" {
					return true
				}
			}
		}
	}
	return false
}

func decorateNodeWithDebugInformation(component *WebComponent, content *html.Node) *html.Node {
	stringOrEmpty := func(input *string) string {
		if input != nil {
			return *input
		}
		return ""
	}

	timeOrEmpty := func(input *time.Time) int64 {
		if input != nil {
			return input.Unix() * 1000
		}
		return -1
	}

	durationOrEmpty := func(input *time.Duration) int64 {
		if input != nil {
			return input.Milliseconds()
		}
		return -1
	}

	wrapperRawString := `
<div style="background-color: rgba(%d,%d,%d,0.3); border-style: dotted;">
	<img 
		onclick="showDebug(this)" 
		class="webc-debug-icon" 
		src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAMAAAAoLQ9TAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAAgY0hSTQAAeiYAAICEAAD6AAAAgOgAAHUwAADqYAAAOpgAABdwnLpRPAAAAURQTFRFAAAAMj1DMTtCKytVMj1BLztBLkZGMTxCMDxCMEBARERARERANTtCMTxCMj1DLUBAMjxDNztAMj1BMTxCMzNNRERARERARERAMz1CMjxCMTxCMj1CMj1BO0JAMz5DMD1BMD1DLzxCNjZDMz5BMj1CNzdDRERAMjxCMTxCMTtDMjxDMjtDOkJAMzxDMzxBMz1BOj9DND5BMT1COT5DND5BNT1BNj8+Mz1CMTxCMjxCMz1CN0A+RERAMTxEND9CMjxCMj1BMjtCMTxCMT1CMT1BMj1BDw9yAAAAMT1CMTtDMz1DMjtCAACAMT5CMDtCMTxCMzxDMTxBMTxCMDxBMjxBMjtDMjxDMjxCMTxCMD1BMDxBMD1CMTxBJElJMz1CMjxCMTxCKEhIMj1CMDpDLEBALjpAMz5EK0BANDxAMTxCMjxC////FlnL7gAAAGl0Uk5TAC5JBkcrC1mpIAQFJ6tcDaAxLqIKAQINvuvruXEiWVBQURN2ixQPlP6wtIogjzenL3huLak+G5v09JwdCUFETaaftrWgpAMBbbG4bAI+o+C13qZaZmdN/f1LamWYB2Dv7gibNQ0sLQw8ogoMdwAAAAFiS0dEa1JlpZgAAAAJcEhZcwAACxMAAAsTAQCanBgAAADjSURBVBjTNY93V8JQDMVvRQFfFQe2uAeo0OfeC0VQQaoIintvTb7/BzCvPeaPnNxfcpIbALBaIpIRaW1DGNFYvF3ZHZ2JKP6jq7unN9lnKifQbqp/YHAo0AoYFj0yOjY+kZZOxgYmp1LT2ZynczNwZ+dkYn5hcYnI00TLK6trskOtb2yGYGt7x+y087sUAOY9V3Rhv1gKADMfHB4VYJUrxwKqPjOdVE4tc7wmQPtMVAttJs7qxA2f6PyiaZxeXuH6hrWm2zvcP4iPxycn//zi6dcY8PYuTj8cZavPr+8f88mv+gOqACqSsd0yfAAAACV0RVh0ZGF0ZTpjcmVhdGUAMjAxOC0wNS0wNVQxNzoyNTo0Ni0wNTowMN/NlCcAAAAldEVYdGRhdGU6bW9kaWZ5ADIwMTgtMDUtMDVUMTc6MjU6NDYtMDU6MDCukCybAAAAAElFTkSuQmCC"
		data-source-id="%s"
		data-method="%s"
		data-url="%s"
		data-name="%s"
		data-body="%s"
		data-cached-until="%d"
		data-loading-time="%d"
		>
</div>
`
	wrapperString := fmt.Sprintf(
		wrapperRawString,
		rand.Intn(255),
		rand.Intn(255),
		rand.Intn(255),
		stringOrEmpty(component.source.id),
		stringOrEmpty(component.source.method),
		stringOrEmpty(component.source.url),
		stringOrEmpty(component.name),
		stringOrEmpty(component.source.body),
		timeOrEmpty(component.source.cachedUntil),
		durationOrEmpty(component.source.loadTime),
	)

	doc, _ := parseString(&wrapperString)
	wrapper := cascadia.MustCompile("div").MatchFirst(doc)

	appendContent(wrapper, content)
	return wrapper
}

func debugStyleNode() *html.Node {
	wrapperString := `
<style>
.webc-debug-icon { 
	width: 16px; 
	float: left; 
	background-color: white; 
	border: solid 1px 
}

.webc-debug-modal {
  display: none; /* Hidden by default */
  position: fixed; /* Stay in place */
  z-index: 1; /* Sit on top */
  padding-top: 100px; /* Location of the box */
  left: 0;
  top: 0;
  width: 100%; /* Full width */
  height: 100%; /* Full height */
  overflow: auto; /* Enable scroll if needed */
  background-color: rgb(0,0,0); /* Fallback color */
  background-color: rgba(0,0,0,0.4); /* Black w/ opacity */
}

.webc-debug-modal table th{
	text-align: left;
}
.webc-debug-modal table td{
	text-align: left;
	padding-left: 10px;
}

/* Modal Content */
.webc-debug-modal-content {
  background-color: #fefefe;
  margin: auto;
  padding: 20px;
  border: 1px solid #888;
  width: 80%;
}

/* The Close Button */
.webc-debug-close {
  color: #aaaaaa;
  float: right;
  font-size: 28px;
  font-weight: bold;
  top: -20px;
  position: relative;
  right: -10px;
}

.webc-debug-close:hover,
.webc-debug-close:focus {
  color: #000;
  text-decoration: none;
  cursor: pointer;
}
</style>
`
	doc, _ := parseString(&wrapperString)
	return cascadia.MustCompile("style").MatchFirst(doc)
}

func debugScriptNode() *html.Node {
	wrapperString := `
<script>
var webcDebugModal = document.getElementById("webc-debug-modal");

function closeModal() {
	webcDebugModal.style.display = "none";
}

window.onclick = function(event) {
  if (event.target == webcDebugModal) {
    closeModal();
  }
}

function showDebug(x) {
	function getAttribute(name){
		return x.getAttribute("data-" + name);
	}
	function renderCachedUntil(value){
		let numberValue = parseInt(value);

		if(isNaN(numberValue) || numberValue < 0){
			return "Not cached";
		}else{
			const options = { year: 'numeric', month: 'long', day: 'numeric', hour: 'numeric', minute: 'numeric', second: 'numeric' };
			const dateTimeFormat = new Intl.DateTimeFormat('en-GB', options);
			return dateTimeFormat.format(new Date(numberValue));
		}
	}

	function renderBody(value){
		if(value == ""){
			return "No body";
		} else {
			return value;
		}
	}

	function renderData(name, value){
		let element = document.getElementById("webc-debug-" + name);
		element.innerHTML = value;
	}

	let sourceId = getAttribute("source-id");
	let method = getAttribute("method");
	let url = getAttribute("url");
	let name = getAttribute("name");
	let body = getAttribute("body");
	let loadTime = getAttribute("loading-time");
	let cacheUntil = getAttribute("cached-until");

	renderData("source-id", sourceId);
	renderData("method", method); 
	renderData("url", url);
	renderData("body", renderBody(body));
	renderData("loading-time", loadTime + "ms");
	renderData("cached-until", renderCachedUntil(cacheUntil));
	renderData("name", name);

  	webcDebugModal.style.display = "block";
}
</script>
`
	doc, _ := parseString(&wrapperString)
	return cascadia.MustCompile("script").MatchFirst(doc)
}

func debugModalNode() *html.Node {
	wrapperString := `
<div id="webc-debug-modal" class="webc-debug-modal">

  <!-- Modal content -->
  <div class="webc-debug-modal-content">
    <span class="webc-debug-close" onclick="closeModal()">&times;</span>
  <table>
  <tr>
	  <th>Source Id:</th>
	  <td id="webc-debug-source-id">-</td>
  </tr>
  <tr>
	  <th>Source Method:</th>
	  <td id="webc-debug-method">-</td>
  </tr>
  <tr>
	  <th>Source Url:</th>
	  <td id="webc-debug-url">-</td>
  </tr>
  <tr>
	  <th>Source Body:</th>
	  <td id="webc-debug-body">-</td>
  </tr>
  <tr>
	  <th>Source loading time:</th>
	  <td id="webc-debug-loading-time">-</td>
  </tr>
  <tr>
	  <th>Source Cached until time:</th>
	  <td id="webc-debug-cached-until">-</td>
  </tr>
  <tr>
	  <th>Component name:</th>
	  <td id="webc-debug-name">-</td>
  </tr>
  </table>
  </div>
</div>
`
	doc, _ := parseString(&wrapperString)
	return cascadia.MustCompile("div").MatchFirst(doc)
}