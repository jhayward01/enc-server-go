package utils

type Server interface {

	// Start server.
	Start() (err error)
}

type ClientBE interface {

	// This endpoint accepts requests to store a record associated with a user ID.
	StoreRecord(id, record []byte) (err error)

	// This endpoint accepts requests for record retrieval via a user ID.
	RetrieveRecord(id []byte) (record []byte, err error)

	// This endpoint accepts requests for record deletion via a user ID.
	DeleteRecord(id []byte) (err error)
}

type ClientFE interface {

	// This endpoint accepts requests to store a record associated with a user ID.
	StoreRecord(id, record []byte) (key []byte, err error)

	// This endpoint accepts requests for record retrieval via a user ID.
	RetrieveRecord(id []byte, key []byte) (record []byte, err error)

	// This endpoint accepts requests for record deletion via a user ID.
	DeleteRecord(id []byte) (err error)
}
