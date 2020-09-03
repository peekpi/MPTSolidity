package types

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
)

func toHexStr(b []byte) string {
	return fmt.Sprintf("0x%x", b)
}

type _DerivableList interface {
	Len() int
	GetRlp(i int) []byte
}

// same logic as DeriveSha(...)
func _DeriveSha(list _DerivableList) (*trie.Trie, common.Hash) {
	keybuf := new(bytes.Buffer)
	trie := new(trie.Trie)
	for i := 0; i < list.Len(); i++ {
		keybuf.Reset()
		rlp.Encode(keybuf, uint(i))
		trie.Update(keybuf.Bytes(), list.GetRlp(i))
	}
	return trie, trie.Hash()
}

type RawRLP []byte

func (raw RawRLP) EncodeRLP(w io.Writer) error {
	w.Write(raw)
	return nil
}

type MemDB struct {
	keys   [][]byte
	values []RawRLP
}

func (db *MemDB) Put(key []byte, value []byte) error {
	db.keys = append(db.keys, key)
	db.values = append(db.values, RawRLP(value))
	return nil
}

// Delete removes the key from the key-value data store.
func (db *MemDB) Delete(key []byte) error {
	panic("Delete")
}

func (db *MemDB) Has(key []byte) (bool, error) {
	panic("Has")
}

// Get retrieves the given key if it's present in the key-value data store.
func (db *MemDB) Get(key []byte) ([]byte, error) {
	for i, dbkey := range db.keys {
		if bytes.Compare(dbkey, key) == 0 {
			return []byte(db.values[i]), nil
		}
	}
	panic("not found")
}

func (db *MemDB) ToProof() string {
	b, _ := rlp.EncodeToBytes(db.values)
	return toHexStr(b)
}

// copy from trie.keybytesToHex
func keybytesToHex(str []byte) []byte {
	l := len(str)*2 + 1
	var nibbles = make([]byte, l)
	for i, b := range str {
		nibbles[i*2] = b / 16
		nibbles[i*2+1] = b % 16
	}
	nibbles[l-1] = 16
	return nibbles
}

type Inputs struct {
	RootHash string `json:"rootHash"`
	Keys     string `json:"keys"`
	Proof    string `json:"proof"`
}

type TestCase struct {
	Result bool   `json:"result"`
	Return string `json:"return"`
	Inputs Inputs `json:"inputs"`
}

func TestReceiptMPT(t *testing.T) {
	testcases := make([]TestCase, 0, 11111)
	testcases = append(testcases, randomCases(1)...)
	testcases = append(testcases, randomCases(10)...)
	testcases = append(testcases, randomCases(100)...)
	testcases = append(testcases, randomCases(1000)...)
	testcases = append(testcases, randomCases(10000)...)

	tj, _ := json.Marshal(testcases)
	err := ioutil.WriteFile("/tmp/testcases.json", tj, 0666)
	if err != nil {
		panic("write error")
	}
}

func randomCases(elemLen int) []TestCase {
	cases := make([]TestCase, 0, elemLen)
	rs := make(Receipts, 0, elemLen)
	for i := 0; i < elemLen; i++ {
		var rh [32]byte
		rand.Read(rh[:])
		r := NewReceipt(rh[:], i&1 == 0, uint64(i))
		rs = append(rs, r)
	}
	tree, rootHash := _DeriveSha(rs)
	for i := 0; i < elemLen; i++ {
		b, _ := rlp.EncodeToBytes(uint(i))
		var db MemDB
		tree.Prove(b, 0, &db) // generate proof data in db.values
		ret, _, err := trie.VerifyProof(rootHash, b, &db)
		cases = append(cases, TestCase{
			Result: err == nil,
			Return: toHexStr(ret),
			Inputs: Inputs{
				RootHash: rootHash.String(),
				Keys:     toHexStr(keybytesToHex(b)),
				Proof:    db.ToProof(),
			},
		})
	}
	return cases
}
