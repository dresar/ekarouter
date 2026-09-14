## Description

Provide a clear description of the changes introduced by this pull request.

---

## Type of Change

- [ ] `feat`: New feature or provider adapter
- [ ] `fix`: Bug fix or failover improvement
- [ ] `refactor`: Code refactoring without behavior change
- [ ] `perf`: Performance optimization
- [ ] `test`: New or updated tests
- [ ] `docs`: Documentation updates

---

## Verification Checklist

- [ ] All tests pass locally (`go test -count=1 ./...`).
- [ ] The application builds cleanly (`go build -o bin/ekarouter.exe ./cmd/ekarouter`).
- [ ] **Strict Nokomen Verified**: Zero comments (`//` or `/* */`) added to any Go (`*.go`) files.
- [ ] Pure Go preserved (No CGO dependencies introduced).
- [ ] Secrets and credentials are not hardcoded or committed.
- [ ] Documentation updated in `docs/` if architectural or endpoint changes were made.
