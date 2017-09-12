INSERT INTO "system_states" ("rb_id") VALUES ('0');

INSERT INTO "1_contracts" ("value", "wallet_id","active", "conditions") VALUES 
('contract MoneyTransfer {
    data {
        Recipient string
        Amount    string
        Comment     string "optional"
    }
    conditions {
        $recipient = AddressToId($Recipient)
        if $recipient == 0 {
            error Sprintf("Recipient %%s is invalid", $Recipient)
        }
        var total money
        $amount = Money($Amount) 
        if $amount == 0 {
            error "Amount is zero"
        }
        total = Money(DBString(Table(`keys`), `amount`, $wallet))
        if $amount >= total {
            error Sprintf("Money is not enough %%v < %%v",total, $amount)
        }
    }
    action {
        DBUpdate(Table(`keys`), $wallet,`-amount`, $amount)
        DBUpdate(Table(`keys`), $recipient,`+amount`, $amount)
    }
}', '%[1]d', '1', 'ContractConditions(`MainCondition`)');
