package godo

import (
	"testing"
)

func TestStringify_Map(t *testing.T) {
	result := Stringify(map[int]*DropletBackupPolicy{
		1: {DropletID: 1, BackupEnabled: true},
		2: {DropletID: 2},
		3: {DropletID: 3, BackupEnabled: true},
	})

	expected := `map[1:godo.DropletBackupPolicy{DropletID:1, BackupEnabled:true}, 2:godo.DropletBackupPolicy{DropletID:2, BackupEnabled:false}, 3:godo.DropletBackupPolicy{DropletID:3, BackupEnabled:true}]`
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}
