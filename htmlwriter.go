package main

import (
	"bytes"
	"fmt"
	"os"
)

//CHECK is the unicode charater for a tick (check) mark
const CHECK = "&#x2713"

//HTMLWriter contains a reference to the Coverage Check that needs to be written out as an HTML file
type HTMLWriter struct {
	CovCheckerInfo *CovCheckerInfo
	Buffer         *bytes.Buffer
}

//NewHTMLWriter returns a new instance of the HTMLWriter
func NewHTMLWriter(covChecker *CovCheckerInfo) *HTMLWriter {
	return &HTMLWriter{
		CovCheckerInfo: covChecker,
		Buffer:         new(bytes.Buffer),
	}
}

//Write the HTML into the buffer
func (hw *HTMLWriter) Write(outfilename string) error {
	hw.AddStaticContent()
	hw.PrintTotalRow()
	hw.IterateServices()
	hw.AddTrailingContent()

	f, err := os.Create(outfilename)
	if err != nil {
		return err
	}
	defer f.Close()
	s := hw.Buffer.String()
	_, err = f.Write([]byte(s))
	return err
}

//AddStaticContent adds all the head, style, and script tags into the HTML
func (hw *HTMLWriter) AddStaticContent() {
	fmt.Fprintf(hw.Buffer, `
<!DOCTYPE html>
<html>
<head>
	<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.4.1/jquery.min.js"></script>
</head>
<meta name="viewport" content="width=device-width, initial-scale=1">
<style>

p, h1, h2, td, th, span{
	font-family: Calibri;
}

th, td {
	border-bottom: 1px solid lightgrey;
	padding-right: 6px;
	font-size: 16px;
	vertical-align: bottom;
	border-collapse: collapse;
}

th {
	text-align: left;
	padding-left: 6px;
}

table {
	border-top: 1px solid lightgrey;
	border-left: 1px solid lightgrey;
	border-right: 1px solid lightgrey;
	border-collapse: separate;
	border-spacing: 0;
}

.tableHeader {
	padding: 2px 0;
	position: sticky;
	top: 0;
	background: white;
	z-index: 10;
	padding-left: 6px;
	padding-right: 6px;
}

.covCol {
	padding-left: 6px;
}

.docCol {
	padding-left: 6px;
}

.verbDetail {
	font-style: italic;
}
.verbCounts {
	font-style: italic;
	text-align: center;
}

.check {
	color: green;
	text-align: center;
}

.caret {
	cursor: pointer;
	-webkit-user-select: none; /* Safari 3.1+ */
	-moz-user-select: none; /* Firefox 2+ */
	-ms-user-select: none; /* IE 10+ */
	user-select: none;
}

.expand .caret::before {
	content: "\23F5";
	color: grey;
	display: inline-block;
	margin-right: 2px;
}

.collapse .caret::before {
	content: "\23F7";
	color: grey;
	display: inline-block;
	margin-right: 2px;  
}

.level1 td:first-child {
	padding-left: 19px;
}
.level2 td:first-child {
	padding-left: 37px;
}
.level3 td:first-child {
	padding-left: 55px;
}
.level1{
	display: none;
}
.level2 {
	display: none;
}
.level3 {
	display: none;
}
meter { opacity: 0.6; }
.meter-value {
	display: block; height: 0px;
	position: relative;
	text-align: center;
	top: -19px;
}
</style>
<body>
<script>
	$(function() {
		$('#covTable').on('click', '.caret', function () {
			//Gets all <tr>'s  of greater depth
			//below element in the table
			var findChildren = function (tr) {
				var depth = tr.data('depth');
				return tr.nextUntil($('tr').filter(function () {
					return $(this).data('depth') <= depth;
				}));
			};
	
			var el = $(this);
			var tr = el.closest('tr'); //Get <tr> parent of toggle button
			var children = findChildren(tr);
	
			//Remove already collapsed nodes from children so that we don't
			//make them visible. 
			//(Confused? Remove this code and close Item 2, close Item 1 
			//then open Item 1 again, then you will understand)
			var subnodes = children.filter('.expand');
			subnodes.each(function () {
				var subnode = $(this);
				var subnodeChildren = findChildren(subnode);
				children = children.not(subnodeChildren);
			});
	
			//Change icon and hide/show children
			if (tr.hasClass('collapse')) {
				tr.removeClass('collapse').addClass('expand');
				children.hide();
			} else {
				tr.removeClass('expand').addClass('collapse');
				children.show();
			}
			return children;
		});
	});
</script>
<table id="covTable">
	<thead>
		<tr>
			<th class="tableHeader">Path</th>
			<th class="tableHeader covCol">Coverage</th>
			<th class="tableHeader docCol">Documented</th>
		</tr>
	</thead>
	<tbody>	
`)
}

//PrintTotalRow adds the total stats into the table
func (hw *HTMLWriter) PrintTotalRow() {
	cc := hw.CovCheckerInfo
	fmt.Fprintf(hw.Buffer, `
	<tr data-depth="0">
		<th >Total</th>
		<td class="covCol"><meter min="0" max="1" low="0.8" high="0.8" optimum="1" value="%3.2f"></meter><span class="meter-value">%3.2f%%</span></td>
		<td class="docCol"><meter min="0" max="1" low=".9999" high=".9999" optimum="1" value="%3.2f"></meter><span class="meter-value">%3.2f%%</span></td>
	</tr>
`, cc.Coverage, cc.Coverage*100, 1-cc.Undocumented, (1-cc.Undocumented)*100)
}

//IterateServices iterates over all the services printing content rows for each endpoint
func (hw *HTMLWriter) IterateServices() {
	for _, ss := range hw.CovCheckerInfo.ServiceStats {
		hw.PrintServiceCoverage(ss)
		for _, ep := range ss.Endpoints {
			hw.PrintEndpointCoverage(ep)
			for _, verb := range ep.Verbs {
				hw.PrintVerbCoverage(verb)
				hw.PrintResponsesHeader(verb.Responses)
				for _, resp := range verb.Responses {
					c := resp.Covered > 0
					hw.PrintDetailRow(resp.Response, c, resp.Documented)
				}
				hw.PrintQueriesHeader(verb.Parameters)
				for _, param := range verb.Parameters {
					c := param.Covered > 0
					hw.PrintDetailRow(param.Key, c, param.Documented)
				}
			}
		}
	}
}

//PrintServiceCoverage adds the total stats for a specific service into the table
func (hw *HTMLWriter) PrintServiceCoverage(ss *ServiceStat) {
	fmt.Fprintf(hw.Buffer, `
    <tr data-depth="0">
        <th>%s</th>
        <td class="covCol"><meter min="0" max="1" low="0.8" high="0.8" optimum="1" value="%3.2f"></meter><span class="meter-value">%3.2f%%</span></td>
        <td class="docCol"><meter min="0" max="1" low=".9999" high=".9999" optimum="1" value="%3.2f"></meter><span class="meter-value">%3.2f%%</span></td>
    </tr>
`, ss.Name, ss.Coverage, ss.Coverage*100, 1-ss.Undocumented, (1-ss.Undocumented)*100)
}

//PrintEndpointCoverage adds the stats for a endpoint in a service into the table
func (hw *HTMLWriter) PrintEndpointCoverage(ep EndpointStat) {
	fmt.Fprintf(hw.Buffer, `
    <tr data-depth="0" class="expand level0">
        <td><span class="caret expand"></span>%s</td>
        <td class="covCol"><meter min="0" max="1" low="0.8" high="0.8" optimum="1" value="%3.2f"></meter><span class="meter-value">%3.2f%%</span></td>
        <td class="docCol"><meter min="0" max="1" low=".9999" high=".9999" optimum="1" value="%3.2f"></meter><span class="meter-value">%3.2f%%</span></td>
    </tr>
`, ep.Path, ep.Coverage, ep.Coverage*100, 1-ep.Undocumented, (1-ep.Undocumented)*100)
}

//PrintVerbCoverage adds the stats for a endpoint in a service into the table
func (hw *HTMLWriter) PrintVerbCoverage(verb VerbStat) {
	coverage := float64(verb.Covered) / float64(verb.Total)
	undocumented := float64(verb.Undocumented) / float64(verb.Total)
	fmt.Fprintf(hw.Buffer, `
    <tr data-depth="1" class="expand level1">
        <td><span class="caret expand"></span>%s</td>
        <td class="covCol"><meter min="0" max="1" low="0.8" high="0.8" optimum="1" value="%3.2f"></meter><span class="meter-value">%3.2f%%</span></td>
        <td class="docCol"><meter min="0" max="1" low=".9999" high=".9999" optimum="1" value="%3.2f"></meter><span class="meter-value">%3.2f%%</span></td>
    </tr>
`, verb.Method, coverage, coverage*100, 1-undocumented, (1-undocumented)*100)
}

//PrintResponsesHeader prints the row for the responses into the table
func (hw *HTMLWriter) PrintResponsesHeader(r map[string]*Response) {
	c := 0
	d := 0
	for _, resp := range r {
		if resp.Covered > 0 {
			c++
		}
		if resp.Documented {
			d++
		}
	}
	fmt.Fprintf(hw.Buffer, `
    <tr data-depth="2" class="expand level2">
        <td class="verbDetail"><span class="caret expand"></span>Responses</td>
        <td class="verbCounts">%d/%d</td>
        <td class="verbCounts">%d/%d</td>
    </tr>
`, c, len(r), d, len(r))
}

//PrintQueriesHeader the row for the responses into the table
func (hw *HTMLWriter) PrintQueriesHeader(p map[string]*QueryParameter) {
	c := 0
	d := 0
	for _, resp := range p {
		if resp.Covered > 0 {
			c++
		}
		if resp.Documented {
			d++
		}
	}
	fmt.Fprintf(hw.Buffer, `
    <tr data-depth="2" class="expand level2">
        <td class="verbDetail"><span class="caret expand"></span>Parameters</td>
        <td class="verbCounts">%d/%d</td>
        <td class="verbCounts">%d/%d</td>
    </tr>
`, c, len(p), d, len(p))
}

//PrintDetailRow prints a detail row for a response or query into the table
func (hw *HTMLWriter) PrintDetailRow(name string, covered bool, documented bool) {
	checkString := fmt.Sprintf(" check\">%s</td>", CHECK)
	noCheckString := "\"></td>"
	cs := noCheckString
	if covered {
		cs = checkString
	}
	ds := noCheckString
	if documented {
		ds = checkString
	}
	fmt.Fprintf(hw.Buffer, `
    <tr data-depth="3" class="expand level3">
        <td>%s</td>
        <td class="covCol%s
        <td class="docCol%s
    </tr>
`, name, cs, ds)
}

//AddTrailingContent adds all the html close tags at the end of the file
func (hw *HTMLWriter) AddTrailingContent() {
	fmt.Fprintf(hw.Buffer, `
</tbody>
</table>
</body>
</html>
`)
}

//PrintStats prints the calculated coverage stats into an HTML document
func (cc *CovCheckerInfo) PrintStats() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, `<html>
<h1>Summary</h1>
<table>
<tr><td>Overall API coverage</td><td><meter min="0" max="1" low="0.8" high="0.8" optimum="1" value="%3.2f"></meter> %3.2f%%</td></tr>
<tr><td>Overall API documentation</td><td><meter min="0" max="1" low=".9999" high=".9999" optimum="1" value="%3.2f"></meter> %3.2f%%</td></tr>
</table>
<h2>Coverage summary by Service</h2>
<table>
`, cc.Coverage, cc.Coverage*100, 1-cc.Undocumented, (1-cc.Undocumented)*100)

	//First let's print the summary of services
	for _, ss := range cc.ServiceStats {
		fmt.Fprintf(&buf,
			`<tr>
  <td>%s</td>
  <td><meter min="0" max="1" low="0.8" high="0.8" optimum="1" value="%3.2f"></meter> Coverage: %3.2f%%<br/>
  <meter min="0"  max="1" low=".9999" high=".9999" optimum="1" value="%3.2f"></meter> Documented: %3.2f%%
  </td>
</tr>
`, ss.Name, ss.Coverage, ss.Coverage*100, 1-ss.Undocumented, (1-ss.Undocumented)*100)
	}
	fmt.Fprintln(&buf, `</table>`)

	for _, ss := range cc.ServiceStats {
		fmt.Fprintf(&buf, `<h1>%s</h1>
<table>
`, ss.Name)

		for _, ep := range ss.Endpoints {
			fmt.Fprintf(&buf,
				`<tr>
  <td>%s</td>
  <td><meter min="0" max="1" low="0.8" high="0.8" optimum="1" value="%3.2f"></meter> Coverage: %3.2f%%<br/>
  <meter min="0" max="1" low=".9999" high=".9999" optimum="1" value="%3.2f"></meter> Documented: %3.2f%%
  </td>
</tr>
`, ep.Path, ep.Coverage, ep.Coverage*100, 1-ep.Undocumented, (1-ep.Undocumented)*100)
		}

		fmt.Fprintln(&buf, `</table>`)
	}
	fmt.Fprintln(&buf, `</html>`)
	return buf.String()
}
