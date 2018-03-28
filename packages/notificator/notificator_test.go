package notificator

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ParseRecipientNotificationsItem struct {
	Input []map[string]string
	Want  map[int64]*[]notificationRecord
}

func TestParseRecipientNotifications(t *testing.T) {
	table := []ParseRecipientNotificationsItem{
		ParseRecipientNotificationsItem{
			Input: []map[string]string{
				map[string]string{
					"recipient_id": "1",
					"role_id":      "1",
					"cnt":          "2",
				},
				map[string]string{
					"recipient_id": "1",
					"role_id":      "2",
					"cnt":          "1",
				},
				map[string]string{
					"recipient_id": "2",
					"role_id":      "3",
					"cnt":          "4",
				},
				map[string]string{
					"recipient_id": "2",
					"role_id":      "4",
					"cnt":          "3",
				},
			},
			Want: map[int64]*[]notificationRecord{
				1: &[]notificationRecord{
					notificationRecord{
						EcosystemID:  1,
						RoleID:       1,
						RecordsCount: 2,
					},
					notificationRecord{
						EcosystemID:  1,
						RoleID:       2,
						RecordsCount: 1,
					},
				},
				2: &[]notificationRecord{
					notificationRecord{
						EcosystemID:  1,
						RoleID:       3,
						RecordsCount: 4,
					},
					notificationRecord{
						EcosystemID:  1,
						RoleID:       4,
						RecordsCount: 3,
					},
				},
			},
		},
		//========= new item =========
	}

	for i, item := range table {
		result := parseRecipientNotification(item.Input, 1)

		if err := compareNotificationRecordResult(result, item.Want); err != nil {
			t.Errorf("on item %d err: %v\n", i, err)
		}
	}
}

func compareNotificationRecordResult(have, want map[int64]*[]notificationRecord) error {
	for wRecipient, wRecords := range want {
		hRecords, ok := have[wRecipient]
		if !ok {
			return fmt.Errorf("Have does'nt contains %d recipient", wRecipient)
		}

		for _, rec := range *wRecords {
			if !containsNotificationRecord(*hRecords, rec) {
				return fmt.Errorf("recipient %d does'nt contains %+v", wRecipient, rec)
			}
		}
	}

	return nil
}

func containsNotificationRecord(slice []notificationRecord, rec notificationRecord) bool {
	for _, item := range slice {
		if item == rec {
			return true
		}
	}

	return false
}

func TestOutputFormat(t *testing.T) {
	records := []notificationRecord{
		notificationRecord{
			EcosystemID:  1,
			RoleID:       1,
			RecordsCount: 2,
		},
		notificationRecord{
			EcosystemID:  2,
			RoleID:       2,
			RecordsCount: 1,
		},
	}

	want := `[{"ecosystem":1,"role_id":1,"count":2},{"ecosystem":2,"role_id":2,"count":1}]`
	bts, err := json.Marshal(records)
	if assert.NoError(t, err) {
		assert.Equal(t, string(bts), want, "marshaled not equal")
	}
}

func TestStatsChanged(t *testing.T) {
	type tsc struct {
		old    []notificationRecord
		new    []notificationRecord
		result bool
	}

	table := []tsc{
		// new role added
		tsc{
			old: []notificationRecord{
				notificationRecord{EcosystemID: 1, RoleID: 1, RecordsCount: 1},
				notificationRecord{EcosystemID: 1, RoleID: 2, RecordsCount: 1},
				notificationRecord{EcosystemID: 1, RoleID: 3, RecordsCount: 1},
			},

			new: []notificationRecord{
				notificationRecord{EcosystemID: 1, RoleID: 1, RecordsCount: 1},
				notificationRecord{EcosystemID: 1, RoleID: 2, RecordsCount: 1},
				notificationRecord{EcosystemID: 1, RoleID: 4, RecordsCount: 1}, //new role added
			},
			result: true,
		},
		// count changed
		tsc{
			old: []notificationRecord{
				notificationRecord{EcosystemID: 1, RoleID: 1, RecordsCount: 1},
				notificationRecord{EcosystemID: 1, RoleID: 2, RecordsCount: 1},
				notificationRecord{EcosystemID: 1, RoleID: 3, RecordsCount: 1},
			},

			new: []notificationRecord{
				notificationRecord{EcosystemID: 1, RoleID: 1, RecordsCount: 1},
				notificationRecord{EcosystemID: 1, RoleID: 2, RecordsCount: 2}, //records count changed
				notificationRecord{EcosystemID: 1, RoleID: 3, RecordsCount: 1},
			},
			result: true,
		},
		// not changed
		tsc{
			old: []notificationRecord{
				notificationRecord{EcosystemID: 1, RoleID: 1, RecordsCount: 1},
				notificationRecord{EcosystemID: 1, RoleID: 2, RecordsCount: 1},
				notificationRecord{EcosystemID: 1, RoleID: 3, RecordsCount: 1},
			},

			new: []notificationRecord{
				notificationRecord{EcosystemID: 1, RoleID: 1, RecordsCount: 1},
				notificationRecord{EcosystemID: 1, RoleID: 2, RecordsCount: 1},
				notificationRecord{EcosystemID: 1, RoleID: 3, RecordsCount: 1},
			},
			result: false,
		},
		// no old records add new record
		tsc{
			old: []notificationRecord{},
			new: []notificationRecord{
				notificationRecord{EcosystemID: 1, RoleID: 1, RecordsCount: 1},
				notificationRecord{EcosystemID: 1, RoleID: 2, RecordsCount: 1},
				notificationRecord{EcosystemID: 1, RoleID: 3, RecordsCount: 1},
			},
			result: true,
		},
		// old is nil add new record
		tsc{
			old: nil,
			new: []notificationRecord{
				notificationRecord{EcosystemID: 1, RoleID: 1, RecordsCount: 1},
				notificationRecord{EcosystemID: 1, RoleID: 2, RecordsCount: 1},
				notificationRecord{EcosystemID: 1, RoleID: 3, RecordsCount: 1},
			},
			result: true,
		},
		// old has value - now nil
		tsc{
			old: []notificationRecord{
				notificationRecord{EcosystemID: 1, RoleID: 1, RecordsCount: 1},
				notificationRecord{EcosystemID: 1, RoleID: 2, RecordsCount: 1},
				notificationRecord{EcosystemID: 1, RoleID: 3, RecordsCount: 1},
			},

			new:    nil,
			result: true,
		},
	}

	for i, record := range table {
		if assert.Equal(t, record.result, statsChanged(record.old, record.new)) != true {
			t.Errorf("step %d the result is not the expected", i)
		}
	}
}
