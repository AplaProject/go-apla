package modes

import "github.com/GenesisKernel/go-genesis/packages/conf"

func GetDaemonsToStart() []string {
	if conf.Config.IsSupportingOBS() {
		return []string{
			"Scheduler",
		}
	}

	return []string{
		"BlocksCollection",
		"BlockGenerator",
		"QueueParserTx",
		"QueueParserBlocks",
		"Disseminator",
		"Confirmations",
		"Scheduler",
	}
}
