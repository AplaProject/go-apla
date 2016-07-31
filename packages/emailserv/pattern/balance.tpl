{{define "balanceSubject"}}Dcoin balance{{end}}

{{define "balanceHTML"}}

<h1 class="align-center" style="font-family: 'Helvetica Neue', Helvetica, Arial, 'Lucida Grande', sans-serif; font-weight: 300; line-height: 1.4em; margin: 0; margin-bottom:20px; font-size: 22px; text-align: center; color: #48545d !important;">Your balance:</h1>
<table cellpadding="0" cellspacing="0" style="box-sizing: border-box; mso-table-lspace: 0pt; mso-table-rspace: 0pt; border-collapse: separate !important;" align="center">
{{range .List}}{{if .Summary}}
	<tr>
  		<td height="35" align="right" valign="middle" style="color:#0071B1; font-size:28px; vertical-align:middle; padding-bottom:10px;">{{.Top}}</td>
		<td height="35" align="left" valign="middle" style="color:#0071B1; font-size:28px; vertical-align:middle; padding-bottom:10px;">&nbsp;<font size="4px">d{{.Currency}}</font></td>
	</tr>{{end}}{{end}}
</table>

<table width="100%" border="0" cellspacing="0" cellpadding="0">
											<tbody>
												<tr>
													<td style="font-family: 'Helvetica Neue', Helvetica, Arial, 'Lucida Grande', sans-serif; font-size: 14px; line-height:20px; font-weight: normal; padding-top: 20px; color:#48545d;">
You can sell and buy dcoins on the stock exchange 
<a href="https://github.com/democratic-coin/dcoin-go/wiki/Exchange" style="color:#0071B1;">stock exchange</a>. Also, you can exchange dcoins from miners with a built-in p2p technology.
</td>
												</tr>
											</tbody>
										</table>

	{{if ne .Money 1}}
	<table width="100%" cellspacing="0" cellpadding="3" border="1" bordercolor="#8c8e8e" style="font-family: 'Helvetica Neue', Helvetica, Arial, 'Lucida Grande', sans-serif; font-size: 11px; line-height:16px; font-weight: normal; margin: 0; margin-top: 20px; color:#8c8e8e;">
			<tbody align="left" valign="top">
				<tr>													<th>Currency</th>
														<th>Wallet</th>
														<th>Mining</th>
														<th>Bonuses</th>
														<th>Promised<br />Amount</th>
														<th>Summary</th>
													</tr>
	{{range .List}}
	<tr><td>d{{.Currency}}</td><td>{{.Wallet}}</td><td>{{.Tdc}}</td><td>{{.Restricted}}</td><td>{{.Promised}}</td><td>{{.Summary}}</td></tr>
	{{end}}
	</tbody>
	</table>
	{{end}}
{{end}}

{{define "balanceSubject42"}}Dcoin баланс{{end}}

{{define "balanceHTML42"}}

<h1 class="align-center" style="font-family: 'Helvetica Neue', Helvetica, Arial, 'Lucida Grande', sans-serif; font-weight: 300; line-height: 1.4em; margin: 0; margin-bottom:20px; font-size: 22px; text-align: center; color: #48545d !important;">Ваш баланс:</h1>
<table cellpadding="0" cellspacing="0" style="box-sizing: border-box; mso-table-lspace: 0pt; mso-table-rspace: 0pt; border-collapse: separate !important;" align="center">
{{range .List}}{{if .Summary}}
	<tr>
  		<td height="35" align="right" valign="middle" style="color:#0071B1; font-size:28px; vertical-align:middle; padding-bottom:10px;">{{.Top}}</td>
		<td height="35" align="left" valign="top" style="color:#{{if gt .Dif 0.0}}007f66{{else}}c0392b{{end}}; font-size:14px; vertical-align:top; padding-bottom:10px;">&nbsp;{{if ne .Dif 0.0}}({{if gt .Dif 0.0}}+{{else}}-{{end}}{{.Dif}}){{end}}</td>
		<td height="35" align="left" valign="middle" style="color:#0071B1; font-size:28px; vertical-align:middle; padding-bottom:10px;">&nbsp;<font size="4px">d{{.Currency}}</font></td>
	</tr>{{end}}{{end}}
</table>

<table width="100%" border="0" cellspacing="0" cellpadding="0">
											<tbody>
												<tr>
													<td style="font-family: 'Helvetica Neue', Helvetica, Arial, 'Lucida Grande', sans-serif; font-size: 14px; line-height:20px; font-weight: normal; padding-top: 20px; color:#48545d;">
													
Вы можете продать или купить Dcoin-ы на <a href="https://github.com/democratic-coin/dcoin-go/wiki/Exchange" style="color:#0071B1;">Бирже</a> или обменять у майнеров при помощи встроенного p2p маханизма.</td>
												</tr>
											</tbody>
										</table>
{{if ne .Money 1}}
<table width="100%" cellspacing="0" cellpadding="3" border="1" bordercolor="#8c8e8e" style="font-family: 'Helvetica Neue', Helvetica, Arial, 'Lucida Grande', sans-serif; font-size: 11px; line-height:16px; font-weight: normal; margin: 0; margin-top: 20px; color:#8c8e8e;">
		<tbody align="left" valign="top">
			<tr>													<th>Currency</th>
													<th>Wallet</th>
													<th>Mining</th>
													<th>Bonuses</th>
													<th>Promised<br />Amount</th>
													<th>Summary</th>
												</tr>
{{range .List}}
<tr><td>d{{.Currency}}</td><td>{{.Wallet}}</td><td>{{.Tdc}}</td><td>{{.Restricted}}</td><td>{{if ne .CurrencyId 1}}{{.Promised}}{{end}}</td><td>{{.Summary}}</td></tr>
{{end}}
</tbody>
</table>
{{end}}
{{end}}

