SetVar(
	global = 0,
	type_new_page_id = TxId(NewPage),
	type_new_contract_id = TxId(NewContract),
	type_new_table_id = TxId(NewTable),	
	sc_conditions = "$citizen == #wallet_id#")
SetVar(sc_newForexOrder = `contract newForexOrder {
	data {
		SellTable string
		SellRate float
		Amount money
		BuyTable string
		Direction string
	}


	func conditions {

  		if $SellTable == $BuyTable {
			warning "$SellTable == $BuyTable"
		}
		if $SellRate == 0 || $Amount == 0 {
			warning "$SellRate == 0"
		}
		var total money
		total = $Amount * $SellRate
		Println(total)
		if total < 1 {
			warning total
		}
		var amount money
		amount = DBAmount($SellTable, "citizen_id", $citizen)
		Println("amount", amount)

		var forexAmount money
		Println("SellTable", $SellTable)
		Println("citizen", $citizen)
		var account_id int
		account_id = DBIntExt($SellTable, "id", $citizen, "citizen_id")
		Println("account_id", account_id)
		forexAmount = DBIntWhere("global_forex_orders", "sum(amount)","sell_table = $ AND sell_table_account_id = $", $SellTable, account_id)
		Println("forexAmount", forexAmount)
		if amount+forexAmount < $Amount {
		    warning "not enough money "
		}
	    
	}
	
	func action {
		var reverseRate float
		reverseRate = 1.0 / $SellRate
		var totalSellAmount money
		totalSellAmount = $Amount

		Println("totalSellAmount", totalSellAmount)
		var list array
		var data1 map
		var i int
		var len int

		Println("sell_rate>=", reverseRate)
		Println("$SellTable", $SellTable)
		Println("reverseRate", reverseRate)
		Println("$BuyTable", $BuyTable)
		list = DBGetList("global_forex_orders", "id,sell_rate,amount,sell_table_account_id,sell_table,buy_table_account_id,buy_table", 0, 100000, "id desc", "buy_table=$ AND sell_rate>=$ AND sell_table=$ AND (empty_block_id=0 OR empty_block_id is NULL)", $SellTable, reverseRate, $BuyTable)
		len = Len(list)
		Println("len", len)
		Println("list", list)
		while i < len {
			Println("i", i, "list", list)
			data1 = list[i]
			i = i + 1
			Println(data1)
			var readyToBuy money
			readyToBuy = totalSellAmount * data1["sell_rate"]
			var sellerSellAmount money
			if readyToBuy >= data1["amount"] {
				sellerSellAmount = data1["amount"] // ордер будет закрыт, а мы продолжим искать новые
			} else {
				sellerSellAmount = readyToBuy // данный ордер удовлетворяет наш запрос целиком
			}
			if data1["amount"] - sellerSellAmount < 1 { // ордер опустошили
				Println($block)
				Println(data1["id"])
				DBUpdate("global_forex_orders", Int(data1["id"]), "amount,empty_block_id", "0", $block)
				DBInsert("global_forex_history2", "price,amount,total,direction,currency,timestamp time", data1["sell_rate"], data1["amount"], Money(data1["amount"])*Float(data1["sell_rate"]), $Direction, $BuyTable, $block_time)

			} else {
				// вычитаем забранную сумму из ордера
				var rowAmount money
				rowAmount = DBIntExt("global_forex_orders", "amount", data1["id"], "id")
				DBUpdate("global_forex_orders", Int(data1["id"]), "amount", rowAmount - sellerSellAmount)
				DBInsert("global_forex_history2", "price,amount,total,direction,currency,timestamp time", data1["sell_rate"], sellerSellAmount, Money(sellerSellAmount)*Float(data1["sell_rate"]), $Direction, $BuyTable, $block_time)
			}
			var sellerBuyAmount money
			sellerBuyAmount = sellerSellAmount * (1.0 / data1["sell_rate"])

			Println("001")
			var sender_id int
			sender_id = data1["sell_table_account_id"]

			var recipient_id int
			recipient_id = DBIntExt(data1["sell_table"], "id", $citizen, "citizen_id")
			Println("recipient_id", recipient_id, "sender_id", sender_id, data1["sell_table"], sellerSellAmount, sellerBuyAmount, data1["sell_rate"])
			DBTransfer(data1["sell_table"], "amount,id", Int(sender_id), recipient_id, sellerSellAmount)

			Println("003")
			sender_id = DBIntExt(data1["buy_table"], "id", $citizen, "citizen_id")
			Println("sender_id", sender_id)
			recipient_id = data1["buy_table_account_id"]
			Println("recipient_id", recipient_id, "sender_id", sender_id, data1["buy_table"], sellerBuyAmount)
            DBTransfer(data1["buy_table"], "amount,id", sender_id, Int(recipient_id), sellerBuyAmount)

			Println("004")
				// вычитаем с нашего баланса сумму, которую потратили на данный ордер
			totalSellAmount = totalSellAmount - sellerBuyAmount
			Println("0041", totalSellAmount)
			if totalSellAmount < 1 {
				Println("0042")
				break // проход по ордерам прекращаем, т.к. наш запрос удовлетворен
			}
		}

		Println("005")
		if totalSellAmount >= 0.01 {
			var sell_account_id int
			Println("006")
			sell_account_id = DBIntExt($SellTable, "id", $citizen, "citizen_id")
			var buy_account_id int
			Println("007")
			buy_account_id = DBIntExt($BuyTable, "id", $citizen, "citizen_id")
			Println("008")
			DBInsert("global_forex_orders", "sell_rate,amount,sell_table_account_id,sell_table,buy_table_account_id,buy_table", $SellRate, totalSellAmount, sell_account_id, $SellTable, buy_account_id, $BuyTable)
		}
	}
}`)
TextHidden( sc_newForexOrder)
SetVar(`p_newForexOrder #= FullScreen(1)
Title: Forex

Navigation(LiTemplate(dashboard_default, Citizen))

Form()

        
        
    Divs(md-12, panel panel-default panel-body text-center)
    BtnTemplate(newForexOrder,EURO/USD,"Table1:'global_euro',Table2:'1_accounts',global:1,Currency1:'EURO',Currency2:'USD'")
    DivsEnd:


Divs(md-6, panel panel-default panel-body)

    Legend(" ", "BUY #Currency1#")

    Form()
        Input(DirectionBuy, "hidden", text, text, buy)
        Input(SellTable, "hidden", text, text, Param(Table2))
        Input(BuyTable, "hidden", text, text, Param(Table1))
        Divs(form-group)
            Divs(col-md-4 form-horizontal control-label text-right pr0 pt-sm)
                MarkDown: Your balance
            DivsEnd:
            Divs(col-md-4 form-horizontal control-label text-left pr0 pt-sm)
                MarkDown: GetOne(amount, #Table2#, "citizen_id", #citizen#)
            DivsEnd:
            Divs(col-md-4 pl0 pt-sm)
                MarkDown: #Currency2#
            DivsEnd:
            Divs(clearfix)
            DivsEnd:
        DivsEnd:
        
        Divs(form-group)
            Divs(col-md-4 form-horizontal control-label text-right pr0 pt-sm)
                MarkDown: Amount #Currency1#
            DivsEnd:
            Divs(col-md-4 pr-sm pl-sm)
                Input(Amount0, "form-control")
            DivsEnd:
            Divs(col-md-4 pl0 pt-sm)
            DivsEnd:
            Divs(clearfix)
            DivsEnd:
        DivsEnd:
        
        Divs(form-group)
            
            Divs(col-md-4 form-horizontal control-label text-right pr0 pt-sm)
                MarkDown: Price per #Currency1#
            DivsEnd:
            Divs(col-md-4 pr-sm pl-sm)
                Input(Rate0, "form-control")
            DivsEnd:
            Divs(col-md-4 pl0 pt-sm)
                MarkDown: #Currency2#
            DivsEnd:
            Divs(clearfix)
            DivsEnd:
        DivsEnd:
        TxButton{ Contract: @newForexOrder, Inputs: "Direction=DirectionBuy,SellTable=SellTable,SellRate=Rate0,Amount=Amount0,BuyTable=BuyTable",OnSuccess: "template,newForexOrder,global:1,Table1:'global_euro',Table2:'1_accounts',Currency1:'EURO',Currency2:'USD'" }
    
    FormEnd:
DivsEnd:


Divs(md-6, panel panel-default panel-body)

    Legend(" ", "SELL #Currency1#")

    Form()
        Input(DirectionSell, "hidden", text, text, sell)
        
        Divs(form-group)
            Divs(col-md-4 form-horizontal control-label text-right pr0 pt-sm)
                MarkDown: Your balance
            DivsEnd:
            Divs(col-md-4 form-horizontal control-label text-left pr0 pt-sm)
                MarkDown: GetOne(amount, #Table1#, "citizen_id", #citizen#)

            DivsEnd:
            Divs(col-md-4 pl0 pt-sm)
                MarkDown: #Currency1#
            DivsEnd:
            Divs(clearfix)
            DivsEnd:
        DivsEnd:
        
        Divs(form-group)
            Divs(col-md-4 form-horizontal control-label text-right pr0 pt-sm)
                MarkDown: Amount #Currency1#
            DivsEnd:
            Divs(col-md-4 pr-sm pl-sm)
                Input(Amount1, "form-control")
            DivsEnd:
            Divs(col-md-4 pl0 pt-sm)
            DivsEnd:
            Divs(clearfix)
            DivsEnd:
        DivsEnd:
        
        Divs(form-group)
            
            Divs(col-md-4 form-horizontal control-label text-right pr0 pt-sm)
                MarkDown: Price per #Currency1#
            DivsEnd:
            Divs(col-md-4 pr-sm pl-sm)
                Input(Rate1, "form-control")
            DivsEnd:
            Divs(col-md-4 pl0 pt-sm)
                MarkDown: #Currency2#
            DivsEnd:
            Divs(clearfix)
            DivsEnd:
        DivsEnd:
        TxButton{ Contract: @newForexOrder, Inputs: "Direction=DirectionSell,SellTable=BuyTable,SellRate=Rate1,Amount=Amount1,BuyTable=SellTable",OnSuccess: "template,newForexOrder,global:1,Table1:'global_euro',Table2:'1_accounts',Currency1:'EURO',Currency2:'USD'" }
    
    FormEnd:
DivsEnd:

Divs(md-6, panel panel-default panel-body)
    Legend(" ", "Sell orders")
    Divs()
    Table {
    	Table: global_forex_orders
    	Order: sell_rate ASC
    	Where: sell_table='#Table2#' and (empty_block_id is NULL or empty_block_id=0)
    	Columns: [
    		[Price, #sell_rate#],
    		[Amount, #amount# #Currency1#],
    		[Total, #Mult(#amount#, #sell_rate#) #Currency2#]
    	]
    }
    DivsEnd:
DivsEnd:


Divs(md-6, panel panel-default panel-body)
    Legend(" ", "Buy orders")
    Divs()
    Table {
    	Table: global_forex_orders
    	Order: sell_rate DESC
    	Where: sell_table='#Table1#' and (empty_block_id is NULL or empty_block_id=0)
    	Columns: [
    		[Price, #sell_rate#],
    		[Amount, #amount# #Currency1#],
    		[Total, #Mult(#amount#, #sell_rate#) #Currency2#]
    	]
    }
    DivsEnd:
DivsEnd:

Divs(md-12, panel panel-default)
    Divs(panel-heading)
        Divs(panel-title)
           MarkDown: Trade history
        DivsEnd:
    DivsEnd:
    
    Divs(panel-body)
    Table {
    	Table: global_forex_history2
    	Order: time DESC
    	Where: currency='#Table1#' or currency='#Table2#'
    	Columns: [
    		[Time, DateTime(#time#)],
    		[Type, If(#direction#=="sell", P(text-bold  text-danger, sell), P(text-bold  text-success, buy))],
    		[direction, #direction#],
    		[Price, #price# #Currency2#],
    		[Amount, #amount#  #Currency1#],
    		[Total, #total#  #Currency2#]
    	]
    }
    DivsEnd:
DivsEnd:

PageEnd:`)
TextHidden( p_newForexOrder)
Json(`Head: "Forex",
Desc: "Forex",
		Img: "/static/img/apps/forex.png",
		OnSuccess: {
			script: 'template',
			page: 'government',
			parameters: {}
		},
		TX: [{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 1,
			table_name : "forex_history2",
			columns: '[["price", "money", "1"],["total", "money", "1"],["amount", "money", "1"],["currency", "hash", "1"],["direction", "hash", "1"],["time", "time", "1"]]',
			permissions: "$citizen == #wallet_id#"
			}
	   },
{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: 1,
			table_name : "forex_orders",
			columns: '[["sell_table_account_id", "int64", "1"],["amount", "money", "1"],["buy_table", "hash", "1"],["sell_rate", "hash", "1"],["sell_table", "hash", "1"],["empty_block_id", "int64", "1"],["buy_table_account_id", "int64", "1"]]',
			permissions: "$citizen == #wallet_id#"
			}
	   },
{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: 1,
			name: "newForexOrder",
			value: $("#sc_newForexOrder").val(),
			conditions: $("#sc_conditions").val()
			}
	   },
{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "newForexOrder",
			menu: "menu_default",
			value: $("#p_newForexOrder").val(),
			global: 1,
			conditions: "$citizen == #wallet_id#",
			}
	   }]`
)