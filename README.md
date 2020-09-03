# Install
`yarn` or `npm install`
# Compilation
`truffle compile`
# Deploy contract
Truffle Develop:
1. start Truffle Develop(a simple ethereum node)
```
truffle develop
```
2. deploy contract
```
truffle deploy --reset --network=develop
```
Harmony network
...

# MPT proof data generate
run `func TestReceiptMPT(t *testing.T)` in `go/derive_sha_test.go`, will generate `/tmp/testcases.json`.

# MPT proof contract test
```
cd scripts
ln -s /tmp/testcases.json
truffle --network=develop exec mpt.js
```