package protoutil

import (
	"log"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
)

func MustParseTSFromString(format string, ts string) *timestamp.Timestamp {
	t, err := time.Parse(format, ts)
	check(err)
	protots, err := ptypes.TimestampProto(t)
	check(err)
	return protots
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
