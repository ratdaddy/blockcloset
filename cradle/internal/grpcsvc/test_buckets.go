package grpcsvc

// Special bucket names for manual testing. These trigger specific behaviors
// without requiring special setup.

// checkTestBucket checks if the bucket name is a special test name and returns
// a modified bytes written value if so. Returns the actual bytes if normal
// processing should continue.
func checkTestBucket(bucket string, actualBytes int64) int64 {
	switch bucket {
	case "size-mismatch":
		// Report incorrect bytes written (half of actual)
		return actualBytes / 2
	}

	return actualBytes
}
