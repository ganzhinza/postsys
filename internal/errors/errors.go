package errors

type DomainError struct {
	msg string
}

func (de *DomainError) Error() string {
	return de.msg
}

func New(msg string) *DomainError {
	return &DomainError{
		msg: msg,
	}
}

var (
	InternalError                           error = New("internal error")
	PostNotFoundError                       error = New("post not found")
	CommentNotFoundError                    error = New("comment not found")
	PathNotFoundError                       error = New("path not found")
	CommentsDisabledError                   error = New("comments disabled")
	CommentTooLongError                     error = New("comment too long")
	NotOwnerUpdatesCommentAvailabilityError error = New("only post owner can update comment availability")
)
