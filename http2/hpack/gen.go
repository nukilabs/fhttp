// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"sort"

	"github.com/robin-samuel/fhttp/http2/hpack"
)

// http://tools.ietf.org/html/draft-ietf-httpbis-header-compression-07#appendix-B
var staticTableEntries = [...]hpack.HeaderField{
	{Name: ":authority"},
	{Name: ":method", Value: "GET"},
	{Name: ":method", Value: "POST"},
	{Name: ":path", Value: "/"},
	{Name: ":path", Value: "/index.html"},
	{Name: ":scheme", Value: "http"},
	{Name: ":scheme", Value: "https"},
	{Name: ":status", Value: "200"},
	{Name: ":status", Value: "204"},
	{Name: ":status", Value: "206"},
	{Name: ":status", Value: "304"},
	{Name: ":status", Value: "400"},
	{Name: ":status", Value: "404"},
	{Name: ":status", Value: "500"},
	{Name: "accept-charset"},
	{Name: "accept-encoding", Value: "gzip, deflate"},
	{Name: "accept-language"},
	{Name: "accept-ranges"},
	{Name: "accept"},
	{Name: "access-control-allow-origin"},
	{Name: "age"},
	{Name: "allow"},
	{Name: "authorization"},
	{Name: "cache-control"},
	{Name: "content-disposition"},
	{Name: "content-encoding"},
	{Name: "content-language"},
	{Name: "content-length"},
	{Name: "content-location"},
	{Name: "content-range"},
	{Name: "content-type"},
	{Name: "cookie"},
	{Name: "date"},
	{Name: "etag"},
	{Name: "expect"},
	{Name: "expires"},
	{Name: "from"},
	{Name: "host"},
	{Name: "if-match"},
	{Name: "if-modified-since"},
	{Name: "if-none-match"},
	{Name: "if-range"},
	{Name: "if-unmodified-since"},
	{Name: "last-modified"},
	{Name: "link"},
	{Name: "location"},
	{Name: "max-forwards"},
	{Name: "proxy-authenticate"},
	{Name: "proxy-authorization"},
	{Name: "range"},
	{Name: "referer"},
	{Name: "refresh"},
	{Name: "retry-after"},
	{Name: "server"},
	{Name: "set-cookie"},
	{Name: "strict-transport-security"},
	{Name: "transfer-encoding"},
	{Name: "user-agent"},
	{Name: "vary"},
	{Name: "via"},
	{Name: "www-authenticate"},
}

type pairNameValue struct {
	name, value string
}

type byNameItem struct {
	name string
	id   uint64
}

type byNameValueItem struct {
	pairNameValue
	id uint64
}

func headerFieldToString(f hpack.HeaderField) string {
	return fmt.Sprintf("{Name: \"%s\", Value:\"%s\", Sensitive: %t}", f.Name, f.Value, f.Sensitive)
}

func pairNameValueToString(v pairNameValue) string {
	return fmt.Sprintf("{name: \"%s\", value:\"%s\"}", v.name, v.value)
}

const header = `
// go generate gen.go
// Code generated by the command above; DO NOT EDIT.

package hpack

var staticTable = &headerFieldTable{
	evictCount: 0,
	byName: map[string]uint64{
`

//go:generate go run gen.go
func main() {
	var bb bytes.Buffer
	fmt.Fprintf(&bb, header)
	byName := make(map[string]uint64)
	byNameValue := make(map[pairNameValue]uint64)
	for index, entry := range staticTableEntries {
		id := uint64(index) + 1
		byName[entry.Name] = id
		byNameValue[pairNameValue{entry.Name, entry.Value}] = id
	}
	// Sort maps for deterministic generation.
	byNameItems := sortByName(byName)
	byNameValueItems := sortByNameValue(byNameValue)

	for _, item := range byNameItems {
		fmt.Fprintf(&bb, "\"%s\":%d,\n", item.name, item.id)
	}
	fmt.Fprintf(&bb, "},\n")
	fmt.Fprintf(&bb, "byNameValue: map[pairNameValue]uint64{\n")
	for _, item := range byNameValueItems {
		fmt.Fprintf(&bb, "%s:%d,\n", pairNameValueToString(item.pairNameValue), item.id)
	}
	fmt.Fprintf(&bb, "},\n")
	fmt.Fprintf(&bb, "ents: []HeaderField{\n")
	for _, value := range staticTableEntries {
		fmt.Fprintf(&bb, "%s,\n", headerFieldToString(value))
	}
	fmt.Fprintf(&bb, "},\n")
	fmt.Fprintf(&bb, "}\n")
	genFile("static_table.go", &bb)
}

func sortByNameValue(byNameValue map[pairNameValue]uint64) []byNameValueItem {
	var byNameValueItems []byNameValueItem
	for k, v := range byNameValue {
		byNameValueItems = append(byNameValueItems, byNameValueItem{k, v})
	}
	sort.Slice(byNameValueItems, func(i, j int) bool {
		return byNameValueItems[i].id < byNameValueItems[j].id
	})
	return byNameValueItems
}

func sortByName(byName map[string]uint64) []byNameItem {
	var byNameItems []byNameItem
	for k, v := range byName {
		byNameItems = append(byNameItems, byNameItem{k, v})
	}
	sort.Slice(byNameItems, func(i, j int) bool {
		return byNameItems[i].id < byNameItems[j].id
	})
	return byNameItems
}

func genFile(name string, buf *bytes.Buffer) {
	b, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := ioutil.WriteFile(name, b, 0644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
