{{define "cashreqHTML"}}
<p>You"ve got the request for {{.Amount}} {{.Currency}}. It has to be repaid within the next 48 hours.</p>
{{end}}

{{define "cashreqHTML42"}}
<p>Уважаемый Dcoin-майнер!</p>

<p>На Вашу обещанную сумму {{.Amount}} {{.Currency}} пришел запрос от майнера {{.FromUserId}}.<br>
Вам нужно зайти в свой Dcoin-кошелек, далее во "Входящие запросы", там Вы сможете узнать контакты майнера {{.FromUserId}}.<br>
После передачи денег майнеру {{.FromUserId}}, он должен передать Вам специальный код, который разблокирует Ваши обещанные суммы и начислит Вам {{.Amount}} d{{.Currency}}.</p>
<p>Если в течение 48 часов Вы не введете код, то Вы больше не сможете создавать новые  Dcoin-ы из обещанных сумм.<br>
На <a href="https://github.com/democratic-coin/dcoin-go/wiki/Биржи">биржах</a> Вы в любой момент можете продать dUSD или отправить запрос на обмен другому майнеру.
</p>
{{end}}

