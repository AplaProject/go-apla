package migration

var blocksDataSQL = `INSERT INTO "%[1]d_blocks" (id, name, value, conditions) VALUES
		(1, 'admin_link', 'If(#sort#==1){
	SetVar(sort_name, "id asc")
}.ElseIf(#sort#==2){
	SetVar(sort_name, "id desc")
}.ElseIf(#sort#==3){
	SetVar(sort_name, "name asc")
}.ElseIf(#sort#==4){
	SetVar(sort_name, "name desc")
}.Else{
	SetVar(sort, "1")
	SetVar(sort_name, "id asc") 
}

If(Or(#width#==12,#width#==6,#width#==4)){
}.Else{
	SetVar(width, "12")
}

Form(){
	Div(clearfix){
		Div(pull-left){
			DBFind(applications,apps)
			Select(Name:AppId, Source:apps, NameColumn: name, ValueColumn: id, Value: #buffer_value_app_id#, Class: bg-gray)
		}
		Div(pull-left){
			Span(Button(Body: Em(Class: fa fa-play), Class: btn bg-gray, Page: #admin_page#, PageParams: "sort=#sort#,width=#width#,current_page=#current_page#", Contract: @1ExportNewApp, Params: "ApplicationId=Val(AppId)")).Style(margin-left:3px;)
		}
		Div(pull-right){
			If(#sort#==1){
				Span(Button(Body: Em(Class: fa fa-long-arrow-down) Sort by ID, Class: btn bg-gray-lighter, Page: #admin_page#, PageParams: "sort=2,width=#width#,current_page=#current_page#")).Style(margin-left:5px;)
			}.ElseIf(#sort#==2){
				Span(Button(Body: Em(Class: fa fa-long-arrow-up) Sort by ID, Class: btn bg-gray-lighter, Page: #admin_page#, PageParams: "sort=1,width=#width#,current_page=#current_page#")).Style(margin-left:5px;)
			}.Else{
				Span(Button(Body: Sort by ID, Class: btn bg-gray, Page: #admin_page#, PageParams: "sort=1,width=#width#,current_page=#current_page#")).Style(margin-left:5px;)
			}
			If(#sort#==3){
				Span(Button(Body: Em(Class: fa fa-long-arrow-down) Sort by NAME, Class: btn bg-gray-lighter, Page: #admin_page#, PageParams: "sort=4,width=#width#,current_page=#current_page#")).Style(margin-left:5px;)
			}.ElseIf(#sort#==4){
				Span(Button(Body: Em(Class: fa fa-long-arrow-up) Sort by NAME, Class: btn bg-gray-lighter, Page: #admin_page#, PageParams: "sort=3,width=#width#,current_page=#current_page#")).Style(margin-left:5px;)
			}.Else{
				Span(Button(Body: Sort by NAME, Class: btn bg-gray, Page: #admin_page#, PageParams: "sort=3,width=#width#,current_page=#current_page#")).Style(margin-left:5px;)
			}
		}
		Div(pull-right){
			If(#width#==12){
				Span(Button(Body: Em(Class: fa fa-bars), Class: btn bg-gray-lighter, Page: #admin_page#, PageParams: "sort=#sort#,width=12,current_page=#current_page#")).Style(margin-right:5px;)
			}.Else{
				Span(Button(Body: Em(Class: fa fa-bars), Class: btn bg-gray, Page: #admin_page#, PageParams: "sort=#sort#,width=12,current_page=#current_page#")).Style(margin-right:5px;)
			}
			If(#width#==6){
				Span(Button(Body: Em(Class: fa fa-th-large), Class: btn bg-gray-lighter, Page: #admin_page#, PageParams: "sort=#sort#,width=6,current_page=#current_page#")).Style(margin-right:5px;)
			}.Else{
				Span(Button(Body: Em(Class: fa fa-th-large), Class: btn bg-gray, Page: #admin_page#, PageParams: "sort=#sort#,width=6,current_page=#current_page#")).Style(margin-right:5px;)
			}
			If(#width#==4){
				Span(Button(Body: Em(Class: fa fa-th), Class: btn bg-gray-lighter, Page: #admin_page#, PageParams: "sort=#sort#,width=4,current_page=#current_page#")).Style(margin-right:5px;)
			}.Else{
				Span(Button(Body: Em(Class: fa fa-th), Class: btn bg-gray, Page: #admin_page#, PageParams: "sort=#sort#,width=4,current_page=#current_page#")).Style(margin-right:5px;)
			}
		}
	}
}', 'ContractConditions("MainCondition")'),
		(2, 'export_info', 'DBFind(Name: buffer_data, Source: src_buffer).Columns("value->app_id,value->app_name,value->menu_name,value->menu_id,value->count_menu").Where("key=''export'' and member_id=#key_id#").Vars(buffer)

If(#buffer_value_app_id# > 0){
	DBFind(pages, src_pages).Where("app_id=#buffer_value_app_id#").Limit(250).Order("name").Count(count_pages)
	DBFind(blocks, src_blocks).Where("app_id=#buffer_value_app_id#").Limit(250).Order("name").Count(count_blocks)
	DBFind(app_params, src_parameters).Where("app_id=#buffer_value_app_id#").Limit(250).Order("name").Count(count_parameters)
	DBFind(languages, src_languages).Where("app_id=#buffer_value_app_id#").Limit(250).Order("name").Count(count_languages)
	DBFind(contracts, src_contracts).Where("app_id=#buffer_value_app_id#").Limit(250).Order("name").Count(count_contracts)
	DBFind(tables, src_tables).Where("app_id=#buffer_value_app_id#").Limit(250).Order("name").Count(count_tables)
}

Div(panel panel-primary){
	If(#buffer_value_app_id# > 0){
		Div(panel-heading, "Export - #buffer_value_app_name#")
	}.Else{
		Div(panel-heading, "Export") 
	}
	Form(){
		Div(list-group-item){
			Div(row){
				Div(col-md-10 mc-sm text-left){
					Span("Pages")
				}
				Div(col-md-2 mc-sm text-right){
					If(#count_pages# > 0){
						Span("(#count_pages#)")
					}.Else{
						Span("(0)")  
					}
				} 
			}
			Div(row){
				Div(col-md-12 mc-sm text-left){
					If(#count_pages# > 0){
						ForList(src_pages){
							Span(Class: text-muted h6, Body: "#name#, ")
						}
					}.Else{
						Span(Class: text-muted h6, Body: "Nothing selected")
					}
				}
			}
		}
		Div(list-group-item){
			Div(row){
				Div(col-md-10 mc-sm text-left){
					Span("Blocks")
				}
				Div(col-md-2 mc-sm text-right){
					If(#count_blocks# > 0){
						Span("(#count_blocks#)")
					}.Else{
						Span("(0)")  
					}
				} 
			}
			Div(row){
				Div(col-md-12 mc-sm text-left){
					If(#count_blocks# > 0){
						ForList(src_blocks){
							Span(Class: text-muted h6, Body: "#name#, ")
						}
					}.Else{
						Span(Class: text-muted h6, Body: "Nothing selected")
					}
				}
			}
		}
		Div(list-group-item){
			Div(row){
				Div(col-md-10 mc-sm text-left){
					Span("Menu")
				}
				Div(col-md-2 mc-sm text-right){
					If(#buffer_value_app_id# > 0){
						Span("(#buffer_value_count_menu#)")
					}.Else{
						Span("(0)")
					}
				} 
			}
			Div(row){
				Div(col-md-12 mc-sm text-left){
					If(And(#buffer_value_app_id#>0,#buffer_value_count_menu#>0)){
						Span(Class: text-muted h6, Body:"#buffer_value_menu_name#")
					}.Else{
						Span(Class: text-muted h6, Body:"Nothing selected")
					}
				}
			}
		}
		Div(list-group-item){
			Div(row){
				Div(col-md-10 mc-sm text-left){
					Span("Parameters")
				}
				Div(col-md-2 mc-sm text-right){
					If(#count_parameters# > 0){
						Span("(#count_parameters#)")
					}.Else{
						Span("(0)")  
					}
				} 
			}
			Div(row){
				Div(col-md-12 mc-sm text-left){
					If(#count_parameters# > 0){
						ForList(src_parameters){
							Span(Class: text-muted h6, Body: "#name#, ")
						}
					}.Else{
						Span(Class: text-muted h6, Body: "Nothing selected")
					}
				}
			}
		}
		Div(list-group-item){
			Div(row){
				Div(col-md-10 mc-sm text-left){
					Span("Language resources")
				}
				Div(col-md-2 mc-sm text-right){
					If(#count_languages# > 0){
						Span("(#count_languages#)")
					}.Else{
						Span("(0)")  
					}
				} 
			}
			Div(row){
				Div(col-md-12 mc-sm text-left){
					If(#count_languages# > 0){
						ForList(src_languages){
							Span(Class: text-muted h6, Body: "#name#, ")
						}
					}.Else{
						Span(Class: text-muted h6, Body: "Nothing selected")
					}
				}
			}
		}
		Div(list-group-item){
			Div(row){
				Div(col-md-10 mc-sm text-left){
					Span("Contracts")
				}
				Div(col-md-2 mc-sm text-right){
					If(#count_contracts# > 0){
						Span("(#count_contracts#)")
					}.Else{
						Span("(0)")  
					}
				} 
			}
			Div(row){
				Div(col-md-12 mc-sm text-left){
					If(#count_contracts# > 0){
						ForList(src_contracts){
							Span(Class: text-muted h6, Body: "#name#, ")
						}
					}.Else{
						Span(Class: text-muted h6, Body: "Nothing selected")
					}
				}
			}
		}
		Div(list-group-item){
			Div(row){
				Div(col-md-10 mc-sm text-left){
					Span("Tables")
				}
				Div(col-md-2 mc-sm text-right){
					If(#count_tables# > 0){
						Span("(#count_tables#)")
					}.Else{
						Span("(0)")  
					}
				} 
			}
			Div(row){
				Div(col-md-12 mc-sm text-left){
					If(#count_tables# > 0){
						ForList(src_tables){
							Span(Class: text-muted h6, Body: "#name#, ")
						}
					}.Else{
						Span(Class: text-muted h6, Body: "Nothing selected")
					}
				}
			}
		}
		If(#buffer_value_app_id# > 0){
			Div(panel-footer clearfix){
				Div(pull-left){
					Button(Body: Em(Class: fa fa-refresh), Class: btn btn-default, Contract: @1ExportNewApp, Params: "ApplicationId=#buffer_value_app_id#", Page: export_resources)
				}
				Div(pull-right){
					Button(Body: Export, Class: btn btn-primary, Page: export_download, Contract: @1Export)
				}
			}
		}
	}
}', 'ContractConditions("MainCondition")'),
		(3, 'export_link', 'If(And(#res_type#!="pages",#res_type#!="blocks",#res_type#!="menu",#res_type#!="parameters",#res_type#!="languages",#res_type#!="contracts",#res_type#!="tables")){
	SetVar(res_type, "pages")
}

Div(breadcrumb){
	If(#res_type#=="pages"){
		Span(Class: text-muted, Body: "Pages")
	}.Else{
		LinkPage(Body: "Pages", Page: export_resources,, "res_type=pages")
	}
	Span(|).Style(margin-right: 10px; margin-left: 10px;)
	If(#res_type#=="blocks"){
		Span(Class: text-muted, Body: "Blocks")
	}.Else{
		LinkPage(Body: "Blocks", Page: export_resources,, "res_type=blocks")
	}
	Span(|).Style(margin-right: 10px; margin-left: 10px;)
	If(#res_type#=="menu"){
		Span(Class: text-muted, Body: "Menu")
	}.Else{
	   LinkPage(Body: "Menu", Page: export_resources,, "res_type=menu")
	}
	Span(|).Style(margin-right: 10px; margin-left: 10px;)
	If(#res_type#=="parameters"){
		Span(Class: text-muted, Body: "Application parameters")
	}.Else{
	   LinkPage(Body: "Application parameters", Page: export_resources,, "res_type=parameters")
	}
	Span(|).Style(margin-right: 10px; margin-left: 10px;)
	If(#res_type#=="languages"){
		Span(Class: text-muted, Body: "Language resources")
	}.Else{
	   LinkPage(Body: "Language resources", Page: export_resources,, "res_type=languages")
	}
	Span(|).Style(margin-right: 10px; margin-left: 10px;)
	If(#res_type#=="contracts"){
		Span(Class: text-muted, Body: "Contracts")
	}.Else{
	   LinkPage(Body: "Contracts", Page: export_resources,, "res_type=contracts")
	} 
	Span(|).Style(margin-right: 10px; margin-left: 10px;)
	If(#res_type#=="tables"){
		Span(Class: text-muted, Body: "Tables")
	}.Else{
	   LinkPage(Body: "Tables", Page: export_resources,, "res_type=tables")
	}
}', 'ContractConditions("MainCondition")'),
		(4, 'pager', 'DBFind(#pager_table#, src_records).Where(#pager_where#).Count(records_count)
	
SetVar(previous_page, Calculate(Exp: #current_page# - 1, Type: int))
SetVar(next_page, Calculate(Exp: #current_page# + 1, Type: int))
SetVar(count_div_limit_int, Calculate(Exp: (#records_count# / #pager_limit#), Type: int))
SetVar(remainder, Calculate(Exp: (#records_count# / #pager_limit#) - #count_div_limit_int#, Type: float))

If(#remainder# != 0){
	SetVar(last_page, Calculate(Exp: #count_div_limit_int# + 1, Type: int))
}.Else{
	SetVar(last_page, #count_div_limit_int#)
}

SetVar(last_page_plus_one, Calculate(Exp: #last_page# + 1, Type: int))
SetVar(delta_last_page, Calculate(Exp: #last_page# - #current_page#))
SetVar(range_l, Calculate(Exp: #current_page# - 4, Type: int))
SetVar(range_r, Calculate(Exp: #current_page# + 6, Type: int))
SetVar(range_l_max, Calculate(Exp: #last_page# - #pager_limit#, Type: int))
SetVar(pager_limit_plus_one, Calculate(Exp: #pager_limit# + 1, Type: int))

If(#current_page# < 5){
	If(#last_page# >= 10){
		Range(src_pages, 1, 11)
	}.Else{
		Range(src_pages, 1, #last_page_plus_one#) 
	}
}.ElseIf(#delta_last_page# < 6){
	If(#range_l_max# > 0){
		Range(src_pages, #range_l_max#, #last_page_plus_one#)
	}.Else{
		Range(src_pages, 1, #last_page_plus_one#)
	}
}.Else{
	Range(src_pages, #range_l#, #range_r#)
}

Div(){
	Span(){
		If(#current_page# == 1){
			Button(Body: Em(Class: fa fa-angle-double-left), Class: btn btn-default disabled)
		}.Else{
			Button(Body: Em(Class: fa fa-angle-double-left), Class: btn btn-default, Page: #pager_page#, PageParams: "current_page=1,sort=#sort#,width=#width#")
		}
	}
	Span(){
		If(#current_page# == 1){
			Button(Body: Em(Class: fa fa-angle-left), Class: btn btn-default disabled)
		}.Else{
			Button(Body: Em(Class: fa fa-angle-left), Class: btn btn-default, Page: #pager_page#, PageParams: "current_page=#previous_page#,sort=#sort#,width=#width#")
		}
	}
	ForList(src_pages){
		Span(){
			If(#id# == #current_page#){
				Button(Body: #id#, Class: btn btn-primary float-left, Page: #pager_page#, PageParams: "current_page=#id#,sort=#sort#,width=#width#")
			}.Else{
				Button(Body: #id#, Class: btn btn-default float-left, Page: #pager_page#, PageParams: "current_page=#id#,sort=#sort#,width=#width#")
			}
		}
	}
	Span(){
		If(#current_page# == #last_page#){
			Button(Body: Em(Class: fa fa-angle-right), Class: btn btn-default disabled)
		}.Else{
			Button(Body: Em(Class: fa fa-angle-right), Class: btn btn-default, Page: #pager_page#, PageParams: "current_page=#next_page#,sort=#sort#,width=#width#")
		}
	}
	Span(){
		If(#current_page# == #last_page#){
			Button(Body: Em(Class: fa fa-angle-double-right), Class: btn btn-default disabled)
		}.Else{
			Button(Body: Em(Class: fa fa-angle-double-right), Class: btn btn-default, Page: #pager_page#, PageParams: "current_page=#last_page#,sort=#sort#,width=#width#")
		}
	}
}.Style("div {display:inline-block;}")', 'ContractConditions("MainCondition")'),
		(5, 'pager_header', 'If(#current_page# > 0){}.Else{
	SetVar(current_page, 1)
}
SetVar(pager_offset, Calculate(Exp: (#current_page# - 1) * #pager_limit#, Type: int))
SetVar(current_page, #current_page#)', 'ContractConditions("MainCondition")');
`
