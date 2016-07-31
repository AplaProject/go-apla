{{define "dcsentHTML"}}
<table cellpadding="0" cellspacing="0" style="box-sizing: border-box; mso-table-lspace: 0pt; mso-table-rspace: 0pt; border-collapse: separate !important;" align="center">
		<tr>
			<!--td align="right" valign="middle" style="padding-bottom:40px;"><img src="http://dcoin.club/email_images/logo.png" style="display:block; margin-right:10px;" width="57" height="57" border="0" alt=""/></td-->
			<td align="left" valign="middle" style="color:#0071B1; font-size:48px; vertical-align:middle; padding-bottom:40px;">d{{.Currency}}</td>
			</tr></table>
			<h1 class="align-center" style="font-family: 'Helvetica Neue', Helvetica, Arial, 'Lucida Grande', sans-serif; font-weight: 300; line-height: 1.4em; margin: 0; font-size: 22px; text-align: center; color: #48545d !important;">You have sent</h1>
			<p style="font-family: 'Helvetica Neue', Helvetica, Arial, 'Lucida Grande', sans-serif; font-size: 14px; font-weight: normal; margin: 0; margin-bottom: 40px; color:#8c8e8e;" align="center"><strong>{{.DCSent.Amount}}</strong> d{{.Currency}}</p>
			<p style="font-family: 'Helvetica Neue', Helvetica, Arial, 'Lucida Grande', sans-serif; font-size: 1px; line-height:1px; font-weight: normal; margin: 0; margin-bottom: 40px; height:2px; background-color:#e9e9e9" align="center">&nbsp;</p>
			<p style="font-family: 'Helvetica Neue', Helvetica, Arial, 'Lucida Grande', sans-serif; font-size: 14px; line-height:24px; font-weight: normal; margin: 0; margin-bottom: 40px; color:#48545d;" align="center">Debiting <strong>{{.DCSent.Amount}} d{{.Currency}}</strong>.</p>
{{end}}