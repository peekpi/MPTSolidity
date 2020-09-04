package types

import (
	"fmt"
	"reflect"
	"testing"
)

func getElem(typ reflect.Type) reflect.Type {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ
}

type NameStack struct {
	name []string
}

func (ns *NameStack) push(n string) {
	ns.name = append(ns.name, n)
}
func (ns *NameStack) pop() {
	ns.name = ns.name[:len(ns.name)-1]
}
func (ns *NameStack) top(n int) string {
	return ns.name[len(ns.name)-1-n]
}

var header = `pragma solidity ^0.6.0;

import "./RLPReader.sol";

contract %sReader {
    using RLPReader for RLPReader.RLPItem;
	using RLPReader for bytes;

	function traverse(bytes memory rlpdata) pure public {
`

var footer = `	}
}`

var deep = 2

func spaceN() string {
	ret := make([]byte, deep*4)
	for i := range ret {
		ret[i] = ' '
	}
	return string(ret)
}

type SolidityGenerate struct {
	name    string
	content string
}

func (sg *SolidityGenerate) write(line string) {
	sg.content += spaceN() + line
}

func (sg *SolidityGenerate) String() string {
	head := fmt.Sprintf(header, sg.name)
	return head + sg.content + footer
}

var sg = &SolidityGenerate{}

type receipt struct {
	PostStateOrStatus []byte
	CumulativeGasUsed uint64
	Bloom             Bloom
	Logs              []rlpLog
}

func TestReceiptDecoding(t *testing.T) {
	typ := reflect.TypeOf(receipt{})
	typ = getElem(typ)
	ns.push("stacks")
	defer ns.pop()
	sg.name = typ.Name()
	sg.write(fmt.Sprintf("RLPReader.RLPItem memory %s = rlpdata.toRlpItem();\n", ns.top(0)))
	visitElem(typ.Name(), typ)
	fmt.Println(sg)
}

var ns = &NameStack{}

func visitElem(name string, typ reflect.Type) {
	typ = getElem(typ)
	switch typ.Kind() {
	case reflect.Struct:
		visitStruct(name, typ)
	case reflect.Slice:
		visitSlice(name, typ)
	case reflect.Array:
		visitArray(name, typ)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		visitNum(name, typ)
	case reflect.Bool:
		visitBool(name, typ)
	case reflect.String:
		visitString(name, typ)
	default:
		panic("unkown " + name + " " + typ.Kind().String())
	}
}

func visitStruct(name string, typ reflect.Type) {
	ns.push(name)
	defer ns.pop()
	if typ.String() == "big.Int" {
		sg.write(fmt.Sprintf("uint %s = %s.toUint();\n", ns.top(0), ns.top(1)))
		return
	}
	sg.write(fmt.Sprintf("RLPReader.RLPItem[] memory %s = %s.toList();\n", ns.top(0), ns.top(1)))

	for i := 0; i < typ.NumField(); i++ {
		ns.push(fmt.Sprintf("%s[%d]", ns.top(0), i))
		fieldTyp := typ.Field(i)
		visitElem(fieldTyp.Name, fieldTyp.Type)
		ns.pop()
	}
}
func visitString(name string, typ reflect.Type) {
	ns.push(name)
	defer ns.pop()
	sg.write(fmt.Sprintf("string memory %s = string(%s.toBytes());\n", ns.top(0), ns.top(1)))
}
func visitSlice(name string, typ reflect.Type) {
	ns.push(name)
	defer ns.pop()
	elemTyp := getElem(typ.Elem())
	if elemTyp.Kind() == reflect.Uint8 {
		sg.write(fmt.Sprintf("bytes memory %s = %s.toBytes();\n", ns.top(0), ns.top(1)))
		return
	}
	sg.write(fmt.Sprintf("RLPReader.RLPItem[] memory %s = %s.toList();\n", ns.top(0), ns.top(1)))
	sg.write(fmt.Sprintf("for(uint i = 0; i < %s.length; i++) {\n", ns.top(0)))
	deep++
	ns.push(fmt.Sprintf("%s[i]", ns.top(0)))
	visitElem(elemTyp.Name(), elemTyp)
	ns.pop()
	deep--
	sg.write("}\n")
}
func visitArray(name string, typ reflect.Type) {
	ns.push(name)
	defer ns.pop()
	elemTyp := getElem(typ.Elem())
	if elemTyp.Kind() == reflect.Uint8 {
		if typ.String() == "common.Address" && typ.Len() == 20 {
			sg.write(fmt.Sprintf("address %s = %s.toAddress();\n", ns.top(0), ns.top(1)))
		} else if typ.Len() <= 32 {
			sg.write(fmt.Sprintf("bytes%d %s = bytes%d(%s.toUint());\n", typ.Len(), ns.top(0), typ.Len(), ns.top(1)))
		} else {
			sg.write(fmt.Sprintf("bytes memory %s = %s.toBytes();\n", ns.top(0), ns.top(1)))
		}
		return
	}
	sg.write(fmt.Sprintf("RLPReader.RLPItem[] memory %s = %s.toList();\n", ns.top(0), ns.top(1)))
	sg.write(fmt.Sprintf("for(uint i = 0; i < %d; i++) {\n", typ.Len()))
	deep++
	ns.push(fmt.Sprintf("%s[i]", ns.top(0)))
	visitElem(elemTyp.Name(), elemTyp)
	ns.pop()
	deep--
	sg.write("}\n")
}
func visitNum(name string, typ reflect.Type) {
	ns.push(name)
	defer ns.pop()
	sg.write(fmt.Sprintf("uint %s = %s.toUint();\n", ns.top(0), ns.top(1)))
}
func visitBool(name string, typ reflect.Type) {
	ns.push(name)
	defer ns.pop()
	sg.write(fmt.Sprintf("bool %s = %s.toBoolean();\n", ns.top(0), ns.top(1)))
}
