.PHONY: proto

proto:
	protoc -I=. --go_out=. --go_opt=paths=source_relative wikipedia.proto