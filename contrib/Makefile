BINARIES := issue-109

build: $(BINARIES)
clean:
	$(RM) $(BINARIES)

.PHONY: tests
tests: issue-109
	valgrind ./issue-109

issue-109: issue-109.go
	go build -v issue-109.go

