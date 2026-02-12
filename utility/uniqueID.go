package utility

import "github.com/gofrs/uuid"

var set = make(map[string]struct{})

func GetNewID() string {

	exists := true
	var candidate uuid.UUID
	var err error

	for exists {
		candidate, err = uuid.NewV4()

		if err != nil {
			panic(err)
		}
		_, exists = set[candidate.String()]
	}
	set[candidate.String()] = struct{}{}
	return candidate.String()
}

func FeedFromExternalSource(elm string) {
	set[elm] = struct{}{}
}
