package notificator

import (
	"encoding/json"
	"fmt"
	"testing"
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
					"cnt":          "1", //1
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
						RoleID:       1,
						RecordsCount: 2,
					},
					notificationRecord{
						RoleID:       2,
						RecordsCount: 1,
					},
				},
				2: &[]notificationRecord{
					notificationRecord{
						RoleID:       3,
						RecordsCount: 4,
					},
					notificationRecord{
						RoleID:       4,
						RecordsCount: 3,
					},
				},
			},
		},
		//========= new item =========
	}

	for i, item := range table {
		result, err := parseRecipientNotification(item.Input)
		if err != nil {
			t.Error(err, "on ", i, " item")
		}

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
			RoleID:       1,
			RecordsCount: 2,
		},
		notificationRecord{
			RoleID:       2,
			RecordsCount: 1,
		},
	}

	want := `[{"role_id":1,"count":2},{"role_id":2,"count":1}]`
	bts, err := json.Marshal(records)
	if err != nil {
		t.Error("error on marshal records")
		return
	}

	if string(bts) != want {
		t.Errorf(`marshaled result "%s" not equal to "%s"`, string(bts), want)
	}
}
