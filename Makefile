.PHONY: check be-test be-fmt fe-test fe-analyze

check: be-fmt be-test fe-test fe-analyze

be-test:
	cd server && go test ./...

be-fmt:
	cd server && test -z "$$(gofmt -l .)"

fe-test:
	cd app && flutter test

fe-analyze:
	cd app && flutter analyze
