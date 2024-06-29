package shred_test

import (
	"os"
	"testing"

	shred "github.com/IvanFriendly/canonical-assessment"
)

func TestShred(t *testing.T) {
	t.Run("directory path", func(t *testing.T) {
		err := shred.Shred(".")
		if err == nil {
			t.Errorf("expected failure shredding a directory")
		}
	})

	t.Run("empty path", func(t *testing.T) {
		err := shred.Shred("")
		if err == nil {
			t.Errorf("expected failure shreding an empty path")
		}
	})

	t.Run("non-existent path", func(t *testing.T) {
		err := shred.Shred("thisfiledoesnotexistihope")
		if err == nil {
			t.Errorf("expected failure shreding a non-existent path name")
		}
	})

	t.Run("small size file shredding", func(t *testing.T) {
		file, err := os.CreateTemp(".", "shred_test_a")
		if err != nil {
			t.Fatalf("failed to create a temporary file: %v", err)
		}
		defer os.Remove(file.Name())

		_, err = file.Write([]byte{0})
		if err != nil {
			t.Fatalf("failed to write to the temporary test file: %v", err)
		}

		// Close the file so it can be reopened by the Shred function
		file.Close()

		err = shred.Shred(file.Name())
		if err != nil {
			t.Errorf("expected no failure shredding a small file, but got: %v", err)
		}

		_, err = os.Stat(file.Name())
		if err == nil {
			t.Errorf("shredding should remove the file")
		}
	})

	t.Run("small size file shredding overwrite pass", func(t *testing.T) {
		file, err := os.CreateTemp(".", "shred_test_b")
		if err != nil {
			t.Fatalf("failed to create a temporary file: %v", err)
		}
		defer os.Remove(file.Name())
		defer file.Close()

		_, err = file.Write([]byte{'A'})
		if err != nil {
			t.Fatalf("failed to write to the temporary test file: %v", err)
		}

		fileInfo, err := os.Stat(file.Name())
		if err != nil {
			t.Fatalf("error while getting test file statistics: %v", err)
		}

		err = shred.OverwriteFileContents(file, fileInfo.Size())
		if err != nil {
			t.Errorf("expected no failure shreding a small file, but got %v", err)
		}
	})

	t.Run("medium size file shredding", func(t *testing.T) {
		file, err := os.CreateTemp(".", "shred_test_c")
		if err != nil {
			t.Fatalf("failed to create a temporary file: %v", err)
		}
		defer os.Remove(file.Name())

		// create some easily recognizable pattern, shouldn't matter since the file is intended to be removed
		data := make([]byte, shred.Blocksize+1)
		for i := range data {
			data[i] = 'A'
		}
		_, err = file.Write(data)
		if err != nil {
			t.Fatalf("failed to write to the temporary test file: %v", err)
		}

		// close the file so it can be reopened by the Shred function
		file.Close()

		err = shred.Shred(file.Name())
		if err != nil {
			t.Errorf("expected no failure shredding a medium file, but got: %v", err)
		}

		_, err = os.Stat(file.Name())
		if err == nil {
			t.Errorf("shredding should remove the file")
		}
	})

	t.Run("medium size file shredding overwrite pass", func(t *testing.T) {
		file, err := os.CreateTemp(".", "shred_test")
		if err != nil {
			t.Fatalf("failed to create a temporary file: %v", err)
		}
		defer os.Remove(file.Name())
		defer file.Close()

		// create some easily recognizable pattern, shouldn't matter since the file is intended to be removed
		const testFileSize int64 = shred.Blocksize + 1
		data := make([]byte, testFileSize)
		for i := range data {
			data[i] = 'A'
		}
		_, err = file.Write(data)
		if err != nil {
			t.Fatalf("failed to write to the temporary test file: %v", err)
		}

		fileInfo, err := os.Stat(file.Name())
		if err != nil {
			t.Fatalf("error while getting test file statistics: %v", err)
		}

		err = shred.OverwriteFileContents(file, fileInfo.Size())
		if err != nil {
			t.Errorf("expected no failure overwriting a medium file, but got: %v", err)
		}

		overwrittenData, err := os.ReadFile(file.Name())
		if err != nil {
			t.Fatalf("failed to open file after overwriting: %v", err)
		}
		if len(overwrittenData) != len(data) {
			t.Fatalf("unexpected file size after overwrite pass: %d vs %d", len(overwrittenData), len(data))
		}
		var percentDifferent float32
		for i, v := range overwrittenData {
			if data[i] != v {
				percentDifferent += 1
			}
		}
		percentDifferent /= float32(len(data))
		percentDifferent *= 100.0

		// piggy, ran out of time :(
		if percentDifferent < 95.0 {
			t.Fatalf("the shredding process is not good enough, only %f%% of bytes differ", percentDifferent)
		}
	})

}
