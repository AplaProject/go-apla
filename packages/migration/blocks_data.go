package migration

var blocksDataSQL = `INSERT INTO "1_blocks" (id, name, value, conditions, ecosystem) VALUES
		(next_id('1_blocks'), 'admin_link', 'If(#sort#==1){
    SetVar(sort_name, "{id:1}")
}.ElseIf(#sort#==2){
    SetVar(sort_name, "{id:-1}")
}.ElseIf(#sort#==3){
    SetVar(sort_name, "{name: 1}")
}.ElseIf(#sort#==4){
    SetVar(sort_name, "{name: -1}")
}.Else{
    SetVar(sort, "1")
    SetVar(sort_name, "{id:1}") 
}

If(Or(#width#==12,#width#==6,#width#==4)){
}.Else{
    SetVar(width, "12")
}

Form(){
    Div(clearfix){
        Div(pull-left){
            DBFind(@1applications,apps).Where({ecosystem:#ecosystem_id#})
            Select(Name:AppId, Source:apps, NameColumn: name, ValueColumn: id, Value: #buffer_value_app_id#, Class: bg-gray)
        }
        Div(pull-left){
            Button(Class: fa fa-play btn bg-gray ml-sm, Page: #admin_page#, PageParams: "sort=#sort#,width=#width#,current_page=#current_page#", Contract: @1SelectApp, Params: "ApplicationId=Val(AppId)")
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
            If(#hideLink#==0){
            }.ElseIf(#width#==12){
                Span(Button(Body: Em(Class: fa fa-bars), Class: btn bg-gray-lighter, Page: #admin_page#, PageParams: "sort=#sort#,width=12,current_page=#current_page#")).Style(margin-right:5px;)
            }.Else{
                Span(Button(Body: Em(Class: fa fa-bars), Class: btn bg-gray, Page: #admin_page#, PageParams: "sort=#sort#,width=12,current_page=#current_page#")).Style(margin-right:5px;)
            }
            If(#hideLink#==0){
            }.ElseIf(#width#==6){
                Span(Button(Body: Em(Class: fa fa-th-large), Class: btn bg-gray-lighter, Page: #admin_page#, PageParams: "sort=#sort#,width=6,current_page=#current_page#")).Style(margin-right:5px;)
            }.Else{
                Span(Button(Body: Em(Class: fa fa-th-large), Class: btn bg-gray, Page: #admin_page#, PageParams: "sort=#sort#,width=6,current_page=#current_page#")).Style(margin-right:5px;)
            }
            If(#hideLink#==0){
            }.ElseIf(#width#==4){
                Span(Button(Body: Em(Class: fa fa-th), Class: btn bg-gray-lighter, Page: #admin_page#, PageParams: "sort=#sort#,width=4,current_page=#current_page#")).Style(margin-right:5px;)
            }.Else{
                Span(Button(Body: Em(Class: fa fa-th), Class: btn bg-gray, Page: #admin_page#, PageParams: "sort=#sort#,width=4,current_page=#current_page#")).Style(margin-right:5px;)
            }
        }
    }
}', 'ContractConditions("MainCondition")', '%[1]d'),
		(next_id('1_blocks'), 'export_info', 'DBFind(@1buffer_data).Columns("value->app_id,value->app_name,value->menu_name,value->menu_id,value->count_menu").Where({key:''export'', member_id: #key_id#,ecosystem:#ecosystem_id#}).Vars(buffer)

If(#buffer_value_app_id# > 0){
    DBFind(@1pages, src_pages).Where({app_id: #buffer_value_app_id#,ecosystem:#ecosystem_id#}).Limit(250).Order("name").Count(count_pages)
    DBFind(@1blocks, src_blocks).Where({app_id: #buffer_value_app_id#,ecosystem:#ecosystem_id#}).Limit(250).Order("name").Count(count_blocks)
    DBFind(@1app_params, src_parameters).Where({app_id:#buffer_value_app_id#,ecosystem:#ecosystem_id#}).Limit(250).Order("name").Count(count_parameters)
    DBFind(@1languages, src_languages).Where({ecosystem:#ecosystem_id#}).Limit(250).Order("name").Count(count_languages)
    DBFind(@1contracts, src_contracts).Where({app_id:#buffer_value_app_id#,ecosystem:#ecosystem_id#}).Limit(250).Order("name").Count(count_contracts)
    DBFind(@1tables, src_tables).Where({app_id:#buffer_value_app_id#,ecosystem:#ecosystem_id#}).Limit(250).Order("name").Count(count_tables)
}

Div(panel panel-primary){
    If(#buffer_value_app_id# > 0){
        Div(){
            Button(Body: "Export - #buffer_value_app_name#", Class: btn btn-primary btn-block, Page: @1export_download, Contract: @1Export)
        }
    }.Else{
        Div(panel-heading, "Export")
    }
    Form(){
        Div(list-group-item){
            Div(clearfix){
                Div(pull-left){
                    Span("Pages")
                }
                Div(pull-right){
                    If(#count_pages# > 0){
                        Span("(#count_pages#)")
                    }.Else{
                        Span("(0)")
                    }
                }
            }
            Div(row){
                Div(col-md-12 text-left text-muted){
                    If(#count_pages# > 0){
                        ForList(src_pages){
                            Span(Class: h6, Body: "#name#, ")
                        }
                    }.Else{
                        Span(Class: h6, Body: "Nothing selected")
                    }
                }
            }
        }
        Div(list-group-item){
            Div(clearfix){
                Div(pull-left){
                    Span("Blocks")
                }
                Div(pull-right){
                    If(#count_blocks# > 0){
                        Span("(#count_blocks#)")
                    }.Else{
                        Span("(0)")
                    }
                }
            }
            Div(row){
                Div(col-md-12 text-left text-muted){
                    If(#count_blocks# > 0){
                        ForList(src_blocks){
                            Span(Class: h6, Body: "#name#, ")
                        }
                    }.Else{
                        Span(Class: h6, Body: "Nothing selected")
                    }
                }
            }
        }
        Div(list-group-item){
            Div(clearfix){
                Div(pull-left){
                    Span("Menu")
                }
                Div(pull-right){
                    If(#buffer_value_app_id# > 0){
                        Span("(#buffer_value_count_menu#)")
                    }.Else{
                        Span("(0)")
                    }
                }
            }
            Div(row){
                Div(col-md-12 text-left text-muted){
                    If(And(#buffer_value_app_id#>0,#buffer_value_count_menu#>0)){
                        Span(Class: h6, Body:"#buffer_value_menu_name#")
                    }.Else{
                        Span(Class: h6, Body:"Nothing selected")
                    }
                }
            }
        }
        Div(list-group-item){
            Div(clearfix){
                Div(pull-left){
                    Span("Parameters")
                }
                Div(pull-right){
                    If(#count_parameters# > 0){
                        Span("(#count_parameters#)")
                    }.Else{
                        Span("(0)")
                    }
                }
            }
            Div(row){
                Div(col-md-12 text-left text-muted){
                    If(#count_parameters# > 0){
                        ForList(src_parameters){
                            Span(Class: h6, Body: "#name#, ")
                        }
                    }.Else{
                        Span(Class: h6, Body: "Nothing selected")
                    }
                }
            }
        }
        Div(list-group-item){
            Div(clearfix){
                Div(pull-left){
                    Span("Language resources")
                }
                Div(pull-right){
                    If(#count_languages# > 0){
                        Span("(#count_languages#)")
                    }.Else{
                        Span("(0)")
                    }
                }
            }
            Div(row){
                Div(col-md-12 text-left text-muted){
                    If(#count_languages# > 0){
                        ForList(src_languages){
                            Span(Class: h6, Body: "#name#, ")
                        }
                    }.Else{
                        Span(Class: h6, Body: "Nothing selected")
                    }
                }
            }
        }
        Div(list-group-item){
            Div(clearfix){
                Div(pull-left){
                    Span("Contracts")
                }
                Div(pull-right){
                    If(#count_contracts# > 0){
                        Span("(#count_contracts#)")
                    }.Else{
                        Span("(0)")
                    }
                }
            }
            Div(row){
                Div(col-md-12 text-left text-muted){
                    If(#count_contracts# > 0){
                        ForList(src_contracts){
                            Span(Class: h6, Body: "#name#, ")
                        }
                    }.Else{
                        Span(Class: h6, Body: "Nothing selected")
                    }
                }
            }
        }
        Div(list-group-item){
            Div(clearfix){
                Div(pull-left){
                    Span("Tables")
                }
                Div(pull-right){
                    If(#count_tables# > 0){
                        Span("(#count_tables#)")
                    }.Else{
                        Span("(0)")
                    }
                }
            }
            Div(row){
                Div(col-md-12 text-left text-muted){
                    If(#count_tables# > 0){
                        ForList(src_tables){
                            Span(Class: h6, Body: "#name#, ")
                        }
                    }.Else{
                        Span(Class: h6, Body: "Nothing selected")
                    }
                }
            }
        }
        If(#buffer_value_app_id# > 0){
            Div(panel-footer text-right){
                Button(Body: "Export", Class: btn btn-primary, Page: @1export_download, Contract: @1Export)
            }
        }
    }
}', 'ContractConditions("MainCondition")', '%[1]d'),
		(next_id('1_blocks'), 'export_link', 'If(And(#res_type#!="pages",#res_type#!="blocks",#res_type#!="menu",#res_type#!="parameters",#res_type#!="languages",#res_type#!="contracts",#res_type#!="tables")){
    SetVar(res_type, "pages")
}

Div(breadcrumb){
    If(#res_type#=="pages"){
        Span(Class: text-muted, Body: "Pages")
    }.Else{
        LinkPage(Body: "$@1pages$", Page: @1export_resources,, "res_type=pages")
    }
    Span(|,mh-sm)
    If(#res_type#=="blocks"){
        Span(Class: text-muted, Body: "Blocks")
    }.Else{
        LinkPage(Body: "$@1blocks$", Page: @1export_resources,, "res_type=blocks")
    }
    Span(|,mh-sm)
    If(#res_type#=="menu"){
        Span(Class: text-muted, Body: "Menu")
    }.Else{
        LinkPage(Body: "$@1menu$", Page: @1export_resources,, "res_type=menu")
    }
    Span(|,mh-sm)
    If(#res_type#=="parameters"){
        Span(Class: text-muted, Body: "Application parameters")
    }.Else{
        LinkPage(Body: "$@1app_params$", Page: @1export_resources,, "res_type=parameters")
    }
    Span(|,mh-sm)
    If(#res_type#=="languages"){
        Span(Class: text-muted, Body: "Language resources")
    }.Else{
        LinkPage(Body: "$@1lang_res$", Page: @1export_resources,, "res_type=languages")
    }
    Span(|,mh-sm)
    If(#res_type#=="contracts"){
        Span(Class: text-muted, Body: "Contracts")
    }.Else{
        LinkPage(Body: "$@1contracts$", Page: @1export_resources,, "res_type=contracts")
    }
    Span(|,mh-sm)
    If(#res_type#=="tables"){
        Span(Class: text-muted, Body: "Tables")
    }.Else{
        LinkPage(Body: "$@1tables$", Page: @1export_resources,, "res_type=tables")
    }
}', 'ContractConditions("MainCondition")', '%[1]d'),
		(next_id('1_blocks'), 'pager', 'DBFind(#pager_table#).Where(#pager_where#).Count(records_count)
    
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
            Button(Body: Em(Class: fa fa-angle-double-left), Class: btn btn-default, Page: #pager_page#, PageParams: "current_page=1,sort=#sort#,width=#width#,page_params=#page_params#")
        }
    }
    Span(){
        If(#current_page# == 1){
            Button(Body: Em(Class: fa fa-angle-left), Class: btn btn-default disabled)
        }.Else{
            Button(Body: Em(Class: fa fa-angle-left), Class: btn btn-default, Page: #pager_page#, PageParams: "current_page=#previous_page#,sort=#sort#,width=#width#,page_params=#page_params#")
        }
    }
    ForList(src_pages){
        Span(){
            If(#id# == #current_page#){
                Button(Body: #id#, Class: btn btn-primary float-left, Page: #pager_page#, PageParams: "current_page=#id#,sort=#sort#,width=#width#,page_params=#page_params#")
            }.Else{
                Button(Body: #id#, Class: btn btn-default float-left, Page: #pager_page#, PageParams: "current_page=#id#,sort=#sort#,width=#width#,page_params=#page_params#")
            }
        }
    }
    Span(){
        If(#current_page# == #last_page#){
            Button(Body: Em(Class: fa fa-angle-right), Class: btn btn-default disabled)
        }.Else{
            Button(Body: Em(Class: fa fa-angle-right), Class: btn btn-default, Page: #pager_page#, PageParams: "current_page=#next_page#,sort=#sort#,width=#width#,page_params=#page_params#")
        }
    }
    Span(){
        If(#current_page# == #last_page#){
            Button(Body: Em(Class: fa fa-angle-double-right), Class: btn btn-default disabled)
        }.Else{
            Button(Body: Em(Class: fa fa-angle-double-right), Class: btn btn-default, Page: #pager_page#, PageParams: "current_page=#last_page#,sort=#sort#,width=#width#,page_params=#page_params#")
        }
    }
}.Style("div {display:inline-block;}")', 'ContractConditions("MainCondition")', '%[1]d'),
		(next_id('1_blocks'), 'pager_header', 'If(#current_page# > 0){}.Else{
    SetVar(current_page, 1)
}
SetVar(pager_offset, Calculate(Exp: (#current_page# - 1) * #pager_limit#, Type: int))
SetVar(current_page, #current_page#)', 'ContractConditions("MainCondition")', '%[1]d');
`
