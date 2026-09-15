package producer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProcessFile(t *testing.T) {
	tests := []struct {
		name          string
		data          string
		workers       int
		wantTotal     int
		wantValid     int
		wantInvalid   int
		wantInfo      int
		wantWarn      int
		wantError     int
		wantUnique    int
	}{
		{
			name: "normal data",
			data: `INFO,user-101,login successful
ERROR,user-203,database connection failed
WARN,user-305,request took too long
INFO,user-101,viewed dashboard`,
			workers:     3,
			wantTotal:   4,
			wantValid:   4,
			wantInvalid: 0,
			wantInfo:    2,
			wantWarn:    1,
			wantError:   1,
			wantUnique:  3,
		},
		{
			name: "invalid records",
			data: `INFO,user-101,login
INVALID,user-102,something
ERROR
INFO,user-103`,
			workers:     2,
			wantTotal:   4,
			wantValid:   1,
			wantInvalid: 3,
			wantInfo:    1,
			wantWarn:    0,
			wantError:   0,
			wantUnique:  1,
		},
		{
			name: "duplicate users",
			data: `INFO,user-101,login
INFO,user-101,logout
ERROR,user-101,failed request
WARN,user-202,slow request`,
			workers:     4,
			wantTotal:   4,
			wantValid:   4,
			wantInvalid: 0,
			wantInfo:    2,
			wantWarn:    1,
			wantError:   1,
			wantUnique:  2,
		},
		{
			name: "empty fields",
			data: `INFO,,login
INFO,user-101,
, user-102,message
ERROR,user-103,database failed`,
			workers:     2,
			wantTotal:   4,
			wantValid:   1,
			wantInvalid: 3,
			wantInfo:    0,
			wantWarn:    0,
			wantError:   1,
			wantUnique:  1,
		},
		{
			name:        "empty file",
			data:        "",
			workers:     3,
			wantTotal:   0,
			wantValid:   0,
			wantInvalid: 0,
			wantInfo:    0,
			wantWarn:    0,
			wantError:   0,
			wantUnique:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "data.csv")

			err := os.WriteFile(path, []byte(tt.data), 0644)
			if err != nil {
				t.Fatal(err)
			}

			got, err := ProcessFile(path, tt.workers)
			if err != nil {
				t.Fatal(err)
			}

			if got.TotalLines != tt.wantTotal {
				t.Errorf("TotalLines = %d, want %d", got.TotalLines, tt.wantTotal)
			}

			if got.ValidLines != tt.wantValid {
				t.Errorf("ValidLines = %d, want %d", got.ValidLines, tt.wantValid)
			}

			if got.InvalidLines != tt.wantInvalid {
				t.Errorf("InvalidLines = %d, want %d", got.InvalidLines, tt.wantInvalid)
			}

			if got.INFO != tt.wantInfo {
				t.Errorf("INFO = %d, want %d", got.INFO, tt.wantInfo)
			}

			if got.WARN != tt.wantWarn {
				t.Errorf("WARN = %d, want %d", got.WARN, tt.wantWarn)
			}

			if got.ERROR != tt.wantError {
				t.Errorf("ERROR = %d, want %d", got.ERROR, tt.wantError)
			}

			if got.UniqueUsers != tt.wantUnique {
				t.Errorf("UniqueUsers = %d, want %d", got.UniqueUsers, tt.wantUnique)
			}
		})
	}
}

func TestProcessFileMissingFile(t *testing.T) {
	_, err := ProcessFile("does-not-exist.csv", 3)

	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestProcessFileDifferentWorkerCounts(t *testing.T) {
	data := `INFO,user-1,login
ERROR,user-2,database failed
WARN,user-3,slow request
INFO,user-1,logout
ERROR,user-2,timeout`

	path := filepath.Join(t.TempDir(), "data.csv")

	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	for _, workers := range []int{1, 2, 3, 5, 10} {
		t.Run("workers_"+string(rune('0'+workers)), func(t *testing.T) {
			got, err := ProcessFile(path, workers)
			if err != nil {
				t.Fatal(err)
			}

			if got.TotalLines != 5 {
				t.Errorf("TotalLines = %d, want 5", got.TotalLines)
			}

			if got.ValidLines != 5 {
				t.Errorf("ValidLines = %d, want 5", got.ValidLines)
			}

			if got.INFO != 2 || got.WARN != 1 || got.ERROR != 2 {
				t.Errorf(
					"unexpected counts: INFO=%d WARN=%d ERROR=%d",
					got.INFO,
					got.WARN,
					got.ERROR,
				)
			}

			if got.UniqueUsers != 3 {
				t.Errorf("UniqueUsers = %d, want 3", got.UniqueUsers)
			}
		})
	}
}