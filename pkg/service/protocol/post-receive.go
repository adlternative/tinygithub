package protocol

type PostReceiveRequest struct {
	OldOid  string `json:"old_oid"`
	NewOid  string `json:"new_oid"`
	RefName string `json:"ref_name"`
}

type PostReceiveResponse struct {
	Message string `json:"message"`
}

type PostReceiveError struct {
	Error string `json:"error"`
}
