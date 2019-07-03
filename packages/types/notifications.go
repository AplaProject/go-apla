package types

type Notifications interface {
	AddAccounts(ecosystem int64, accounts ...string)
	AddRoles(ecosystem int64, roles ...int64)
	Size() int
	Send()
}
