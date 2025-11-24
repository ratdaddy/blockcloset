package grpcsvc

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

// Special bucket names for manual testing. These trigger specific error
// responses without requiring database setup.

// checkTestBucket checks if the bucket name is a special test name and returns
// an error if so. Returns nil if normal processing should continue.
func checkTestBucket(bucket string) error {
	switch bucket {
	case "bad":
		return status.Errorf(codes.InvalidArgument, "InvalidBucketName")

	case "taken":
		// For CreateBucket: bucket already exists (owned by someone else)
		conflict := &servicev1.BucketOwnershipConflict{
			Reason: servicev1.BucketOwnershipConflict_REASON_BUCKET_ALREADY_EXISTS,
			Bucket: bucket,
		}
		st := status.New(codes.AlreadyExists, "bucket already exists")
		withDetail, _ := st.WithDetails(conflict)
		return withDetail.Err()

	case "forbidden":
		// For ResolveWrite: access denied to bucket
		detail := &servicev1.ResolveWriteError{
			Reason: servicev1.ResolveWriteError_REASON_BUCKET_ACCESS_DENIED,
			Bucket: bucket,
		}
		st := status.New(codes.PermissionDenied, "access denied")
		withDetail, _ := st.WithDetails(detail)
		return withDetail.Err()

	case "no-cradles":
		detail := &servicev1.ResolveWriteError{
			Reason: servicev1.ResolveWriteError_REASON_NO_CRADLE_SERVERS,
			Bucket: bucket,
		}
		st := status.New(codes.FailedPrecondition, "no cradle servers available")
		withDetail, _ := st.WithDetails(detail)
		return withDetail.Err()

	case "panic":
		panic(status.New(codes.Internal, "intentional test panic"))
	}

	return nil
}
