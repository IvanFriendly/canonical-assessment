package shred

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"os"
)

// For chunked reading of larger files
const Blocksize int64 = 4096

// Export it for testing, black-box testing doesn't provide confidence, since the problem
// statements asks us to delete the file
func OverwriteFileContents(file *os.File, fileSize int64) error {
	// Move the file pointer to the beginning of the file
	_, err := file.Seek(0, 0)
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(file)
	randomData := make([]byte, Blocksize)

	for remainingBytes := fileSize; remainingBytes > 0; remainingBytes -= Blocksize {
		var dataToWrite []byte
		if remainingBytes < Blocksize {
			dataToWrite = randomData[:remainingBytes]
		} else {
			dataToWrite = randomData
		}

		_, err = rand.Read(dataToWrite)
		if err != nil {
			return err
		}

		_, err = writer.Write(dataToWrite)
		if err != nil {
			return err
		}
	}

	if err = writer.Flush(); err != nil {
		return err
	}

	if err = file.Sync(); err != nil {
		return err
	}

	return nil
}

func Shred(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat file %s: %w", path, err)
	}

	if !fileInfo.Mode().IsRegular() {
		return fmt.Errorf("not a regular file: %s", path)
	}

	fileSize := fileInfo.Size()
	// potentially overly defensive check
	if fileSize < 0 {
		return fmt.Errorf("%s has a negative size", path)
	}

	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()

	for pass := 0; pass < 3; pass++ {
		if err := OverwriteFileContents(file, fileSize); err != nil {
			return fmt.Errorf("failed to overwrite file %s on pass %d: %w", path, pass, err)
		}
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file %s: %w", path, err)
	}

	// There may be metadata left in the filesystem after a simple "unlink"
	// In production exhausting all the directory entries using some kind
	// of randomized renaming process would be more secure
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to remove file %s: %w", path, err)
	}

	return nil
}
